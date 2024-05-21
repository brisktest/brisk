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

package cmd

import (
	"brisk-supervisor/api"
	"brisk-supervisor/brisk-cli/cli_utils"
	"brisk-supervisor/brisk-cli/utilities"
	pb "brisk-supervisor/brisk-supervisor"
	"brisk-supervisor/shared/auth"
	"brisk-supervisor/shared/constants"
	"brisk-supervisor/shared/fsWatch"
	"brisk-supervisor/shared/honeycomb"
	"log"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"syscall"
	"unsafe"

	"hash/fnv"

	"github.com/morikuni/aec"

	"go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	otel_codes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	grpc_status "google.golang.org/grpc/status"

	. "brisk-supervisor/shared/logger"
	"brisk-supervisor/shared/types"
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"errors"

	"github.com/bugsnag/bugsnag-go"
	"github.com/moby/term"
	"github.com/spf13/viper"

	. "brisk-supervisor/shared"
	"os"
	"strings"
	"time"

	//	ui "github.com/gizak/termui/v3"

	"github.com/acarl005/stripansi"

	durafmt "github.com/hako/durafmt"
	"github.com/mattn/go-tty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var globalUsername string
var uniqueInstanceId string

const name = "brisk"

func getConn(ctx context.Context, super *api.Super) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	var size int = 1024 * 1024 * 10
	timeout := viper.GetDuration("CLI_STREAM_TIMEOUT")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var credentialOpts grpc.DialOption
	if !IsDev() {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		credentialOpts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))

	} else {
		credentialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
		grpc_retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			fmt.Printf("retry.. %v \n", attempt)
			Logger(ctx).Errorf("retrying after error: %v attempt # %v while connecting to %+v", err, attempt, super)
		}),
	}

	Logger(ctx).Debugf("Dialing %v", super.ExternalEndpoint)

	conn, err = grpc.DialContext(ctx, super.ExternalEndpoint, grpc.WithDefaultCallOptions(
		grpc.UseCompressor(gzip.Name),
		grpc.MaxCallRecvMsgSize(size)), credentialOpts, grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second,
			Timeout: 10 * time.Second,
		}), grpc.WithChainStreamInterceptor(

			grpc_retry.StreamClientInterceptor(opts...),
			otelgrpc.StreamClientInterceptor()),
		grpc.WithChainUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(opts...),
			otelgrpc.UnaryClientInterceptor(),
			BugsnagClientUnaryInterceptor))
	Logger(ctx).Debugf("Dialed %v", super.ExternalEndpoint)

	return conn, err
}

func connectToServer(ctx context.Context, config Config, super *api.Super, outputChan chan *pb.Output, cancelChan chan string, controlChan chan string, syncHost string, workingDirectory string, projectToken string, apiKey string, apiToken string) error {
	ctx, cancel := context.WithCancel(ctx)
	ctx, span := otel.Tracer(name).Start(ctx, "connectToServer")
	defer CancelCtx(ctx, cancel, "connectToServer finished so we are cancelling the context")

	defer span.End()

	authCreds := auth.AuthCreds{ProjectToken: config.ProjectToken, ApiToken: viper.GetString("apiToken"), ApiKey: viper.GetString("apiKey")}

	Logger(ctx).Debugf("Connecting to supervisor : %v", super.ExternalEndpoint)

	ctx, ctxError := prepareOutgoingContextMin(ctx, projectToken, authCreds, uniqueInstanceId)
	if ctxError != nil {
		return ctxError
	}

	conn, err := getConn(ctx, super)
	if err != nil {
		mesg := fmt.Sprintf("%s : Could not connect to %v", err, super.ExternalEndpoint)
		Logger(ctx).Error(mesg)
		return errors.New(mesg)
	}
	Logger(ctx).Debugf("Connected to %v", super.ExternalEndpoint)

	defer conn.Close()

	c := pb.NewBriskSupervisorClient(conn)
	listData = append(listData, "connected...")

	userDetails := pb.UserDetails{ApiToken: apiToken, ApiKey: apiKey, ProjectToken: projectToken}

	var privateKey string

	buildCommands, _ := parseCommandsFromConfig(ctx, config)

	outputChan <- &pb.Output{Response: "Waiting on project concurrency", Created: timestamppb.Now()}

	privateKey, globalUsername, err = setupKeys(ctx, config, authCreds, c, &userDetails, buildCommands)

	if err != nil {
		Logger(ctx).Errorf("Error during setup connecting to %v :  %v ", super.ExternalEndpoint, err)
		fmt.Printf(".")
		return err
	}

	listData = append(listData, "initial sync....")
	// For CI we do a single run and then exit

	if viper.GetBool("CI") {
		Logger(ctx).Debugf("Running in CI mode")
		repoInfo, err := GetGitInfo(ctx, workingDirectory)
		if err != nil {
			Logger(ctx).Errorf("Error getting git info %v", err)
		}

		outputChan <- &pb.Output{Response: "Starting run in CI mode", Created: timestamppb.Now()}

		startTime, err := singleTestRun(ctx, super, privateKey, syncHost, workingDirectory, outputChan, authCreds, true, repoInfo)
		if err != nil {
			Logger(ctx).Errorf("Error during run %v ", err.Error())

			if err.Error() == "server closed the stream without sending trailers" {
				Logger(ctx).Errorf("Server closed the stream without sending trailers")
				return errors.New("internal server error")
			}
			return err

		}
		endTime := time.Now()
		totalTime := endTime.Sub(startTime)
		duration, derr := durafmt.ParseString(totalTime.Round(time.Second).String())

		if derr != nil {
			fmt.Println(derr)
		}

		outputChan <- &pb.Output{Response: fmt.Sprintf("Finished Run in %s", duration), Created: timestamppb.Now()}

		return err
	} else {

		Logger(ctx).Debugf("Running in normal mode")

		for {

			switch val := <-controlChan; val {

			case "AutoSave":
				Logger(ctx).Debug("Got save waiting %v milliseconds", fsWatchDelay)
				outputChan <- &pb.Output{Response: "File change detected", Created: timestamppb.Now()}

				time.Sleep(fsWatchDelay * time.Millisecond)

				outputChan <- &pb.Output{Response: "Starting run", Created: timestamppb.Now()}

				startTime, err := singleTestRun(ctx, super, privateKey, syncHost, workingDirectory, outputChan, authCreds, true, nil)

				if err != nil {
					Logger(ctx).Debugf("Error during run %+v ", err)
				}

				endTime := time.Now()
				totalTime := endTime.Sub(startTime)
				duration, timeErr := durafmt.ParseString(totalTime.Round(time.Second).String())

				if timeErr != nil {
					fmt.Println(timeErr)
					Logger(ctx).Errorf("Error getting time %+v ", timeErr)
				}

				outputChan <- &pb.Output{Response: fmt.Sprintf("Completed in %v ", duration), Created: timestamppb.Now()}

				for len(controlChan) > 0 {
					Logger(ctx).Debug("Emptying controlChan after Autosave run")
					<-controlChan
				}
				if err != nil {
					Logger(ctx).Debugf("Receieved error from test run : %v", err)
					outputChan <- &pb.Output{Response: "Test run failed", Created: timestamppb.Now()}
				}

			case "R", "\r", "\n":

				outputChan <- &pb.Output{Response: "Starting run", Created: timestamppb.Now()}
				Logger(ctx).Debug("Connectivity : %v", conn.GetState())
				startTime, err := singleTestRun(ctx, super, privateKey, syncHost, workingDirectory, outputChan, authCreds, true, nil)
				if err != nil && !errors.Is(err, context.Canceled) {
					fmt.Println(err)
					Logger(ctx).Errorf("Error during run %+v ", err)
					_ = bugsnag.Notify(err)

				}
				endTime := time.Now()
				totalTime := endTime.Sub(startTime)
				duration, timeErr := durafmt.ParseString(totalTime.Round(time.Second).String())
				if timeErr != nil {
					fmt.Println(timeErr)
					Logger(ctx).Errorf("Error getting time %+v ", timeErr)
				}
				outputChan <- &pb.Output{Response: fmt.Sprintf("Completed in %v ", duration), Created: timestamppb.Now()}
				for len(controlChan) > 0 {
					Logger(ctx).Debugf("Emptying control channel")
					<-controlChan
				}

				if err != nil {
					Logger(ctx).Debugf("Receieved error from test run : %v", err)
					outputChan <- &pb.Output{Response: "Test run failed", Created: timestamppb.Now()}

				}

			case "Q":
				Logger(ctx).Debug("Got quit")

				return nil

			case "X", "x":
				Logger(ctx).Debug("Got clear workers")
				clearErr := ClearWorkersForProject(ctx, viper.GetString("ApiEndpoint"), "")
				if clearErr != nil {
					Logger(ctx).Errorf("Got error removing workers : %v", clearErr)
					outputChan <- &pb.Output{Response: fmt.Sprintf("Got error rebuilding - %v", clearErr)}

				} else {
					// should I quit? hmm
					outputChan <- &pb.Output{Response: "Servers Cleared", Created: timestamppb.Now()}

					return nil
				}
			default:
				outputChan <- &pb.Output{Response: fmt.Sprintf("Key not recognized %v", val), Created: timestamppb.Now()}
				outputChan <- &pb.Output{Response: usageString, Created: timestamppb.Now()}
				Logger(ctx).Debugf("Key not recognized %v", val)

			}
		}
	}

}

