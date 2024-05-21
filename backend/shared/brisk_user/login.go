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

package brisk_user

import (
	"brisk-supervisor/api"
	. "brisk-supervisor/shared/logger"
	"context"
	"errors"
	"fmt"

	. "brisk-supervisor/shared"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/pkg/browser"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func Login(ctx context.Context, nonce string) (*api.Credentials, error) {

	haveOpenedBrowser := false

	var conn *grpc.ClientConn
	endpoint := viper.GetString("ApiEndpoint")

	conn, err := ApiConn(ctx, endpoint)
	if err != nil {
		Logger(ctx).Errorf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	Logger(ctx).Debug("Connected ")
	c := api.NewUsersClient(conn)
	in := api.LoginRequest{Nonce: nonce}

	res, getErr := c.Login(ctx, &in, grpc_retry.WithMax(10))
	if getErr != nil {
		Logger(ctx).Errorf("Error logging in %v", getErr)
		return nil, getErr
	}
	for {

		in, err := res.Recv()
		if in != nil {

			Logger(ctx).Debugf(" %+v", in)
		}

		if err != nil {
			Logger(ctx).Errorf("Error logging in %v", err)
			return nil, err
		}

		if len(in.Url) > 0 && !haveOpenedBrowser {
			fmt.Println("the url is " + in.Url)
			haveOpenedBrowser = true
			browser.OpenURL(in.Url)
		} else if in.Status == "SUCCESS" {
			Logger(ctx).Debug("Login successful")
			fmt.Println("Login successful")
			return in.Credentials, nil

		} else if in.Status == "FAILURE" {
			Logger(ctx).Errorf("Error logging in %v", err)
			return nil, errors.New("Login failed")
		}

	}
}
