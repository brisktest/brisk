// Copyright 2024 Brisk, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"brisk-supervisor/shared/auth"
	"brisk-supervisor/shared/brisk_metrics"
	"brisk-supervisor/shared/constants"
	"brisk-supervisor/shared/env"
	"brisk-supervisor/shared/honeycomb"
	. "brisk-supervisor/shared/logger"
	"brisk-supervisor/shared/nomad"
	"brisk-supervisor/shared/types"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log_service "brisk-supervisor/shared/services/log_service"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/mitchellh/copystructure"

	"github.com/spf13/viper"

	"brisk-supervisor/api"
	brisksupervisor "brisk-supervisor/brisk-supervisor"
	pb "brisk-supervisor/brisk-supervisor"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "brisk-supervisor/shared"

	. "brisk-supervisor/shared/context"

	"github.com/bugsnag/bugsnag-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	grpcotel "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/filters"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"net/http"
	_ "net/http/pprof"
)

var globalProjectLock sync.Mutex
var globalProjectToken string
var globalPublicKey string
var globalPrivateKey string
var globalInstanceID string

type server struct {
	pb.UnimplementedBriskSupervisorServer
}

// only allow one run at a time
var syncChannel = make(chan bool, 1)

var setupMutex sync.Mutex

func (s *server) Setup(ctx context.Context, data *pb.TestOption) (*pb.Response, error) {

	ctx, span := otel.Tracer(name).Start(ctx, "Setup")
	defer span.End()
	defer bugsnag.AutoNotify(ctx)
	Logger(ctx).Debug("Setup++")
	// critical section - we don't want multiple people running setup at the same time
	Logger(ctx).Debug("Waiting for setup mutex")
	setupMutex.Lock()
	Logger(ctx).Debug("Got setup mutex")
	defer setupMutex.Unlock()
	start := time.Now()
	defer Logger(ctx).Debugf("Setup done after %v", time.Since(start))

	ctx = AddMetadataToCtx(ctx)
	defer Logger(ctx).Debug("Setup--")
	// add the additional directoy path (hashed)
	ptErr := CheckGlobalProjectToken(ctx, globalProjectToken)
	if ptErr != nil {
		Logger(ctx).Errorf("Project token error %s", ptErr.Error())
		return nil, ptErr
	}

	project, err := GetProjectWithToken(ctx)
	if err != nil {
		Logger(ctx).Errorf("Can't get project %s", err.Error())
		return nil, err
	}

	if len(project.ProjectToken) == 0 {
		Logger(ctx).Errorf("Project token is empty %+v", project)
		return nil, ProjectTokenError
	}

	Logger(ctx).Debugf("Project token is %s", project.ProjectToken)

	globalProjectLock.Lock()
	// if the global project token is empty, set it to our project
	if len(globalProjectToken) == 0 {
		globalProjectToken = project.ProjectToken

		Logger(ctx).Infof("Set the project token to %s", globalProjectToken)

	}

	globalProjectLock.Unlock()

	if globalProjectToken != project.ProjectToken {

		return nil, errors.New("Project token is already set")
	}

	// we only do this once
	if len(globalPublicKey) == 0 {

		if _, err := os.Stat("/tmp/.my-private-key"); os.IsNotExist(err) {
			Logger(ctx).Debug("No private key found, creating one")

			privateKey, publicKey, err := CreateKey(ctx, project.Name)

			if err != nil {
				return nil, err
			}
			if viper.GetBool("print_keys") {
				Logger(ctx).Debugf("private key = %s", privateKey)
			}
			Logger(ctx).Debugf("public key = %s", publicKey)
			globalPublicKey = publicKey
			globalPrivateKey = privateKey

			WriteKeyToFile(ctx, []byte(privateKey), "/tmp/.my-private-key")
			WriteKeyToFile(ctx, []byte(publicKey), "/tmp/.my-public-key")
		} else {
			if IsDev() {
				Logger(ctx).Info("Private key found - this suggests that we have restarted this process which is dangerous - but we will allow it in Dev mode")
			} else {
				Logger(ctx).Panic("Private key found - this suggests that we have restarted this process which is dangerous our called Setup twice and created duplicate keys even though we have a global variable which should prevent it- which is vaguely confusing")
			}

		}
	}
	return &pb.Response{Message: globalPrivateKey, Username: project.Username}, nil
}

const name = "supervisor"

// lock channel
var lockChannel = make(chan bool, 1)

// so long as we hold a stream to this lock function, we hold the lock
// as soon as we close the stream, we release the lock
// for graceful stop this will continue to hold the lock because graceful stop just prevents new connections
// however when it is finished the lock will be released and super will be able to close (if it isn't terminated before then)
// our timeout for killing the super and our max test timeout should be the same ? (but not sure how long to have for test timeout)
func (s *server) Lock(LockRequest *brisksupervisor.LockRequest, stream brisksupervisor.BriskSupervisor_LockServer) error {
	ctx := stream.Context()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.New("returned from lock stream"))
	ctx, span := otel.Tracer(name).Start(ctx, "Lock")
	defer span.End()
	defer bugsnag.AutoNotify(ctx)

	startTime := time.Now()
	superTimeout := viper.GetDuration("SUPERVISOR_TIMEOUT")
	projectRunTimeout := viper.GetDuration("PROJECT_RUN_TIMEOUT")

	select {

	case <-lockChannel:
		Logger(ctx).Debugf("reserved super with Lock() after %v", time.Since(startTime))
		rs := brisksupervisor.LockResponse{Message: constants.SUPER_LOCKED, Locked: true}
		streamErr := stream.Send(&rs)
		if streamErr != nil {
			Logger(ctx).Errorf("Error sending lock response %v", streamErr)
			// return streamErr
		}

		select {
		case <-ctx.Done():
			Logger(ctx).Infof("Context done so unlocking %s cause is ", ctx.Err(), context.Cause(ctx))
			lockChannel <- true
			return nil
		case <-time.After(projectRunTimeout):
			// Should we kill the super here?
			// We are in a weird state and don't want to be running again.
			// why can't I just have tons of supers and just manage the workers between them

			Logger(ctx).Error("We timed out waiting for the lock to be released")
			lockChannel <- true
			return errors.New(fmt.Sprintf("Time out waiting for test to finish after %v", projectRunTimeout))

		}

	case <-time.After(superTimeout):
		Logger(ctx).Infof("Can't reserve super, timed out after %v", superTimeout)
		return ProjectInUseError

	case <-ctx.Done():
		Logger(ctx).Infof("Context done so returning context error is %v, cause is %v ", ctx.Err(), context.Cause(ctx))
		return nil
	}

}

