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
	"brisk-supervisor/api"
	"brisk-supervisor/shared"
	"brisk-supervisor/shared/brisk_metrics"
	. "brisk-supervisor/shared/context"
	"brisk-supervisor/shared/env"
	"brisk-supervisor/shared/honeycomb"
	. "brisk-supervisor/shared/logger"
	"brisk-supervisor/shared/nomad"
	"brisk-supervisor/shared/types"
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "google.golang.org/grpc/encoding/gzip"

	"github.com/bugsnag/bugsnag-go"
	errors "github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "brisk-supervisor/brisk-supervisor"
	. "brisk-supervisor/shared"
	constants "brisk-supervisor/shared/constants"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	grpcotel "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/filters"
	"go.opentelemetry.io/otel"
)

type server struct {
	pb.UnimplementedCommandRunnerServer
}

var globalProjectToken string

func (s *server) CheckBuild(ctx context.Context, in *pb.CheckBuildMsg) (*pb.CheckBuildResp, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "CheckBuild")
	defer span.End()

	ctx = AddMetadataToCtx(ctx)

	Logger(ctx).Info("In CheckBuild")
	Logger(ctx).Debugf("CheckBuild: %v", in)

	defer func() {
		Logger(ctx).Debug("CheckBuild--")
	}()

	val, err := shared.RebuildRequired(ctx, constants.DEFAULT_SERVER_HASH_FILE, os.Getenv("REMOTE_DIR"), in.Config.RebuildFilePaths)
	Logger(ctx).Infof("the return value from RebuildRequired is %v", val)

	if err != nil {
		Logger(ctx).Errorf("Error checking build: %v", err)
		return &pb.CheckBuildResp{Success: false, Error: err.Error()}, err
	}
	Logger(ctx).Infof("CheckBuild: Success == %v", val)
	return &pb.CheckBuildResp{Success: val}, nil
}

// we check the global project token to make sure it's the same project
// we also check the intended alloc id to make sure it's the same alloc
// we write the public key from the super into the authorized_keys file for the server

func (s *server) Setup(ctx context.Context, testOption *pb.TestOption) (*pb.Response, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "Setup")
	defer span.End()
	ctx = AddMetadataToCtx(ctx)
	Logger(ctx).Debugf("Worker Setup++ %v with public key %v and test option : %+v", nomad.GetNomadAllocId(), testOption.PublicKey, testOption)
	defer func() { Logger(ctx).Debugf("Worker Setup-- %v", nomad.GetNomadAllocId()) }()
	allocErr := CheckIntendedAllocID(ctx)
	if allocErr != nil {
		return nil, allocErr
	}

	ptErr := CheckGlobalProjectToken(ctx, globalProjectToken)
	if ptErr != nil {
		return nil, ptErr
	}
	Logger(ctx).Debugf("Calling Setup for worker %+v", nomad.GetNomadAllocId())
	//adding in the restrictrions
	key := os.Getenv("RRSYNC_COMMAND") + " " + testOption.PublicKey
	sshAuthKeysFile := os.Getenv("PUBLIC_KEY_FILE")
	if len(sshAuthKeysFile) == 0 {
		Logger(ctx).Error("PUBLIC_KEY_FILE needs to be set")
		return nil, errors.New("we need PUBLIC_KEY_FILE to be set")
	}
	err := shared.WriteKeyToFile(ctx, []byte(key), sshAuthKeysFile)
	if err != nil {
		Logger(ctx).Error(err.Error())

		return nil, err

	}
	// read the file back to make sure it's the same
	keyFile, err := shared.ReadKeyFromFile(ctx, sshAuthKeysFile)
	if err != nil {
		Logger(ctx).Error(err.Error())
		return nil, err
	}

	Logger(ctx).Debugf("Setup: Success with public key file now equal to \n %v", string(keyFile))
	return &pb.Response{Value: true}, nil
}

type RunCommandsGate struct {
	mutex  sync.Mutex
	locked bool
}

var runCommandsGate = RunCommandsGate{}