func markSuperAsUnreachable(ctx context.Context, projectToken string, sup *api.Super) error {
	Logger(ctx).Debugf("marking supervisor %d as unreachable ", sup.Id)
	ctx, span := otel.Tracer(name).Start(ctx, "markSuperAsUnreachable")
	defer span.End()

	conn, err := ApiConn(ctx, viper.GetString("ApiEndpoint"))
	if err != nil {
		return err
	}
	defer conn.Close()
	c := api.NewSupersClient(conn)
	_, clientErr := c.MarkSuperAsUnreachable(ctx, &api.UnreacheableReq{ProjectToken: projectToken, Super: sup})
	if clientErr != nil {
		Logger(ctx).Errorf("Error marking super as unreachable %+v ", clientErr)
		return clientErr
	}
	Logger(ctx).Debug("Finishing marking")
	return nil
}
func getSuperAffinity(ctx context.Context) string {
	affinityVar := viper.GetString("AFFINITY_ENV_VAR")
	var affinity string
	if len(affinityVar) > 0 {
		Logger(ctx).Debugf("Getting affinity from env var %s", affinityVar)
		affinity = os.Getenv(affinityVar)
		Logger(ctx).Debugf("Affinity is %s", affinity)
	} else {
		Logger(ctx).Debug("No affinity env var set, affinity will be empty")
	}

	return affinity

}

// add to this a path for the folder so that we can make sure to limit the runs on a per folder/project basis
func getSuperForProject(ctx context.Context, projectToken string, uniqueInstanceId string, config Config) (*api.Super, error) {
	Logger(ctx).Debug("Config is ")
	Logger(ctx).Debug(config)
	Logger(ctx).Debugf("Getting supervisor for project %s", projectToken)
	conn, err := ApiConn(ctx, viper.GetString("ApiEndpoint"))
	if err != nil {
		Logger(ctx).Errorf("Error fetching super for project %v", err)
		return nil, err
	}
	defer conn.Close()
	Logger(ctx).Debug("connecting to client")
	c := api.NewProjectsClient(conn)
	Logger(ctx).Debug("connected")

	affinity := getSuperAffinity(ctx)

	in := api.GetSuperReq{ProjectToken: projectToken, UniqueInstanceId: uniqueInstanceId, Affinity: affinity}
	Logger(ctx).Debug("attempting connection for project")

	clientDeadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(ctx, clientDeadline)

	defer CancelCtx(ctx, cancel, "getSuperForProject returning so cancelling the context")
	out, err := c.GetSuperForProject(ctx, &in)

	if err != nil {
		Logger(ctx).Debugf("got an error %s", err)
		st, ok := status.FromError(err)
		if !ok {
			Logger(ctx).Debug("NOT OK")
			Logger(ctx).Error("Not a status error when fetching super from api %v", err)
			return nil, err
		}

		Logger(ctx).Debug(st.Message())
		Logger(ctx).Debug(st.Code())
		if st.Code() == codes.DeadlineExceeded {
			Logger(ctx).Error("Could not get a super from project")
			return nil, errors.New("could not get a super for project")
		}
		if st.Code() == codes.Unimplemented {
			Logger(ctx).Error("got unimplemeted match")
			return nil, errors.New("unimplemented")
		}
		// Should probably barf here for now
		Logger(ctx).Error("Got an error from the api when getting super %v ", err)
		return nil, err
	}

	Logger(ctx).Debugf("Got super %+v", out.Super)
	Logger(ctx).Debugf("Have super with id %v, ip address %s and endpoint %s", out.Super.Id, out.Super.IpAddress, out.Super.ExternalEndpoint)
	return out.Super, nil
}
func setupDisplay() OutputWriter {
	status := NewStatusScreen()
	status.SetColor(255, 255, 0)
	// status.SetBackgroundColor(255, 255, 255)
	scrolledOutput := NewScrollingScreen()
	scrolledOutput.SetColor(255, 255, 255)
	screens := []*ScreenSect{scrolledOutput, status}
	//screens := []*shared.ScreenSect{&status}
	display := NewDisplay(os.Stdout.Fd(), screens)

	return display
}

