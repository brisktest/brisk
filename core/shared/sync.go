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
	brisksupervisor "brisk-supervisor/brisk-supervisor"
	pb "brisk-supervisor/brisk-supervisor"
	. "brisk-supervisor/shared/logger"
	"brisk-supervisor/shared/types"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"os/exec"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//
// need to tighten this up
// Add in a username
// Implment Worker to Super key sharing (like with the super at the moment)
// Configure rrsync to only allow ssh

// we need to do an initial sync and then watch

func StartSync(ctx context.Context, config Config, localDirectory string, ssh_key string, destHost string, destPort string, destUserName string, outputChan chan *pb.Output) error {
	ctx, span := otel.Tracer(name).Start(ctx, "StartSync")
	defer span.End()

	Logger(ctx).Debugf("The destHost = %v", destHost)
	// execute a sync to remote host
	if len(destHost) == 0 {
		Logger(ctx).Error("We need a host to sync to - no sync host")
		return errors.New("we need a host to sync to - no sync host")
	}
	var keyFileName string
	if len(ssh_key) > 0 {
		Logger(ctx).Debugf("ssh_key is %s", ssh_key)

		keyFileName = "/tmp/.my_temp_key" + time.Now().Format("20060102150405")
		err := ioutil.WriteFile(keyFileName, []byte(ssh_key), 0600)
		if err != nil {
			Logger(ctx).Errorf("Error in StartSync: %v", err)
			outputChan <- &pb.Output{Response: "Error during sync - " + string(err.Error()), Stderr: err.Error(), Control: types.FINISHED, Created: timestamppb.Now()}

			return RSyncError
		}

	} else {
		Logger(ctx).Error("No ssh key provided")
		os.Exit(1)
	}

	sync_host := viper.GetString("SYNC_HOST")
	rsyncTimeout := strconv.Itoa((viper.GetInt("RSYNC_TIMEOUT")))
	var finalArgs []string
	var args []string
	args = append(args, "--timeout="+rsyncTimeout, "--stats", "-avz", "--delete")
	withExclude := addExclusions(args, config.ExcludedFromSync)
	if !strings.HasSuffix(localDirectory, "/") {
		localDirectory = localDirectory + "/"
	}

	if viper.GetBool("NO_BASTION") {

		destHostWithUserAndDir := fmt.Sprintf("%s@%s:%s", destUserName, destHost, "/tmp/remote_dir")
		Logger(ctx).Debugf("The key file name is %s", keyFileName)
		finalArgs = append(withExclude, "-e", fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o IdentitiesOnly=yes -i %s -p %s", keyFileName, destPort), localDirectory, destHostWithUserAndDir)

		if viper.GetBool("DEBUG_RSYNC") {
			fmt.Println("Final args are ", finalArgs)
		}

	} else {

		bastionHost := "bastion@" + sync_host
		destHostWithUserAndDir := fmt.Sprintf("%s@%s:%s", destUserName, destHost, "/")
		loadBalancerSyncPort := "2222" // this is correct cause it's the port that then hits traefik
		//loadBalancerSyncPort := destPort
		rsh := fmt.Sprintf("/usr/bin/ssh -i %s  -p %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o IdentitiesOnly=yes  -o \"ProxyCommand ssh -A -p %s   -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o IdentitiesOnly=yes -i %s  %s -W %%h:%%p\"", keyFileName, destPort, loadBalancerSyncPort, keyFileName, bastionHost)
		Logger(ctx).Debug(rsh)
		//cmd := exec.Command("rsync", "-avz", fmt.Sprintf("-e ssh -o StrictHostKeyChecking=no -i %s", keyFileName), source, destination)

		Logger(ctx).Debugf("rsyncTimeout is %v", rsyncTimeout)

		finalArgs := append(withExclude, "-e", rsh, localDirectory, destHostWithUserAndDir)
		Logger(ctx).Debugf("Final arges are ------------------------------- %v", finalArgs)
	}
	cmd := exec.Command("rsync", finalArgs...)

	Logger(ctx).Debug(cmd)
	//outputChan <- pb.Output{Response: rsh, Stdout: rsh}
	//outputChan <- pb.Output{Response: cmd.String(), Stdout: cmd.String()}

	output, err := cmd.CombinedOutput()

	if err != nil {
		Logger(ctx).Error(string(output))

		outputChan <- &pb.Output{Response: string(output), Stderr: string(output), Stdout: string(output), Control: types.FINISHED, Created: timestamppb.Now()}

		Logger(ctx).Error("Error with cmd ", cmd.String())
		Logger(ctx).Error("Error during sync - " + string(err.Error()))
		outputChan <- &pb.Output{Response: "Error during sync - " + string(err.Error()+string(output)), Stderr: "Error during sync - " + string(err.Error()+string(output)), Control: types.FINISHED, Created: timestamppb.Now()}
		return RSyncError
		//Logger(ctx).Panic(err)
	}
	Logger(ctx).Debug("returning from sync ")
	return err
}

func SuperToWorkerSync(ctx context.Context, destHost string, destSyncPort string, sourceFolder string, destFolder string, destUserName string, config *brisksupervisor.Config) error {
	ctx, span := otel.Tracer(name).Start(ctx, "SuperToWorkerSync")
	defer span.End()
	startTime := time.Now()
	defer func() { Logger(ctx).Debugf("SuperToWorkerSync for %v took %v", destHost, time.Since(startTime)) }()
	destination := fmt.Sprintf("%s@%s:%s", destUserName, destHost, destFolder)
	//keyFileName := "/home/brisk/.ssh/super_key"
	keyFileName := "/tmp/.my-private-key"

	withExclude := addExclusions(nil, config.ExcludedFromSync)
	Logger(ctx).Debugf("Sync to worker excludes are  %+v", withExclude)
	var args []string
	if !strings.HasSuffix(sourceFolder, "/") {
		sourceFolder = sourceFolder + "/"
	}

	rsyncTimeout := strconv.Itoa((viper.GetInt("RSYNC_TIMEOUT")))
	Logger(ctx).Debugf("rsyncTimeout is %v", rsyncTimeout)
	if viper.GetBool("DEBUG_RSYNC") {

		args = append(args, "-vv")
	}
	args = append(args, "--timeout="+rsyncTimeout, "--stats", "-avz", "--delete")
	args = append(args, withExclude...)
	if viper.GetBool("DEBUG_RSYNC") {
		Logger(ctx).Debug("Debugging ssh")
		args = append(args, fmt.Sprintf("-e ssh -vvv -p %s -o StrictHostKeyChecking=no -o IdentitiesOnly=yes -i %s", destSyncPort, keyFileName), sourceFolder, destination)
	} else {
		args = append(args, fmt.Sprintf("-e ssh -p %s -o StrictHostKeyChecking=no -o IdentitiesOnly=yes -i %s", destSyncPort, keyFileName), sourceFolder, destination)
	}
	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("SYNC_TO_WORKER_TIMEOUT"))
	defer cancel()
	cmd := exec.Command("rsync", args...)
	Logger(ctx).Debugf("Sync to worker command is %v", cmd)
	out, err := TermCombinedOutput(ctx, cmd)
	if err != nil {
		Logger(ctx).Errorf("error output from sync - %v", string(out))

		Logger(ctx).Errorf("Error syncing in SuperToWorkerSync %v", err)

		err = fmt.Errorf("error syncing to workers: %w. Command Output:  %v", err, string(out))
	}

	//Logger(ctx).Debug(string(output))
	return err
}

func ParallelSuperToWorkerSync(ctx context.Context, destHosts []*api.Worker, sourceFolder string, destFolder string, destUserName string) error {
	Logger(ctx).Debug("Using parallel-sync...")
	var hosts string
	for _, v := range destHosts {
		hosts += fmt.Sprintf(" -H %s@%s ", destUserName, v.IpAddress)
	}
	keyFileName := "/home/brisk/.ssh/super_key"
	cmd := exec.Command("parallel-rsync", "-avz", hosts, "-x --delete", fmt.Sprintf("-e ssh -o StrictHostKeyChecking=no -o IdentitiesOnly=yes -i %s", keyFileName), sourceFolder, destFolder)
	Logger(ctx).Debug(cmd)
	out, err := cmd.CombinedOutput()
	Logger(ctx).Debug(string(out))
	if err != nil {
		Logger(ctx).Debug(err)
	}

	return err
}

func addExclusions(command []string, exclusions []string) (newCommand []string) {
	newCommand = command
	for _, v := range exclusions {
		newCommand = append(newCommand, "--exclude", v)
	}
	return
}
