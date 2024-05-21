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
	. "brisk-supervisor/shared/logger"
	"context"

	"github.com/bugsnag/bugsnag-go"
	"go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	Logger(ctx).Debugf("Server: %v, method: %v, req: %v", info.Server, info.FullMethod, req)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		Logger(ctx).Errorf("No metadata found in context")
	}

	Logger(ctx).Debugf("Metadata: %+v", md)

	resp, err := handler(ctx, req)
	Logger(ctx).Debugf("Finished -- Server: %v, method: %v, resp: %v", info.Server, info.FullMethod, resp)
	return resp, err
}

func StreamServerLoggingInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	Logger(ctx).Debugf("Server: %v, method: %v", srv, info.FullMethod)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		Logger(ctx).Errorf("No metadata found in context")
	}

	Logger(ctx).Debugf("Metadata: %+v", md)
	err := handler(srv, ss)
	Logger(ctx).Debugf("Finished -- Server: %v, method: %v, err: %v", srv, info.FullMethod, err)
	return err
}

func BugsnagClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		Logger(ctx).Debugf("opening client streaming to the server method: %v", method)
		output, err := streamer(ctx, desc, cc, method)
		if err != nil {
			if CancelledError(err) {
				Logger(ctx).Debug("Context cancelled")
			} else {
				Logger(ctx).Errorf("Error in streamer: %v", err)
				berr := bugsnag.Notify(err)
				if berr != nil {
					Logger(ctx).Errorf("Error in bugsnag: %v", berr)
				}
			}
		}
		Logger(ctx).Debugf("Streamer returned %+v with error %v", output, err)
		return output, err
	}
}

// func BugsnagClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
// 	// Logic before invoking the invoker
// 	// Calls the invoker to execute RPC

// 	Logger(ctx).Debugf("Calling streamer for %v", method)
// 	stream, err := streamer(ctx, desc, cc, method, opts...)
// 	// Logic after invoking the invoker
// 	if err != nil {
// 		Logger(ctx).Errorf("Error in streamer: %v", err)
// 		berr := bugsnag.Notify(err)
// 		if berr != nil {
// 			Logger(ctx).Errorf("Error in bugsnag: %v", berr)
// 		}
// 	}
// 	Logger(ctx).Debugf("Streamer returned %+v with error %v", stream, err)
// 	return stream, err
// }

// non-stream

func BugsnagClientUnaryInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Logic before invoking the invoker

	// Calls the invoker to execute RPC

	Logger(ctx).Debugf("Calling invoker for %v", method)
	Logger(ctx).Debugf("Request: %+v", req)

	err := invoker(ctx, method, req, reply, cc, opts...)
	// Logic after invoking the invoker
	if err != nil {
		Logger(ctx).Errorf("Error in invoker: %v", err)
		if CancelledError(err) {
			Logger(ctx).Debug("Context cancelled not notifying")
		} else {
			Logger(ctx).Debugf("Not a context cancelled error instead %v", err.Error())

			span := trace.SpanFromContext(ctx)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			berr := bugsnag.Notify(err)

			if berr != nil {
				Logger(ctx).Errorf("Error in bugsnag: %v", berr)
			}
		}
	}
	return err
}
