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
	pb "brisk-supervisor/brisk-supervisor"

	"context"
	"reflect"
	"testing"
)

func Test_server_RunCommands(t *testing.T) {
	type fields struct {
		UnimplementedCommandRunnerServer pb.UnimplementedCommandRunnerServer
	}
	type args struct {
		stream pb.CommandRunner_RunCommandsServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				UnimplementedCommandRunnerServer: tt.fields.UnimplementedCommandRunnerServer,
			}
			if err := s.RunCommands(tt.args.stream); (err != nil) != tt.wantErr {
				t.Errorf("server.RunCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_CheckBuild(t *testing.T) {
	type fields struct {
		UnimplementedCommandRunnerServer pb.UnimplementedCommandRunnerServer
	}
	type args struct {
		ctx context.Context
		in  *pb.CheckBuildMsg
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CheckBuildResp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				UnimplementedCommandRunnerServer: tt.fields.UnimplementedCommandRunnerServer,
			}
			got, err := s.CheckBuild(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("server.CheckBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server.CheckBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_Setup(t *testing.T) {
	type fields struct {
		UnimplementedCommandRunnerServer pb.UnimplementedCommandRunnerServer
	}
	type args struct {
		ctx        context.Context
		testOption *pb.TestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.Response
		wantErr bool
	}{
		{name: "test", args: args{ctx: context.Background(),
			testOption: &pb.TestOption{}}, want: &pb.Response{Value: true}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				UnimplementedCommandRunnerServer: tt.fields.UnimplementedCommandRunnerServer,
			}
			got, err := s.Setup(tt.args.ctx, tt.args.testOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("server.Setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server.Setup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapHash(t *testing.T) {
	type args struct {
		hash map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{name: "a = b", args: args{hash: map[string]string{"a": "b"}}, want: []string{"a=b"}},
		{name: "brisk log level", args: args{hash: map[string]string{"BRISK_LOG_LEVEL": "debug"}}, want: []string{"BRISK_LOG_LEVEL=debug"}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapHash(tt.args.hash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRebuildHash(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRebuildHash(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRebuildHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRebuildHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