func (s *server) RunTests(TestOption *pb.TestOption, stream brisksupervisor.BriskSupervisor_RunTestsServer) error {
	ctx := stream.Context()

	ctx, span := otel.Tracer(name).Start(ctx, "RunTests")
	defer span.End()
	//maybe I need to do all of my defers before the cancelCtx
	ctx, cancel := context.WithCancelCause(ctx)
	errChannel := make(chan error, 1)
	defer cancel(errors.New("returned from RunTests stream"))

	if TestOption.RepoInfo == nil {

		TestOption.RepoInfo = &api.RepoInfo{IsGitRepo: false}

	}

	ctx, aErr := auth.PropagateCredentials(ctx)
	if aErr != nil {
		Logger(ctx).Errorf("Auth error %s", aErr.Error())
		return aErr
	}

	Logger(ctx).Debugf("RunTests++")

	defer bugsnag.AutoNotify(ctx)

	startRunTests := time.Now()
	// if reserveOrWaitForSuper(ctx) {
	// 	// shouldn't this be protected by a lock already ??
	// 	Logger(ctx).Debugf("locked super ")
	// 	defer unreserveSuper(ctx)
	// } else {
	// 	Logger(ctx).Errorf("Can't run tests")
	// 	return ProjectInUseError
	// }

	defer func() {
		Logger(ctx).Debugf("TIMING Run tests took %v ", time.Since(startRunTests))
		Logger(ctx).Debugf("RunTests--")
	}()

	Logger(ctx).Debugf("Run tests has TestOption = %+v", TestOption)
	Logger(ctx).Debugf("Environment passed to RunTests = %+v", TestOption.Config.Environment)

	if (TestOption.Config == nil || TestOption.Config == &pb.Config{} || len(TestOption.Config.Framework) == 0) {
		return errors.New("empty config passed to RunTests")
	}
	if len(TestOption.Commands) == 0 {
		return errors.New("no commands passed to RunTests")
	}

	Logger(ctx).Debugf("The context is %+v ", ctx)
	Logger(ctx).Debugf("The test option : %+v", TestOption)

	if len(globalProjectToken) == 0 {
		Logger(ctx).Error("super globalToken missing no project token in global - so setup has not been run")
		return errors.New("super globalToken missing")

	}

	ptErr := CheckGlobalProjectToken(ctx, globalProjectToken)
	if ptErr != nil {
		return ptErr
	}

	contextKeyResponseStream := ContextKey("response-stream")

	ctx = AddMetadataToCtx(ctx)
	var responseStream = make(chan *pb.Output, 20000)
	ctx = context.WithValue(ctx, contextKeyResponseStream, &responseStream)
	defer func(rS chan *pb.Output) {
		if r := recover(); r != nil {
			Logger(ctx).Debug("Recovered in f", r)
			Logger(ctx).Errorf("Error: %v", r)
			Logger(ctx).Errorf("Panic recovered in RunTests %s", r)
			Logger(ctx).Error(string(debug.Stack()))
			rS <- &pb.Output{Response: fmt.Sprintf("Error: %v", r), Stderr: fmt.Sprintf("Error: %v", r), Created: timestamppb.Now()}
			Logger(ctx).Panic("Panic in RunTests - repanicing %v", r)
		}
	}(responseStream)
	// how big is this buffer going to get?
	s3Buffer := &bytes.Buffer{}
	logUid := uuid.New().String()
	Logger(ctx).Infof("logUid for this group of runs is %s", logUid)
	streamInfo := types.LogStreamInfo{LogUid: logUid}
	defer sendBufferToS3(context.Background(), s3Buffer, streamInfo, Logger(ctx))

	go func() {

		for rs := range responseStream {

			output := rs
			if output.Stderr != "" {
				s3Buffer.WriteString(fmt.Sprintf("%v %v : %v ", output.Created.AsTime().Format("2006-01-02 15:04:05"), PrintWorkerDetails(output), output.Stderr))
				s3Buffer.WriteRune('\n')

			}

			if output.Stdout != "" {
				s3Buffer.WriteString(fmt.Sprintf("%v %v : %v ", output.Created.AsTime().Format("2006-01-02 15:04:05"), PrintWorkerDetails(output), output.Stdout))
				s3Buffer.WriteRune('\n')
			}

			if output.Response != "" && output.Response != output.Stderr && output.Response != output.Stdout {
				// only want to share when it's not the same as stderr or stdout
				s3Buffer.WriteString(fmt.Sprintf("%v %v : %v ", output.Created.AsTime().Format("2006-01-02 15:04:05"), PrintWorkerDetails(output), output.Response))
				s3Buffer.WriteRune('\n')
			}

			Logger(ctx).Debugf("sending Response: %+v", rs)

			streamErr := stream.Send(rs)
			if streamErr != nil {
				Logger(ctx).Errorf("Error sending response %v", streamErr)
				errChannel <- streamErr
				return
			}

		}
	}()
	Logger(ctx).Debugf("The environment passed to the Super is %+v", TestOption.Config.Environment)
	Logger(ctx).Debugf("TIMING runTestTheTests started at %v", time.Since(startRunTests))

	protectedCounter := ProtectedCounter{}
	protectedCounter.Add(len(TestOption.Commands))

	for _, c := range TestOption.Commands {
		go func(command api.Command) {
			retry := true
			for retryCount := 0; retry && retryCount < (1+viper.GetInt("REBUILD_RETRY")); retryCount++ {

				var outErr error

				retry, outErr = runTestTheTests(ctx, responseStream, TestOption.BuildCommands, command, TestOption.Config, *TestOption.RepoInfo, logUid)

				if outErr != nil {
					if outErr == context.Canceled {
						Logger(ctx).Debugf("Context was cancelled %v - not notifying bugsnag", outErr)
						if errors.Is(outErr, TestFailedError) {
							Logger(ctx).Debugf("Test failed error %v - not notifying bugsnag", outErr)

						} else {
							bugsnag.Notify(outErr)

						}
					}

					Logger(ctx).Errorf("error from test run %v", outErr)
					responseStream <- &pb.Output{Response: "error from test run", Stdout: "error from test run", Stderr: outErr.Error(), Created: timestamppb.Now()}

					errChannel <- outErr
					return
				}
				if retry {
					Logger(ctx).Info("Need to rebuild workers but waited 15 seconds 15 times so it's not happening")
					responseStream <- &pb.Output{Response: "File change detected in rebuild list - need to rebuild workers", Created: timestamppb.Now()}
					time.Sleep(viper.GetDuration("REBUILD_BACKOFF"))
				}

			}

			if retry {
				// we've left the loop because we've retried too many times
				Logger(ctx).Error("Timed out waiting for workers to rebuild")
				responseStream <- &pb.Output{Response: "Timed out waiting for workers to rebuild", Created: timestamppb.Now()}
				errChannel <- errors.New("Timed out waiting for workers to rebuild")
				return
			}

			errChannel <- nil

		}(*c)
	}

	Logger(ctx).Info("Waiting at end of RunTests")

	for err := range errChannel {

		if err != nil {
			Logger(ctx).Errorf("Got an error from test run %v", err)
			// this is where we could decide to wait for everyone to finish - but we are failing fast now
			return err
		} else {
			Logger(ctx).Info("No error from test run - things finished normally")
			protectedCounter.Done()
			if protectedCounter.Finished() {
				Logger(ctx).Info("All tests finished")
				break
			}
		}

	}
	// we get here if the errChannel is closed without an error being sent to it
	Logger(ctx).Info("No error from test run - things finished normally")

	return nil
}
func CheckIfFirstRun(ctx context.Context, filePath string) error {
	Logger(ctx).Debug("Checking if this is the first run")
	Logger(ctx).Info("in CheckIfFirstRun printing out the stacktrace so we know where we came from")
	debug.PrintStack()
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			Logger(ctx).Debug("This is the first run")
			file, err := os.Create(filePath)
			if err != nil {
				Logger(ctx).Errorf("Error creating firstRun file %v", err)
				return err
			}
			defer file.Close()
			os.WriteFile(filePath, []byte(time.Now().String()), 0644)
			return nil
		}

		Logger(ctx).Errorf("Error checking firstRun file %v", err)
	}
	timestamp, err := os.ReadFile(filePath)
	if err != nil {
		Logger(ctx).Errorf("Error reading firstRun file %v", err)
		return err
	}

	return errors.New("Super has been run before at " + string(timestamp))

}
func init() {
	if os.Getenv("BUGSNAG_API_KEY") == "" {

		bugsnag.Configure(bugsnag.Configuration{
			APIKey: os.Getenv("BUGSNAG_API_KEY"),
			// The import paths for the Go packages containing your source files
			ProjectPackages: []string{"main", "brisk-supervisor"},
			AppType:         "supervisor",
			ReleaseStage:    os.Getenv("RELEASE_STAGE"),
		})
	}
}
func main() {
	fmt.Print("Starting up the Supervisor - in main\n")

	// runtime.SetBlockProfileRate(1)

	ctx := context.Background()
	env.InitServerViper(ctx)
	fmt.Println("init env done")
	Logger(context.Background()).Info("Starting up the Super")
	Logger(context.Background()).Sync()
	var super *api.Super

	firstRunFile := os.Getenv("FIRST_RUN_FILE")
	if len(firstRunFile) == 0 {
		firstRunFile = constants.SUPER_FIRST_RUN_FILE
	}
	if !IsDev() {
		// we only want to check if this is the first run in production
		err := CheckIfFirstRun(ctx, firstRunFile)
		if err != nil {
			Logger(ctx).Errorf("Error checking if this is the first run : %v", err)
			SafeExit(err)
		}
		Logger(ctx).Error("This is the first run")

	}

	cleanup := honeycomb.InitTracer()
	defer cleanup()

	maxMemory := viper.GetInt64("MAX_MEMORY")
	go MemoryKiller(ctx, maxMemory, os.Getpid())

	brisk_metrics.StartPrometheusServer(ctx)

	// go func() {
	// 	OutputRuntimeStats(ctx)
	// }()

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		// if we're in here, we had a panic and have caught it
	// 		Logger(ctx).Debugf("we safely caught the panic: %s\n", err)
	// 		DeRegisterSuper(super)
	// 		SafeExit(err.(error))

	// 	}
	// }()
	Logger(ctx).Info("Init Super")

	serverGracefulStop := listenOn(ctx, ":50050")

	// need to regsiter the machine if is has not already been registered
	// need to do this for k8s

	if viper.GetBool("USE_KUBERNETES") {
		Logger(ctx).Info("Using Kubernetes")
		err := Registerk8Client(ctx)
		if err != nil {
			Logger(ctx).Errorf("Can't register k8 client %v", err)
			SafeExit(err)
		}
	}

	if viper.GetBool("USE_DOCKER_COMPOSE") {
		Logger(ctx).Info("Using Docker Compose")
		err := RegisterLocalDockerClient(ctx)
		if err != nil {
			Logger(ctx).Errorf("Can't register docker compose client %v", err)
			SafeExit(err)
		}
	}

	Logger(ctx).Info("Registering Super")

	for i := 0; i < 10; i++ {
		var err error
		super, err = RegisterSuper(ctx)

		if err != nil {
			Logger(ctx).Errorf("Can't register super : %v", err)

			Logger(ctx).Debug(errors.Wrap(err, 1).ErrorStack())
			Logger(ctx).Debug("Sleeping 5 seconds and trying again")
			time.Sleep(10 * time.Second)
			if i == 9 {
				Logger(ctx).Error("Can't register super - giving up")
				SafeExit(err)
			}
		} else {
			break
		}

	}

	// this starts us off listening for connections from the client
	syncChannel <- true
	lockChannel <- true

	go func() {

		killPeriod := viper.GetDuration("SUPER_KILL_TIME")
		// we jitter up to quarter the kill period
		killPeriod += jitter(int(killPeriod.Minutes() / 4))
		Logger(ctx).Debugf("We will kill super after %v", killPeriod)
		select {
		case <-ctx.Done():
			Logger(ctx).Infof("Context done - exiting , cause is %v", context.Cause(ctx))
			return
		case <-time.After(killPeriod):
			time.Sleep(killPeriod)
			Logger(ctx).Warn("Killing super on schedule after %v", killPeriod)
			fmt.Printf("Killing super on schedule after %v", killPeriod)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}

	}()

	WaitForSignals(ctx)
	DeRegisterSuper(ctx, super)
	Logger(ctx).Warn("Deregistered super")
	Logger(ctx).Warn("Calling graceful stop")
	Logger(ctx).Sync()
	stopTime := time.Now()
	serverGracefulStop()
	Logger(ctx).Warnf("Graceful stop took %v for super %v", time.Since(stopTime), super.Uid)
	Logger(ctx).Debug("Graceful stop finished")
	Logger(ctx).Warn("Quitting")
	fmt.Println("Graceful Stop - Quitting")
	Logger(ctx).Sync()

}

// gives us a random jitter of up to t minutes
func jitter(t int) time.Duration {
	return time.Duration(rand.Intn(t)) * time.Minute
}

func listenOn(ctx context.Context, port string) func() {
	Logger(ctx).Debugf("ListenOn, port : %v", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		Logger(ctx).Errorf("failed to listen: %v", err)
		SafeExit(err)
	}
	Logger(ctx).Debugf("Continuing in listenOn")
	var kaep = keepalive.EnforcementPolicy{
		MinTime:             2 * time.Second, // If a client pings more than once every 10 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute, // If a client is idle for 15 minutes, send a GOAWAY
		MaxConnectionAge:      20 * time.Hour,   // If any connection is alive for more than 30 hours, send a GOAWAY
		MaxConnectionAgeGrace: 10 * time.Second, // Allow 10 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  10 * time.Second, // Ping the client if it is idle for 10 seconds to ensure the connection is still active
		Timeout:               5 * time.Second,  // Wait 5 second for the ping ack before assuming the connection is dead
	}

	s := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp),
		grpc.ChainUnaryInterceptor(UnaryServerLoggingInterceptor, grpcotel.UnaryServerInterceptor(otelgrpc.WithInterceptorFilter(
			filters.Not(
				filters.HealthCheck(),
			),
		),
		), grpc_auth.UnaryServerInterceptor(DoAuth),
		),
		grpc.ChainStreamInterceptor(StreamServerLoggingInterceptor, grpcotel.StreamServerInterceptor(otelgrpc.WithInterceptorFilter(
			filters.Not(
				filters.HealthCheck(),
			),
		),
		),
			grpc_recovery.StreamServerInterceptor(),
			grpc_ctxtags.StreamServerInterceptor(),
			//grpc_opentracing.StreamServerInterceptor(),
			// grpc_prometheus.StreamServerInterceptor,
			//grpc_zap.StreamServerInterceptor(zapLogger),
			grpc_auth.StreamServerInterceptor(DoAuth),
		))
	Logger(ctx).Debugf("Listening on port %v", port)
	go func() {
		// adding this for testing purposes - won't be accesible from outside the control pane
		http.ListenAndServe(":8024", nil)
	}()
	pb.RegisterBriskSupervisorServer(s, &server{})

	RegisterHealthCheck(ctx, s)
	go func() {
		if err := s.Serve(lis); err != nil {
			Logger(ctx).Errorf("failed to serve: %v", err)
			SafeExit(err)
		}
	}()
	return s.GracefulStop
}

// might think about setting a different directory for each project
// could help with caching, debuging and provide a tiny bit of obscurity
func addDefaultRemoteDirectory(ctx context.Context, buildCommands []*api.Command, command api.Command) ([]*api.Command, api.Command) {
	for i := range buildCommands {

		if buildCommands[i].WorkDirectory == "" {
			buildCommands[i].WorkDirectory = "/tmp/remote_dir/"
		}
	}
	// for i, _ := range commands {
	// 	if commands[i].WorkDirectory == "" {

	// 		commands[i].WorkDirectory = "/tmp/remote_dir/"
	// 	}
	// }
	command.WorkDirectory = "/tmp/remote_dir/"
	return buildCommands, command

}

func AddDefaultRemoteDirectorybriskSupervisor(ctx context.Context, buildCommands []api.Command, commands []api.Command) ([]api.Command, []api.Command) {

	for i := range buildCommands {

		if buildCommands[i].WorkDirectory == "" {
			fmt.Println("CHANGING buildC")
			buildCommands[i].WorkDirectory = "/tmp/remote_dir/"
		}
	}
	for i := range commands {
		if commands[i].WorkDirectory == "" {

			commands[i].WorkDirectory = "/tmp/remote_dir/"
		}
	}
	return buildCommands, commands
}

