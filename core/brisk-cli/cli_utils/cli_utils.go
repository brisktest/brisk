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

package cli_utils

import (
	"brisk-supervisor/shared/auth"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"

	. "brisk-supervisor/shared"
	. "brisk-supervisor/shared/logger"

	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

func CliAddAuthToCtx(ctx context.Context) (context.Context, error) {
	authCreds := auth.AuthCreds{ProjectToken: viper.GetString("projectToken"), ApiToken: viper.GetString("apiToken"), ApiKey: viper.GetString("apiKey")}
	return auth.AddAuthToCtx(ctx, authCreds)
}

func CliAddTraceKeyToCtx(ctx context.Context, projectToken string) (context.Context, error) {
	traceKey := GetKey(ctx, projectToken)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace-key", traceKey)
	ctx = WithTraceId(ctx, traceKey)
	return ctx, nil
}

func GetMachineUID(ctx context.Context) (string, error) {
	return machineid.ProtectedID("brisk-cli")
}

var uuid string

func GetUniqueInstanceID(ctx context.Context) string {
	if uuid != "" {
		return uuid
	}

	path, perr := os.Getwd()
	if perr != nil {
		Logger(ctx).Errorf("Error getting current working directory: %v", perr)
		return fmt.Sprintf("Error getting current working directory: %v", perr)
	}

	pid, merr := machineid.ProtectedID("brisk-cli")
	if merr != nil {
		Logger(ctx).Errorf("Error getting machine ID: %v", merr)
		return fmt.Sprintf("Error getting machine ID: %v", merr)
	}

	return protect("brisk-cli", fmt.Sprintf("%s-%s", path, pid))
}

func protect(appID, id string) string {
	mac := hmac.New(sha256.New, []byte(id))
	mac.Write([]byte(appID))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
