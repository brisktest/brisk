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

package shared

import (
	"brisk-supervisor/api"
	pb "brisk-supervisor/brisk-supervisor"

	"brisk-supervisor/shared/auth"
	"brisk-supervisor/shared/constants"

	. "brisk-supervisor/shared/logger"
	"brisk-supervisor/shared/nomad"
	. "brisk-supervisor/shared/types"
	"errors"
	"strconv"

	. "brisk-supervisor/shared/context"
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	nomadapi "github.com/hashicorp/nomad/api"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	trace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Retry(ctx context.Context, attempts int, sleep int, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			Logger(ctx).Infof("retrying after error:", err)
			time.Sleep(time.Duration(sleep) * time.Second)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}

	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func RetryWithReturn[T any](ctx context.Context, attempts int, sleep int, f func() (T, error)) (result T, err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			Logger(ctx).Infof("retrying after error:", err)
			fmt.Print(".")
			time.Sleep(time.Duration(sleep) * time.Second)
			sleep *= 2
		}
		result, err = f()
		if err == nil {
			return result, nil
		}
	}
	return result, fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func GetRecentDeregisteredWorkers(ctx context.Context) ([]*api.Worker, error) {
	if viper.GetBool("HONEYCOMB_VERBOSE") {
		var span trace.Span
		ctx, span = otel.Tracer(name).Start(ctx, "GetRecentDeregisteredWorkers")
		defer span.End()
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		return nil, err
	}

	ctx, authErr := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("WORKER_RECENT_DEREG_ROUTE_TOKEN"), ApiKey: os.Getenv("WORKER_RECENT_DEREG_ROUTE_KEY")})
	if authErr != nil {
		panic(authErr)
	}
	defer conn.Close()
	c := api.NewWorkersClient(conn)
	in := api.WorkersReq{}
	out, err := c.GetRecentlyDeregistered(ctx, &in)
	if err != nil {
		return nil, err
	}

	Logger(ctx).Debugf("Workers deregistered : - %v", out.Workers)
	return out.Workers, err
}

func DeRegisterWorkersForProject(ctx context.Context, workers []*api.Worker) error {

	Logger(ctx).Debugf("DeRegisterWorkersForProject %v ", workers)
	ctx, span := otel.Tracer(name).Start(ctx, "DeRegisterWorkersForProject")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		return err
	}

	c := api.NewProjectsClient(conn)
	_, err = c.DeRegisterWorkers(ctx, &api.DeRegisterWorkersReq{Workers: workers})
	if err != nil {
		Logger(ctx).Errorf("Failed to deregister workers for project with workers : %+v error %v", workers, err)
	}
	return err
}

func ClearWorkersForProject(ctx context.Context, api_endpoint string, superID string) error {
	ctx, span := otel.Tracer(name).Start(ctx, "ClearWorkersForProject")
	defer span.End()
	conn, err := ApiConn(ctx, api_endpoint)
	if err != nil {
		return err
	}

	defer conn.Close()
	c := api.NewProjectsClient(conn)
	in := api.ClearWorkersReq{SupervisorUid: superID}
	out, err := c.ClearWorkersForProject(ctx, &in)
	if err != nil {
		Logger(ctx).Errorf("Failed to clear workers for project with super : %v error %v", superID, err)
		return err
	}

	Logger(ctx).Debugf("Workers cleared : - %v", out.Status)
	return nil
}