func setUlimit(ctx context.Context, ulimit int) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}
	Logger(ctx).Debugf("The current rLimit is %v ", rLimit)
	rLimit.Max = 999999
	rLimit.Cur = uint64(ulimit)
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		Logger(ctx).Errorf("Error Setting Rlimit %s", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		Logger(ctx).Errorf("Error Getting Rlimit %s", err)
	}

	Logger(ctx).Info("Rlimit Final", rLimit)
}
func ClearWorkers(ctx context.Context) error {

	ctx, span := otel.Tracer(name).Start(ctx, "ClearWorkers")
	defer span.End()
	workingDirectory, wdErr := os.Getwd()
	if wdErr != nil {
		return wdErr
	}
	setupUUID(ctx)

	config, err := initialLoadConfig(ctx, workingDirectory)
	if err != nil {
		return err
	}
	ctx, _, err = setupAuthCtx(ctx, *config)
	if err != nil {
		fmt.Println("Error during auth: ", err)
		return err
	}

	err = ClearWorkersForProject(ctx, viper.GetString("ApiEndpoint"), "")
	if err != nil {
		fmt.Println("Error clearing workers: ", err)
		return err
	}
	fmt.Println("Cleared workers")
	return nil

}

func initialLoadConfig(ctx context.Context, workingDirectory string) (*Config, error) {
	var err error

	out := viper.GetString("PROJECT_CONFIG_FILE")
	if out != constants.DEFAULT_PROJECT_CONFIG_FILE {
		fmt.Println("Config file is ", out)
	}

	configFile := filepath.Join(workingDirectory, viper.GetString("PROJECT_CONFIG_FILE"))
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		Logger(ctx).Errorf("No project found in current directory - checking path %s", configFile)

		// fmt.Println("no project config file found in current directory - if you haven't run it before use \"brisk project init rails|jest \" to create one")
		return nil, fmt.Errorf("config file %s not found - use \"brisk project init rails|jest \" to create one if this is your first run", configFile)
	}
	config, err := ReadConfig(ctx, viper.GetString("PROJECT_CONFIG_FILE"))

	if err != nil {
		Logger(ctx).Errorf("Error reading config %v", err)
		return nil, err
	}
	Logger(ctx).Debugf("Reading config it is  %+v", config)
	if config.SplitByJUnit {
		files, errFiles := GetFilesByJunitJest(ctx, "unit/junit.xml")
		if errFiles != nil {
			Logger(ctx).Errorf("error reading junit.xml %v", errFiles)
			return nil, errFiles
		}
		config.OrderedFiles = files
	}
	return config, nil
}
func setupAuthCtx(ctx context.Context, config Config) (context.Context, auth.AuthCreds, error) {
	authCreds := auth.AuthCreds{ProjectToken: config.ProjectToken, ApiToken: viper.GetString("apiToken"), ApiKey: viper.GetString("apiKey")}

	ctx, authErr := prepareOutgoingContext(ctx, config.ProjectToken, authCreds, uniqueInstanceId)
	if authErr != nil {
		Logger(ctx).Errorf("Error preparing outgoing context %v", authErr)

	}

	return ctx, authCreds, authErr

}

func setupUUID(ctx context.Context) {

	uniqueInstanceId = cli_utils.GetUniqueInstanceID(ctx)

	Logger(ctx).Debugf("Set uuinstance id to  %v", uniqueInstanceId)

}
func isGatewayError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "502") {
		return true
	}
	return false
}

