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
	"fmt"
	"time"

	trace "go.opentelemetry.io/otel/trace"
)

func GetKey(ctx context.Context, token string) string {
	span := trace.SpanFromContext(ctx)

	honeycomb_key := span.SpanContext().TraceID().String()

	if len(honeycomb_key) > 10 {
		Logger(ctx).Debugf("using honeycomb trace key : %v", honeycomb_key)
		return honeycomb_key
	} else {
		hx := ReversibleHash([]byte(token))
		traceKey := fmt.Sprintf("%v-%x", time.Now().Unix(), hx)
		//traceKey := fmt.Sprintf("%v", time.Now().Unix())
		Logger(ctx).Infof("Generating trace key of %v", traceKey)
		Logger(ctx).Debugf("using our own trace key ", traceKey)

		return traceKey
	}

}

func ReversibleHash(token []byte) []byte {
	var val []byte
	for _, b := range token {
		val = append(val, ^(b))
	}
	return val
}
