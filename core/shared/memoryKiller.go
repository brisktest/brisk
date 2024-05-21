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
	"context"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// this library is used to kill processes that are using too much memory
// we have an ENV variable that is used by nomad to OOM kill
// we want to send a signal to the process to gracefully shutdown before that happens
// we would also like to send a message to the user that the process is being killed
// log it
// and log the memory usage so we can see why we are hitting the upper bounds

// gets the current memory usage of the process
func getCurrentMemoryUsage(ctx context.Context) uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	Logger(ctx).Debugf("Alloc = %v MiB", bToMb(m.Alloc))
	return bToMb(m.Alloc)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func MemoryKiller(ctx context.Context, memoryLimit int64, pid int) {
	if memoryLimit == 0 {
		Logger(ctx).Warn("Memory limit is 0 - not checking memory usage")
		return
	}
	if memoryUsageOver(ctx, memoryLimit) {
		Logger(ctx).Errorf("Memory usage is over the limit sending SIGTERM to %d", pid)
		syscall.Kill(pid, syscall.SIGTERM)
		Logger(ctx).Info("Sent SIGTERM to process - now returning let the main take it from here")
	}

}

func memoryUsageOver(ctx context.Context, memoryLimit int64) bool {
	sleepTime := viper.GetDuration("MEMORY_CHECK_INTERVAL")
	for {
		usage := getCurrentMemoryUsage(ctx)
		if usage > uint64(float64(memoryLimit)*0.8) {
			Logger(ctx).Errorf("Memory usage (%d) is over the limit of 80%% of  %d ", usage, memoryLimit)
			PrintMemoryUsage(ctx)
			return true
		}

		if viper.GetBool("DEBUG_MEMORY_USAGE") {
			PrintMemoryUsage(ctx)
		}

		time.Sleep(sleepTime)
	}

}

func PrintMemoryUsage(ctx context.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	Logger(ctx).Infof("MemStats: Alloc = %v MiB", bToMb(m.Alloc))
	Logger(ctx).Infof("MemStats:  TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	Logger(ctx).Infof("MemStats:  Sys = %v MiB", bToMb(m.Sys))
	Logger(ctx).Infof("MemStats:  NumGC = %v", m.NumGC)
	Logger(ctx).Info("MemStats: Number Of Goroutines: ", runtime.NumGoroutine())
	if runtime.NumGoroutine() > 20 {
		buf := make([]byte, 100056)
		length := runtime.Stack(buf, true)
		Logger(ctx).Infof("MemStats: length is %d", length)
		Logger(ctx).Infof("MemStats: Stack Trace of all goroutines: %v", string(buf[:length]))

	}
}