func runTestTheTests(ctx context.Context, responseStream chan *pb.Output, buildCommands []*api.Command, command api.Command, config *brisksupervisor.Config, repoInfo api.RepoInfo, logUid string) (bool, error) {
	ctx, cancel := context.WithCancelCause(ctx)

	// if len(commands) == 0 {
	// 	Logger(ctx).Errorf("Incorrect Config provided : commands is  %+v", commands)
	// 	return false, errors.New("Incorrect Config: commands can not be blank ")
	// }
	defer cancel(errors.New("returned from runTestTheTests"))
	ctx, runTestTheTestsSpan := otel.Tracer(name).Start(ctx, "runTestTheTests")
	defer runTestTheTestsSpan.End()

	fmt.Printf("Logger(ctx): %+v\n", Logger(ctx))
	startRunTestTheTests := time.Now()
	defer func() {
		Logger(ctx).Debugf("runTestTheTests took %v", time.Since(startRunTestTheTests))
	}()

	if config.WorkerImage == "" {
		Logger(ctx).Errorf("Incorrect Config provided : config is  %+v", config)
		return false, errors.New("Incorrect Config: image can not be blank ")
	}

	buildCommands, command = addDefaultRemoteDirectory(ctx, buildCommands, command)
	// traceKey := md["trace-key"][0]
	// ConfigureLogger(traceKey)

	Logger(ctx).Debugf("Run test the %v tests has buildCommands with values: %+v", len(buildCommands), buildCommands)

	ctx = WithNomadAllocId(ctx, nomad.GetSmallNomadAllocId())

	ctx = context.WithValue(ctx, "outputStream", types.CtxOutputStream{OutputChannel: responseStream})

	envString := os.Getenv("MAX_WORKERS")
	var requestedWorkers int
	if command.CommandConcurrency != 0 {
		requestedWorkers = int(command.CommandConcurrency)
	} else {

		requestedWorkers = int(config.Concurrency)
	}
	maxWorkers, err := strconv.Atoi(envString)

	if err != nil {
		Logger(ctx).Errorf("Can't parse MAX_WORKERS %v", err)
		maxWorkers = 5

	}
	if requestedWorkers > maxWorkers {
		requestedWorkers = maxWorkers
	}

	request := command
	startWorkerTime := time.Now()
	var jrunStatus api.JobRunStatus = api.JobRunStatus_failed
	var exitCode int32 = -1
	var output = ""
	var errorOutput = ""
	var jobrunId int32 = -1
	var finalWorkerCount int32 = -1
	var failingWorkers []*api.Worker

	// authcreds, authErr := auth.GetAuthCredsFromMd(ctx)
	// unCancelled, authErr = auth.AddAuthToCtx(unCancelled, authcreds)
	// I set jrunStatus to success at the end
	defer func(ctx context.Context, logger *BriskLogger) {
		// little pause to let the LogRuns finish
		// time.Sleep(1 * time.Second)
		logger.Debugf("Defer func for FinishRun values are jobRunId: %v,	jrunStatus: %v, exitCode: %v, output: %v, errorOutput: %v , finalWorkerCount %v, failingWorkers %v", jobrunId, jrunStatus, exitCode, output, errorOutput, finalWorkerCount, failingWorkers)

		if jobrunId == -1 {
			logger.Errorf("jobrunId is -1 so we cannot finish the run")
			return
		}
		unCancelled := context.Background()
		authcreds, authErr := auth.GetAuthCredsFromMd(ctx)
		if authErr != nil {
			logger.Errorf("Error getting auth creds from md in Finish %v", authErr)
		}
		unCancelled, authErr = auth.AddAuthToCtx(unCancelled, authcreds)
		if authErr != nil {
			logger.Errorf("Error adding auth to ctx in Finish %v", authErr)
		}
		unCancelled, cancel = context.WithCancelCause(unCancelled)
		defer cancel(errors.New("returned from go func the logs the run"))
		runError := FinishRun(unCancelled, jobrunId, jrunStatus, exitCode, output, errorOutput, SuperUID(ctx), finalWorkerCount, failingWorkers, logger)
		if runError != nil {
			logger.Errorf("Error finishing run %v", runError)
		}
	}(ctx, Logger(ctx))

	goodWorkers, jobrunId, jobrunLink, workerErr := getWorkingWorkers(ctx, requestedWorkers, config.WorkerImage, config.RebuildFilePaths, repoInfo, logUid)

	finalWorkerCount = int32(len(goodWorkers))
	if workerErr != nil && strings.Contains(workerErr.Error(), "We don't have enough workers for this run") {
		Logger(ctx).Errorf("Error getting workers for project %v", workerErr.Error())
		timeout := viper.GetDuration("NO_WORKER_TIMEOUT")
		responseStream <- &pb.Output{Response: "Waiting on workers for project - trying again in " + timeout.String(), Stderr: "Waiting on workers for project", Created: timestamppb.Now()}
		time.Sleep(timeout)
		goodWorkers, jobrunId, jobrunLink, workerErr = getWorkingWorkers(ctx, requestedWorkers, config.WorkerImage, config.RebuildFilePaths, repoInfo, logUid)

	}

	printLinkToStream(ctx, responseStream, jobrunLink)
	defer printLinkToStream(ctx, responseStream, jobrunLink)

	if workerErr != nil {
		Logger(ctx).Errorf("Error getting workers for project %v", workerErr.Error())
		responseStream <- &pb.Output{Response: workerErr.Error(), Stderr: workerErr.Error(), Created: timestamppb.Now()}
		return false, workerErr
	}
	endWorkerTime := time.Now()

	Logger(ctx).Debugf("TIMING Got %v workers in %v", len(goodWorkers), endWorkerTime.Sub(startWorkerTime))

	responseStream <- &pb.Output{Response: fmt.Sprintf("Assigned %v servers", len(goodWorkers)), Created: timestamppb.Now()}
	Logger(ctx).Debugf("TIMING Starting sync after %v ", time.Since(startRunTestTheTests))

	beforeSyncCount := len(goodWorkers)
	var syncErr error
	goodWorkers, failingWorkers, syncErr = syncToWorkers(ctx, goodWorkers, config)
	finalWorkerCount = int32(len(goodWorkers))
	if len(failingWorkers) > 0 {
		Logger(ctx).Errorf("Got %v failing workers after sync they are %+v", len(failingWorkers), failingWorkers)
	} else {
		Logger(ctx).Debugf("Got no failing workers after sync")
	}

	if len(goodWorkers) == 0 {
		Logger(ctx).Errorf("Got no good workers after sync - before sync was %v", beforeSyncCount)
		Logger(ctx).Errorf("Sync error is %v ", syncErr)
		responseStream <- &pb.Output{Response: "Got an Error syncing - not recoverable, no workers", Stdout: syncErr.Error(), Stderr: syncErr.Error(), Created: timestamppb.Now()}
		errorOutput = syncErr.Error()
		return false, syncErr
	}

	if syncErr != nil && float64(len(goodWorkers)) < 0.85*float64(beforeSyncCount) {
		mesg := fmt.Sprintf("Got an Error syncing - not recoverable, less than 85%% of workers. Workers count was %v, now %v", beforeSyncCount, len(goodWorkers))
		Logger(ctx).Errorf(mesg)
		responseStream <- &pb.Output{Response: mesg, Stdout: syncErr.Error(), Stderr: syncErr.Error(), Created: timestamppb.Now()}
		errorOutput = syncErr.Error()
		return false, syncErr
	}

	if syncErr != nil {
		Logger(ctx).Errorf("Sync error is %v  Workers count was %v, now %v", syncErr, beforeSyncCount, len(goodWorkers))
		responseStream <- &pb.Output{Response: fmt.Sprintf("Got an Error syncing - recoverable. Workers count was %v, now %v", beforeSyncCount, len(goodWorkers)), Stdout: syncErr.Error(), Stderr: syncErr.Error(), Created: timestamppb.Now()}
		//TODO we need to update the jobrun with the new assigned concurrency - or change the logic at the end of FinishRun - because at present
		// We think it has failed because not all of the workers report back
		// SO maybe UpdateJobrunConcurrency(  new concurrency)
	}
	Logger(ctx).Debugf("TIMING Sync done after %v", time.Since(startRunTestTheTests))

	beforeRebuildCount := len(goodWorkers)
	// only check rebuild if config isn't set
	if os.Getenv("CHECK_WORKER_HASH") == "false" {
		Logger(ctx).Debugf("Skipping worker hash check")
	} else {
		Logger(ctx).Debugf("Checking worker hash")
		ctx, checkSpan := otel.Tracer(name).Start(ctx, "CheckWorkersForRebuild")

		var badWorkers []*api.Worker
		var rebuilderr error
		goodWorkers, badWorkers, rebuilderr = checkWorkersForRebuild(ctx, goodWorkers, config)
		finalWorkerCount = int32(len(goodWorkers))
		if rebuilderr != nil {
			checkSpan.RecordError(rebuilderr)
		}
		checkSpan.End()

		if len(badWorkers) > 0 {

			//we do this here because if we return we will free the workers from the super
			creds, authErr := auth.GetAuthCredsFromMd(ctx)
			if authErr != nil {
				Logger(ctx).Errorf("Error getting auth creds from md in Finish %v", authErr)
				return false, authErr
			}

			ctx, authErr := auth.AddAuthToCtx(ctx, creds)
			if authErr != nil {
				Logger(ctx).Errorf("Error adding auth to ctx  %v", authErr)
				return false, authErr
			}

			clearErr := DeRegisterWorkersForProject(ctx, badWorkers)
			if clearErr != nil {
				Logger(ctx).Errorf("Error clearing workers %v", clearErr)
				return true, clearErr
			}

		}

	}

	//now that we've checked the workers for rebuild
	if float64(len(goodWorkers)) < 0.85*float64(beforeRebuildCount) {
		mesg := fmt.Sprintf("Checking for rebuild: more than 15%% of workers need to be rebuilt.  Workers count was %v, now %v", beforeRebuildCount, len(goodWorkers))
		Logger(ctx).Errorf(mesg)
		responseStream <- &pb.Output{Response: mesg, Created: timestamppb.Now()}
		return true, nil
	} else {
		Logger(ctx).Debugf("Checking for rebuild: less than 15%% of workers need to be rebuilt. Workers count was %v, now %v", beforeRebuildCount, len(goodWorkers))
	}

	// we don't need to warm any workers if we have less than 1

	var warmedWorkers []*api.Worker
	var warmingError []error
	var alreadyWarm []*api.Worker
	var warmingWaitGroup sync.WaitGroup
	warmingWaitGroup.Add(1)
	go func() {

		if len(goodWorkers) > 1 {

			needBuildCommands := []api.Worker{}
			for _, w := range goodWorkers[1 : len(goodWorkers)-1] {
				if w.BuildCommandsRunAt == nil {
					needBuildCommands = append(needBuildCommands, *w)
				} else {
					alreadyWarm = append(alreadyWarm, w)
				}
			}

			if len(needBuildCommands) > 0 {
				warmedWorkers, warmingError = warmOtherWorkers(ctx, needBuildCommands, config, responseStream)
				if len(warmingError) != 0 {
					Logger(ctx).Errorf("Error warming workers %+v", warmingError)
					for _, e := range warmingError {
						responseStream <- &pb.Output{Response: fmt.Sprintf("Error pre-building workers : %v", e), Stderr: e.Error(), Created: timestamppb.Now()}
					}
				}
				if len(warmedWorkers) > 0 {
					Logger(ctx).Debugf("We have prewarmed the following workers %+v", warmedWorkers)
				}
			}

		} else {
			Logger(ctx).Debugf("Skipping warming workers as we have less than 2 workers")
		}
		warmingWaitGroup.Done()
	}()

	// so for non test runs we don't split and we just run the commands
	// going to check the last command maybe we'll want to use the first commands for setup or something
	var myFiles [][]string
	var totalTestCount int

	if command.NoTestFiles {
		Logger(ctx).Debugf("Not a test run so not splitting files")
		totalTestCount = len(goodWorkers)
	} else {
		myFiles, totalTestCount, err = splitTestFiles(ctx, len(goodWorkers), goodWorkers[0], config, responseStream, !config.SkipRecalcFiles)
	}
	if err != nil {
		Logger(ctx).Errorf("Error splitting files: %v", err)

		responseStream <- &pb.Output{Response: "Error splitting tests", Stderr: err.Error(), Created: timestamppb.Now()}

		errorOutput = err.Error()
		return false, err
	}

	warmingWaitGroup.Wait()
	if len(warmingError) > 0 {
		// we literally just checked these workers - so something flakey is going on
		Logger(ctx).Errorf("Error warming workers that were previously checked so failing :  %+v", warmingError)
		// also this means we can assume our goodworkers length is correct and continue with the split
		return false, warmingError[0]
	}

	Logger(ctx).Debugf("TIMING test split done after %v ", time.Since(startRunTestTheTests))

	Logger(ctx).Debugf("About to wait for warming group to finish")
	Logger(ctx).Debugf("Warming group finished")

	Logger(ctx).Debugf("After split test files and warming , good workers are %+v", goodWorkers)

	var wg sync.WaitGroup
	wg.Add(len(goodWorkers))
	exitChannel := make(chan error, 1)
	doneChannel := make(chan bool, 1)

	cs := countStruct{TotalTestCount: totalTestCount}
	Logger(ctx).Debug("about to run tests")
	for i, worker := range goodWorkers {

		go func(i int, worker api.Worker) {
			defer bugsnag.AutoNotify(ctx)
			defer wg.Done()

			var files = []string{}
			if len(myFiles) > i && myFiles[i] != nil && len(myFiles[i]) > 0 {
				Logger(ctx).Debugf("i is %v", i)
				Logger(ctx).Debugf("Files I'm sending are %v", myFiles[i])
				files = myFiles[i]
			} else {
				Logger(ctx).Errorf("We have no files for this worker number %v", i)
				Logger(ctx).Errorf("command is %+v", command)
				// have to have a check here for the number of files with a command
				if config.Framework == string(types.Python) || command.NoTestFiles {
					Logger(ctx).Debug("No test files for this run")
					files = []string{""}
				} else {
					Logger(ctx).Debug("We are not python, so we are not going to send a blank file")
					responseStream <- &pb.Output{Response: "Error running tests - we have more workers than test files and so we are sending an empty file list to one or more workers - reduce BRISK_CONCURRENCY to be less than or equal to the number of tests", Stderr: "Error running tests - we have more workers than test files and so we are sending an empty file list to one or more workers - reduce BRISK_CONCURRENCY to be less than or equal to the number of tests", Created: timestamppb.Now()}

					//TODO verify what happens here  - I think we are panicing in this instance which isn't good
					files = []string{""}
					// wg.Done()
					return
				}
			}

			// prevents weird bugs cause we change these later
			var copiedRequests []api.Command
			// for _, v := range requests {
			copVI, err := copystructure.Copy(request)
			if err != nil {
				Logger(ctx).Errorf("Error copying request %v", err)
				exitChannel <- err
				return
			}
			copV := copVI.(api.Command)
			copV.WorkerNumber = int32(i)
			copV.TotalWorkerCount = int32(len(goodWorkers))
			copV.Stage = "Run"
			// copiedRequests = append(copiedRequests, *copV)
			copiedRequest := copV
			// }

			Logger(ctx).Infof("Copied Requests %+v", copiedRequest)
			// if len(copiedRequests) > 0 {
			// 	Logger(ctx).Infof("Copied Request commandline %v coppiedd Request NoTestFiles %v", copiedRequests[0].Commandline, copiedRequests[0].NoTestFiles)
			// } else {
			// 	Logger(ctx).Infof("Copied Request is empty")
			// }

			var copiedBuildCmds []api.Command
			if len(os.Getenv("ALWAYS_RUN_BUILDCOMMANDS")) > 0 || worker.BuildCommandsRunAt == nil {
				Logger(ctx).Debug("Copying build commands")
				Logger(ctx).Debugf("Build Commands before %+v", buildCommands)

				for _, v := range buildCommands {
					copVI, err := copystructure.Copy(v)
					if err != nil {
						Logger(ctx).Errorf("Error copying build command %v", err)
						exitChannel <- err
						return
					}
					copV := copVI.(*api.Command)
					copV.WorkerNumber = int32(i)
					copV.TotalWorkerCount = int32(len(goodWorkers))
					copV.Stage = "Build"
					copiedBuildCmds = append(copiedBuildCmds, *copV)
				}
				Logger(ctx).Debugf("Copied Build Commands %+v", copiedBuildCmds)

			} else {
				Logger(ctx).Debugf("Build commands already run for worker with id %v so skipping", worker.Id)
			}
			// if we split too much we can have empty files here which causes all tests to be run
			Logger(ctx).Debugf("TIMING about to run connectToServer after %v", time.Since(startRunTestTheTests))
			start_run := timestamppb.Now()
			if err != nil {
				Logger(ctx).Error(err.Error())
				responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Exitcode: exitCode, Created: timestamppb.Now()}
			}
			uid := uuid.New().String()
			//later we'll fill this in

			buffer := &bytes.Buffer{}
			bufferStream := make(chan *pb.Output)
			defer close(bufferStream)
			// we read from the responseStream and write to the buffer and the actual response stream
			go func() {
				for output := range bufferStream {

					if output.Stderr != "" {
						buffer.WriteString(fmt.Sprintf("%v : %v ", PrintWorkerDetails(output), output.Stderr))
						buffer.WriteRune('\n')

					}

					if output.Stdout != "" {
						buffer.WriteString(fmt.Sprintf("%v : %v ", PrintWorkerDetails(output), output.Stdout))
						buffer.WriteRune('\n')
					}

					if output.Response != "" && output.Response != output.Stderr && output.Response != output.Stdout {
						// only want to share when it's not the same as stderr or stdout
						buffer.WriteString(fmt.Sprintf("%v : %v ", PrintWorkerDetails(output), output.Response))
						buffer.WriteRune('\n')
					}
					responseStream <- output

				}
			}()

			streamInfo := types.LogStreamInfo{WorkerRunInfoUID: uid}

			unCancelledContext := context.Background()
			authcreds, authErr := auth.GetAuthCredsFromMd(ctx)
			if authErr != nil {
				Logger(ctx).Errorf("Error getting auth creds from md in Finish %v", authErr)
			}
			unCancelledContext, authErr = auth.AddAuthToCtx(unCancelledContext, authcreds)
			if authErr != nil {
				Logger(ctx).Errorf("Error adding auth to ctx in Finish %v", authErr)
			}
			unCancelledContext, cancelUnCancel := context.WithCancel(unCancelledContext)

			var executionInfo *api.ExecutionInfo
			copiedRequests = append(copiedRequests, copiedRequest)
			executionInfos, err := connectToServer(ctx, files, copiedRequests, copiedBuildCmds, totalTestCount, &cs, bufferStream, worker, config)
			if len(executionInfos) > 0 {
				executionInfo = executionInfos[len(executionInfos)-1]
			} else {
				executionInfo = &api.ExecutionInfo{}
			}

			if err != nil {
				Logger(ctx).Error(err.Error())
				responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Exitcode: exitCode, Created: timestamppb.Now()}
				exitCode = 1
				executionInfo.ExitCode = 1

			} else {

				// we only want to update the exit code if it's not already set to a failure
				if exitCode <= 0 && executionInfo.ExitCode > 0 {
					exitCode = executionInfo.ExitCode
				}
			}
			// we pass the logger cause we still want to associate logs with this context trace key etc
			logLocation, logErr := sendBufferToS3(unCancelledContext, buffer, streamInfo, Logger(ctx))
			if logErr != nil {
				Logger(ctx).Errorf("Error sending buffer to s3 %v", logErr)
			}

			end_run := timestamppb.Now()
			Logger(ctx).Debugf("TIMING Finished connectToServer start_run.AsTime() after %v", time.Since(start_run.AsTime()))
			Logger(ctx).Debugf("TIMING Finished connectToServer startRunTestTheTests after %v", time.Since(startRunTestTheTests))
			Logger(ctx).Debugf("The start_run is %+v", start_run)
			Logger(ctx).Debugf("The end_run is %+v", end_run)
			Logger(ctx).Debugf("The exitCode is %+v", executionInfo.ExitCode)
			Logger(ctx).Debugf("The err is %+v", err)
			Logger(ctx).Debugf("The worker is %+v", worker)

			var errorString string
			if err != nil {
				errorString = err.Error()
			} else {
				errorString = ""
			}

			ri := api.RunInfo{JobrunId: uint32(jobrunId), RebuildHash: executionInfo.RebuildHash, LogLocation: logLocation, Uid: uid, Files: files, WorkerId: uint32(worker.Id), ExitCode: strconv.Itoa(int(executionInfo.ExitCode)), Error: errorString, StartedAt: executionInfo.Started, FinishedAt: executionInfo.Finished, ExecutionInfos: executionInfos}

			Logger(ctx).Debugf("The run info is %+v", ri)

			// go func() {
			logRunErr := LogRun(unCancelledContext, &ri, Logger(ctx), &command)
			if logRunErr != nil {
				Logger(ctx).Errorf("Error logging run %v - run info is %+v", logRunErr, ri)
				responseStream <- &pb.Output{Stderr: logRunErr.Error(), Created: timestamppb.Now()}
			}
			defer cancelUnCancel()
			// }()

			if err != nil {
				Logger(ctx).Error("Error at end of loop: ", err.Error())
				responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Exitcode: executionInfo.ExitCode, Created: timestamppb.Now()}
				if config.NoFailFast {
					Logger(ctx).Debugf("Not failing fast so not killing other workers")
				} else {
					exitChannel <- err

				}
			}

			Logger(ctx).Debug("Finished with worker - Done %v", worker.Id)
			if !config.NoFailFast && executionInfo.ExitCode != 0 {
				Logger(ctx).Debugf("Failing fast so killing other workers")
				exitChannel <- TestFailedError
			}

		}(i, *worker)

	}

	go func() {
		Logger(ctx).Debug("Waiting for all workers to finish")
		wg.Wait()
		if exitCode == -1 {
			Logger(ctx).Debug("No exit code change set so setting to 0 (success)")
			exitCode = 0
		}
		Logger(ctx).Debug("Finished waiting for workers to finish")

		close(doneChannel)
		Logger(ctx).Debug("Closed done channel")
	}()

	select {

	// we probably do want to respond to the stream getting cancelled somewhere
	// case c := <-ctx.Done():
	// 	Logger(ctx).Debugf("Got context done %v", c)
	// 	Logger(ctx).Infof("Context done with cause %v", context.Cause(ctx))

	// 	return false, ctx.Err()
	case e := <-exitChannel:
		Logger(ctx).Debugf("Got exit channel with error %v", e)
		return false, e
	case <-doneChannel:
		Logger(ctx).Debugf("Got done channel")
		if exitCode == 0 {
			jrunStatus = api.JobRunStatus_completed

		} else {
			jrunStatus = api.JobRunStatus_failed
		}
		Logger(ctx).Debugf("Got all done")
		return false, nil
		// this seeems to cause problems with stuff not erroring out when it fails/timeouts
		// case <-ctx.Done():
		// 	Logger(ctx).Debugf("Got context done")
		// 	return false, nil

	}

}