// TODO
// Create certs and use them for all comms between the super and the worker.
// Should be all internal traffic but it's another layer.
func SetupWorker(ctx context.Context, w *api.Worker, publicKey string) error {
	ctx, span := otel.Tracer(name).Start(ctx, "SetupWorker")
	defer span.End()

	startTime := time.Now()
	defer func() { Logger(ctx).Debugf("TIMING SetupWorker for %v took %v", w.Uid, time.Since(startTime)) }()

	endpoint := w.IpAddress + ":" + w.Port
	if len(endpoint) <= 0 {

		Logger(ctx).Error("Need an endpoint")

		return errors.New("need an endpoint for the worker")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	workerSetupTimeout := viper.GetDuration("WORKER_SETUP_TIMEOUT")
	ctx, cancel := context.WithTimeout(ctx, workerSetupTimeout)
	defer cancel()

	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(3),
		grpc_retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			Logger(ctx).Errorf("retrying after error: %v attempt # %v", err, attempt)
		}),
	}
	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), grpc_retry.StreamClientInterceptor(opts...), BugsnagClientInterceptor()),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), grpc_retry.UnaryClientInterceptor(opts...), BugsnagClientUnaryInterceptor),
		grpc.WithDefaultCallOptions(), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithBlock())

	if err != nil {
		if status, ok := status.FromError(err); ok {
			Logger(ctx).Debugf("Exit Status: %d", status.Message())
			Logger(ctx).Debugf("Exit Code: %d", status.Code())

		} else {
			Logger(ctx).Errorf("Failed to connect to worker in SetupWorker worker: %+v error %v", w.Uid, err)
		}
		return err
	}
	Logger(ctx).Debug("Connected to worker in worker setup")

	defer conn.Close()
	c := pb.NewCommandRunnerClient(conn)
	response, setupErr := c.Setup(ctx, &pb.TestOption{PublicKey: publicKey})
	if setupErr != nil {
		Logger(ctx).Error("error running setup on worker")
		return setupErr
	}
	Logger(ctx).Debugf("Response is %+v", response)
	return nil

}

func RecordSetup(ctx context.Context, s *api.Super) error {
	ctx, span := otel.Tracer(name).Start(ctx, "Record Setup")
	defer span.End()
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		return err
	}

	defer conn.Close()
	c := api.NewSupersClient(conn)
	in := api.SuperReq{Id: uint64(s.Id)}
	out, err := c.RecordSetup(ctx, &in)
	if err != nil {
		return err
	}

	Logger(ctx).Debug("Setup Recorded for ", out.Super.Id)
	return nil

}

const name = "brisk-shared"

func DeRegisterSuper(ctx context.Context, w *api.Super) error {
	ctx, span := otel.Tracer(name).Start(ctx, "Deregister Super")
	defer span.End()
	if IsDev() {
		Logger(ctx).Debug("Not de-registering super in dev mode")
		return nil
	}

	Logger(ctx).Infof("Deregistering super %+v", w)

	ctx, err := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("SUPER_DEREG_ROUTE_TOKEN"), ApiKey: os.Getenv("SUPER_DEREG_ROUTE_KEY")})
	if err != nil {
		panic(err)
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		return err
	}

	defer conn.Close()
	c := api.NewSupersClient(conn)
	in := api.SuperReq{Id: uint64(w.Id)}
	out, err := c.DeRegister(ctx, &in)
	if err != nil {
		return err
	}

	Logger(ctx).Debug("DeRegistered super with id ", out.Super.Id)
	return nil
}

func DeRegisterWorker(ctx context.Context, w *api.Worker) error {
	ctx, span := otel.Tracer(name).Start(ctx, "DeRegister Worker")
	defer span.End()
	if IsDev() {
		return nil
	}

	Logger(ctx).Debug("Deregistering worker ", w.Uid)

	ctx, err := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("WORKER_DEREG_ROUTE_TOKEN"), ApiKey: os.Getenv("WORKER_DEREG_ROUTE_KEY")})
	if err != nil {
		panic(err)
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))

	if err != nil {
		Logger(ctx).Errorf("Error connecting to api in DeRegisterWorker %v", err)
		return err
	}

	defer conn.Close()
	c := api.NewWorkersClient(conn)
	in := api.WorkerReq{Uid: w.Uid}
	out, err := c.DeRegister(ctx, &in)
	if err != nil {
		Logger(ctx).Errorf("Error deregistering %v : %v", w.Uid, err)
		return err
	}
	if out != nil && out.Worker != nil {
		Logger(ctx).Debugf("DeRegistered worker with id %v", out.Worker.Id)
	} else {
		Logger(ctx).Debugf("Tried to DeRegister worker with id %v but got no response", w.Uid)
	}
	return nil
}
func WaitForSignals(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	signal.Notify(c, signals...)
	Logger(ctx).Debugf("Waiting for signals %+v", signals)
	sig := <-c
	Logger(ctx).Debugf("\r Shutdown:- received signal %v", sig.String())

}

func RegisterHealthCheck(ctx context.Context, grpcServer grpc.ServiceRegistrar) {
	ctx, span := otel.Tracer(name).Start(ctx, "Register Health Check")
	defer span.End()
	Logger(ctx).Debug("Registering health check")
	healthService := NewHealthChecker()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	Logger(ctx).Debug("Registered health check")
}

