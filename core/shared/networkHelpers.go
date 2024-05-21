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
	"brisk-supervisor/shared/auth"
	. "brisk-supervisor/shared/logger"
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func GetOutboundIP(ctx context.Context) (net.IP, error) {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		Logger(ctx).Errorf("Error getting outbound IP: %s", err)
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	Logger(ctx).Debug("IPAdress = ", localAddr.IP.String())
	return localAddr.IP, nil
}

func ApiConn(ctx context.Context, endpoint string) (conn *grpc.ClientConn, err error) {

	if endpoint == "" {
		return nil, errors.New("no endpoint provided to ApiConn")
	}

	if !auth.OutgoingContextHasAuth(ctx) && auth.MdHasAuth(ctx) {
		Logger(ctx).Info("No auth in outgoing context")
		authCreds, err := auth.GetAuthCredsFromMd(ctx)
		if err != nil {

			Logger(ctx).Info("Error getting auth creds from md -but ignoring for now")
			Logger(ctx).Info(err)
			// return nil, err
		} else {
			ctx, err = auth.AddAuthToCtx(ctx, authCreds)
			if err != nil {
				Logger(ctx).Info(err)
				Logger(ctx).Info("Error adding auth creds to Ctx  -but ignoring for now")

				//return nil, err
			}
		}
	}

	Logger(ctx).Debugf("Attempting to connect to %s", endpoint)

	var credentialOpts grpc.DialOption
	if !IsDev() {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		credentialOpts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))

	} else {
		credentialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	count := 0
	startTime := time.Now()
	for count < 5 && conn == nil {
		Logger(ctx).Debug("About to start dialing")

		connectCtx, _ := context.WithTimeout(ctx, 3*time.Second)
		opts := []grpc_retry.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponential(200 * time.Millisecond)),
			grpc_retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
				Logger(ctx).Errorf("retrying after error: %v attempt # %v connecting to %v after %v ", err, attempt, endpoint, time.Since(startTime))
			}),
		}
		clientOpts := []grpc.DialOption{

			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff:           backoff.DefaultConfig,
				MinConnectTimeout: 5 * time.Second,
			}),
			grpc.WithBlock(),
			grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor(), grpc_retry.StreamClientInterceptor(opts...), BugsnagClientInterceptor()),
			grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor(), grpc_retry.UnaryClientInterceptor(opts...), BugsnagClientUnaryInterceptor),
			grpc.WithDefaultCallOptions(),
			credentialOpts,
		}

		conn, err = grpc.DialContext(connectCtx, endpoint, clientOpts...)
		Logger(ctx).Debug("Dialing over")
		if err != nil {
			Logger(ctx).Errorf("Error dialing to %v, retrying in 5 seconds : %v", endpoint, err)
			count++
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		Logger(ctx).Errorf("did not connect: %v", err)
		return nil, err
	}

	Logger(ctx).Debug("Connected to api")
	return conn, nil
}
