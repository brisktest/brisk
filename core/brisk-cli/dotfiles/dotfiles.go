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

package dotfiles

import (
	constants "brisk-supervisor/shared/constants"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	. "brisk-supervisor/shared/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//package to manage our dotifle

func initDefaultValues(ctx context.Context) {
	viper.SetDefault("LOG_LEVEL", "error")
	viper.SetDefault("NoRebuild", false)
	viper.SetDefault("HashFilePath", constants.DEFAULT_REBUILD_HASH_FILE)
	viper.SetDefault("RebuildWatchPaths", []string{})
	viper.SetDefault("ApiEndpoint", constants.DEFAULT_API_ENDPOINT)
	viper.SetDefault("REBUILD_HASH_FILE", constants.DEFAULT_REBUILD_HASH_FILE)
	viper.SetDefault("PROJECT_IN_USE_TIMEOUT", 10)
	viper.SetDefault("PROJECT_CONFIG_FILE", constants.DEFAULT_PROJECT_CONFIG_FILE)
	viper.SetDefault("watch", true)
	viper.SetDefault("print_keys", false)
	viper.SetDefault("TELEMETRY_ENABLED", true)
	viper.SetDefault("SYNC_HOST", constants.DEFAULT_SYNC_HOST)
	viper.SetDefault("RETRY_TIMEOUT", "5s")
	viper.SetDefault("PRINT_TRACE_KEY", true)
	viper.SetDefault("CONFIG_WARNINGS", true)
	viper.SetDefault("RETRY_COUNT", 5)
	viper.SetDefault("LOG_FILE", defaultForOS())
	viper.SetDefault("CREDENTIALS_CONFIG", constants.DEFAULT_CONFIG_DIR)
	// this is an int cause it's passed straight to rsync
	viper.SetDefault("RSYNC_TIMEOUT", 60)
	viper.SetDefault("CLI_STREAM_TIMEOUT", 20*time.Second)
}

func defaultForOS() string {
	os := runtime.GOOS
	if os == "darwin" {
		return constants.DEFAULT_OSX_LOG_FILE
	} else {
		return constants.DEFAULT_LINUX_LOG_FILE
	}
}

func InitCLIViper(ctx context.Context) {
	viper.SetEnvPrefix("BRISK")
	initDefaultValues(ctx)
	viper.AutomaticEnv()

	if viper.GetString("LOG_FILE") != defaultForOS() && (viper.GetString("LOG_LEVEL") == "error" || viper.GetString("LOG_LEVEL") == "warn") {
		if viper.GetBool("CONFIG_WARNINGS") {
			Logger(ctx).Error("Log file is changed but log level is still set to error or warn.")
		}
	}

	// fmt.Println("The apiEndpoint is " + os.Getenv("BRISK_APIENDPOINT"))
	// fmt.Println(" viper endpoint is " + viper.GetString("apiEndpoint"))

	viper.SetConfigName(constants.DEFAULT_CONFIG_FILE)     // name of config file (without extension)
	viper.SetConfigPermissions(0600)                       // read permissions
	viper.SetConfigType(constants.DEFAULT_CONFIG_FILE_EXT) // REQUIRED if the config file does not have the extension in the name
	home, err := os.UserHomeDir()
	if err != nil {
		Logger(ctx).Fatal(err)
	}
	viper.AddConfigPath(home + "/" + viper.GetString("CREDENTIALS_CONFIG"))

	Logger(ctx).Debug("Watching config at ", home+"/"+viper.GetString("CREDENTIALS_CONFIG")+constants.DEFAULT_CONFIG_FILE+"."+constants.DEFAULT_CONFIG_FILE_EXT)

	EnsureDirectory(ctx, home+"/"+viper.GetString("CREDENTIALS_CONFIG"))
	filename := home + "/" + viper.GetString("CREDENTIALS_CONFIG") + constants.DEFAULT_CONFIG_FILE + "." + constants.DEFAULT_CONFIG_FILE_EXT
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = viper.SafeWriteConfigAs(filename)
		if err != nil {
			fmt.Println(err)
			Logger(ctx).Fatal(err)
		}
	}

	err = viper.MergeInConfig()
	if err != nil {
		fmt.Println(err)
		Logger(ctx).Fatal(err)
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		Logger(ctx).Infof("Config file updated: %v", e.Name)
		Logger(ctx).Info("Reloading...")
	})

	// now we over write any read in standard config with env var stuff

	if viper.GetString("API_TOKEN") != "" && viper.GetBool("CONFIG_WARNINGS") {
		if viper.GetString("ApiKey") != "" {
			Logger(ctx).Warn("API_TOKEN is set in environment. This will override any value in the config file.")
		}
		viper.Set("ApiToken", viper.GetString("API_TOKEN"))
	}

	if viper.GetString("API_KEY") != "" && viper.GetBool("CONFIG_WARNINGS") {
		if viper.GetString("ApiKey") != "" {
			Logger(ctx).Warn("API_KEY is set in environment. This will override any value in the config file.")
		}
		viper.Set("ApiKey", viper.GetString("API_KEY"))
	}

	viper.WatchConfig()
	// viper.Debug()

}

// // this is never called
// func ReadConfig(ctx context.Context) error {
// 	panic("ReadConfig is never called")
// 	viper.AddConfigPath("$HOME/.config/brisk/") // call multiple times to add many search paths
// 	cfile, err := os.Open(viper.GetString("CREDENTIALS_CONFIG"))
// 	if err != nil {
// 		Logger(ctx).Errorf(err.Error())
// 		return err
// 	}
// 	viper.ReadConfig(cfile)
// 	// err := viper.ReadInConfig() // Find and read the config file

// 	return err

// }

func PrintConfig(ctx context.Context) {
	filename := viper.GetViper().ConfigFileUsed()

	fmt.Printf("Config file is  %v \n", filename)
	fmt.Printf("Config: \n")

	for k, v := range viper.AllSettings() {
		fmt.Println(k, "=", v)
	}
}

func WriteConfig(ctx context.Context) error {
	filename := viper.GetViper().ConfigFileUsed()

	fmt.Printf("Writing config to %v", filename)

	return viper.WriteConfig()
}

func EnsureDirectory(ctx context.Context, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Logger(ctx).Infof("Creating directory %v", path)
		err := os.MkdirAll(path, 0700)
		if err != nil {
			Logger(ctx).Fatal(err)
		}
	}
}

func AddToConfig(ctx context.Context, key string, value interface{}) {
	viper.Set(key, value)
}

// func ReadDotFile(ctx context.Context, file_uri string) (string, error) {

// 	dat, err := os.ReadFile(file_uri)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(dat), err

// }

// func WriteDotFile(ctx context.Context, file_uri string, data string) error {

// 	err := os.MkdirAll(file_uri, 0700)
// 	if err != nil {
// 		return err
// 	}

// 	err = os.WriteFile(file_uri, []byte(data), 0600)
// 	if err != nil {
// 		return err
// 	}
// 	return err
// }

// func CheckDotFileOrDir(ctx context.Context, dir_uri string) bool {

// 	if _, err := os.Stat(dir_uri); os.IsNotExist(err) {
// 		return false
// 	} else {
// 		return true
// 	}
// }