func GetUniqueInstanceIDFromContext(ctx context.Context) string {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {

		return ""
	}

	instance_id := md.Get("unique_instance_id")
	if len(instance_id) == 0 {
		Logger(ctx).Error("no unique instance id found in context")
		return ""
	} else {
		Logger(ctx).Debugf("Found unique instance id %v", instance_id[0])
		return instance_id[0]
	}

}

func CheckGlobalInstanceID(ctx context.Context, globaleInstanceID string) error {
	Logger(ctx).Debug("Checking Global Instance ID")

	uniqueInstanceId := GetUniqueInstanceIDFromContext(ctx)
	if uniqueInstanceId == "" {
		return errors.New("unique instance id not found")
	}

	if uniqueInstanceId == globaleInstanceID {
		return nil
	} else {
		Logger(ctx).Errorf("The unique instance id is not the same as the global instance id %v != %v", uniqueInstanceId, globaleInstanceID)
		return errors.New("e5001: incorrect instance ID")
	}

}

func CheckGlobalProjectToken(ctx context.Context, globalProjectToken string) error {

	if len(globalProjectToken) > 0 {
		token, err := GetAuthenticatedProjectToken(ctx)
		if err != nil {
			return err
		}
		if globalProjectToken != token {
			Logger(ctx).Errorf("Project Mix Up : calling Setup for project token %v when global project token is %v", token, globalProjectToken)
			return errors.New("can't call function on  assigned supervisor")
		}

	} else {
		Logger(ctx).Debug("No global project token found")
	}
	return nil
}
func GetAuthenticatedProjectToken(ctx context.Context) (string, error) {

	if v := ctx.Value(ContextKey("project_token")); v != nil {
		Logger(ctx).Debugf("The context key for project is %v", v)
		if projectToken, ok := v.([]string); ok {
			return projectToken[0], nil
		}
		if projectToken, ok := v.(string); ok {
			return projectToken, nil
		}
	}
	Logger(ctx).Error("no project token found in context")
	return "", errors.New("no project token found in context")
}

func AddIntendedAllocIdTo(ctx context.Context, allocId string) context.Context {
	ctx = context.WithValue(ctx, ContextKey("worker_alloc_id"), allocId)
	ctx = metadata.AppendToOutgoingContext(ctx, "worker_alloc_id", allocId)
	return ctx
}
func GetIntendedAllocIdFrom(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md.Get("worker_alloc_id")[0], nil
	}
	Logger(ctx).Errorf("Cannot find the alloc id in context, context is %+v", ctx)
	return "", errors.New("wrong worker no alloc id provided") // if v := ctx.Value(ContextKey("worker_alloc_id")); v != nil {

}
func CheckIntendedAllocID(ctx context.Context) error {

	if IsDev() {
		return nil
	}

	alloc_id, err := GetIntendedAllocIdFrom(ctx)
	if err != nil {
		return err
	}
	if alloc_id != os.Getenv("NOMAD_ALLOC_ID") {
		Logger(ctx).Error("Incorrect alloc id %v for %v", alloc_id, os.Getenv("NOMAD_ALLOC_ID"))
		return errors.New("incorrect id for worker")
	}
	return nil
}

func connectAndGetProjectClient(ctx context.Context) (api.ProjectsClient, *grpc.ClientConn, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "connectAndGetProjectClient")
	defer span.End()
	Logger(ctx).Debug("Connecting to Project Service")
	defer Logger(ctx).Debug("Done Connecting to Project Service")
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var err error
	endpoint := os.Getenv("BRISK_API")
	var conn *grpc.ClientConn
	size := 1024 * 1024 * 1024 * 4

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
		grpc_retry.WithMax(3),
		grpc_retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			Logger(ctx).Errorf("retrying after error: %v attempt # %v", err, attempt)
		}),
	}
	conn, err = grpc.DialContext(ctx, endpoint,
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), grpc_retry.StreamClientInterceptor(opts...), BugsnagClientInterceptor()),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), grpc_retry.UnaryClientInterceptor(opts...), BugsnagClientUnaryInterceptor),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(size)), credentialOpts, grpc.WithBlock())

	if err != nil {
		Logger(ctx).Errorf("did not connect: %v", err)
		return nil, nil, err
	}
	Logger(ctx).Debug("Connected ")

	c := api.NewProjectsClient(conn)
	return c, conn, nil
}

