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

package context

import (
	pb "brisk-supervisor/brisk-supervisor"
	. "brisk-supervisor/shared/logger"
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

type ContextKey string

func (c ContextKey) String() string {
	return "mypackage context key " + string(c)
}

func ResponseStream(ctx context.Context) (*chan pb.Output, error) {
	contextKeyResponseStream := ContextKey("response-stream")

	responseStream, ok := ctx.Value(contextKeyResponseStream).(*chan pb.Output)
	var err error
	err = nil
	if !ok {
		Logger(ctx).Error("Cannot retreive response stream from context")
		err = errors.New("cannot retreive response stream from context")
	}
	return responseStream, err
}

func AddMetadataToCtx(ctx context.Context) context.Context {

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		Logger(ctx).Debugf("Print the md from incoming context %+v", md)

		tkeys := md.Get("trace-key")
		for _, tk := range tkeys {
			ctx = WithTraceId(ctx, tk)
			ctx = metadata.AppendToOutgoingContext(ctx, "trace-key", tk)
			Logger(ctx).Debugf("Added trace id %v", tk)
		}

		version := md.Get("brisk_api_version")
		if len(version) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "brisk_api_version", version[0])
		} else {
			Logger(ctx).Debug("No brisk_api_version found in metadata")
		}
		instance_id := md.Get("unique_instance_id")
		if len(instance_id) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "unique_instance_id", instance_id[0])
		} else {
			Logger(ctx).Debug("No unique_instance_id found in metadata")
		}

	} else {
		Logger(ctx).Debug("No metadata provided")
	}
	return ctx
}