func RunBrisk(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	ctx, span := otel.Tracer(name).Start(ctx, "RunBrisk")

	cleanup := func() {
		cancel()
		span.AddEvent("cleanup called")
		span.End()
		honeycomb.ShutdownTracer()

	}
	defer cleanup()

	firstRunMessage()
	fmt.Println()
	var err error
	setupUUID(ctx)
	if err != nil {
		return err
	}

	if !viper.GetBool("CI") {
		//don't watch in CI
		setUlimit(ctx, 99999)
	}

	//check if the config file exists
	workingDirectory, wdErr := os.Getwd()
	if wdErr != nil {
		return wdErr
	}

	config, err := initialLoadConfig(ctx, workingDirectory)
	if err != nil {
		return err

	}
	var display OutputWriter
	var status *ScreenSect
	// if no terminal
	if viper.GetBool("NO_TERM") || !term.IsTerminal(os.Stdout.Fd()) {

		Logger(ctx).Debug("Not a terminal")
		display = &TerminalWriter{Output: os.Stdout}
		status = &ScreenSect{}
	} else {
		Logger(ctx).Debug("Is a terminal")

		display = setupDisplay()
		// scroll := display.Screens[0]
		status = display.GetStatusScreen()

		status.Set("⚡ Initializing")
	}
	Logger(ctx).Infof("Brisk CLI version:  %s", constants.VERSION)
	Logger(ctx).Debug(config)

	if err != nil {
		Logger(ctx).Errorf("Error loading config %v", err)
		return err

	}

	var logOuputChannel = make(chan *pb.Output, 10)
	controlChan := make(chan string, 2)
	controlChan <- "AutoSave"
	cancelChan := make(chan string, 1)

	if term.IsTerminal(os.Stdin.Fd()) {
		go func() { readInput(ctx, controlChan, cancelChan) }()
	} else {
		Logger(ctx).Debug("No /dev/tty so not reading input")
	}
	go catchSignals(ctx, cancelChan, cleanup)
	ctx, authCreds, authErr := setupAuthCtx(ctx, *config)
	if authErr != nil {
		Logger(ctx).Error(authErr)
		return authErr
	}
	Logger(ctx).Debug(config)
	listData = append(listData, "starting...")

	go printOutputLoop(ctx, config, logOuputChannel, listData, display, status)

	fmt.Print("connecting to brisk...")
	super, err := getSuperForProject(ctx, config.ProjectToken, uniqueInstanceId, *config)
	fmt.Print("connected \n")
	if err != nil {
		Logger(ctx).Errorf("Error getting supervisor %v", err)

		if errors.Is(err, context.Canceled) {
			fmt.Println("\n Could not connect to api when getting supervisor - context canceled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("\n Timeout - Could not connect to api when getting supervisor - deadline exceeded")
		}

		return err
	}

	if viper.GetBool("watch") && !viper.GetBool("CI") {
		Logger(ctx).Debug("Watching files...")
		go watchFiles(ctx, workingDirectory, logOuputChannel, controlChan, config)
	}

	Logger(ctx).Debugf("sync endpoint is %v", super.SyncEndpoint)

	timeout := viper.GetDuration("RETRY_TIMEOUT")
	if timeout < time.Duration(5*time.Second) {
		timeout = time.Duration(5 * time.Second)
	}
	retryCount := 0
	go func() {
		for {
			ctx, forSpan := otel.Tracer(name).Start(ctx, "RunBrisk:connectToServer")
			err = connectToServer(ctx, *config, super, logOuputChannel, cancelChan, controlChan, super.SyncEndpoint, workingDirectory, config.ProjectToken, viper.GetString("apiKey"), viper.GetString("apiToken"))
			if err == nil {
				forSpan.End()
				cancelChan <- "finished normally"
				return

			}
			if err == TestFailedError {

				Logger(ctx).Error(err.Error())
				// logOuputChannel <- pb.Output{Response: fmt.Sprintf("Error %v", err), Stderr: err.Error(), Created: timestamppb.Now()}
				// time.Sleep(1 * time.Second)

				forSpan.End()
				cancelChan <- "tests failed"
				return

			}

			if CancelledError(err) {
				forSpan.End()
				cancelChan <- "context canceled"
				return

			}
			forSpan.RecordError(err)
			forSpan.SetStatus(otel_codes.Error, err.Error())
			_ = bugsnag.Notify(err)

			if errors.Is(err, ProjectInUseError) {
				Logger(ctx).Info(err)
				status.Set("")
				fmt.Println("Project is already in use - waiting.")
				fmt.Println("")

				timeout := viper.GetInt("PROJECT_IN_USE_TIMEOUT")
				retryCount = 0
				time.Sleep(time.Duration(timeout) * time.Second)

			} else if retryCount > viper.GetInt("RETRY_COUNT") {
				Logger(ctx).Info(err)
				status.Set("")
				fmt.Println("Too many errors - exiting")
				forSpan.End()
				cancelChan <- "too many errors - exiting"
				return

			} else if errors.Is(err, context.DeadlineExceeded) {
				Logger(ctx).Info(err)
				status.Set("")
				retryCount++
				Logger(ctx).Infof("Timeout - retrying in %v seconds", (time.Duration(retryCount+1) * timeout))
				fmt.Printf("Timeout - retrying in %v seconds \n", (time.Duration(retryCount+1) * timeout))
				time.Sleep(time.Duration(retryCount+1) * timeout)
			} else if grpc_status.Convert(err).Code() == codes.Unavailable || grpc_status.Convert(err).Code() == codes.NotFound || grpc_status.Convert(err).Code() == codes.Internal || grpc_status.Convert(err).Code() == codes.Unauthenticated || err == RSyncError || isGatewayError(err) {
				Logger(ctx).Infof("The error is %v with code %v", err, grpc_status.Convert(err).Code())
				Logger(ctx).Infof("Network Error - retrying in %v seconds", (time.Duration(retryCount+1) * timeout))
				fmt.Printf("Network Error - retrying in %v seconds \n", (time.Duration(retryCount+1) * timeout))
				retryCount++
				time.Sleep(time.Duration(retryCount+1) * timeout)
			} else {
				Logger(ctx).Errorf("error is not ProjectInUseError it is %v", err)

				Logger(ctx).Error("when error converted it is %v", grpc_status.Convert(err).Code())
				fromError, ok := grpc_status.FromError(err)
				if ok {
					Logger(ctx).Error("From errors in ok  %v", fromError)
				} else {
					Logger(ctx).Error("From errors not ok is  %v", fromError)
				}

				if err.Error() == SUPER_NOT_REACHABLE {
					Logger(ctx).Info("SUPER_NOT_REACHABLE")
					superErr := Retry(ctx, 5, 1, func() error { return markSuperAsUnreachable(ctx, config.ProjectToken, super) })
					if superErr != nil {
						Logger(ctx).Debug(superErr)
						Logger(ctx).Debug("can't mark super")

						forSpan.End()
						cancelChan <- "can't mark super - exiting"
						return

					}
					Logger(ctx).Error("Can't reach supervisor retrying")
					time.Sleep(time.Second * 1)
					var newSuperErr error
					super, newSuperErr = RetryWithReturn(ctx, 5, 1, func() (*api.Super, error) {
						Logger(ctx).Debug("getting new super")

						return getSuperForProject(ctx, authCreds.ProjectToken, uniqueInstanceId, *config)
					})
					if newSuperErr != nil {
						Logger(ctx).Error("Can't connect to servers %v", newSuperErr)
						fmt.Println("Can't connect to servers")

						forSpan.End()
						cancelChan <- "can't connect to servers - exiting"
						return
					}

					Logger(ctx).Infof("got new super %v", super.Id)

				}
			}
			forSpan.End()
		}
	}()

	val := <-cancelChan

	Logger(ctx).Debugf("cancelChan received %v", val)
	if val == "finished normally" || val == "Q" || val == "q" {
		return nil
	} else {
		return errors.New(val)
	}
}

var DEFAULT_JEST_FILTER_LIST = []string{"node ./scripts/jest/jest-cli.js", "cannot set terminal process group", "jest-haste-map: duplicate manual mock",
	"bash: no job control in this shell", "The following files share their name; please delete one of them:", "* <rootDir>/packages/", "Running tests for default (www-modern)",
	"NODE_ENV=development RELEASE_CHANNEL=experimental", "run a command as administrator", "man sudo_root", "Running command /usr/bin/bash -c -i", "Not Building Worker", "Sending command", "yarn run",
	"node ./scripts/jest/jest-cli.js", "Ran all test suites matching", "Time:", "Tests:", "Snapshots:", "Test Suites:", "Force exiting Jest: Have you considered using", "Test results written to:", "Done in"}

func getFilterList(config Config) []string {

	if config.FilterList != nil && len(config.FilterList) > 0 {
		return config.FilterList
	}
	return DEFAULT_JEST_FILTER_LIST
}

func shouldFilter(output *pb.Output, config Config) bool {
	if os.Getenv("NO_FILTER") == "true" {
		return false
	}
	response := stripansi.Strip(output.Response)
	if len(response) == 0 {
		return true
	}

	for _, s := range getFilterList(config) {
		if strings.Contains(response, s) {
			return true
		}
	}
	return false

}

func readInput(ctx context.Context, controlChan chan string, cancelChan chan string) {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	for {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		Logger(ctx).Debug("I got the byte ", r, "("+string(r)+")")
		val := strings.ToUpper(string(r))
		if val == "Q" {
			cancelChan <- val
		}

		controlChan <- val
		// handle key event
	}

}
func PrintContextInternals(ctx interface{}, inner bool) {
	contextValues := reflect.ValueOf(ctx).Elem()
	contextKeys := reflect.TypeOf(ctx).Elem()

	if !inner {
		fmt.Printf("\nFields for %s.%s\n", contextKeys.PkgPath(), contextKeys.Name())
	}

	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			reflectValue := contextValues.Field(i)
			reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()

			reflectField := contextKeys.Field(i)

			if reflectField.Name == "Context" {
				PrintContextInternals(reflectValue.Interface(), true)
			} else {
				fmt.Printf("field name: %+v\n", reflectField.Name)
				fmt.Printf("value: %+v\n", reflectValue.Interface())
			}
		}
	} else {
		fmt.Printf("context is empty (int)\n")
	}

}
func setupKeys(ctx context.Context, config Config, authCreds auth.AuthCreds, c pb.BriskSupervisorClient, userDetails *pb.UserDetails, buildCommands []*api.Command) (string, string, error) {
	// ctx = context.Background()
	ctx, authErr := auth.AddAuthToCtx(ctx, authCreds)
	if authErr != nil {
		return "", "", authErr
	}
	ctx, span := otel.Tracer(name).Start(ctx, "setupKeys")
	defer span.End()

	// TODO MAYBE MAKE THESE ACTUAL COMMANDS WITH PROPER ENV
	initMessage := pb.TestOption{Command: "init", UserDetails: userDetails}
	Logger(ctx).Debug("in setupKeys")
	defer Logger(ctx).Debug("exit setupKeys")
	Logger(ctx).Debug("Before calling Setup")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.Setup(ctx, &initMessage)
	Logger(ctx).Debug("After calling Setup")
	if err != nil {
		Logger(ctx).Infof("Error when calling Setup: %v", err)

		fmt.Println(".")
		st, ok := status.FromError(err)
		if !ok {
			Logger(ctx).Error("NOT OK")
			return "", "", err
		}

		Logger(ctx).Debug(st.Message())
		Logger(ctx).Debug(st.Code())
		if st.Code() == codes.Unimplemented {
			Logger(ctx).Debug("got unimplemented no super here")
			return "", "", errors.New(SUPER_NOT_REACHABLE)
		}

		if st.Code() == codes.Unavailable {
			Logger(ctx).Debug("got unavailable no super here")
			return "", "", errors.New(SUPER_NOT_REACHABLE)
		}
		Logger(ctx).Errorf("Error when calling Setup: %v", err)
		return "", "", err
	}

	if viper.GetBool("print_keys") {
		Logger(ctx).Debugf("got response %v", response)
	}
	privateKey := response.Message
	if len(privateKey) == 0 {
		Logger(ctx).Error("No private key returned")
		fmt.Println("Internal Error: No private key returned from server - please contact support")
		os.Exit(1)
	}

	username := "brisk"
	Logger(ctx).Debug(username)
	return privateKey, username, nil
}