func connectAndGetSplitterClient(ctx context.Context) (api.SplittingClient, *grpc.ClientConn, error) {
	ctx, span := otel.Tracer("brisk").Start(ctx, "connectAndGetSplitterClient")
	defer span.End()
	var err error
	endpoint := os.Getenv("BRISK_API")
	var conn *grpc.ClientConn
	var credentialOpts grpc.DialOption
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
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
		grpc_retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			Logger(ctx).Errorf("retrying after error: %v attempt # %v", err, attempt)
		}),
	}
	conn, err = grpc.DialContext(ctx, endpoint,
		grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), grpc_retry.StreamClientInterceptor(opts...), BugsnagClientInterceptor()),
		grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), grpc_retry.UnaryClientInterceptor(opts...), BugsnagClientUnaryInterceptor),
		grpc.WithDefaultCallOptions(), credentialOpts, grpc.WithBlock())

	if err != nil {
		Logger(ctx).Errorf("did not connect: %v", err)
		return nil, nil, err
	}
	Logger(ctx).Debug("Connected ")

	c := api.NewSplittingClient(conn)
	return c, conn, nil
}
func LogRun(ctx context.Context, ri *api.RunInfo, logger *BriskLogger, command *api.Command) error {
	ctx, span := otel.Tracer(name).Start(ctx, "LogRun")
	defer span.End()
	logger.Debugf("Logging run %v - %+v", ri, ri)
	client, conn, err := connectAndGetProjectClient(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	frr := api.LogRunReq{WorkerRunInfo: ri, Command: command}
	logRunResp, err := client.LogRun(ctx, &frr)
	if err != nil {
		logger.Errorf("Error logging run %v", err)
		return err
	}
	logger.Debugf("Logged run with response %+v", logRunResp)
	return nil
}

func FinishRun(ctx context.Context, jobrun int32, status api.JobRunStatus, exitCode int32, output string, errorString string, supervisorUid string, finalWorkerCount int32, failingWorkers []*api.Worker, logger *BriskLogger) error {
	ctx, span := otel.Tracer(name).Start(ctx, "FinishRun")
	defer span.End()
	logger.Debugf("Finishing run %v", jobrun)
	client, conn, err := connectAndGetProjectClient(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	frr := api.FinishRunRequest{SupervisorUid: supervisorUid, JobrunId: jobrun, Status: status, ExitCode: exitCode, Output: output, Error: errorString, FinalWorkerCount: finalWorkerCount, SyncFailedWorkers: failingWorkers}
	f, err := client.FinishRun(ctx, &frr)
	if err != nil {
		logger.Errorf("Error finishing run %v", err)
		return err
	}
	logger.Debugf("Finished run with response %+v", f)
	return nil
}

// TODO add the uid of the caller to this worker so we don't have overlap
func GetWorkersForProject(ctx context.Context, workersRequired int, workerImage string, rebuildFilePaths []string, repoInfo api.RepoInfo, projectPriority int32, logUid string) ([]*api.Worker, int32, string, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "GetWorkersForProject")
	defer span.End()
	Logger(ctx).Debugf("getWorkersForProject called with workersRequired %v and workerImage %v", workersRequired, workerImage)
	c, conn, err := connectAndGetProjectClient(ctx)
	if err != nil {
		return nil, -1, "", err
	}
	defer conn.Close()

	rebuildHash, hashErr := HashFiles(ctx, constants.DEFAULT_SERVER_HASH_FILE, os.Getenv("REMOTE_DIR"), rebuildFilePaths)
	if hashErr != nil {
		Logger(ctx).Errorf("Error getting hash for rebuild files %v", hashErr)
		return nil, -1, "", hashErr
	}

	baseTimeout := viper.GetDuration("GET_WORKER_TIMEOUT")
	timeout := baseTimeout * time.Duration(projectPriority) * 10

	in := api.GetWorkersReq{RebuildHash: rebuildHash, NumWorkers: uint64(workersRequired), WorkerImage: workerImage, SupervisorUid: SuperUID(ctx), RepoInfo: &repoInfo, LogUid: logUid}
	out, err := c.GetWorkersForProject(ctx, &in, grpc_retry.WithMax(5), grpc_retry.WithBackoff(grpc_retry.BackoffLinear(timeout)))
	if err != nil {
		Logger(ctx).Errorf("Error getting workers %v", err)
		if out != nil {
			return nil, int32(out.JobrunId), string(out.JobrunLink), err
		} else {
			return nil, -1, "", err
		}
	}
	for _, w := range out.Workers {
		Logger(ctx).Debugf("Got worker %+v", w)
		if w.State == "finished" {
			return nil, int32(out.JobrunId), out.JobrunLink, errors.New("Internal Error: invalid worker data #EW223487")
		}
	}

	return out.Workers, int32(out.JobrunId), out.JobrunLink, nil
}

// We use the token in the context
func GetProjectWithToken(ctx context.Context) (*api.Project, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "GetProjectWithToken")
	defer span.End()
	endpoint := os.Getenv("BRISK_API")

	var conn *grpc.ClientConn

	Logger(ctx).Debug("GetProjectWithToken")
	conn, err := ApiConn(ctx, endpoint)
	if err != nil {
		Logger(ctx).Errorf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	Logger(ctx).Debug("Connected ")
	c := api.NewProjectsClient(conn)
	in := api.GetProjectReq{}

	ctx, innerSpan := otel.Tracer(name).Start(ctx, "GetProjectWithToken:call")
	res, getErr := c.GetProject(ctx, &in, grpc_retry.WithMax(10))
	innerSpan.End()
	if getErr != nil {
		Logger(ctx).Errorf("Error getting project %v", getErr)
		return nil, getErr
	}

	return res.Project, nil
}

func RegisterWorker(ctx context.Context, port Port, ipAddress IpAddress, hostIp IpAddress, uid string, workerImage string, hostUid string, syncPort string) (*api.Worker, error) {

	// if IsDev() {
	// 	return &api.Worker{}, nil
	// }
	//hostIp := os.Getenv("HOST_IP")
	// lets the DB know we exist
	// this should use a public API
	// or come from Nomad or whoever starts up the services?
	if viper.GetBool("HONEYCOMB_VERBOSE") {
		var span trace.Span
		ctx, span = otel.Tracer(name).Start(ctx, "RegisterWorker")
		defer span.End()
	}
	ctx, err := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("WORKER_REG_ROUTE_TOKEN"), ApiKey: os.Getenv("WORKER_REG_ROUTE_KEY")})
	if err != nil {
		panic(err)
	}
	//ipAddress, err := GetOutboundIP(ctx)

	if err != nil {
		panic(err)
	}

	// opts := []grpc_retry.CallOption{
	// 	grpc_retry.WithBackoff(grpc_retry.BackoffExponential(10 * time.Millisecond)),
	// }
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return nil, err
	}

	defer conn.Close()
	c := api.NewWorkersClient(conn)
	in := api.WorkerRegReq{HostUid: hostUid, IpAddress: string(ipAddress), HostIp: string(hostIp), Port: string(port), Uid: uid, WorkerImage: workerImage, SyncPort: syncPort}
	Logger(ctx).Debugf("Registering worker with WorkerRegReq %+v", in)
	out, err := c.Register(ctx, &in, grpc_retry.WithMax(5))

	if err != nil {
		Logger(ctx).Errorf("Error registering worker %v", err)
		return nil, err
	}

	Logger(ctx).Debug("Registered Worker")
	return out.Worker, nil
}

