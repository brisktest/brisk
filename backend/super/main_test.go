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
	brisksupervisor "brisk-supervisor/brisk-supervisor"
	pb "brisk-supervisor/brisk-supervisor"

	"context"
	_ "net/http/pprof"
	"testing"
)

func Test_runTestTheTests(t *testing.T) {
	type args struct {
		ctx            context.Context
		responseStream chan pb.Output
		buildCommands  []*api.Command
		commands       []*api.Command
		config         *brisksupervisor.Config
		repoInfo       api.RepoInfo
		logUid         string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test runTestTheTests",
			args: args{
				ctx:            (context.Background()),
				responseStream: make(chan pb.Output),
				buildCommands:  []*api.Command{},
				commands:       []*api.Command{},
				config:         &brisksupervisor.Config{},
				repoInfo:       api.RepoInfo{},
				logUid:         "test",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Test runTestTheTests",
			args: args{
				ctx:            (context.Background()),
				responseStream: make(chan pb.Output),
				buildCommands:  []*api.Command{},
				commands:       []*api.Command{{Commandline: "ls -al | wc -l"}, {Commandline: "ls -al"}},
				config:         &brisksupervisor.Config{WorkerImage: "rails"},
				repoInfo:       api.RepoInfo{},
				logUid:         "test",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runTestTheTests(tt.args.ctx, tt.args.responseStream, tt.args.buildCommands, tt.args.commands, tt.args.config, tt.args.repoInfo, tt.args.logUid)
			if (err != nil) != tt.wantErr {
				t.Errorf("runTestTheTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("runTestTheTests() = %v, want %v", got, tt.want)
			}
		})
	}
}