// api.CommandRunner_RunCommandsServer
func (s *server) RunCommands(stream pb.CommandRunner_RunCommandsServer) error {

	ctx, cancel := context.WithDeadline(stream.Context(), time.Now().Add(viper.GetDuration("RUNNER_COMMAND_TIMEOUT")))
	defer DebugCancelCtx(ctx, cancel, "RunCommands finished so cancelling context")

	ctx, span := otel.Tracer(name).Start(ctx, "RunCommands")
	Logger(ctx).Debugf("Before runcommands lock")

	if runCommandsGate.locked {
		// this is just a sanity check to make sure we aren't running multiple commands at once in the workers
		Logger(ctx).Error("RunCommands is locked so returning - globalProjectToken is %v", globalProjectToken)
		return errors.New("concurrent access to worker")
	}

	runCommandsGate.mutex.Lock()
	defer runCommandsGate.mutex.Unlock()
	if runCommandsGate.locked {
		Logger(ctx).Error("RunCommands is locked so returning - step 2")
		return errors.New("concurrent access to worker step 2")
	}
	runCommandsGate.locked = true
	defer func() { runCommandsGate.locked = false }()
	Logger(ctx).Debugf("After runcommands lock")

	ctx = AddMetadataToCtx(ctx)
	defer span.End()
	defer bugsnag.AutoNotify(ctx)
	defer LogPanic(ctx)

	Logger(ctx).Debug("RunCommands++")
	defer func() { Logger(ctx).Debug("RunCommands--") }()

	allocErr := CheckIntendedAllocID(ctx)
	if allocErr != nil {
		Logger(ctx).Error(allocErr.Error())
		return allocErr
	}
	ptErr := CheckGlobalProjectToken(ctx, globalProjectToken)
	if ptErr != nil {
		Logger(ctx).Errorf("CheckGlobalProjectToken error %v", ptErr)
		return ptErr
	}
	ctx = WithNomadAllocId(ctx, nomad.GetSmallNomadAllocId())

	// 1000 is the max number of commands we can queue up
	commandChan := make(chan *api.Command, 1000)
	errChan := make(chan error, 1)
	finishChan := make(chan bool, 1)

	go executeCommandStream(commandChan, ctx, errChan, stream, finishChan)

	go readCommands(ctx, stream, errChan, commandChan)

	select {

	case <-finishChan:
		{
			Logger(ctx).Debug("RunCommands is finished - everything looks good")
			Logger(ctx).Debug("Returning from RunCommands")
			return nil
		}

	case err := <-errChan:
		{
			Logger(ctx).Errorf("RunCommands error %v", err)

			return err
		}
	case <-ctx.Done():
		{
			Logger(ctx).Debug("RunCommands Context cancelled - probably the other end of the stream")
			return nil
		}
	}

}