func RegisterSuper(ctx context.Context) (*api.Super, error) {

	// lets the DB know we exist
	// this should use a public API
	// or come from Nomad or whoever starts up the services?
	ctx, span := otel.Tracer(name).Start(ctx, "RegisterSuper")
	defer span.End()
	ctx, err := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("SUPER_REG_ROUTE_TOKEN"), ApiKey: os.Getenv("SUPER_REG_ROUTE_KEY")})
	if err != nil {
		panic(err)
	}
	//ipAddress, err := GetOutboundIP(ctx)
	ipAddress := os.Getenv("NOMAD_IP_sync_port")

	if err != nil {
		return nil, err
	}

	// opts := []grpc_retry.CallOption{
	// 	grpc_retry.WithBackoff(grpc_retry.BackoffExponential(10 * time.Millisecond)),
	// }
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Error("Failing API Conn in super")
		return nil, err
	}

	defer conn.Close()
	c := api.NewSupersClient(conn)
	var syncPort string
	var port string
	if IsDev() {
		syncPort = "2222"
		port = os.Getenv("BRISK_SUPER_PORT")
	} else {
		syncPort = os.Getenv("NOMAD_HOST_PORT_sync_port")
		port = os.Getenv("NOMAD_HOST_PORT_super_port")
	}

	var externalEndpoint string
	var syncEndpoint string

	if IsDev() {

		if viper.GetBool("USE_DOCKER_COMPOSE") {
			ipAddress = "0.0.0.0"
			Logger(ctx).Info("Running on docker compose so using 0.0.0.0 for ip because we cannot connect to container on osx")
		} else {

			ip, err := GetOutboundIP(ctx)
			ipAddress = ip.String()
			if err != nil {
				return nil, err
			}
		}
		externalEndpoint = fmt.Sprintf("%s:%s", ipAddress, os.Getenv("BRISK_SUPER_PORT"))
		syncEndpoint = ipAddress

	} else {
		allocId := nomad.GetNomadAllocId()

		externalEndpoint = fmt.Sprintf("%s.%s:%s", allocId, os.Getenv("BRISK_DOMAIN_NAME"), os.Getenv("BRISK_SUPER_PORT"))
		syncEndpoint = ipAddress
	}
	uid := SuperUID(ctx)
	hostIp := os.Getenv("HOST_IP")
	hostUid := GetHostUID(ctx)

	in := api.SuperRegReq{HostUid: hostUid, IpAddress: ipAddress, HostIp: hostIp, Port: port, SyncPort: syncPort, ExternalEndpoint: externalEndpoint, SyncEndpoint: syncEndpoint, Uid: uid}
	Logger(ctx).Infof("Registering Super with %+v", in)
	out, err := c.Register(ctx, &in, grpc_retry.WithMax(2))
	if err != nil {
		Logger(ctx).Errorf("Error registering super %v", err)
		return nil, err
	}

	Logger(ctx).Debug("Registered Super")
	return out.Super, nil
}

