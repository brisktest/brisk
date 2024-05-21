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
	"errors"
	"fmt"
	"io/ioutil"

	"brisk-supervisor/api"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

type CircleCIConfig struct {
	Version   int       `yaml:"version"`
	Jobs      Jobs      `yaml:"jobs"`
	Workflows Workflows `yaml:"workflows"`
}
type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Docker struct {
	Image       string            `yaml:"image"`
	Auth        Auth              `yaml:"auth"`
	Command     []string          `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

type Build struct {
	Docker           []Docker          `yaml:"docker"`
	Environment      map[string]string `yaml:"environment"`
	WorkingDirectory string            `yaml:"working_directory"`
	Steps            []interface{}     `yaml:"steps"`
}

type Setup struct {
	Docker           []Docker          `yaml:"docker"`
	Environment      map[string]string `yaml:"environment"`
	WorkingDirectory string            `yaml:"working_directory"`
	Steps            []interface{}     `yaml:"steps"`
}

type Run struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}
type Steps struct {
	Run Run `yaml:"run"`
}
type DeployStage struct {
	Docker           []Docker `yaml:"docker"`
	WorkingDirectory string   `yaml:"working_directory"`
	Steps            []Steps  `yaml:"steps"`
}
type DeployProd struct {
	Docker           []Docker `yaml:"docker"`
	WorkingDirectory string   `yaml:"working_directory"`
	Steps            []Steps  `yaml:"steps"`
}
type Jobs struct {
	Build       Build       `yaml:"build"`
	Setup       Setup       `yaml:"setup"`
	DeployStage DeployStage `yaml:"deploy-stage"`
	DeployProd  DeployProd  `yaml:"deploy-prod"`
}
type BuildDeploy struct {
	Jobs []Jobs `yaml:"jobs"`
}
type Workflows struct {
	Version     int         `yaml:"version"`
	BuildDeploy BuildDeploy `yaml:"build-deploy"`
}

// ReadCircleConfig reads our config file
func ReadCircleConfig(ctx context.Context, fileName string) CircleCIConfig {
	// Open our yamlFile
	yamlFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Print(err)
		os.Exit(2)
	}

	fmt.Printf("Successfully Opened %v", fileName)
	// defer the closing of our yamlFile so that we can parse it later on
	defer yamlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		Logger(ctx).Debug("Error reading")
		Logger(ctx).Fatal(err)
	}

	// we initialize our Config array
	var config CircleCIConfig

	// we unmarshal our byteArray which contains our
	// yamlFile's content into 'users' which we defined above
	uErr := yaml.Unmarshal(byteValue, &config)

	if uErr != nil {
		Logger(ctx).Fatal(uErr)
	}

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	//fmt.Print(commands)
	// for i := 0; i < len(commands.Commands); i++ {
	// 	spew.Dump(commands.Commands[i])
	// 	fmt.Println("Comands: " + commands.Commands[i].Commandline)
	// 	fmt.Println("Workdir: " + commands.Commands[i].WorkDirectory)
	// 	// for _, r := range commands.Commands[i].args {
	// 	// 	fmt.Print(r)
	// 	// }

	// }
	Logger(ctx).Debug("The Config is")
	Logger(ctx).Debug(config)
	//buildSteps := make([]map[interface{}]interface{}, 10)

	dockerConfig := config.Jobs.Build.Docker
	Logger(ctx).Debug(dockerConfig)

	var runCommands []api.Command

	Logger(ctx).Debug("--------------------------steps----------------------------")
	//for _, step := range config.Jobs.Build.Steps {
	for _, step := range config.Jobs.Setup.Steps {
		if properStep, ok := step.(map[interface{}]interface{}); ok {

			for k, v := range properStep {
				// Logger(ctx).Debug("k .. ")
				// Logger(ctx).Debug(k)
				// Logger(ctx).Debug("v ..")
				// Logger(ctx).Debug(v)
				if rangedValue, ok := v.(map[interface{}]interface{}); ok {

					for k2, v2 := range rangedValue {
						Logger(ctx).Debug("k2 ..")

						Logger(ctx).Debug(k2)
						Logger(ctx).Debug("v2 ..")
						Logger(ctx).Debug(v2)

						// should take this out into a function and parse each node separately
						if k == "run" {
							run, err := parseRun(ctx, v)
							if err != nil {
								Logger(ctx).Fatal(err)
							}
							runCommands = append(runCommands, run)
						}
					}
				}
			}
		} else {
			fmt.Printf("ok is %v for %v", ok, step)
			Logger(ctx).Debug()

			/* not string */
		}

	}
	// Logger(ctx).Debug("the build steps are")
	// Logger(ctx).Debug(buildSteps)
	Logger(ctx).Debug("-----------------------------::::::::::::::::::::::::-------------------")

	for i := 0; i < len(runCommands); i++ {
		Logger(ctx).Debug(runCommands[i].Commandline)
		for k, v := range runCommands[i].Environment {
			Logger(ctx).Debugf(" %v : %v ", k, v)
		}
	}
	Logger(ctx).Debugf("the jobs in total are %v", config.Jobs)

	Logger(ctx).Debug("return from reader")
	return config
}

// parses a run node into a pb Command
func parseRun(ctx context.Context, node interface{}) (returnValue api.Command, e error) {
	Logger(ctx).Debugf("the run value for v is %v", node)
	returnValue = api.Command{Environment: map[string]string{}}
	switch node.(type) {
	case string:
		if stringValue, ok := node.(string); ok {
			returnValue.Commandline = stringValue
			return
		}
	case map[interface{}]interface{}:
		if line, ok := node.(map[interface{}]interface{}); ok {
			Logger(ctx).Debug("we in the run line parse")
			for k, value := range line {
				Logger(ctx).Debugf("the key is %v", k)
				if k == "environment" {

					Logger(ctx).Debugf("the environment is %v, %T", value, value)
					fmt.Println(reflect.TypeOf(value), value)
					if mv, ok := value.(map[interface{}]interface{}); ok {
						for k3, v3 := range mv {
							fmt.Println(reflect.TypeOf(v3), v3)

							fmt.Println(reflect.TypeOf(k3), k3)

							returnValue.Environment[k3.(string)] = v3.(string)
						}
					}

				}
				if stringValue, ok := value.(string); ok {

					if k == "command" {
						// for _, s := range strings.Split(stringValue, "\n") {
						// 	c.Commandline = string(s)
						returnValue.Commandline = stringValue
						// }

						return
					}
				}
			}
		} else {
			e = errors.New("Failed to parse : map wrong type")
		}
	default:
		e = errors.New("Failed to parse : type of initial value not recognized")

	}
	return
}
