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

package auth

import (
	. "brisk-supervisor/shared/logger"
	"context"
	"encoding/json"

	"github.com/go-errors/errors"
	"google.golang.org/grpc/metadata"

	. "brisk-supervisor/shared/context"
)

func MdHasAuth(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	auth_keys := md.Get("authorization")
	if len(auth_keys) == 0 {
		return false
	}
	if auth_keys[0] == "" {
		return false
	}
	return true
}

func OutgoingContextHasAuth(ctx context.Context) bool {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return false
	}
	auth_keys := md.Get("authorization")
	if len(auth_keys) == 0 {
		return false
	}
	if auth_keys[0] == "" {
		return false
	}
	return true
}

func AddAuthToCtx(ctx context.Context, authCreds AuthCreds) (context.Context, error) {
	//only add auth if it is not already there
	// it seems multiple authorizatin headers make it go nuts
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok && md.Get("authorization") != nil {
		return ctx, nil
	}
	encoded, err := authCreds.Encode(ctx)

	return metadata.AppendToOutgoingContext(ctx, "authorization", encoded), err
}

type AuthCreds struct {
	ProjectToken string
	ApiToken     string
	ApiKey       string
}

func PropagateCredentials(ctx context.Context) (context.Context, error) {

	authCreds, err := GetAuthCredsFromMd(ctx)

	if err != nil {
		return ctx, err
	}
	ctx = context.WithValue(ctx, ContextKey("project_token"), authCreds.ProjectToken)

	return AddAuthToCtx(ctx, authCreds)

}

func GetAuthCredsFromMd(ctx context.Context) (AuthCreds, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return AuthCreds{}, errors.New("failed to get metadata from context ")
	}
	auth_keys := md.Get("authorization")
	if len(auth_keys) == 0 {
		Logger(ctx).Error("No auth keys provided - exiting")
		Logger(ctx).PrintStackTrace()
		return AuthCreds{}, errors.New("No auth keys provided")

	}
	if auth_keys[0] == "" {
		return AuthCreds{}, errors.New("no credentials provided")
	}
	var authCreds AuthCreds
	jsonErr := json.Unmarshal([]byte(auth_keys[0]), &authCreds)
	return authCreds, jsonErr

}

func (a AuthCreds) Encode(ctx context.Context) (string, error) {
	j, err := json.Marshal(a)
	if err != nil {
		Logger(ctx).Debugf("Error encoding auth %v", err)
		return "", err
	}

	// return base64.StdEncoding.EncodeToString(j), nil
	return string(j), nil
}