// returns true if we need to rebuild
func CheckRebuild(ctx context.Context, worker *api.Worker, config *pb.Config) (bool, error) {
	Logger(ctx).Debug("Super Checking for rebuild")
	conn, err := createConnectionToWorkerNoRetry(ctx, worker, config)
	if err != nil {
		Logger(ctx).Errorf("Error creating connection to worker %v", err)
		return false, err
	}

	defer conn.Close()
	server := brisksupervisor.NewCommandRunnerClient(conn)

	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("CHECK_REBUILD_TIMEOUT"))
	defer CancelCtx(ctx, cancel, "CheckRebuild finished so we are cancelling the context")

	req := &brisksupervisor.CheckBuildMsg{Config: config}
	res, err := server.CheckBuild(ctx, req)
	if err != nil {
		Logger(ctx).Errorf("Super error checking for rebuild %v", err)
		return false, err
	}
	Logger(ctx).Infof("The response from CheckBuild is %v with error %v", res.Success, err)

	Logger(ctx).Debugf("Super Check for rebuild response %v", res)
	return res.Success, nil

}

func checkWorkersForRebuild(ctx context.Context, goodWorkers []*api.Worker, config *pb.Config) ([]*api.Worker, []*api.Worker, error) {
	Logger(ctx).Debug("Checking for rebuild")
	startOfCheckWorkers := time.Now()
	defer func() {
		Logger(ctx).Debugf("TIMING: Checking for rebuild took %v", time.Since(startOfCheckWorkers))
	}()

	// Create a channel to receive results from the goroutines.
	workersThatNeedRebuilding := []*api.Worker{}
	safeWorkers := []*api.Worker{}
	errCh := make(chan error, len(goodWorkers))

	// Use a WaitGroup to wait for all the goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(len(goodWorkers))
	var appendMutex sync.Mutex
	// Run CheckRebuild concurrently for each worker.
	for _, worker := range goodWorkers {
		go func(w *api.Worker) {
			defer wg.Done()

			rebuild, err := CheckRebuild(ctx, w, config)
			if err != nil || rebuild {
				// if we error or need to rebuild, add the worker to the list of rebuilds
				appendMutex.Lock()
				workersThatNeedRebuilding = append(workersThatNeedRebuilding, w)
				appendMutex.Unlock()
			} else {
				Logger(ctx).Debugf("Worker %v does not need to be rebuilt", w.Id)
				appendMutex.Lock()
				safeWorkers = append(safeWorkers, w)
				appendMutex.Unlock()
				return
			}

			if err != nil {
				Logger(ctx).Errorf("Error checking for rebuild error : %v", err)
				errCh <- err
				return
			}
			if rebuild {
				Logger(ctx).Debugf("Worker %v needs to be rebuilt", w.Id)
				return
			}

		}(worker)
	}

	// Wait for all the goroutines to finish.
	wg.Wait()
	close(errCh)

	// Check the results from the channel to see if any workers need to be rebuilt.
	var errs []error

	for err := range errCh {
		Logger(ctx).Errorf("Encountered %v during CheckRebuild", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		Logger(ctx).Errorf("Encountered %d errors during CheckRebuild - returning first - %v", len(errs), errs[0])
	}

	Logger(ctx).Debug("No workers need to be rebuilt")

	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	return safeWorkers, workersThatNeedRebuilding, err
}

func createConnectionToWorkerNoRetry(ctx context.Context, worker *api.Worker, config *pb.Config) (*grpc.ClientConn, error) {

	// Set up a connection to the server.

	endpoint := worker.IpAddress + ":" + worker.Port
	if len(endpoint) <= 0 {
		Logger(ctx).Errorf("No endpoint for worker %+v", worker)

		SafeExit(errors.Errorf("Need an endpoint for worker + %v", worker))
	}
	ctx = AddIntendedAllocIdTo(ctx, worker.Uid)

	var dialOpt grpc.DialOption
	if IsDev() {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		tlsConfig := &tls.Config{
			//we aren't running a certificate authority so we need to disable this
			//this traffic is also internal so we don't need to worry about it being MITM'd
			InsecureSkipVerify: true,
		}
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	Logger(ctx).Debugf("Connecting to endpoint %v", endpoint)

	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), BugsnagClientInterceptor()),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), BugsnagClientUnaryInterceptor),
		grpc.WithDefaultCallOptions(), grpc.WithBlock(), grpc.WithTimeout(1000*time.Millisecond),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
		dialOpt,
	)

	if err != nil {
		return nil, err
	}
	Logger(ctx).Debug("Connected ")

	return conn, nil

}

