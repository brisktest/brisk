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
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpc_status "google.golang.org/grpc/status"
)

// our custom errors

var ProjectInUseError = status.Error(codes.Unavailable, "project is in use")
var ProjectTokenError = status.Error(codes.InvalidArgument, "project token is invalid")
var RSyncError = status.Error(codes.Internal, "rsync error")
var TestFailedError = errors.New("test run failed")

func CancelledError(err error) bool {
	return errors.Is(err, context.Canceled) || grpc_status.Code(err) == codes.Canceled

}