const SUPER_NOT_REACHABLE = "super not reachable"

func grpcConfigFromConfig(ctx context.Context, config Config) *pb.Config {

	Logger(ctx).Debugf("Config in is %+v", config)
	var buildCommands []*api.Command
	for _, c := range config.BuildCommands {
		buildCommands = append(buildCommands, &api.Command{Commandline: c.Commandline, WorkDirectory: c.WorkDirectory, Args: c.Args, Background: c.Background})
	}
	out := pb.Config{AutomaticSplitting: config.AutomaticSplitting, NoFailFast: config.NoFailFast, SkipRecalcFiles: config.SkipRecalcFiles, ListTestCommand: config.ListTestCommand, PreListTestCommand: config.PreListTestCommand, BuildCommands: buildCommands, Framework: config.Framework, ExcludedFromSync: config.ExcludedFromSync, WorkerImage: config.WorkerImage, Environment: config.Environment, Concurrency: int32(config.Concurrency), SplitByJUnit: config.SplitByJUnit, OrderedFiles: config.OrderedFiles, RebuildFilePaths: config.RebuildFilePaths}

	return &out

}

// we seem to call this a lot of times
func prepareOutgoingContextMin(ctx context.Context, projectToken string, authCreds auth.AuthCreds, uniqueInstanceId string) (context.Context, error) {
	if uniqueInstanceId == "" {
		Logger(ctx).Errorf("uniqueInstanceId is empty")
		debug.PrintStack()

		return nil, errors.New("need a machine uuid")
	}

	traceKey := GetKey(ctx, projectToken)
	// if viper.GetBool("PRINT_TRACE_KEY") {
	// 	fmt.Printf("run: %v \n", traceKey)
	// }
	ctx = metadata.AppendToOutgoingContext(ctx, "trace-key", traceKey)
	ctx = WithTraceId(ctx, traceKey)

	ctx = metadata.AppendToOutgoingContext(ctx, "unique_instance_id", uniqueInstanceId)

	ctx = metadata.AppendToOutgoingContext(ctx, "brisk_api_version", constants.VERSION)

	//TODO why is this causing crazy crashes - 400s were we  don't even hit the LB
	// ctx, authErr := auth.AddAuthToCtx(ctx, authCreds)

	// if authErr != nil {
	// 	return nil, fmt.Errorf("Need credentials +v", authErr)
	// }
	return ctx, nil
}
func prepareOutgoingContext(ctx context.Context, projectToken string, authCreds auth.AuthCreds, uniqueInstanceId string) (context.Context, error) {
	if uniqueInstanceId == "" {
		Logger(ctx).Errorf("uniqueInstanceId is empty")
		debug.PrintStack()

		return nil, errors.New("need a machine uuid")
	}

	traceKey := GetKey(ctx, projectToken)
	if viper.GetBool("PRINT_TRACE_KEY") {
		fmt.Println("")
		fmt.Println(traceKey)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "trace-key", traceKey)
	ctx = WithTraceId(ctx, traceKey)

	ctx = metadata.AppendToOutgoingContext(ctx, "unique_instance_id", uniqueInstanceId)

	ctx = metadata.AppendToOutgoingContext(ctx, "brisk_api_version", constants.VERSION)
	ctx, authErr := auth.AddAuthToCtx(ctx, authCreds)

	if authErr != nil {
		return nil, fmt.Errorf("Need credentials +%v", authErr)
	}
	return ctx, nil
}

