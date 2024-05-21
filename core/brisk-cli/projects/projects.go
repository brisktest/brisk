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

package projects

import (
	"brisk-supervisor/api"
	"brisk-supervisor/shared"
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/spf13/viper"
)

func CheckForBriskJsonFile(ctx context.Context, path string) bool {
	// checks to see if we have a brisk.json file in the current directory
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)

}

//creates a shell brisk.json file for a node project

func createJestProjectFile(ctx context.Context, projectToken string, concurrency int) error {

	config := shared.Config{}
	config.Framework = "Jest"
	config.WorkerImage = "node-lts"

	config.ExcludedFromSync = []string{"log/", ".git/", "node_modules", ".rvm"}
	config.ExcludedFromWatch = []string{"log/", ".git/", "log", ".git", "node_modules"}
	config.ListTestCommand = "yarn -s test --listTests --json"
	config.Commands = []shared.Command{{Commandline: "FORCE_COLOR=true yarn test  --json "}}
	config.BuildCommands = []shared.Command{{Commandline: "nvm install 14.18.0"}, {Commandline: "nvm alias default 14.18.0 && nvm use default"}, {Commandline: "yarn"}}
	config.Environment = map[string]string{"MY_ENV": "empty"}
	config.FilterList = []string{"SOME_FILTER"}
	config.ProjectToken = projectToken
	config.Concurrency = concurrency

	writeErr := config.WriteConfig(ctx)
	return writeErr
}
func createPythonProjectFile(ctx context.Context, projectToken string, concurrency int) error {

	config := shared.Config{}
	config.Framework = "Python"
	config.WorkerImage = "python"

	config.ExcludedFromSync = []string{"log/", ".git/", "node_modules", ".rvm"}
	config.ExcludedFromWatch = []string{"log/", ".git/", "log", ".git", "node_modules"}
	config.ListTestCommand = "echo 'list tests'"
	config.Commands = []shared.Command{{Commandline: "pytest --splits $BRISK_NODE_TOTAL --group $((BRISK_NODE_INDEX+1))"}}
	config.BuildCommands = []shared.Command{{Commandline: ""}}
	config.Environment = map[string]string{"MY_ENV": "empty"}
	config.FilterList = []string{"SOME_FILTER"}
	config.ProjectToken = projectToken
	config.Concurrency = concurrency

	writeErr := config.WriteConfig(ctx)
	return writeErr
}
func createRawProjectFile(ctx context.Context, projectToken string, concurrency int) error {

	config := shared.Config{}
	config.Framework = "Raw"
	config.WorkerImage = "raw"

	config.ExcludedFromSync = []string{"log/", ".git/", "node_modules", ".rvm"}
	config.ExcludedFromWatch = []string{"log/", ".git/", "log", ".git", "node_modules"}
	config.ListTestCommand = "echo 'list tests'"
	config.Commands = []shared.Command{{Commandline: ""}}
	config.BuildCommands = []shared.Command{{Commandline: ""}}
	config.Environment = map[string]string{"MY_ENV": "empty"}
	config.FilterList = []string{"SOME_FILTER"}
	config.ProjectToken = projectToken
	config.Concurrency = concurrency

	writeErr := config.WriteConfig(ctx)
	return writeErr
}

// creates a shell brisk.json file for a rails project
func createRailsProjectFile(ctx context.Context, prokectToken string, concurrency int) error {

	config := shared.Config{}
	config.Framework = "Rails"
	config.WorkerImage = "rails"

	config.ExcludedFromSync = []string{"log/", ".git/", "node_modules", ".rvm"}
	config.ExcludedFromWatch = []string{"log/", ".git/", "log", ".git", "node_modules"}
	config.ListTestCommand = `bundle exec rspec --dry-run --format json`
	config.Commands = []shared.Command{{Commandline: "bin/test"}}
	config.Environment = map[string]string{"MY_ENV": "empty"}
	config.FilterList = []string{"SOME_FILTER"}
	config.ProjectToken = prokectToken
	config.Concurrency = concurrency

	writeErr := config.WriteConfig(ctx)
	return writeErr
}
func CreateRailsProject(ctx context.Context) error {

	endpoint := viper.GetString("ApiEndpoint")
	conn, err := shared.ApiConn(ctx, endpoint)

	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewProjectsClient(conn)
	resp, initErr := c.InitProject(ctx, &api.InitProjectReq{Framework: "Rails"})

	if initErr != nil {
		return initErr
	}
	projectToken := resp.Project.ProjectToken
	concurrency := resp.Project.Concurrency
	err = createRailsProjectFile(ctx, projectToken, int(concurrency))
	return err
}

func CreateJestProject(ctx context.Context) error {

	endpoint := viper.GetString("ApiEndpoint")
	conn, err := shared.ApiConn(ctx, endpoint)

	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewProjectsClient(conn)
	resp, initErr := c.InitProject(ctx, &api.InitProjectReq{Framework: "Jest"})

	if initErr != nil {
		return initErr
	}
	projectToken := resp.Project.ProjectToken
	concurrency := resp.Project.Concurrency
	err = createJestProjectFile(ctx, projectToken, int(concurrency))
	return err
}

func CreateRawProject(ctx context.Context) error {

	endpoint := viper.GetString("ApiEndpoint")
	conn, err := shared.ApiConn(ctx, endpoint)

	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewProjectsClient(conn)
	resp, initErr := c.InitProject(ctx, &api.InitProjectReq{Framework: "Raw"})

	if initErr != nil {
		return initErr
	}
	projectToken := resp.Project.ProjectToken
	concurrency := resp.Project.Concurrency
	err = createRawProjectFile(ctx, projectToken, int(concurrency))
	return err
}

func CreatePythonProject(ctx context.Context) error {

	endpoint := viper.GetString("ApiEndpoint")
	conn, err := shared.ApiConn(ctx, endpoint)

	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewProjectsClient(conn)
	resp, initErr := c.InitProject(ctx, &api.InitProjectReq{Framework: "Python"})

	if initErr != nil {
		return initErr
	}
	projectToken := resp.Project.ProjectToken
	concurrency := resp.Project.Concurrency
	err = createPythonProjectFile(ctx, projectToken, int(concurrency))
	return err
}

func ListAll(ctx context.Context) ([]*api.Project, error) {

	endpoint := viper.GetString("ApiEndpoint")
	conn, err := shared.ApiConn(ctx, endpoint)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := api.NewProjectsClient(conn)

	resp, err := c.GetAllProjects(ctx, &api.GetAllProjectsReq{})
	if err != nil {
		return nil, err
	}
	return resp.Projects, nil

}