func executeCommandStream(commandChan chan *api.Command, ctx context.Context, errChan chan error, stream pb.CommandRunner_RunCommandsServer, finishChan chan bool) bool {
	for in := range commandChan {
		ctx, forSpan := otel.Tracer(name).Start(ctx, "RunCommands-Recv-loop")
		defer forSpan.End()

		Logger(ctx).Debugf("The params from the super are %+v", in)
		Logger(ctx).Debugf("The command incoming is %+v", in)

		Logger(ctx).Debugf("The environment being passed in is %+v", in.Environment)

		commandNumber := in.SequenceNumber
		Logger(ctx).Debugf("Command sequence number is %v", commandNumber)

		var resultsFilename = fmt.Sprintf("/tmp/test-output-test-run-%v", time.Now().Unix())
		command, commandErr := setCommand(in, ctx, resultsFilename, errChan)

		if commandErr != nil {
			Logger(ctx).Errorf("Error setting command %v", commandErr)
			return false
		}

		Logger(ctx).Debugf("the Command we are about to run is %v", command)

		cmd := createCmd(ctx, command, in)

		streamErr := stream.Send(&pb.Output{Response: fmt.Sprintf("Running command %v", cmd.String()), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
		if streamErr != nil {
			Logger(ctx).Errorf("Error sending to stream %v", streamErr)
			errChan <- streamErr
			return false
		}
		Logger(ctx).Debugf("Running command %v", cmd.String())

		stderrRC, errStd := cmd.StderrPipe()
		if errStd != nil {
			streamErr := stream.Send(&pb.Output{ExecutionInfo: &api.ExecutionInfo{ExitCode: int32(cmd.ProcessState.ExitCode())}, Control: types.FAILED, Response: errStd.Error(), Command: in, Stderr: errStd.Error(), Exitcode: int32(cmd.ProcessState.ExitCode()), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
			if streamErr != nil {
				Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
				errChan <- streamErr
			}
			Logger(ctx).Errorf("RunCommands Got an Error %v", errStd)
			errChan <- errStd
			return true
		}
		stderr := bufio.NewReaderSize(stderrRC, 10*1024)

		var wg sync.WaitGroup

		wg.Add(2)

		go scanStdErr(ctx, &wg, stderr, in, commandNumber, stream)

		var lastStdout string

		stdoutRC, err1 := cmd.StdoutPipe()

		if err1 != nil {
			streamErr := stream.Send(&pb.Output{ExecutionInfo: &api.ExecutionInfo{ExitCode: int32(cmd.ProcessState.ExitCode())}, Control: types.FAILED, Response: err1.Error(), Command: in, Stderr: err1.Error(), Exitcode: int32(cmd.ProcessState.ExitCode()), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
			if streamErr != nil {
				Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
				errChan <- streamErr
			}
			Logger(ctx).Errorf("RunCommands got an Error %v", err1)
			errChan <- err1
			lastStdout = ""
			return false
		}
		stdout := bufio.NewReaderSize(stdoutRC, 10*1024*1024)
		stdOutScanner := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		stdOutScanner.Buffer(buf, 10*1024*1024)
		stdOutScanner.Split(bufio.ScanLines)
		go func() {
			defer bugsnag.AutoNotify(ctx)
			defer wg.Done()
			for stdOutScanner.Scan() {

				m := stdOutScanner.Text()
				Logger(ctx).Debugf("StdOut: %v", m)
				lastStdout = m
				rs := pb.Output{Response: string(m), Stdout: string(m), Command: in, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()}
				Logger(ctx).Debug("Waiting to send to the stream in stdout scanner")
				streamErr := stream.Send(&rs)
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
					// it's possible this might be recoverable but lets barf
					errChan <- streamErr
					return
				}
				Logger(ctx).Debug("Sent to the stream in stdout scanner")
			}

			stdoutErr := stdOutScanner.Err()
			if stdoutErr != nil {
				Logger(ctx).Debugf("Scanner error is %v", stdoutErr.Error())
				m := "error scanning stdout output:-" + stdoutErr.Error()
				rs := pb.Output{Response: string(m), Stdout: string(m), Command: in, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()}
				Logger(ctx).Debug("Waiting to send error to the stream in stdout scanner")
				streamErr := stream.Send(&rs)
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
				}
				Logger(ctx).Debug("Sent error to the stream in stdout scanner")
				errChan <- stdoutErr

			}
			Logger(ctx).Debugf("Stdout is done")

		}()

		Logger(ctx).Debug("Before start")
		startTime := time.Now()
		cmdWait, startErr := TermStart(ctx, cmd)

		if startErr != nil {
			Logger(ctx).Debugf("The start error is %v", startErr)
			Logger(ctx).Error(cmd.String())
			out, err := cmd.CombinedOutput()
			if err != nil {
				Logger(ctx).Debugf("Error getting combined output %v", err)
			} else {
				Logger(ctx).Debugf("Combined output is %v", out)
			}

			errorResponse := pb.Output{ExecutionInfo: &api.ExecutionInfo{ExitCode: int32(cmd.ProcessState.ExitCode())}, Control: types.FAILED, Response: startErr.Error(), Command: in, Stderr: startErr.Error(), Exitcode: int32(cmd.ProcessState.ExitCode()), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()}
			streamErr := stream.Send(&errorResponse)
			if streamErr != nil {
				Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)

			}
			errChan <- startErr
			return true
		}

		if in.Background {
			Logger(ctx).Debug("Not waiting because command is a background job")

			go func() {
				cmdWait()

			}()

			Logger(ctx).Debug("About to respond for the background job")
			hash, err := getRebuildHash(ctx)
			if err != nil {
				Logger(ctx).Errorf("Error getting rebuild hash %v", err)
				errChan <- err
				return true
			}
			Logger(ctx).Debug("Got the rebuild hash - %v", hash)

			streamErr := stream.Send(&pb.Output{ExecutionInfo: &api.ExecutionInfo{Command: in, RebuildHash: hash, Started: timestamppb.New(startTime), Finished: timestamppb.New(time.Now())}, Control: types.FINISHED, Response: "Not waiting for background job", Stdout: "Not waiting for background job", Command: in, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
			if streamErr != nil {
				Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
				errChan <- streamErr
			}
		} else {
			Logger(ctx).Debug("Not a background job so going to wait")
			wg.Wait()
			Logger(ctx).Debug("wait group is finished - now we are waiting on the cmd to finish")
			err := cmdWait()
			endTime := time.Now()
			hash, hashErr := getRebuildHash(ctx)
			if hashErr != nil {
				Logger(ctx).Errorf("Error getting rebuild hash %v", hashErr)
				errChan <- hashErr
				return true
			}
			Logger(ctx).Debug("Got the rebuild hash - %v", hash)

			executionInfo := api.ExecutionInfo{Command: in, RebuildHash: hash, Started: timestamppb.New(startTime), Finished: timestamppb.New(endTime)}
			Logger(ctx).Debugf("The error after cmd.Wait is %v", err)
			if err != nil {
				if in.IsTestRun && !in.NoTestFiles && in.TestFramework != string(types.Cypress) {
					Logger(ctx).Debug("error and checking to see if we have a file")

					results, readErr := os.ReadFile(resultsFilename)
					if readErr != nil {
						Logger(ctx).Debug("Can't read a file after error")
					} else {
						if len(os.Getenv("SUPRESS_FILE_CONTENTS")) == 0 {
							streamErr := stream.Send(&pb.Output{Response: "Output file contents:", Stdout: "Output file contents:", Command: in, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
							if streamErr != nil {
								Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
								errChan <- streamErr
							}

							streamErr = stream.Send(&pb.Output{Response: string(results), Stdout: string(results), Command: in, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
							if streamErr != nil {
								Logger(ctx).Errorf("executeCommandStream Got an Error %v sending to stream", streamErr)
								errChan <- streamErr
							}

						}
					}
				}
			}
			if err != nil {
				Logger(ctx).Errorf("RunCommands sending back exit error", err)

				streamErr := stream.Send(&pb.Output{Command: in, Response: err.Error(), Stderr: err.Error(), CmdSqNum: commandNumber, Created: timestamppb.Now()})
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
					errChan <- streamErr
				}
				if exiterr, ok := err.(*exec.ExitError); ok {
					Logger(ctx).Debugf("More error is  %+v", exiterr.Stderr)

					streamErr := stream.Send(&pb.Output{Command: in, Response: exiterr.Error(), Stderr: exiterr.Error(), CmdSqNum: commandNumber, Created: timestamppb.Now()})
					if streamErr != nil {
						Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
						errChan <- streamErr
					}
				}
				cmdErrResponse := pb.Output{Control: types.FAILED, Command: in, Response: "command failed :- " + err.Error(), Stderr: err.Error(), Exitcode: int32(cmd.ProcessState.ExitCode()), CmdSqNum: commandNumber, Created: timestamppb.Now()}
				streamErr = stream.Send(&cmdErrResponse)
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
				}
				errChan <- err
				return true
			}
			Logger(ctx).Debug("finished Waiting")
			executionInfo.ExitCode = int32(cmd.ProcessState.ExitCode())
			var cmdState string
			if executionInfo.ExitCode == 0 {
				cmdState = types.FINISHED
			} else {
				cmdState = types.FAILED
			}
			Logger(ctx).Debugf("Finishing worker exit code is %v and exit status is %v ", executionInfo.ExitCode, cmdState)
			Logger(ctx).Debugf("The process state is %v\n", cmd.ProcessState)
			Logger(ctx).Debugf("the string is %v", cmd.String())

			if in.IsTestRun && !in.NoTestFiles && (in.TestFramework != string(types.Cypress) && in.TestFramework != string(types.Rails) && in.TestFramework != string(types.Python) && in.TestFramework != string(types.Raw)) {
				Logger(ctx).Debug("go check to see if the file is there")

				results, readErr := os.ReadFile(resultsFilename)
				if readErr != nil {
					mesg := fmt.Sprintf("Error reading test results from %v : %v", resultsFilename, readErr)
					Logger(ctx).Debug(mesg)
					streamErr := stream.Send(&pb.Output{ExecutionInfo: &executionInfo, Stdout: mesg, Stderr: mesg, Command: in, Control: cmdState, Exitcode: 255, CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
					if streamErr != nil {
						Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
						errChan <- streamErr
					}
				} else {
					Logger(ctx).Debugf("Stream Send: %v We are sending back the results from the worker with executionInfo %v with rebuild hash %v", in.Commandline, executionInfo, hash)

					streamErr := stream.Send(&pb.Output{ExecutionInfo: &executionInfo, Stdout: "Finished running command", Command: in, Control: cmdState, Exitcode: int32(cmd.ProcessState.ExitCode()), JsonResults: string(results), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
					if streamErr != nil {
						Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
						errChan <- streamErr
					}

				}

			} else if in.IsListTest {
				Logger(ctx).Debugf("We are finished running the list test so the JSON results we are returning from StdOut are %v", string(lastStdout))
				streamErr := stream.Send(&pb.Output{ExecutionInfo: &executionInfo, Stdout: "Finished running command", Command: in, Control: cmdState, Exitcode: int32(cmd.ProcessState.ExitCode()), JsonResults: string(lastStdout), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
					errChan <- streamErr
				}

			} else {
				Logger(ctx).Debug("Not a test run so not fetching json file")
				streamErr := stream.Send(&pb.Output{ExecutionInfo: &executionInfo, Stdout: "Finished running command", Command: in, Control: cmdState, Exitcode: int32(cmd.ProcessState.ExitCode()), CmdSqNum: commandNumber, Worker: &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()}, Stage: in.Stage, Created: timestamppb.Now()})
				if streamErr != nil {
					Logger(ctx).Errorf("executeCommandStream Got an Error %v", streamErr)
					errChan <- streamErr
				}

			}
			Logger(ctx).Infof("Command finished %#v", in)

			if cmd.ProcessState.ExitCode() != 0 {
				Logger(ctx).Error("We have a non negative Exit code so we are returning from the command runner %+v", cmd.ProcessState)
				errChan <- err
				return true
			}

			if in.LastCommand {

				Logger(ctx).Debug("We are a test run so we are returning from the command runner")

				finishChan <- true
				return true
			} else {
				Logger(ctx).Debug("We are not the last command so we are not returning from the command runner")
			}
		}
		Logger(ctx).Debug("Finishing loop going back to recv")
		forSpan.End()
	}
	return false
}

// https://pkg.go.dev/gg-scm.io/tool/internal/sigterm
func createCmd(ctx context.Context, command string, in *api.Command) (cmd *exec.Cmd) {
	if os.Getenv("INTERACTIVE_SHELL") == "true" {
		cmd = exec.Command("bash", "-c", "-i", "-l", command)
	} else {
		cmd = exec.Command("bash", "-c", "-l", command)
	}

	cmd.Dir = in.WorkDirectory
	env := os.Environ()

	env = append(env, mapHash(in.Environment)...)

	env = append(env, "BRISK_NODE_INDEX="+strconv.Itoa(int(in.WorkerNumber)))
	env = append(env, "BRISK_NODE_TOTAL="+strconv.Itoa(int(in.TotalWorkerCount)))
	if len(viper.GetString("HTTP_PROXY")) > 0 {
		Logger(ctx).Warnf("Using http_proxy for command with value %v", viper.GetString("HTTP_PROXY"))
		env = append([]string{"http_proxy=" + viper.GetString("HTTP_PROXY")}, env...)
	}
	Logger(ctx).Debugf("the environment we are passing to the command is %+v", env)

	cmd.Env = env
	return cmd
}

func setCommand(in *api.Command, ctx context.Context, resultsFilename string, errChan chan error) (string, error) {

	var command string
	if in.NoTestFiles {
		Logger(ctx).Debug("NoTestFiles is true so we are just running things without an output file")
		command = in.Commandline + " " + strings.Join(in.Args, " ")
		return command, nil
	}

	// otherwise

	if in.IsTestRun {

		switch in.TestFramework {
		case string(types.Jest):

			Logger(ctx).Info("TestFramework says this is a Jest run")

			command = in.Commandline + " --forceExit" + " --outputFile=" + resultsFilename + " " + strings.Join(in.Args, " ")
		case string(types.Cypress):
			Logger(ctx).Info("TestFramework says this is a Cypress run")

			command = in.Commandline + " --spec " + strings.Join(in.Args, ",")
		case string(types.Rails):
			Logger(ctx).Info("TestFramework says this is a Rails run")

			command = in.Commandline + " " + strings.Join(in.Args, " ")

		case string(types.Rspec):
			command = in.Commandline + " -o " + resultsFilename + " " + strings.Join(in.Args, " ")
		case string(types.Python):
			Logger(ctx).Info("TestFramework says this is a Python run")
			command = in.Commandline
		case string(types.Raw):
			Logger(ctx).Info("TestFramework says this is a Raw run")
			command = in.Commandline
		default:
			Logger(ctx).Error("Framework not recognized")

			Logger(ctx).Errorf("RunCommands framework \"%v\" not recognized", in.TestFramework)
			errChan <- errors.Errorf("framework \"%v\" not recognized", in.TestFramework)
			return "", errors.Errorf("framework \"%v\" not recognized", in.TestFramework)
		}
	} else if in.IsListTest {
		Logger(ctx).Debugf("IsListTest is true so we are just running things without an output file")
		command = in.Commandline + " " + strings.Join(in.Args, " ")
	} else {
		command = in.Commandline + " " + strings.Join(in.Args, " ")

	}

	return command, nil
}

func scanStdErr(ctx context.Context, wg *sync.WaitGroup, stderr *bufio.Reader, in *api.Command, commandNumber int32, stream pb.CommandRunner_RunCommandsServer) {

	defer bugsnag.AutoNotify(ctx)

	defer wg.Done()
	scanner := bufio.NewScanner(stderr)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {

		m := scanner.Text()
		Logger(ctx).Debugf("StdErr: %v", m)
		rs := pb.Output{
			Response:       string(m),
			Stdout:         "",
			Stderr:         string(m),
			Control:        "",
			Exitcode:       0,
			TotalTestCount: 0,
			TotalTestFail:  0,
			TotalTestPass:  0,
			TotalTestSkip:  0,
			TotalTestError: 0,
			JsonResults:    "",
			BriskError:     &pb.BriskError{},
			CmdSqNum:       commandNumber,
			Worker:         &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()},
			Stage:          in.Stage,
			Command:        in,
			Created:        timestamppb.Now(),
			ExecutionInfo:  &api.ExecutionInfo{},
		}
		Logger(ctx).Debug("Waiting to send to the stream in stderr scanner")
		streamError := stream.Send(&rs)
		if streamError != nil {
			Logger(ctx).Errorf("Error sending to the stream %v", streamError)
			// it's possible this might be recoverable but lets barf

			return
		}
		Logger(ctx).Debug("Sent to the stream in stderr scanner")

	}
	stderrErr := scanner.Err()
	if stderrErr != nil {
		sendError(ctx, stderrErr, commandNumber, in, stream)
	}
	Logger(ctx).Debugf("Stderr is done")

}

func sendError(ctx context.Context, stderrErr error, commandNumber int32, in *api.Command, stream pb.CommandRunner_RunCommandsServer) {
	Logger(ctx).Errorf("Scanner error is %v", stderrErr.Error())

	m := "error scanning stderr output:-" + stderrErr.Error()

	rs := pb.Output{
		Response:       string(m),
		Stdout:         "",
		Stderr:         string(m),
		Control:        "",
		Exitcode:       0,
		TotalTestCount: 0,
		TotalTestFail:  0,
		TotalTestPass:  0,
		TotalTestSkip:  0,
		TotalTestError: 0,
		JsonResults:    "",
		BriskError:     &pb.BriskError{},
		CmdSqNum:       commandNumber,
		Worker:         &pb.Worker{Number: in.WorkerNumber, Uid: nomad.GetSmallNomadAllocId()},
		Stage:          in.Stage,
		Command:        in,
		Created:        timestamppb.Now(),
		ExecutionInfo:  &api.ExecutionInfo{},
	}
	Logger(ctx).Debug("Waiting to send error to the stream in stderr scanner")
	streamErr := stream.Send(&rs)
	if streamErr != nil {
		Logger(ctx).Errorf("Error sending to the stream %v", streamErr)
		// it's possible this might be recoverable but lets barf
		return
	}
	Logger(ctx).Debug("Sent error to the stream in stderr scanner")
	Logger(ctx).Errorf("RunCommands Error scanning stderr %v", stderrErr)
	panic(stderrErr)
}

func readCommands(ctx context.Context, stream pb.CommandRunner_RunCommandsServer, errChan chan error, commandChan chan *api.Command) {

	for {
		ctx, forSpan := otel.Tracer(name).Start(ctx, "RunCommands-readCommands")

		Logger(ctx).Debug("Waiting at recv")
		in, err := stream.Recv()
		Logger(ctx).Debug("After recv")

		if ctx.Err() != nil {
			Logger(ctx).Errorf("RunCommands context error %v if this is cancelled it is expected", ctx.Err())
			// I think this means we have been cancelled
			return
		}

		if err == io.EOF {
			Logger(ctx).Debugf("Got EOF so this transmission is over %v, %+v", err, in)
			errChan <- nil
			return
		}
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				Logger(ctx).Errorf("RunCommands Not a status %v", err)
				errChan <- err
			}
			Logger(ctx).Errorf("RunCommands status %v", st)
			errChan <- err
		}
		Logger(ctx).Debugf("RunCommands: %v", in)
		commandChan <- in
		forSpan.End()
		if err != nil {
			close(commandChan)
			return
		}
	}

}

const name = "worker"

func init() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: os.Getenv("BUGSNAG_API_KEY"),
		// The import paths for the Go packages containing your source files
		ProjectPackages: []string{"main", "brisk-supervisor/"},
		AppType:         "worker",
		ReleaseStage:    os.Getenv("RELEASE_STAGE"),
	})
}
func main() {

	ctx := context.Background()
	env.InitServerViper(ctx)

	cleanup := honeycomb.InitTracer()
	defer cleanup()
	ctx, span := otel.Tracer(name).Start(ctx, "Main")
	defer span.End()

	brisk_metrics.StartPrometheusServer(ctx)

	go func() {
		OutputRuntimeStats(ctx)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	gracefulStop, err := listenOn(ctx, string(types.WORKER_PORT))
	if err != nil {
		Logger(ctx).Errorf("Error listening %v: %v", types.WORKER_PORT, err)

	} else {

		//not recommended for multi-tenant environments
		if viper.GetBool("SELF_REGISTER") {
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
			workerUid := WorkerUID(ctx)

			w, err := shared.RegisterWorker(ctx, myPort(ctx), myIPAddress(ctx), hostIPAddress(ctx), workerUid, getWorkerImage(ctx), GetHostUID(ctx), "2222")
			if err != nil {
				Logger(ctx).Fatalf("Error registering worker %v", err)
			} else {
				Logger(ctx).Infof("Registered worker %v", w)
			}
		}

		health_check_port := os.Getenv("HEALTH_CHECK_PORT")
		healthCheckGracefulStop, err := listenForHealthCheck(ctx, health_check_port)
		if err != nil {
			Logger(ctx).Errorf("Error listening for health check on %v: %v", health_check_port, err)
		} else {

			WaitForSignals(ctx)
			Logger(ctx).Info("Shutdown: Shutting down gracefully")
			gtime := time.Now()
			healthCheckGracefulStop()
			gracefulStop()
			Logger(ctx).Infof("Shutdown: Graceful shutdown complete in ", time.Since(gtime).String())
		}
	}

}

func listenForHealthCheck(ctx context.Context, port string) (func(), error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		Logger(ctx).Errorf("failed to listen: %v", err)
		return nil, err
	}
	s := grpc.NewServer()
	RegisterHealthCheck(ctx, s)

	go func() {
		if err := s.Serve(lis); err != nil {
			Logger(ctx).Errorf("failed to serve for healthcheck: %v", err)
		}
	}()
	return s.Stop, nil
}

func listenOn(ctx context.Context, port string) (func(), error) {

	serverOpts := []grpc.ServerOption{}
	if !IsDev() {
		tlsCredentials, err := LoadTLSCredentials()
		if err != nil {
			Logger(ctx).Errorf("cannot load TLS credentials: %v ", err)
			shared.SafeExit(err)
		} else {
			serverOpts = append(serverOpts, grpc.Creds(tlsCredentials))

		}
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		Logger(ctx).Errorf("failed to listen: %v", err)
		return nil, err
	}

	var kaep = keepalive.EnforcementPolicy{
		MinTime:             1 * time.Second, // If a client pings more than once every second, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute, // If a client is idle for 15 minutes, send a GOAWAY
		MaxConnectionAge:      30 * time.Minute, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  2 * time.Second,  // Ping the client if it is idle for 2 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	}
	serverOpts = append(serverOpts, grpc.KeepaliveEnforcementPolicy(kaep))
	serverOpts = append(serverOpts, grpc.KeepaliveParams(kasp))
	serverOpts = append(serverOpts,
		grpc.ChainUnaryInterceptor(grpcotel.UnaryServerInterceptor(otelgrpc.WithInterceptorFilter(
			filters.Not(
				filters.HealthCheck(),
			),
		),
		), grpc_auth.UnaryServerInterceptor(DoAuth)))

	serverOpts = append(serverOpts,
		grpc.ChainStreamInterceptor(
			grpcotel.StreamServerInterceptor(otelgrpc.WithInterceptorFilter(
				filters.Not(
					filters.HealthCheck(),
				),
			),
			),
			grpc_recovery.StreamServerInterceptor(),
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(DoAuth),
		))

	s := grpc.NewServer(
		serverOpts...,
	)

	pb.RegisterCommandRunnerServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			Logger(ctx).Errorf("failed to serve: %v", err)
			shared.SafeExit(err)
		}
	}()
	return s.GracefulStop, err
}

// use this to get ENV=value for environment
func mapHash(hash map[string]string) []string {
	var out []string
	for k, v := range hash {
		out = append(out, fmt.Sprintf("%v=%v", k, v))
	}

	return out
}

func getRebuildHash(ctx context.Context) (string, error) {
	// get the hash of the rebuild
	hash, err := shared.ReadRebuildHash(ctx, constants.DEFAULT_SERVER_HASH_FILE)
	if err != nil {
		Logger(ctx).Errorf("Error reading rebuild hash %v", err)
	}
	return hash, err
}

func myIPAddress(ctx context.Context) types.IpAddress {
	myIPAddress := os.Getenv("POD_IP")
	if len(myIPAddress) == 0 {
		ip, err := GetOutboundIP(ctx)
		if err != nil {
			Logger(ctx).Errorf("Error getting outbound IP %v", err)
			shared.SafeExit(err)
		}
		return types.IpAddress(ip.String())
	}
	return types.IpAddress(myIPAddress)
}

// on k8s this is the Port
func myPort(ctx context.Context) types.Port {
	return "50051"
}

func hostIPAddress(ctx context.Context) types.IpAddress {
	hostIPAddress := os.Getenv("HOST_IP")
	if len(hostIPAddress) == 0 {
		Logger(ctx).Warn("It is helpful to set the HOST_IP environment variable to the host IP address to reduce contention on hosts - setting to a random string")
		hostIPAddress = uuid.New().String()
	}
	return types.IpAddress(hostIPAddress)
}

func getMyHostname(ctx context.Context) string {
	//k8s sets this hostname env to the name of the pod
	// in nomad I guess it's alloc id
	hostname := os.Getenv("HOSTNAME")
	if len(hostname) == 0 {
		Logger(ctx).Error("HOSTNAME needs to be set")
		shared.SafeExit(errors.New("HOSTNAME needs to be set"))
	}
	return hostname
}

func getWorkerImage(ctx context.Context) string {
	workerImage := os.Getenv("WORKER_IMAGE")
	if len(workerImage) == 0 {
		Logger(ctx).Error("WORKER_IMAGE needs to be set")
		shared.SafeExit(errors.New("WORKER_IMAGE needs to be set"))
	}
	return workerImage
}