func apiCommandFromConfigCommand(command Command, env map[string]string, framework string) *api.Command {
	return &api.Command{Environment: env, TestFramework: framework, Commandline: command.Commandline, WorkDirectory: command.WorkDirectory, Args: command.Args, Background: command.Background, CommandConcurrency: int32(command.CommandConcurrency), CommandId: command.CommandId, NoTestFiles: command.NoTestFiles}
}

func parseCommandsFromConfig(ctx context.Context, config Config) ([]*api.Command, []*api.Command) {
	var buildCommands []*api.Command
	var commands []*api.Command
	env := config.Environment
	for _, v := range config.BuildCommands {
		buildCommands = append(buildCommands, apiCommandFromConfigCommand(v, env, config.Framework))
	}
	Logger(ctx).Debugf("Build commands are %v", buildCommands)
	for _, v := range config.Commands {
		commands = append(commands, apiCommandFromConfigCommand(v, env, config.Framework))
	}
	Logger(ctx).Debugf("Test Commands are %v", commands)
	return buildCommands, commands
}

func singleTestRun(ctx context.Context, super *api.Super, privateKey string, syncHost string, workingDirectory string, outputChan chan *pb.Output, authCreds auth.AuthCreds, sync bool, repoInfo *api.RepoInfo) (time.Time, error) {
	ctx, cancel := context.WithCancel(ctx)
	startTime := time.Now() // we reset this later but we need something for errors

	defer CancelCtx(ctx, cancel, "singleTestRun returning so cancelling context")
	ctx, span := otel.Tracer(name).Start(ctx, "singleTestRun")
	defer span.End()
	conn, err := getConn(ctx, super)
	if err != nil {
		return startTime, err

	}
	c := pb.NewBriskSupervisorClient(conn)

	config, err := ReadConfig(ctx, viper.GetString("PROJECT_CONFIG_FILE"))

	if err != nil {
		Logger(ctx).Errorf("Error reading config file %v", err)
		return startTime, err
	}
	projectToken := config.ProjectToken
	buildCommands, commands := parseCommandsFromConfig(ctx, *config)

	err = checkBuildCommands(ctx, buildCommands, outputChan)
	if err != nil {
		return startTime, err
	}

	outputChan <- &pb.Output{Response: "Waiting on project supervisor...", Created: timestamppb.Now()}
	Logger(ctx).Debugf("Calling lock on super")
	fmt.Print("Acquiring lock on supervisor...")
	lockClient, lockErr := c.Lock(ctx, &pb.LockRequest{})
	defer lockClient.CloseSend()
	if lockErr != nil {
		Logger(ctx).Errorf("Error connecing to super for lock  -  %v", lockErr)
		return startTime, lockErr
	}
	Logger(ctx).Debugf("Got lock client %+v", lockClient)
	Logger(ctx).Debugf("Now waiting for lock")

	mesg, lockErr := lockClient.Recv() // wait for lock
	if lockErr != nil {
		Logger(ctx).Errorf("Error getting lock: %v", lockErr)

		return startTime, lockErr
	}
	Logger(ctx).Debugf("Got lock message %v", mesg.Message)

	if mesg.Message != constants.SUPER_LOCKED {
		Logger(ctx).Errorf("Expected lock message but got %v", mesg.Message)
		return startTime, fmt.Errorf("expected lock message but got %v", mesg.Message)
	}

	Logger(ctx).Debugf("Got lock message %v", mesg.Message)
	// so now we have a lock
	// but if the stream goes down we want to bail

	go func() {
		for {
			content, err := lockClient.Recv()

			if err == context.Canceled {
				Logger(ctx).Debugf("Lock context cancelled")
				return
			}

			if CancelledError(err) {
				Logger(ctx).Debugf("Lock cancelled")
				return
			}

			if err != nil {
				Logger(ctx).Errorf("Error getting lock message %v", err)
				cancel()
				return
			}
			Logger(ctx).Debugf("Got lock message - although we weren't expecing one %v", content.Message)
		}
	}()

	startTime = time.Now()
	if sync {
		outputChan <- &pb.Output{Response: "Starting sync...", Created: timestamppb.Now()}
		syncErr := StartSync(ctx, *config, workingDirectory, privateKey, syncHost, super.SyncPort, globalUsername, outputChan)
		if syncErr != nil {
			Logger(ctx).Error(syncErr)
			outputChan <- &pb.Output{Response: fmt.Sprintf("Error during sync:- %v", syncErr), Created: timestamppb.Now()}

			return startTime, syncErr
		}
		outputChan <- &pb.Output{Response: "Synced", Created: timestamppb.Now()}

		listData = append(listData, "synced...")
	}
	Logger(ctx).Debug("TIMING synced at  %v", time.Since(startTime))

	if os.Getenv("QUIT_AFTER_SYNC") == "true" {
		Logger(ctx).Debug("QUIT_AFTER_SYNC is true so quitting")
		return startTime, nil
	}
	Logger(ctx).Debugf("Build commands I'm sending are %+v ", buildCommands)
	conf := grpcConfigFromConfig(ctx, *config)
	to := pb.TestOption{Command: "runn the tests", BuildCommands: buildCommands, Commands: commands, Config: conf, RepoInfo: repoInfo}
	Logger(ctx).Debugf("The config I'm sending is %+v", to.Config)
	Logger(ctx).Debugf("The environment I'm sending is %+v", to.Config.Environment)

	//for {

	ctx, err = prepareOutgoingContext(ctx, projectToken, authCreds, uniqueInstanceId)
	if err != nil {
		Logger(ctx).Errorf("Error preparing outgoing context %v", err)
		return startTime, err
	}

	//ctx = addAuthToMD(ctx, config)

	Logger(ctx).Info("Init test run")
	Logger(ctx).Debugf("TIMING Starting test run at %v", time.Since(
		startTime))
	Logger(ctx).Debugf("the parameters to RunTests are %#v ", &to)
	outputChan <- &pb.Output{Response: "Starting test run...", Created: timestamppb.Now()}
	r, err := c.RunTests(ctx, &to)
	defer r.CloseSend()
	if err != nil {
		Logger(ctx).Error("Recieved error from test run setup %v", err)
		outputChan <- &pb.Output{Response: "Error from test run", Created: timestamppb.Now()}
		Logger(ctx).Debug(err)
		return startTime, err
	}

	for {
		Logger(ctx).Debug("Before receive")

		in, err := r.Recv()
		if in != nil {
			outputChan <- in
			Logger(ctx).Debugf(" %+v", in)
		}
		if err == nil && in == nil {
			Logger(ctx).Debug("We have a problem")
			return startTime, errors.New("input is nil and no error")
		}
		if in != nil && in.Control == types.FAILED {
			Logger(ctx).Debugf("Got Failed from server %+v", in)
			Logger(ctx).Debugf("Exit code is %v", in.Exitcode)
			return startTime, TestFailedError

		}
		if in != nil && in.Control == types.FINISHED {
			Logger(ctx).Debug("Got finished from server")
			Logger(ctx).Debugf("Exit code is %v", in.Exitcode)
			// Logger(ctx).Debug("we really aught to count these down so we know when we are finished")
			Logger(ctx).Debugf("The final result from this test run is %+v", in)

		}
		if err != nil {

			switch err {
			case io.EOF:
				{
					Logger(ctx).Debugf("Recv'd done %v", err)
					Logger(ctx).Debugf("Returning from single test run with : %v", err)
					outputChan <- &pb.Output{Response: "Finished Run", Created: timestamppb.Now()}

					return startTime, nil
				}

			case io.ErrUnexpectedEOF:
				{
					// read done.
					Logger(ctx).Debugf("Recv'd done %v", err)
					//close(waitc)
					return startTime, err
				}

			default:
				{
					if status.Code(err) == codes.Unavailable {
						fmt.Print("server is unavailable")
						Logger(ctx).Debugf("Server is unavailable %v", err)

						return startTime, err
					}

					if errors.Is(err, context.DeadlineExceeded) {
						Logger(ctx).Debugf("Deadline Exceeded %v", err)
						return startTime, err
					}
					if CancelledError(err) {
						Logger(ctx).Debugf("Context Cancelled %v", err)
						return startTime, err
					}

					if status.Code(err) == 1 {
						Logger(ctx).Debug("Run cancelled")
						outputChan <- &pb.Output{Response: "Cancelled", Created: timestamppb.Now()}

						return startTime, err
					} else {
						Logger(ctx).Errorf("Received Error during test run %v", err)
						outputChan <- &pb.Output{Response: "Error during test run:", Created: timestamppb.Now()}
						outputChan <- &pb.Output{Response: err.Error(), Created: timestamppb.Now()}

						Logger(ctx).Debug(status.Code(err))

						Logger(ctx).Debugf("where the error happened is : %s", debug.Stack())
						return startTime, err
					}
				}
			}
		}
		//Logger(ctx).Debugf("in is %v", in)
		//Logger(ctx).Debugf("response is %v", in.Response)

	}
}