// we get a list of files
// we respond with the final exit code and how long the final command took
// we are ignoring build commands for that calculcation
func connectToServer(ctx context.Context, files []string, requests []api.Command, buildCommands []api.Command, totalFileCount int, cs *countStruct, responseStream chan *pb.Output, worker api.Worker, config *brisksupervisor.Config) ([]*api.ExecutionInfo, error) {

	// We need to check to see if context is cancelled and if it is we need to return from this function - now I think we are just hanging out here forever and that is not good

	//long ass timeout
	ctx, span := otel.Tracer(name).Start(ctx, "connectToServer(worker)")
	defer span.End()

	//this is going to be empty until the end

	//10 minute + 10 seconds  timeout for the whole run

	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("COMMAND_TIMEOUT"))
	defer CancelCtx(ctx, cancel, "connectToServer has finished so we are cancelling the context")

	// Set up a connection to the server.
	Logger(ctx).Debugf("Connect to server for worker %v/%v has build commands %+v", worker.Id, worker.Uid, buildCommands)
	endpoint := worker.IpAddress + ":" + worker.Port
	if len(endpoint) <= 0 {
		Logger(ctx).Debugf("the type of %T", endpoint)
		SafeExit(errors.Errorf("Need an endpoint for worker + %+v", worker))

	}
	ctx = AddIntendedAllocIdTo(ctx, worker.Uid)
	size := 1024 * 1024 * 1024 * 4

	Logger(ctx).Debugf("the files we have are of length %v", len(files))
	Logger(ctx).Debugf("Connecting to endpoint %v", endpoint)

	var dialOpt grpc.DialOption

	if IsDev() {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		tlsConfig := &tls.Config{
			//we aren't running a certificate authority so we need to disable this
			InsecureSkipVerify: true,
		}
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	startConn := time.Now()
	connectCount := 5
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < connectCount; i++ {
		err = nil
		dialCtx, dialCancel := context.WithTimeout(ctx, 20*time.Second)
		defer dialCancel()
		conn, err = grpc.DialContext(dialCtx, endpoint,
			grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), BugsnagClientInterceptor()),
			grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), BugsnagClientUnaryInterceptor),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(size)),
			dialOpt, grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    10 * time.Second,
				Timeout: 5 * time.Second,
			}))

		if err != nil {
			Logger(ctx).Errorf("Error from DialContext when trying to connect to worker err: (%+v) worker: (%+v) time getting connection took %v on attempt %v ", err, worker, time.Since(startConn), i)
			Logger(ctx).Errorf("I think this might leave these workers in an inconsistent state so maybe we need to remove them")
			if status, ok := status.FromError(err); ok {

				Logger(ctx).Debugf("Exit Status: %d", status.Message())
				Logger(ctx).Debugf("Exit Code: %d", status.Code())
				Logger(ctx).Errorf("Error (%v) from DialContext with status : %v", err, status.Message())

			} else {
				Logger(ctx).Errorf("Error from DialContext: %v ", err)

			}
			connectCount++

		} else {
			// conn must be the connection
			break
		}

	}
	if err != nil {
		Logger(ctx).Errorf("Error from DialContext when trying to connect to worker err: (%+v) worker: (%+v) time getting connection took %v", err, worker, time.Since(startConn))
		responseStream <- &pb.Output{Control: types.FAILED, BriskError: &pb.BriskError{Error: "internal error - could not connect to workers"}, Created: timestamppb.Now()}

		return []*api.ExecutionInfo{{ExitCode: 50, Output: "internal error - could not connect to worker"}}, err
	}
	if conn == nil {
		Logger(ctx).Error("conn is nil - this should not happen")
		return []*api.ExecutionInfo{{ExitCode: 50, Output: "internal error - could not create connection to worker"}}, err

	}

	Logger(ctx).Debugf("Connected in %v", time.Since(startConn))

	defer conn.Close()
	c := pb.NewCommandRunnerClient(conn)

	Logger(ctx).Debug("before creating RunCommands ")

	r, err := c.RunCommands(ctx)

	if err != nil {
		Logger(ctx).Errorf("Error from Runcommands setup: %v", err)
		// SafeExit(err)
		return []*api.ExecutionInfo{{ExitCode: 51, Output: "Error running commnads: " + err.Error()}}, err
	}
	defer r.CloseSend()

	// Going to buffer 100 responses - unsure if it makes things faster or what happens...
	responseChannel := make(chan ResponseOutput, 100)

	go readToChannel(ctx, r, responseChannel)

	filenames := files
	if len(requests) == 0 {
		Logger(ctx).Debug("No commands to run!")
		return []*api.ExecutionInfo{{ExitCode: 11, Output: "No commands"}}, errors.New("no commands to run")
	}
	// we add the filenames to the first of the requests, then we do it after the builds
	requests[0].Args = append(requests[0].Args, filenames...)
	if !requests[0].IsListTest {
		requests[0].IsTestRun = true
	}
	Logger(ctx).Debugf("The build commands have count %v and values =  %+v", len(buildCommands), buildCommands)
	if worker.BuildCommandsRunAt == nil {
		responseStream <- &pb.Output{Stdout: "Building Worker", Response: "Building Worker", Created: timestamppb.Now()}

		requests = append(buildCommands, requests...)

		responseStream <- &pb.Output{Stdout: fmt.Sprintf("All the commands are %+v", requests), Response: fmt.Sprintf("Build commands are %+v", requests), Created: timestamppb.Now()}
	} else {
		responseStream <- &pb.Output{Stdout: "Not Building Worker", Response: "Not Building Worker", Created: timestamppb.Now()}
	}
	startTime := time.Now()
	// this exit code is used internally - but we pass back the results from the worker
	exitCode := int32(-1)
	// executionInfo := &api.ExecutionInfo{ExitCode: 12}
	// this is where we keep all the execution infos
	allExecutionInfo := []*api.ExecutionInfo{}

	for i, command := range requests {
		//exitCode is confusing here - we are setting it before in this loop cause we want it to have a value of -1
		// this 0 can come from build commands being run, thats what is happening here - if build commands were run set it
		if exitCode == 0 {
			if worker.BuildCommandsRunAt == nil && command.IsTestRun {
				// if we are running a command we must have run the build commands
				Logger(ctx).Debugf("Setting the build commands run at to %v for %v", startTime, worker.Uid)
				go setBuildCommandsRunAt(ctx, worker)
				Logger(ctx).Debug("Not waiting for build commands set to run")
			}
		} else {
			Logger(ctx).Debugf("Exit code is %v - not setting build commands for worker %v", exitCode, worker.Uid)
		}

		Logger(ctx).Debugf("%v:- TIMING at top of request loop time since start is %v and command is %v  ", worker.Uid, time.Since(startTime), command.Commandline)

		Logger(ctx).Debugf("In command loop we have looped %v times", i)
		command.SequenceNumber = int32(i)
		Logger(ctx).Debugf("The command I'm sending is %+v", command)
		responseStream <- &pb.Output{Stdout: fmt.Sprintf("Sending command %+v", command), Response: fmt.Sprintf("Sending command %v", command.Commandline), Created: timestamppb.Now()}

		if i == len(requests)-1 {
			command.LastCommand = true
		}
		Logger(ctx).Infof("Sending command line %v with NoTestFiles %v and CommandConcurrency %v to worker %v", command.Commandline, command.NoTestFiles, command.CommandConcurrency, worker.Uid)
		cmdErr := r.Send(&command)
		if cmdErr != nil {
			Logger(ctx).Errorf("Error sending command %v", cmdErr)
			return []*api.ExecutionInfo{{ExitCode: 12, Output: fmt.Sprintf("Error (%v) sending command ", cmdErr)}}, cmdErr
		}

		//go func() {
	CommandFor:
		for {

			select {
			case <-time.After(viper.GetDuration("COMMAND_TIMEOUT")):
				Logger(ctx).Errorf("Timed out after %v", viper.GetDuration("COMMAND_TIMEOUT"))
				return append(allExecutionInfo, &api.ExecutionInfo{ExitCode: 13, Output: fmt.Sprintf("Timed out after %v", viper.GetDuration("COMMAND_TIMEOUT"))}), errors.New("command timedout")
			case <-ctx.Done():
				Logger(ctx).Info("Context done - closing the connection cause is %v", context.Cause(ctx))
				r.CloseSend()

			case input := <-responseChannel:

				if (input == ResponseOutput{}) {
					Logger(ctx).Debugf("Recv'd empty response - done %v", err)
					return allExecutionInfo, nil
				}
				in := input.output
				err := input.error

				if err == io.EOF {
					// read done.
					Logger(ctx).Debugf("Recv'd EOF - done %v", err)
					return allExecutionInfo, nil
				}
				if in != nil {
					if in.ExecutionInfo != nil {
						allExecutionInfo = append(allExecutionInfo, in.ExecutionInfo)
					}
					// sequence check
					if command.SequenceNumber != in.CmdSqNum {
						// this could be to do with a background job being run
						Logger(ctx).Debugf("Out of sequence - command is %v and response is %v", command, in)
					}
					exitCode = in.Exitcode
				}
				if err != nil {
					exitCode = -2
				}

				if errors.Is(err, context.Canceled) {
					Logger(ctx).Debugf("Seems like this context is already cancelled - returning nil %v", err)
					return allExecutionInfo, nil
				}
				if status.Code(err) == codes.Canceled {
					Logger(ctx).Debugf("Seems like the stream is cancelled with %v - returning nil unsure if this is correct is there a valid reason for the server to cancel a stream and for us to swallow the error", err)
					return allExecutionInfo, nil
				}

				if err == io.ErrUnexpectedEOF {
					// read done.
					Logger(ctx).Debugf("Recv'd done with unexpectedEOF %v", err)

				}

				if err != nil && status.Convert(err).Code() == codes.Unavailable {
					responseStream <- &pb.Output{Stderr: "Error connecting to worker " + worker.Uid + " " + err.Error(), Created: timestamppb.Now()}

					return allExecutionInfo, err

				}

				if err != nil {

					bugsnag.Notify(err)

					Logger(ctx).Errorf("Got error : %v", err)
					Logger(ctx).Errorf("Worker info is: %+v", worker)
					if in != nil {
						responseStream <- in

					}
					if status, ok := status.FromError(err); ok {
						Logger(ctx).Debugf("Got status from error :- %+v", status.Message())
						responseStream <- &pb.Output{Stdout: status.Message() + " " + err.Error(), Response: status.Message() + " " + err.Error(), Stderr: err.Error(), Created: timestamppb.Now()}

					} else {
						responseStream <- &pb.Output{Stdout: err.Error(), Response: err.Error(), Stderr: err.Error(), Created: timestamppb.Now()}
					}

					return allExecutionInfo, err

				}
				Logger(ctx).Debug(in)
				Logger(ctx).Debug("trying to send output to channel")
				if cs != nil {
					updateCounts(ctx, in, cs, types.Framework(config.Framework))
					in.TotalTestFail = int32(cs.FailCount)
					in.TotalTestPass = int32(cs.PassCount)
					in.TotalTestSkip = int32(cs.SkipCount)

					in.TotalTestCount = int32(cs.TotalTestCount)
				}
				responseStream <- in
				Logger(ctx).Debug("sent output to channel")

				if in.Control == types.FINISHED {
					Logger(ctx).Debug("Got finished from server")
					Logger(ctx).Debugf("Exit code is %v", in.Exitcode)
					allExecutionInfo[len(allExecutionInfo)-1].ExitCode = in.Exitcode
					Logger(ctx).Debug("The command we were running is %+v", command)
					// I think we should check shit
					Logger(ctx).Debug("Breaking out because we are finished")

					break CommandFor
				}
				if in.Control == types.FAILED {
					Logger(ctx).Debug("Got failed from server")
					Logger(ctx).Debugf("Exit code is %v", in.Exitcode)
					return allExecutionInfo, errors.Errorf("failed with exit code %v", in.Exitcode)

				}

			}

			// Logger(ctx).Debugf("finished command %v", command)
			// if we are looping - it would be good to figure out how long building takes
			// var ei api.ExecutionInfo
			// if len(allExecutionInfo) > 0 {
			// 	ei = *allExecutionInfo[len(allExecutionInfo)-1]
			// } else {
			// 	ei = api.ExecutionInfo{}
			// }

			// Logger(ctx).Debugf("The execution info we have back is %v it's stage is %v and commandLine %v and rebuildHash", ei, command.Stage, command.Commandline, ei.RebuildHash)
		}
	}
	Logger(ctx).Debugf("TIMING after requests are finished time is now %v", time.Since(startTime))

	Logger(ctx).Debug("Super connectToServer returning nil")

	Logger(ctx).Debugf("The execution infos for the entire command are %v", allExecutionInfo)
	Logger(ctx).Debugf("We have %v execution infos", len(allExecutionInfo))
	return allExecutionInfo, err
}

