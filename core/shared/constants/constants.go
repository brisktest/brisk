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

package constants

import "time"

const VERSION = "0.1.70"

const DEFAULT_CONFIG_DIR = ".config/brisk/"

const DEFAULT_CONFIG_FILE = "config"

const DEFAULT_CONFIG_FILE_EXT = "toml"

const UPDATE_SERVER_URL = "https://update.brisktest.com/"

const DEFAULT_API_ENDPOINT = "api.brisktest.com:50052"

const DEFAULT_REBUILD_HASH_FILE = ".rebuild_hash"

const DEFAULT_SERVER_HASH_FILE = "/tmp/.rebuild"

const DEFAULT_LOCK_TIMEOUT = 40

const DEFAULT_PROJECT_CONFIG_FILE = "brisk.json"

const NOMAD_DEPLOY_WORKER = "deploy_worker"

const NOMAD_DEPLOY_WORKER_RAILS = "deploy_worker_rails"

const NOMAD_DEPLOY_WORKER_NODE = "deploy_worker_node"

const NOMAD_DEPLOY_WORKER_PYTHON = "deploy_worker_python"

const NOMAD_DEPLOY_WORKER_RAW = "deploy_worker_raw"

const DEFAULT_SYNC_HOST = "sync.brisktest.com"

const DEFAULT_OSX_LOG_FILE = "/tmp/brisk.log"

const DEFAULT_LINUX_LOG_FILE = "/tmp/brisk.log"

const BUGSNAG_API_KEY = "9d1e83d0c051f715113b99dba0722cbf"

// const HONEYCOMB_API_KEY = "7N6eGnnGJsMZSX7qwdEZQG"
const HONEYCOMB_API_KEY = "gXBWBY6nd9NrpeGIDaOjxB"

const DEFAULT_SUPERVISOR_TIMEOUT = 6 * time.Minute

const SUPER_LOCKED = "SUPER_LOCKED"

const UNLOCKED_SUPER = "UNLOCKED_SUPER"

const SUPER_READY_CONTROL = "SUPER_READY_CONTROL"

const LOGSERVICE_PORT = "60061"

const DEFAULT_REGION = "us-east-1"

const DEFAULT_LOG_BUCKET = "brisk-output-logs"

// wee need these to slowly cancel from leafs to trunk

const PROJECT_RUN_TIMEOUT = COMMAND_TIMEOUT + 10*time.Second

const COMMAND_TIMEOUT = RUNNER_COMMAND_TIMEOUT + 10*time.Second

// we want the runners to quit first
const RUNNER_COMMAND_TIMEOUT = 10 * time.Minute

const SUPER_FIRST_RUN_FILE = "/home/brisk/brisk_first_run_at"