// // checks the git hash and rebuilds if required
// func rebuildRequired(ctx context.Context, config Config) (bool, error) {
// 	if viper.GetBool("NoRebuild") {
// 		Logger(ctx).Info(" Not Rebuilding")
// 		return false, nil
// 	}
// 	Logger(ctx).Debug("Checking if we need to rebuild")
// 	Logger(ctx).Debugf("Hash file path is %v", viper.GetString("HashFilePath"))
// 	Logger(ctx).Debugf("RebuildWatchPaths is %v", viper.GetStringSlice("RebuildWatchPaths"))
// 	path, err := os.Getwd()

// 	if err != nil {
// 		Logger(ctx).Errorf("Error getting current working directory %v", err)
// 		return false, err
// 	}

// 	val, err := git.RebuildRequired(ctx, viper.GetString("HashFilePath"), path, config.RebuildFilePaths)

// 	if err != nil {
// 		Logger(ctx).Error("Error checking if we need to rebuild %v", err)
// 		return false, err
// 	}
// 	if val {
// 		Logger(ctx).Info("Needs Rebuild")
// 		return true, nil

// 	} else {
// 		Logger(ctx).Info("Not Rebuilding")
// 		return false, nil
// 	}
// }

func firstRunMessage() {
	if (viper.GetTime("FIRST_RUN_AT") != time.Time{}) {
		return
	}

	if viper.GetBool("CI") {
		return
	}
	fmt.Println(`
	
Welcome to Brisk 👋👋👋
	
I'm excited you are trying out my project.

I would love to hear your feedback on any part of the process.

If you have any comments or questions or just to reach out

please email me at sean@brisktest.com, or tweet me at @Brisk_testing,
	
Happy testing!

Sean



	`)

	viper.Set("FIRST_RUN_AT", time.Now())
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println("Error writing config")
	}
}

