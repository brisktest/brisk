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
	"brisk-supervisor/shared/types"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/go-errors/errors"

	"os"

	"github.com/spf13/viper"
)

// Config struct lists commands
type Config struct {
	Commands           []Command         `json:"commands"`
	BuildCommands      []Command         `json:"buildCommands"`
	Concurrency        int               `json:"concurrency"`
	ExcludedFromSync   []string          `json:"excludedFromSync"`
	ExcludedFromWatch  []string          `json:"excludedFromWatch"`
	ProjectToken       string            `json:"projectToken"`
	ApiEndpoint        string            `json:"apiEndpoint"`
	ReconnectDelay     int               `json:"reconnectDelay"`
	Framework          string            `json:"framework"`
	ListTestCommand    string            `json:"listTestCommand"`
	PreListTestCommand string            `json:"preListTestCommand"`
	Environment        map[string]string `json:"environment"`
	WorkerImage        string            `json:"image"`
	SplitByJUnit       bool              `json:"splitByJUnit"`
	OrderedFiles       []string          `json:"orderedFiles"`
	FilterList         []string          `json:"filterList"`
	RebuildFilePaths   []string          `json:"rebuildFilePaths"`
	GithubRepo         string            `json:"githubRepo"`
	SkipRecalcFiles    bool              `json:"skipRecalcFiles"`
	Verbose            bool              `json:"verbose"`
	NoFailFast         bool              `json:"noFailFast"`
	AutomaticSplitting bool              `json:"automaticSplitting"`
}

type Server struct {
	Address string `json:"address"`
}

// Command represents each Command to be run
type Command struct {
	Commandline        string   `json:"commandline"`
	Args               []string `json:"args"`
	WorkDirectory      string   `json:"workDirectory"`
	Background         bool     `json:"background"`
	CommandConcurrency int      `json:"commandConcurrency"`
	CommandId          string   `json:"commandId"`
	NoTestFiles        bool     `json:"noTestFiles"`
}

func (config Config) WriteConfig(ctx context.Context) error {
	err := validateConfig(config)
	if err != nil {
		return err
	}
	config = cleanConfig(ctx, config)
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	// if file exists return error
	if _, err := os.Stat(viper.GetString("PROJECT_CONFIG_FILE")); err == nil {
		return errors.Errorf("config file already exists at %v", viper.GetString("PROJECT_CONFIG_FILE"))
	} else {
		writeErr := ioutil.WriteFile(viper.GetString("PROJECT_CONFIG_FILE"), jsonBytes, 0644)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// ReadConfig reads our config file
func ReadConfig(ctx context.Context, fileName string) (*Config, error) {
	// Open our jsonFile
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		Logger(ctx).Error(err)

		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Config array
	var config Config
	Logger(ctx).Debug("the config values are ", string(byteValue))

	error := json.Unmarshal(byteValue, &config)
	if error != nil {
		fmt.Printf("Error reading config : - %v \n", string(byteValue))
		fmt.Println(error)
		fmt.Println("hint: check your config file for syntax errors")
		Logger(ctx).Fatal(error)

	}

	Logger(ctx).Debugf("the config values after unmarshal  are %+v", config)

	config = cleanConfig(ctx, config)

	config = mergeInEnvVars(ctx, config)

	return &config, validateConfig(config)
}

func mergeInEnvVars(ctx context.Context, config Config) Config {

	if viper.GetString("PROJECT_TOKEN") != "" && viper.GetBool("CONFIG_WARNINGS") {
		if config.ProjectToken != "" {
			Logger(ctx).Warnf("BRISK_PROJECT_TOKEN %v env var is overriding projectToken in %v - remove the project token from the config file to avoid this warning", viper.GetString("PROJECT_TOKEN"), viper.GetString("PROJECT_CONFIG_FILE"))
			fmt.Printf("\n WARNING: BRISK_PROJECT_TOKEN env var is overriding projectToken in %v - remove the project token from the config file to avoid this warning \n", viper.GetString("PROJECT_CONFIG_FILE"))
		}
		config.ProjectToken = viper.GetString("PROJECT_TOKEN")
	}

	if viper.GetString("CONCURRENCY") != "" && viper.GetBool("CONFIG_WARNINGS") {
		if config.Concurrency != 0 {
			Logger(ctx).Warnf("BRISK_CONCURRENCY %v env var is overriding concurrency in %v - remove the concurrency from the config file to avoid this warning", viper.GetString("CONCURRENCY"), viper.GetString("PROJECT_CONFIG_FILE"))
			fmt.Printf("\n WARNING: BRISK_CONCURRENCY env var is overriding concurrency in %v - remove the concurrency from the config file to avoid this warning \n", viper.GetString("PROJECT_CONFIG_FILE"))
		}
		concurrency, err := strconv.Atoi(viper.GetString("CONCURRENCY"))
		if err != nil {
			Logger(ctx).Warn("BRISK_CONCURRENCY env var is not a valid integer - ignoring")
			fmt.Println("\n WARNING: BRISK_CONCURRENCY env var is not a valid integer - ignoring")
		} else {
			config.Concurrency = concurrency
		}
	}

	return config
}
func validateConfig(config Config) error {
	if config.WorkerImage == "" {
		return errors.New("config error: image can not be blank")
	}
	return nil
}

func cleanConfig(ctx context.Context, config Config) Config {

	switch strings.ToUpper(config.Framework) {
	case "JEST":
		config.Framework = string(types.Jest)
	case "RSPEC":
		config.Framework = string(types.Rspec)
	case "CYPRESS":
		config.Framework = string(types.Cypress)
	case "RAILS":
		config.Framework = string(types.Rails)
	case "RAW":
		config.Framework = string(types.Raw)
	case "PYTHON":
		config.Framework = string(types.Python)
	default:
		fmt.Printf("Config error - unrecognized framework : %v", config.Framework)
		Logger(ctx).Fatalf("Unrecognized framework %v", config.Framework)
	}

	return config
}
