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

package types

import pb "brisk-supervisor/brisk-supervisor"

// a struct for withValue?
type CtxOutputStream struct {
	OutputChannel chan *pb.Output
}

type QuitChan struct {
	Quit chan bool
}
type LogStreamInfo struct {
	WorkerRunInfoUID string
	LogEncryptionKey string
	ProjectToken     string
	// optional log uid for when we want to save something that isn't a worker run info
	LogUid string
}
type Port string

type IpAddress string

const FAILED string = "Failed"
const FINISHED string = "Finished"

const WORKER_PORT Port = "50051"

type Framework string

const (
	Jest    Framework = "Jest"
	Rspec   Framework = "Rspec"
	Cypress Framework = "Cypress"
	Rails   Framework = "Rails"
	Python  Framework = "Python"
	Raw     Framework = "Raw"
)

type workerRunInfoUID string

const WORKER_RUN_INFO_MD string = "WORKER_RUN_INFO_MD"
const ENCRYPTION_KEY_MD string = "ENCRYPTION_KEY_MD"