func readToChannel(ctx context.Context, r pb.CommandRunner_RunCommandsClient, response chan ResponseOutput) {
	for {

		in, err := r.Recv()
		response <- ResponseOutput{output: in, error: err}
		if err != nil {
			close(response)
			return
		}
	}
}

func sendBufferToS3(ctx context.Context, buffer *bytes.Buffer, streamInfo types.LogStreamInfo, logger *BriskLogger) (string, error) {
	if IsDev() {
		Logger(ctx).Debug("DEV mode so not sending to s3")
	} else {
		logger.Debug("sendBufferToS3")
		// send the buffer to s3
		logger.Debugf("Uploading logs to s3 we have %v bytes", buffer.Len())
		var ls *log_service.LogService
		if streamInfo.LogUid != "" {
			ls = log_service.New(streamInfo.ProjectToken, streamInfo.LogUid)
		} else {
			ls = log_service.New(streamInfo.ProjectToken, streamInfo.WorkerRunInfoUID)
		}
		location, err := ls.UploadContent(ctx, io.Reader(buffer), logger)
		if err != nil {
			logger.Errorf("Error uploading logs to s3 %v", err)
			return "", err
		}
		logger.Debugf("Uploaded logs to %v", location)
		return location, nil
	}
	return "", nil
}

type countStruct struct {
	TotalTestCount int
	FailCount      int
	PassCount      int
	SkipCount      int
	mu             sync.Mutex
}

func setBuildCommandsRunAt(ctx context.Context, worker api.Worker) {

	endpoint := os.Getenv("BRISK_API")

	var conn *grpc.ClientConn
	var err error
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	Logger(ctx).Debug("setBuildCommandsRunAt")
	if !IsDev() {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts := []grpc_retry.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		}
		conn, err = grpc.DialContext(ctx, endpoint,
			grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), grpc_retry.StreamClientInterceptor(opts...), BugsnagClientInterceptor()),
			grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), grpc_retry.UnaryClientInterceptor(opts...), BugsnagClientUnaryInterceptor),
			grpc.WithDefaultCallOptions(),
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    10 * time.Second,
				Timeout: 3 * time.Second}))
	} else {
		conn, err = grpc.DialContext(ctx, endpoint, grpc.WithStreamInterceptor(BugsnagClientInterceptor()), grpc.WithUnaryInterceptor(BugsnagClientUnaryInterceptor),
			grpc.WithDefaultCallOptions(),
			grpc.WithInsecure(), grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    30 * time.Second,
				Timeout: 5 * time.Second,
			}))
	}
	if err != nil {
		Logger(ctx).Errorf("setBuildCommandsRunAt: did not connect: %v with cause %v", err, context.Cause(ctx))
		return
	}
	Logger(ctx).Debug("Connected ")

	defer conn.Close()
	c := api.NewWorkersClient(conn)
	in := api.CommandsRunReq{Id: worker.Id}
	_, runErr := c.BuildCommandsRun(ctx, &in)
	if runErr != nil {

		bugsnag.Notify(runErr)
		Logger(ctx).Errorf("Problem saving buildcommands run at %v", runErr)
		return

	}

}
func updateCountsJest(ctx context.Context, out *pb.Output, cs *countStruct) {
	Logger(ctx).Debugf("Using the string %v to check for Pass Fail", out.Stderr)
	switch res := out.Stderr; {
	case strings.Contains(res, "PASS"):
		Logger(ctx).Debug("Looks like we have recognized a PASS")
		cs.PassCount = cs.PassCount + 1
	case strings.Contains(res, "skipped test"):
		Logger(ctx).Debug("Looks like we have recognized a Skip")
		cs.PassCount = cs.SkipCount + 1
	case strings.Contains(res, "FAIL"):
		Logger(ctx).Debug("Looks like we have recognized a FAIL")

		cs.FailCount = cs.FailCount + 1
	}

}

// figure out exactly what an rspec output says
func updateCountsRspec(ctx context.Context, out *pb.Output, cs *countStruct) error {
	if out.JsonResults != "" {
		Logger(ctx).Debug("waiting on lock")
		cs.mu.Lock()
		Logger(ctx).Debug("acquired lock")

		parsed, err := ParseRspecJsonResults(out.JsonResults)
		if err != nil {
			Logger(ctx).Errorf("Error parsing rspec json results %v", err)
			return err
		}
		cs.FailCount += parsed.Summary.FailureCount
		cs.SkipCount = parsed.Summary.PendingCount

		cs.PassCount += (len(parsed.Examples) - parsed.Summary.FailureCount) //- parsed.Summary.PendingCount)
		Logger(ctx).Debugf("The summary is %v", parsed.SummaryLine)
		cs.mu.Unlock()
	} else {
		Logger(ctx).Debug("updateCountsRspec no JSON results!")
	}
	return nil
}
func updateCounts(ctx context.Context, out *pb.Output, cs *countStruct, testFramework types.Framework) {

	if testFramework == types.Jest {
		updateCountsJest(ctx, out, cs)
	} else if testFramework == types.Cypress {
		Logger(ctx).Debug("No countrstruct for cypress")
	} else {
		updateCountsRspec(ctx, out, cs)
	}
	Logger(ctx).Debugf("updated countStruct is %+v", cs)

}

func MapWorkers(vs []*api.Worker, f func(*api.Worker) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func SplitFilesByRaw(ctx context.Context, worker *api.Worker, remote_dir string, numServers int, responseStream chan *pb.Output, config *brisksupervisor.Config) ([][]string, int, error) {

	num_buckets := numServers

	info, err := getAllFilesRawRemote(ctx, worker, remote_dir, responseStream, config)

	if err != nil {
		Logger(ctx).Errorf("Error getting all files %v", err)
		return nil, 0, err
	}
	Logger(ctx).Debugf("Got all files %v", info)
	files := info.Filenames
	result := make([][]string, num_buckets)
	for i := 0; i < num_buckets; i++ {
		for j := i; j < len(files); j = j + num_buckets {
			localPath := files[j]
			remotePath := strings.Replace(localPath, remote_dir, "", 1)
			result[i] = append(result[i], string(remotePath))
		}
	}
	return result, len(result), nil
}

func splitForRaw(ctx context.Context, jsonString string) (PreTestInfo, error) {
	Logger(ctx).Debug("splitForRaw++")
	Logger(ctx).Debugf("the json string is %v", jsonString)
	var result JestFiles

	err := json.Unmarshal([]byte(jsonString), &result.Files)
	if err != nil {
		Logger(ctx).Debugf("Error parsing json %v", err)
		return PreTestInfo{}, fmt.Errorf("error parsing json: %v", err)
	}
	var fileNames []string
	Logger(ctx).Debugf("the result.Examples length is %v", len(result.Files))
	for i := 0; i < len(result.Files); i++ {
		fileNames = append(fileNames, result.Files[i])

	}

	Logger(ctx).Debug("splitForJest--")
	return PreTestInfo{Filenames: RemoveDuplicatesFromSlice(ctx, fileNames), TotalTestCount: len(result.Files), TotalSkipCount: 0}, nil
}

func getAllFilesRawRemote(ctx context.Context, worker *api.Worker, workDir string, responseStream chan *pb.Output, config *brisksupervisor.Config) (PreTestInfo, error) {
	Logger(ctx).Debugf("getAllFilesRawRemote with worker %v", worker.Id)

	var buildCmds []api.Command
	for _, c := range config.BuildCommands {
		copy := *c
		copy.Environment = config.Environment
		copy.Stage = "Build"

		buildCmds = append(buildCmds, copy)
	}
	commands := []api.Command{{Commandline: config.ListTestCommand, WorkDirectory: workDir, IsListTest: true, IsTestRun: false, Environment: config.Environment, TestFramework: config.Framework}}

	editedStream := make(chan *pb.Output, 100)
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {

		buildCmds, commands = AddDefaultRemoteDirectorybriskSupervisor(ctx, buildCmds, commands)

		executionInfos, err := connectToServer(ctx, nil, commands, buildCmds, 1, nil, editedStream, *worker, config)
		if err != nil {
			Logger(ctx).Error(err.Error())
			responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Created: timestamppb.Now()}
			cancel(fmt.Errorf("error from connectToServer: %v", err))
		}
		executionInfo := executionInfos[len(executionInfos)-1]

		Logger(ctx).Debugf("exitCode is %v", executionInfo.ExitCode)
	}()

	for out := range editedStream {
		if out.Control == types.FINISHED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v for cmd %v", out.Control, out.CmdSqNum)
			Logger(ctx).Info(out.JsonResults)
			if out.Exitcode != 0 {
				Logger(ctx).Debugf("Error running command response is %+v", out)
				return PreTestInfo{}, errors.Errorf("error running command exit code is %+v", out.Exitcode)
			}
			if len(out.JsonResults) > 0 {
				Logger(ctx).Debugf("We have json  test results they are %v", out.JsonResults)
				testInfo, parseErr := splitForRaw(ctx, out.JsonResults)
				if parseErr != nil {
					Logger(ctx).Debugf("Error parsing json for jest, %v", parseErr.Error())
					return PreTestInfo{}, parseErr
				}
				return testInfo, nil
			}
		}

		if out.Control == types.FAILED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v", out.Control)

			responseStream <- out
			// lets see if this helps the errors deliver

			return PreTestInfo{}, errors.Errorf("error running splitting files exit code is %+v", out.Exitcode)

		}
		Logger(ctx).Debugf("in parsing stream - %v", out)

		responseStream <- out

	}
	Logger(ctx).Debug("SHOULD NEVER GET HERE")
	return PreTestInfo{}, nil

}

