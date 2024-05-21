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
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/go-errors/errors"
	"github.com/spf13/viper"

	"github.com/bugsnag/bugsnag-go"
)

func SafeExit(failingErr error) {

	if failingErr != nil {
		Logger(context.Background()).Errorf("In SafeExit failing with err %s", failingErr)
	} else {
		Logger(context.Background()).Infof("No error in SafeExit")
	}

	if failingErr != nil {
		Logger(context.Background()).Error(failingErr)
		Logger(context.Background()).Debug(errors.Wrap(failingErr, 1).ErrorStack())
		bugsnag.Notify(failingErr)

	}
	Logger(context.Background()).Sync()
	time.Sleep(time.Millisecond * 100)
	if err := recover(); err != nil {
		// if we're in here, we had a panic and have caught it
		Logger(context.Background()).Errorf("we safely caught the panic: %s\n", err)
		Logger(context.Background()).Error(string(debug.Stack()))
		Logger(context.Background()).Error("^^^^^^^^ stack trace of panic ^^^^^^^^^^")
		Logger(context.Background()).Error("exiting with -1")
		Logger(context.Background()).Sync()
		bugsnag.Notify(err.(error))

		os.Exit(-1)
	} else {
		Logger(context.Background()).Error("exiting with -1 - no panic but still exiting")
		Logger(context.Background()).Sync()
		os.Exit(-1)
	}
	Logger(context.Background()).Debug("no panic so exiting normally")
	os.Exit(0)
}
func LogPanic(ctx context.Context) {
	Logger(ctx).Debug("LogPanic++")
	if err := recover(); err != nil {
		// if we're in here, we had a panic and have caught it
		Logger(ctx).Errorf("we caught the panic: %s\n", err)
		Logger(ctx).Error(string(debug.Stack()))
		Logger(ctx).Error("^^^^^^^^ stack trace of panic ^^^^^^^^^^")
		panic(err)
	} else {
		Logger(ctx).Info("no panic so exiting normally LogPanic--")
	}

}

func LogAllSignals(ctx context.Context) {
	c := make(chan os.Signal)

	signal.Notify(c)
	go func() {
		defer bugsnag.AutoNotify()

		sig := <-c
		Logger(ctx).Debugf("\r- received signal %v", sig.String())
		Logger(ctx).Debug("NOP")

	}()
}

func DebugCancelCtx(ctx context.Context, cancel func(), msg string) {
	Logger(ctx).Infof("cancelling context %+v because %v", ctx, msg)
	// Logger(ctx).Debug(string(debug.Stack()))

	cancel()
}
func CancelCtx(ctx context.Context, cancel func(), msg string) {
	if len(os.Getenv("SHOW_CTX_CANCEL")) > 0 {
		Logger(ctx).Infof("cancelling context %+v : %v", ctx, msg)
		Logger(ctx).Debug(string(debug.Stack()))
	}
	cancel()
}

func IsDev() bool {
	return os.Getenv("DEV") == "true" || viper.GetBool("DEV")
}

func OutputRuntimeStats(ctx context.Context) {

	viper.SetDefault("STATS_INTERVAL", "120")
	viper.AutomaticEnv()
	statsInterval := viper.GetInt("STATS_INTERVAL")
	for {

		select {
		case <-ctx.Done():
			Logger(ctx).Debug("OutputRuntimeStats: context done")
			return

		case <-time.After(time.Minute * time.Duration(statsInterval)):

			memStats := runtime.MemStats{}
			runtime.ReadMemStats(&memStats)
			gcstats := debug.GCStats{}
			debug.ReadGCStats(&gcstats)
			Logger(ctx).Info("************************************ Runtime Stats Start *******************************************")
			Logger(ctx).Infof("runtime memstats: %+v ", memStats)
			Logger(ctx).Infof("The number of goroutines: %d", runtime.NumGoroutine())
			Logger(ctx).Infof("The number of CPUs: %d", runtime.NumCPU())
			Logger(ctx).Infof("The GC stats: %+v", gcstats)
			Logger(ctx).Info("************************************ Runtime Stats End ********************************************")
		}
	}
}
