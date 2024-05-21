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
	. "brisk-supervisor/shared/logger"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func terminate(proc *os.Process) error {
	// proc.Signal(unix.SIGINT)
	syscall.Kill(-proc.Pid, syscall.SIGINT)

	Logger(context.Background()).Debugf("Sent SIGINT to process group -%d - waiting 5 seconds", proc.Pid)
	time.Sleep(5 * time.Second)

	Logger(context.Background()).Infof("Cleaning up processes - Sending SIGKILL to process group -%d", proc.Pid)

	syscall.Kill(-proc.Pid, syscall.SIGKILL)

	return nil
}

// Start is like calling Start on os/exec.CommandContext but uses
// SIGTERM on Unix-based systems.
func TermStart(ctx context.Context, c *exec.Cmd) (wait func() error, err error) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := c.Start(); err != nil {
		return nil, err
	}

	// Set the niceness to 10 so our tests don't choke out the server
	syscall.Setpriority(syscall.PRIO_PROCESS, c.Process.Pid, 10)

	waitDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			Logger(ctx).Info("Context done - terminating process")
			terminate(c.Process)
		case <-waitDone:
		}
	}()
	return func() error {
		defer close(waitDone)
		return c.Wait()
	}, nil
}

func TermRun(ctx context.Context, c *exec.Cmd) error {
	wait, err := TermStart(ctx, c)
	if err != nil {
		return err
	}
	return wait()
}

func TermCombinedOutput(ctx context.Context, c *exec.Cmd) ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := TermRun(ctx, c)
	return b.Bytes(), err
}