func GetAllFilesJestRemote(ctx context.Context, worker *api.Worker, workDir string, responseStream chan *pb.Output, config *brisksupervisor.Config) (PreTestInfo, error) {
	Logger(ctx).Debug("GetAllFilesJest")
	Logger(ctx).Debugf("Config passed to GetAllFilesJestRemote is : %+v", config)
	//TODO lets make this work right - remove these and get them from the config file
	var buildCmds []api.Command
	for _, c := range config.BuildCommands {
		copy := *c
		copy.Environment = config.Environment
		copy.Stage = "Build"

		buildCmds = append(buildCmds, copy)
	}
	commands := []api.Command{{Commandline: config.ListTestCommand, WorkDirectory: workDir, IsListTest: true, IsTestRun: false, Environment: config.Environment, TestFramework: config.Framework}}

	editedStream := make(chan *pb.Output, 100)
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		Logger(ctx).Debugf("in GetAllFilesJestRemote before addDefaultRemoteDirectorybriskSupervisor buildCmds: %+v  and commands %+v", buildCmds, commands)

		buildCmds, commands = AddDefaultRemoteDirectorybriskSupervisor(ctx, buildCmds, commands)
		Logger(ctx).Debugf("in GetAllFilesJestRemote after addDefaultRemoteDirectorybriskSupervisor buildCmds: %+v  and commands %+v", buildCmds, commands)

		Logger(ctx).Debugf("buildCmds in GetAllFilesJestRemote : %+v", buildCmds)
		exitCode, err := connectToServer(ctx, nil, commands, buildCmds, 1, nil, editedStream, *worker, config)
		if err != nil {
			Logger(ctx).Error(err.Error())
			responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Created: timestamppb.Now()}
			cancel(fmt.Errorf("error from connectToServer: %v", err))
		}
		Logger(ctx).Debugf("exitCode is %v", exitCode)
	}()

	for out := range editedStream {
		if out.Control == types.FINISHED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v for cmd %v", out.Control, out.CmdSqNum)
			Logger(ctx).Infof("Json results from the command are %v", out.JsonResults)
			if out.Exitcode != 0 {
				Logger(ctx).Debugf("Error running command response is %+v", out)
				return PreTestInfo{}, errors.Errorf("error running command exit code is %+v", out.Exitcode)
			}
			if len(out.JsonResults) > 0 {
				Logger(ctx).Debugf("We have json  test results they are %v", out.JsonResults)
				testInfo, parseErr := SplitForJest(ctx, out.JsonResults)
				if parseErr != nil {
					Logger(ctx).Errorf("Error parsing json for jest, %v", parseErr.Error())
					return PreTestInfo{}, parseErr
				}
				return testInfo, nil
			}
		}

		if out.Control == types.FAILED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v", out.Control)

			responseStream <- out
			// lets see if this helps the errors deliver

			return PreTestInfo{}, errors.Errorf("error running splitting files exit code is %+v", out.Exitcode)

		}
		Logger(ctx).Debugf("in parsing stream - %v", out)

		responseStream <- out

	}
	Logger(ctx).Debug("SHOULD NEVER GET HERE")
	return PreTestInfo{}, nil

}

func GetAllFilesRspecRemote(ctx context.Context, worker *api.Worker, workDir string, responseStream chan *pb.Output, config *brisksupervisor.Config) (PreTestInfo, error) {
	defer bugsnag.AutoNotify(ctx)
	Logger(ctx).Debug("GetAllFilesRspec")
	//TODO lets make this work right - remove these and get them from the config file
	var buildCmds []api.Command
	for _, c := range config.BuildCommands {
		copy := *c
		copy.Environment = config.Environment
		copy.Stage = "Build"

		buildCmds = append(buildCmds, copy)
	}
	if len(config.ListTestCommand) == 0 {
		config.ListTestCommand = os.Getenv("RSPEC_LIST_TEST_CMD")
	}

	Logger(ctx).Debugf("The list test command is %v", config.ListTestCommand)
	Logger(ctx).Debugf("The environment passed in is  %v", config.Environment)
	Logger(ctx).Debugf("The config passed in is %+v", config)
	// setting  IsTestRun to true because we want to run with a file output to get our json back
	commands := []api.Command{{Commandline: config.ListTestCommand, WorkDirectory: workDir, IsTestRun: true, Environment: config.Environment, TestFramework: config.Framework}}

	editedStream := make(chan *pb.Output, 100)

	ctx, cancel := context.WithCancelCause(ctx)
	Logger(ctx).Debug("in GetAllFilesRspecRemote before gofunc connecting to server")
	go func() {
		defer bugsnag.AutoNotify(ctx)
		Logger(ctx).Debugf("in GetAllFilesRspecRemote before addDefaultRemoteDirectorybriskSupervisor buildCmds: %+v  and commands %+v", buildCmds, commands)
		buildCmds, commands = AddDefaultRemoteDirectorybriskSupervisor(ctx, buildCmds, commands)
		Logger(ctx).Debugf("in GetAllFilesRspecRemote after addDefaultRemoteDirectorybriskSupervisor buildCmds: %+v  and commands %+v", buildCmds, commands)
		executionInfos, err := connectToServer(ctx, nil, commands, buildCmds, 1, nil, editedStream, *worker, config)

		Logger(ctx).Debug("in GetAllFilesRspecRemote after connectToServer")
		if err != nil {
			Logger(ctx).Error(err.Error())
			responseStream <- &pb.Output{Response: err.Error(), Stderr: err.Error(), Created: timestamppb.Now()}
			cancel(fmt.Errorf("error from connectToServer: %v", err))
		}
		executionInfo := executionInfos[len(executionInfos)-1]

		Logger(ctx).Debugf("exitCode is %v", executionInfo.ExitCode)
	}()

	for out := range editedStream {
		if out.Control == types.FINISHED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v for cmd %v", out.Control, out.CmdSqNum)
			Logger(ctx).Info(out.JsonResults)
			if out.Exitcode != 0 {
				Logger(ctx).Debugf("Error running command response is %+v", out)
				return PreTestInfo{}, errors.Errorf("error running command exit code is %+v", out.Exitcode)
			}
			if len(out.JsonResults) > 0 {
				Logger(ctx).Debugf("We have json  test results they are %v", out.JsonResults)
				testInfo, parseErr := SplitForRSpec(ctx, out.JsonResults)
				if parseErr != nil {
					Logger(ctx).Debugf("Error parsing json for rspec, %v", parseErr.Error())
					return PreTestInfo{}, parseErr
				}
				return testInfo, nil
			}
		}

		if out.Control == types.FAILED {
			Logger(ctx).Debugf("Received %+v from worker", out)
			Logger(ctx).Infof("Received control from connectToServer %v", out.Control)

			responseStream <- out
			// lets see if this helps the errors deliver
			Logger(ctx).Error("Error running command response is %+v", out)
			return PreTestInfo{}, errors.Errorf("error running splitting files exit code is %+v output %v", out.Exitcode, out.Stderr)

		}
		Logger(ctx).Debugf("in parsing stream - %v", out)

		responseStream <- out

	}
	Logger(ctx).Debug("SHOULD NEVER GET HERE")
	return PreTestInfo{}, nil

}

func warmOtherWorkers(ctx context.Context, workers []api.Worker, config *brisksupervisor.Config, responseStream chan *pb.Output) ([]*api.Worker, []error) {

	workDir := os.Getenv("REMOTE_DIR")
	Logger(ctx).Debugf("In warmOtherWorkers with workDir with workers %+v", workers)

	var buildCmds []api.Command
	for _, c := range config.BuildCommands {
		copy := *c
		copy.Environment = config.Environment
		copy.Stage = "Build"

		buildCmds = append(buildCmds, copy)
	}
	// commands := []api.Command{{Commandline: config.ListTestCommand, WorkDirectory: workDir, IsListTest: true, IsTestRun: false, Environment: config.Environment, TestFramework: config.Framework}}
	commands := []api.Command{{Commandline: "echo 'build commands run for worker'", WorkDirectory: workDir, IsListTest: true, IsTestRun: false, Environment: config.Environment, TestFramework: config.Framework}}
	editedStream := make(chan *pb.Output, 1000)
	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("DEFAULT_BUILD_TIMEOUT"))

	defer cancel()
	goodWorkers := make(chan *api.Worker, len(workers))
	errChan := make(chan error, len(workers))

	var workerGroup sync.WaitGroup
	workerGroup.Add(len(workers))
	for _, w := range workers {
		go func(worker api.Worker) {
			buildCmds, commands = AddDefaultRemoteDirectorybriskSupervisor(ctx, buildCmds, commands)
			Logger(ctx).Debugf("in warmOtherWorkers before connectToServer with worker id %v", worker.Uid)
			executionInfo, err := connectToServer(ctx, nil, commands, buildCmds, 1, nil, editedStream, worker, config)
			if err != nil {
				Logger(ctx).Error(err.Error())
				responseStream <- &pb.Output{Response: fmt.Sprintf("Error when running build commands %v removing worker %v", err.Error(), worker.Uid), Stderr: err.Error(), Created: timestamppb.Now()}
				if status.Convert(err).Code() == codes.Unavailable {
					Logger(ctx).Error("Error connecting to worker %v - going to deregister", worker.Uid)
					go DeRegisterWorker(ctx, &worker)

				}
				errChan <- err
			} else {
				if len(executionInfo) > 0 && executionInfo[len(executionInfo)-1].ExitCode != 0 {
					Logger(ctx).Error(err.Error())
					responseStream <- &pb.Output{Response: fmt.Sprintf("Error when running build commands exit code is %v removing worker %v", executionInfo[len(executionInfo)-1].ExitCode, worker.Uid), Stderr: err.Error(), Created: timestamppb.Now()}
				} else {
					go setBuildCommandsRunAt(ctx, worker)
					worker.BuildCommandsRunAt = timestamppb.Now()

					goodWorkers <- &worker
				}
			}

			workerGroup.Done()

		}(w)
	}
	go func() {
		for out := range editedStream {
			if !viper.GetBool("FILTER_BUILD_COMMANDS") {
				responseStream <- out
			}

		}
	}()

	workerGroup.Wait()
	close(goodWorkers)
	close(editedStream)
	close(errChan)
	errors := []error{}
	for e := range errChan {
		errors = append(errors, e)
	}

	returningWorkers := []*api.Worker{}

	for w := range goodWorkers {
		returningWorkers = append(returningWorkers, w)
	}
	return returningWorkers, errors

}

var mPretestInfo PreTestInfo
var mFiles [][]string
var lastListTestCommand string