var superUID string

func SuperUID(ctx context.Context) string {
	if viper.GetBool("USE_DOCKER_COMPOSE") {
		if superUID == "" {
			Logger(ctx).Info("Running on docker compose so using random id for super uid")
			superUID = uuid.New().String()
		}
		return superUID
	} else {
		return nomad.GetNomadAllocId()
	}
}

var workerUID string

func WorkerUID(ctx context.Context) string {
	if viper.GetBool("USE_DOCKER_COMPOSE") {
		if workerUID == "" {
			Logger(ctx).Info("Running on docker compose so using random id for super uid")
			workerUID = uuid.New().String()
		}
		return workerUID

	}
	return "" // not sure why we are using this here - we use an external service to do this in production
}

func DoAuth(ctx context.Context) (context.Context, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "DoAuth")
	defer span.End()
	// hit the API to log in the user
	// I don't actually do any auth here. All my calls should be authed though.
	// maybe I should do one at the begining ?

	Logger(ctx).Debugf("In the DoAuth the ctx is %+v", ctx)
	ctx, err := auth.PropagateCredentials(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth:  %v", err)
	}
	authCreds, err := auth.GetAuthCredsFromMd(ctx)
	Logger(ctx).Debug("The authcreds token we have is %+v", authCreds.ApiToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth:  %v", err)
	}
	ctx = context.WithValue(ctx, ContextKey("project_token"), authCreds.ProjectToken)

	project, authErr := GetProjectWithToken(ctx)
	if authErr != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth:  %v", authErr)
	}
	Logger(ctx).Debugf("Doing auth for project %+v", project)

	return ctx, authErr

}

