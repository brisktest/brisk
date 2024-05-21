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

package env

import (
	"brisk-supervisor/shared/constants"
	"context"
	"time"

	"github.com/spf13/viper"
)

func initDefaultValues(ctx context.Context) {
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("print_keys", false)
	viper.SetDefault("SUPERVISOR_TIMEOUT", constants.DEFAULT_SUPERVISOR_TIMEOUT)
	viper.SetDefault("PROJECT_RUN_TIMEOUT", constants.PROJECT_RUN_TIMEOUT)
	viper.SetDefault("RUNNER_COMMAND_TIMEOUT", constants.RUNNER_COMMAND_TIMEOUT)
	viper.SetDefault("COMMAND_TIMEOUT", constants.COMMAND_TIMEOUT)
	viper.SetDefault("REBUILD_BACKOFF", 5*time.Second)
	viper.SetDefault("REBUILD_RETRY", 20)
	viper.SetDefault("NO_WORKER_TIMEOUT", 60*time.Second)
	viper.SetDefault("SYNC_TO_WORKER_TIMEOUT", 45*time.Second)
	viper.SetDefault("WORKER_SETUP_TIMEOUT", 15*time.Second)
	viper.SetDefault("CHECK_REBUILD_TIMEOUT", 15*time.Second)
	viper.SetDefault("SUPER_KILL_TIME", 2*time.Hour) //TODO make this work better.
	viper.SetDefault("FILTER_BUILD_COMMANDS", false)
	viper.SetDefault("DEFAULT_BUILD_TIMEOUT", 7*time.Minute)
	// viper.SetDefault("HTTP_PROXY", "http://172.31.252.198:3128")
	viper.SetDefault("HTTP_PROXY", "")
	viper.SetDefault("RSYNC_TIMEOUT", 10)
	viper.SetDefault("MEMORY_CHECK_INTERVAL", 5*time.Second)
}

func InitServerViper(ctx context.Context) {

	viper.SetEnvPrefix("BRISK")
	initDefaultValues(ctx)
	viper.AutomaticEnv()

}