func SplitFilesByRspec(ctx context.Context, worker *api.Worker, workDir string, num_buckets int, responseStream chan *pb.Output, reCalcFiles bool, config *brisksupervisor.Config) ([][]string, PreTestInfo, error) {

	//memoize
	if mFiles != nil && !reCalcFiles && config.ListTestCommand == lastListTestCommand {
		return mFiles, mPretestInfo, nil
	} else {
		lastListTestCommand = config.ListTestCommand
		mFiles = nil
		mPretestInfo = PreTestInfo{}
	}
	Logger(ctx).Debug("SplitFilesByRspec++")
	Logger(ctx).Debugf("The params are %v | %v | %v", worker, workDir, num_buckets)
	preTestInfo, err := GetAllFilesRspecRemote(ctx, worker, workDir, responseStream, config)

	Logger(ctx).Debug("after GetAllFilesRspec")
	if err != nil {
		Logger(ctx).Debug(err)
		return nil, PreTestInfo{}, errors.Errorf("Error trying to split tests - %v", err.Error())
	}
	files := preTestInfo.Filenames
	Logger(ctx).Debugf("All the files from rspec are %v", files)
	Logger(ctx).Debugf("The files length is %v", len(files))

	result := make([][]string, num_buckets)
	for i := 0; i < num_buckets; i++ {
		for j := i; j < len(files); j = j + num_buckets {
			localPath := files[j]
			remotePath := strings.Replace(localPath, workDir, "", 1)
			result[i] = append(result[i], string(remotePath))
		}
	}
	for i := 0; i < num_buckets; i++ {
		Logger(ctx).Debugf("the bucket full of files is %v", result[i])
		for _, v := range result {
			Logger(ctx).Debugf("The results have size %v", len(v))
		}
	}
	Logger(ctx).Debug("SplitFilesByRspec--")
	mFiles = result
	mPretestInfo = preTestInfo
	return result, preTestInfo, err
}

func getWorkingWorkers(ctx context.Context, numWorkers int, workerImage string, rebuildFilePaths []string, repoInfo api.RepoInfo, logUid string) ([]*api.Worker, int32, string, error) {

	workers, jobrunId, jobrunLink, err := GetWorkersForProject(ctx, numWorkers, workerImage, rebuildFilePaths, repoInfo, 1, logUid)
	if err != nil {
		Logger(ctx).Errorf("Error getting workers for project %v", err.Error())
		return nil, jobrunId, jobrunLink, err
	}

	if len(workers) == 0 {
		Logger(ctx).Error("No workers returned from GetWorkersForProject")
		Logger(ctx).Error("workerImage is %v", workerImage)
		return nil, jobrunId, jobrunLink, errors.Errorf("no workers available for project")
	}
	// really reducing this cause it should be caught on the other end - and we don't want to check in two places
	if len(workers) < numWorkers*2/10 {
		Logger(ctx).Errorf("Not enough workers returned from GetWorkersForProject we need at least 20% we got %v of %v requested", len(workers), numWorkers)
		Logger(ctx).Error("workerImage is %v", workerImage)
		return nil, jobrunId, jobrunLink, errors.Errorf("not enough workers available for project requested %v got %v", numWorkers, len(workers))
	}

	if IsDev() || os.Getenv("SKIP_CHECK_WORKERS") == "true" {
		return workers, jobrunId, jobrunLink, err
	}

	var checkWg sync.WaitGroup

	goodChannel := make(chan *api.Worker, len(workers))
	checkWg.Add(len(workers))
	Logger(ctx).Debug("Before checking all the workers")
	for _, w := range workers {
		go func(w *api.Worker) {
			defer bugsnag.AutoNotify(ctx)

			Logger(ctx).Debug("Checking the worker")
			result, err := nomad.CheckWorkerNomad(ctx, *w)
			if err != nil {
				Logger(ctx).Errorf("Error checking worker %+v", err)

			} else {
				if result {

					goodChannel <- w
				} else {
					Logger(ctx).Debugf("Worker needs to be deregistered %+v", w)
					go func() { DeRegisterWorker(ctx, w) }()
				}
			}
			Logger(ctx).Debug("now I'm done with this worker")
			checkWg.Done()
		}(w)
	}
	Logger(ctx).Debug("Waiting for workers to finish checking")
	checkWg.Wait()

	var goodWorkers []*api.Worker
	close(goodChannel)
	for goodW := range goodChannel {
		Logger(ctx).Debugf("In the range on channel with good worker %v", goodW.IpAddress)
		goodWorkers = append(goodWorkers, goodW)
	}
	numServers := len(goodWorkers)

	if numServers <= 0 {
		projectToken, err := GetAuthenticatedProjectToken(ctx)
		if err != nil {
			return nil, jobrunId, jobrunLink, err
		}
		Logger(ctx).Debugf("No Servers available for project with token %v", projectToken)
		return nil, jobrunId, jobrunLink, errors.Errorf("No Servers available for project with token %v", projectToken)
	} else {
		Logger(ctx).Debugf("Reserved %v servers for the project", numServers)
	}
	return goodWorkers, jobrunId, jobrunLink, nil
}

// func addError(syncErr *syncError, err error) {
// 	syncErr.m.Lock()
// 	syncErr.errs = append(syncErr.errs, err)
// 	syncErr.m.Unlock()

// }

// type syncError struct {
// 	m    sync.Mutex
// 	errs []error
// }

func syncToWorkers(ctx context.Context, workers []*api.Worker, config *brisksupervisor.Config) ([]*api.Worker, []*api.Worker, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "syncToWorkers")
	defer span.End()
	var syncWg sync.WaitGroup
	syncWg.Add(len(workers))

	var returnWorkers []*api.Worker
	var failingWorkers []*api.Worker
	var returnErr error
	returnErr = nil

	returnWorkersMutex := &sync.Mutex{}

	waitCh := make(chan struct{})
	go func() {
		for _, worker := range workers {

			go func(worker *api.Worker) {
				// we want to always hit the Sync
				defer syncWg.Done()
				Logger(ctx).Debugf("The worker is %v", worker)
				Logger(ctx).Debugf("In worker sync go routine")
				setupErr := SetupWorker(AddIntendedAllocIdTo(ctx, worker.Uid), worker, globalPublicKey)

				if setupErr != nil {
					Logger(ctx).Errorf("Error setting up workers %v - worker details %+v", setupErr, worker)
					Logger(ctx).Error("Worker %v sync finished with error %v", worker.Uid, setupErr)
					returnErr = setupErr

					returnWorkersMutex.Lock()
					failingWorkers = append(failingWorkers, worker)
					returnWorkersMutex.Unlock()

				} else {

					defer bugsnag.AutoNotify(ctx)

					Logger(ctx).Debug("before sync to worker")
					// default timeout of 45 seconds
					if os.Getenv("SUPER_LOCAL_DIR") == "" {
						os.Setenv("SUPER_LOCAL_DIR", "/tmp/remote_dir/")
					}
					var destFolder string
					if IsDev() {
						destFolder = "/tmp/remote_dir/"
					} else {
						// when using a bastion we include this on the other side of the rsync
						destFolder = "/"
					}

					syncErr := SuperToWorkerSync(ctx, worker.IpAddress, worker.SyncPort, os.Getenv("SUPER_LOCAL_DIR"), destFolder, "brisk", config)

					if syncErr != nil {
						Logger(ctx).Errorf("Worker %v sync finished with error %v", worker.Uid, syncErr)
						returnErr = syncErr

						returnWorkersMutex.Lock()

						failingWorkers = append(failingWorkers, worker)
						returnWorkersMutex.Unlock()

					} else {
						returnWorkersMutex.Lock()
						returnWorkers = append(returnWorkers, worker)
						returnWorkersMutex.Unlock()
					}
				}

			}(worker)

		}
		syncWg.Wait()
		close(waitCh)
	}()

	Logger(ctx).Debug("Waiting for worker sync to finish")

	select {
	case <-waitCh:

		Logger(ctx).Debugf("Worker sync finished with returnWorkers count %v and failingWorkers count %v and error %v", len(returnWorkers), len(failingWorkers), returnErr)

		return returnWorkers, failingWorkers, returnErr
	case <-time.After(2 * time.Minute):
		Logger(ctx).Error("Timed out syncing to workers")
		return returnWorkers, failingWorkers, errors.Errorf("Timed out syncing to workers after 2 minutes")
	}

}

func splitTestFiles(ctx context.Context, numServers int, worker *api.Worker, config *brisksupervisor.Config, responseStream chan *pb.Output, reCalcFiles bool) ([][]string, int, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "splitTestFiles")
	defer span.End()
	Logger(ctx).Debug("splitTestFiles++")
	var myFiles [][]string
	var totalTestCount int
	var err error

	var preTestInfo PreTestInfo

	var allFiles PreTestInfo
	var errFiles error
	switch config.Framework {
	case string(types.Rspec):
		responseStream <- &pb.Output{Response: "Rspec detected", Created: timestamppb.Now()}
		Logger(ctx).Debug("Splitting by rspec")
		myFiles, preTestInfo, err = SplitFilesByRspec(ctx, worker, os.Getenv("REMOTE_DIR"), numServers, responseStream, reCalcFiles, config)
		Logger(ctx).Debug("Split by rspec results are %+v %+v %+v", myFiles, preTestInfo, err)

		if err != nil {
			// need to stream the info back instead of a one off command. makes more sense anyway
			Logger(ctx).Errorf("received error from SplitFilesByRspec :- %v", err)

			return nil, 0, err
		}
		totalTestCount = preTestInfo.TotalTestCount
		Logger(ctx).Debugf("the total test count is %v", totalTestCount)
		Logger(ctx).Debugf("myFiles are %v", myFiles)
	case string(types.Python):
		responseStream <- &pb.Output{Response: "Python detected", Created: timestamppb.Now()}
		Logger(ctx).Debug("Splitting by python")
		Logger(ctx).Debug("Nop pytest using pytest-split so nothing required")

	case string(types.Raw):
		responseStream <- &pb.Output{Response: "Raw detected", Created: timestamppb.Now()}
		Logger(ctx).Debug("Splitting using provided list command : " + config.ListTestCommand)

		myFiles, totalTestCount, err = SplitFilesByRaw(ctx, worker, os.Getenv("REMOTE_DIR"), numServers, responseStream, config)

		if err != nil {
			Logger(ctx).Errorf("Error splitting files %s", err)
			return nil, -1, err
		}

	default:
		responseStream <- &pb.Output{Response: "Default test splitter", Created: timestamppb.Now()}

		if config.SplitByJUnit {
			Logger(ctx).Debug("Splitting using provided files")
			if len(config.OrderedFiles) < 1 {
				errFiles = errors.New("Split by JUnit but no files provided")
			} else {

				allFiles = PreTestInfo{Filenames: config.OrderedFiles, TotalTestCount: len(config.OrderedFiles)}
			}

		} else {

			allFiles, errFiles = GetAllFilesJestRemote(ctx, worker, os.Getenv("REMOTE_DIR"), responseStream, config)

		}
		if errFiles != nil {
			//need to parse different and use --outputFile=
			Logger(ctx).Errorf("Error getting files %s", errFiles)
			return nil, -1, errFiles
		}
		if config.AutomaticSplitting {
			Logger(ctx).Debug("Splitting using automatic splitting")
			totalTestCount = len(allFiles.Filenames)
			shortNames := removeWorkDir(allFiles.Filenames, os.Getenv("REMOTE_DIR"))
			var splitMethod string
			myFiles, splitMethod, err = SplitTests(ctx, int32(numServers), shortNames)
			for v := range myFiles {
				Logger(ctx).Debugf("Bucket %v has %v files", v, len(myFiles[v]))
			}
			responseStream <- &pb.Output{Response: "Split method: " + splitMethod, Created: timestamppb.Now()}

		} else {
			Logger(ctx).Debug("Splitting using SplitFilesByJest")
			myFiles, totalTestCount, err = SplitFilesByJest(ctx, os.Getenv("REMOTE_DIR"), numServers, responseStream, allFiles.Filenames)
		}
		if err != nil {
			Logger(ctx).Errorf("Error splitting files %s", err)
			return nil, 0, err
		}
	}

	return myFiles, totalTestCount, nil
}

func removeWorkDir(files []string, workDir string) []string {
	var shortNames []string
	for _, file := range files {

		remotePath := strings.Replace(file, workDir, "", 1)
		shortNames = append(shortNames, remotePath)
	}
	return shortNames
}

func printLinkToStream(ctx context.Context, responseStream chan *pb.Output, jobrunLink string) {

	if jobrunLink != "" {
		responseStream <- &pb.Output{Response: "View online at  " + jobrunLink, Created: timestamppb.Now()}
	}
}

type ResponseOutput struct {
	output *pb.Output
	error  error
}

type ProtectedCounter struct {
	mu    sync.Mutex
	count int
}

func (c *ProtectedCounter) Add(i int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count += i
}

func (c *ProtectedCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
func (c *ProtectedCounter) Done() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count--
}

func (c *ProtectedCounter) Finished() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count == 0
}