func RegisterClient(ctx context.Context, node *nomadapi.Node) error {

	if viper.GetBool("HONEYCOMB_VERBOSE") {
		var span trace.Span
		ctx, span = otel.Tracer(name).Start(ctx, "RegisterClient")
		defer span.End()
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return err
	}
	defer conn.Close()
	c := api.NewInfraClient(conn)

	Logger(ctx).Debugf("Registering node %v", node.ID)
	ctx, authErr := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("INFRA_REG_ROUTE_TOKEN"), ApiKey: os.Getenv("INFRA_REG_ROUTE_KEY")})
	if authErr != nil {
		return authErr
	}

	// for key, value := range node.Attributes {
	// 	Logger(ctx).Debugf("Key: %v , Value: %v", key, value)
	// }
	// Logger(ctx).Debug("Printed the attributes")
	// hasCpu, ok := node.Attributes["cpu"]
	// if !ok {
	// 	Logger(ctx).Errorf("No cpu attribute")
	// }
	// Logger(ctx).Debugf("Has cpu is %+v", hasCpu)

	num_cores := node.Attributes["cpu.numcores"]
	Logger(ctx).Debugf("Num cores is %v", num_cores)

	num_cpus, err := strconv.Atoi(num_cores)
	if err != nil {
		Logger(ctx).Errorf("Could not parse cpus: %v", err)
		return err
	}

	hostUid := node.Attributes["unique.platform.aws.instance-id"]
	//convert this to be maybe strings back to the API and then do the conversions in ruby / the db cause now I'll need to send them over the wire as big ints
	m := api.Machine{
		Uid: hostUid, IpAddress: node.ID,
		HostIp:  node.HTTPAddr,
		HostUid: hostUid,
		Image:   node.Attributes["image"],
		Region:  node.Attributes["region"],
		Cpus:    uint32(num_cpus),
		Memory:  node.Attributes["memory.totalbytes"],
		Disk:    node.Attributes["unique.storage.bytestotal"],
		OsInfo:  node.Attributes["os.name"] + "-" + node.Attributes["os.version"] + "-" + node.Attributes["os.arch"],
	}
	in := api.MachineReq{Machine: &m}
	Logger(ctx).Debugf("Registering machine %+v", in)
	response, err := c.RegisterMachine(ctx, &in, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not register node: %v", err)
		return err
	}
	Logger(ctx).Debugf("Registered node %v", response)
	return err
}

// this is used when we run docker-compose to register the local workers and super
func GetHostUID(ctx context.Context) string {
	hostIp := os.Getenv("HOST_UID")
	if len(hostIp) == 0 {
		hostIp = "no_host_ip_set"
	}
	return hostIp
}

func RegisterLocalDockerClient(ctx context.Context) error {
	ipAddress, err := GetOutboundIP(ctx)
	if err != nil {
		return err
	}

	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return err
	}
	defer conn.Close()
	c := api.NewInfraClient(conn)

	Logger(ctx).Debugf("Registering node %v", ipAddress)

	hostUid := GetHostUID(ctx)
	hostIp := os.Getenv("HOST_IP")
	//convert this to be maybe strings back to the API and then do the conversions in ruby / the db cause now I'll need to send them over the wire as big ints
	m := api.Machine{
		Uid:       hostUid,
		IpAddress: ipAddress.String(),
		HostIp:    hostIp,
		HostUid:   hostUid,
		Image:     os.Getenv("WORKER_IMAGE"),
		Region:    "docker",
		Cpus:      2,
		Memory:    "9999999999999",
		Disk:      "9999999999999",
		OsInfo:    "docker",
	}
	in := api.MachineReq{Machine: &m}
	Logger(ctx).Debugf("Registering machine %+v", in)
	response, err := c.RegisterMachine(ctx, &in, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not register node: %v", err)
		return err
	}
	Logger(ctx).Debugf("Registered node %v", response)
	return err
}

func Registerk8Client(ctx context.Context) error {

	if viper.GetBool("HONEYCOMB_VERBOSE") {
		var span trace.Span
		ctx, span = otel.Tracer(name).Start(ctx, "RegisterClient")
		defer span.End()
	}

	// check all the env vars are set
	if os.Getenv("POD_NAME") == "" {
		Logger(ctx).Error("No pod name set")
		return errors.New("no pod name set")
	}
	if os.Getenv("POD_IP") == "" {
		Logger(ctx).Error("No pod ip set")
		return errors.New("no pod ip set")
	}

	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return err
	}
	defer conn.Close()
	c := api.NewInfraClient(conn)

	Logger(ctx).Debugf("Registering node %v")
	ctx, authErr := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("INFRA_REG_ROUTE_TOKEN"), ApiKey: os.Getenv("INFRA_REG_ROUTE_KEY")})
	if authErr != nil {
		return authErr
	}

	// for key, value := range node.Attributes {
	// 	Logger(ctx).Debugf("Key: %v , Value: %v", key, value)
	// }
	// Logger(ctx).Debug("Printed the attributes")
	// hasCpu, ok := node.Attributes["cpu"]
	// if !ok {
	// 	Logger(ctx).Errorf("No cpu attribute")
	// }
	// Logger(ctx).Debugf("Has cpu is %+v", hasCpu)

	num_cores := 2
	Logger(ctx).Debugf("Num cores is %v", num_cores)

	if err != nil {
		Logger(ctx).Errorf("Could not parse cpus: %v", err)
		return err
	}

	podIpAddress := os.Getenv("POD_IP")
	workerImage := os.Getenv("WORKER_IMAGE")

	//convert this to be maybe strings back to the API and then do the conversions in ruby / the db cause now I'll need to send them over the wire as big ints
	m := api.Machine{
		Uid:       podIpAddress,
		IpAddress: podIpAddress,
		HostIp:    podIpAddress,
		HostUid:   podIpAddress,
		Image:     workerImage,
		Region:    "k8s",
		Cpus:      uint32(num_cores),
		Memory:    "10000000000000000",
		Disk:      "10000000000000000",
		OsInfo:    "k8s",
	}
	in := api.MachineReq{Machine: &m}
	Logger(ctx).Debugf("Registering machine %+v", in)
	response, err := c.RegisterMachine(ctx, &in, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not register node: %v", err)
		return err
	}
	Logger(ctx).Debugf("Registered node %v", response)
	return err
}