func checkBuildCommands(ctx context.Context, buildCommands []*api.Command, outputChan chan *pb.Output) error {
	buildCommandHash, err := utilities.HashBuildCommands(buildCommands)
	if err != nil {
		Logger(ctx).Errorf("Error hashing build commands %v", err)
		return err
	}
	// we check if there is no stored build command hash (For BriskCI)
	// and check if the stored build command hash is different from the current one

	if viper.GetString("BuildCommandHash") != "" && viper.GetString("BuildCommandHash") != buildCommandHash {
		Logger(ctx).Debugf("Build commands have changed, clearing workers")
		viper.Set("BuildCommandHash", buildCommandHash)
		err = viper.WriteConfig()
		if err != nil {
			fmt.Println("Error writing config")
			Logger(ctx).Errorf("Error writing config %v", err)
			return err
		}
		outputChan <- &pb.Output{Response: "Build commands have changed, clearing workers", Created: timestamppb.Now()}
		err = ClearWorkersForProject(ctx, viper.GetString("ApiEndpoint"), "")
		if err != nil {
			Logger(ctx).Errorf("Error clearing workers %v", err)
			return err
		}

	}
	return nil
}
func hash(s string) uint8 {
	h := fnv.New32()
	h.Write([]byte(s))
	u32 := h.Sum32()
	return uint8(u32)
}

func printOutputLoop(ctx context.Context, config *Config, logOuputChannel chan *pb.Output, listData []string, display OutputWriter, status *ScreenSect) {

	for {

		select {
		case output := <-logOuputChannel:
			var outputPrefix string
			if len(config.Commands) > 1 && output.Command != nil && output.Command.CommandId != "" {
				commandId := strings.ToUpper(output.Command.CommandId)
				if !viper.GetBool("NO_COLORS") {
					// want to highlight the command Id in the output in a different background color
					backgroundColor := aec.RGB8Bit(hash(commandId))
					forgroundColor := aec.RGB8Bit(^hash(commandId))

					builder := aec.EmptyBuilder
					label := builder.Color8BitB(backgroundColor).Color8BitF(forgroundColor).Bold().ANSI
					outputPrefix = label.Apply(fmt.Sprintf(" %v ", commandId))

				} else {
					outputPrefix = fmt.Sprintf(" %v ", commandId)

				}
			} else {
				outputPrefix = ""
			}

			if len(output.Response) == 0 && len(output.Stderr) == 0 && len(output.Stdout) == 0 && output.BriskError == nil {
				Logger(ctx).Debugf("Got empty message %v", output)
				break
			}
			if output.Created == nil {
				output.Created = timestamppb.Now()
			}
			if os.Getenv("FILTER_LOGS") == "true" {
				if strings.Contains(output.Response, "PASS") || strings.Contains(output.Response, "FAIL") {
					listData = append(listData, fmt.Sprintf("%v %v, worker %v : %v ", outputPrefix, output.Created.AsTime().Local().Format("15:04:05"), PrintWorkerDetails(output), output.Response))
					//			scroll.Println(fmt.Sprintf("Worker %v - %v  \n", printWorkerDetails(output), output.Response))

				} else {
					Logger(ctx).Debug("Filtered log line ")
					//Logger(ctx).Debug(output.Response)

				}
			} else {

				//Logger(ctx).Debugf("Worker: %v - %v", printWorkerDetails(output), output.Response)
				if output.Stderr != "" {
					Logger(ctx).Debugf("%v, Worker: %v - %v ", output.Created.AsTime().Local().Format("15:04:05"), PrintWorkerDetails(output), output.Stderr)
					if !shouldFilter(output, *config) {
						display.Println(fmt.Sprintf("%v %v : %v ", outputPrefix, PrintWorkerDetails(output), output.Stderr))
					}
					listData = append(listData, fmt.Sprintf("%v %v, %v : %v ", outputPrefix, output.Created.AsTime().Local().Format("15:04:05"), PrintWorkerDetails(output), output.Stderr))
				} else {
					if !shouldFilter(output, *config) {

						display.Println(fmt.Sprintf("%v %v, %v - %v  \n", outputPrefix, output.Created.AsTime().Local().Format("15:04:05"), PrintWorkerDetails(output), output.Response))
					}
					listData = append(listData, fmt.Sprintf("%v %v, %v : %v ", outputPrefix, output.Created.AsTime().Local().Format("15:04:05"), PrintWorkerDetails(output), output.Response))

				}

				if output.BriskError != nil && output.BriskError.Error != "" {

					Logger(ctx).Debug(output.BriskError)
					listData = append(listData, fmt.Sprintf("%v  %v : %v ", outputPrefix, PrintWorkerDetails(output), output.BriskError.Error))
					if !shouldFilter(output, *config) {
						display.Println(fmt.Sprintf("Brisk Error worker %v : %v ", PrintWorkerDetails(output), output.BriskError.Error))
					}

				}

			}

			if config.Framework == string(types.Jest) && len(config.Commands) == 1 {
				status.Set(fmt.Sprintf("TESTS: [%v]     PASSED [%v]", output.TotalTestCount, output.TotalTestPass))
			} else {
				status.Set("")
			}

		}
	}

}

func watchFiles(ctx context.Context, workingDirectory string, logOuputChannel chan *pb.Output, controlChan chan string, config *Config) {

	watchErr := fsWatch.Watch(ctx, workingDirectory, controlChan, *config)
	if watchErr != nil {
		Logger(ctx).Errorf("Error watching %v", watchErr)
		logOuputChannel <- &pb.Output{Stderr: fmt.Sprintf("Error watching %v", watchErr)}
		controlChan <- "Q"

	}

}

var usageString = `Usage: "r" to run, "q" to quit, "x" to clear workers`

const fsWatchDelay = 10

var listData []string

func catchSignals(ctx context.Context, cancelChan chan string, cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt)

	var sig os.Signal = <-c
	go func() {

		time.Sleep(10 * time.Second)
		fmt.Println()
		fmt.Println("Exiting")

		os.Exit(1)
	}()
	fmt.Printf("Signal received %v, exiting \n", sig)
	Logger(ctx).Infof("Got interrupt %v, exiting", sig)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("signal", sig.String()))
	span.AddEvent("Got interrupt", trace.WithAttributes(attribute.String("signal", sig.String())))
	span.End()

	cancelChan <- "Signal:" + sig.String()
	cleanup()
	time.Sleep(100 * time.Millisecond)
	// we error out if brisk cli is ctrl-c'd
	os.Exit(1)

}