func DeRegisterClient(ctx context.Context, client *nomadapi.Node) error {
	ctx, span := otel.Tracer(name).Start(ctx, "DeRegisterClient")
	defer span.End()
	Logger(ctx).Debugf("DeRegistering client %v", client.ID)
	ctx, authErr := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("INFRA_DE_REG_ROUTE_TOKEN"), ApiKey: os.Getenv("INFRA_DE_REG_ROUTE_KEY")})
	if authErr != nil {
		return authErr
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return err
	}
	defer conn.Close()
	c := api.NewInfraClient(conn)
	in := api.MachineReq{Machine: &api.Machine{Uid: client.Attributes["unique.platform.aws.instance-id"]}}
	response, err := c.DeRegisterMachine(ctx, &in, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not deregister client: %v", err)
		return err
	}
	Logger(ctx).Debugf("DeRegistered client %v", response)
	return err

}
func DrainClient(ctx context.Context, client *nomadapi.Node) error {
	ctx, span := otel.Tracer(name).Start(ctx, "DrainClient")
	defer span.End()
	Logger(ctx).Debugf("Draining client %v", client.ID)
	ctx, authErr := auth.AddAuthToCtx(ctx, auth.AuthCreds{ApiToken: os.Getenv("INFRA_DRAIN_ROUTE_TOKEN"), ApiKey: os.Getenv("INFRA_DRAIN_ROUTE_KEY")})
	if authErr != nil {
		return authErr
	}
	conn, err := ApiConn(ctx, os.Getenv("BRISK_API"))
	if err != nil {
		Logger(ctx).Errorf("Could not get connection: %v", err)
		return err
	}
	defer conn.Close()
	c := api.NewInfraClient(conn)
	in := api.MachineReq{Machine: &api.Machine{Uid: client.Attributes["unique.platform.aws.instance-id"]}}
	response, err := c.DrainMachine(ctx, &in, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not drain client: %v", err)
		return err
	}
	Logger(ctx).Debugf("Drained client %v", response)
	return err

}

//now split for some people with this

func SplitTests(ctx context.Context, numServers int32, files []string) ([][]string, string, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "SplitTests")
	defer span.End()

	c, conn, err := connectAndGetSplitterClient(ctx)
	if err != nil {
		return nil, "", err
	}
	defer conn.Close()

	if numServers > int32(len(files)) {
		Logger(ctx).Infof("Num servers is greater than num files, setting num servers to num files: %v", len(files))
		numServers = int32(len(files))
	}
	response, err := c.SplitForProject(ctx, &api.SplitRequest{NumBuckets: numServers, Filenames: files}, grpc_retry.WithMax(5))
	if err != nil {
		Logger(ctx).Errorf("Could not split tests: %v", err)
		return nil, "", err
	}
	Logger(ctx).Debugf("Split tests %v", response)
	for v := range response.FileLists {
		Logger(ctx).Debugf("Bucket %v has %v files", v, len(response.FileLists[v].Filenames))
	}
	return convertSplitResponseTo2DArray(response), response.SplitMethod, nil
}

func convertSplitResponseTo2DArray(response *api.SplitResponse) [][]string {
	var tests [][]string
	for _, bucket := range response.FileLists {
		var testBucket []string
		for _, test := range bucket.Filenames {
			testBucket = append(testBucket, test)
		}
		tests = append(tests, testBucket)
	}
	return tests
}
