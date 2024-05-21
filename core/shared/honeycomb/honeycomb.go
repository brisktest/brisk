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

package honeycomb

import (
	"brisk-supervisor/shared/constants"
	"context"
	"os"

	. "brisk-supervisor/shared/logger"

	"github.com/honeycombio/honeycomb-opentelemetry-go"
	_ "github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-launcher-go/launcher"
	sdk_trace "go.opentelemetry.io/otel/sdk/trace"
)

var sampleRate float64 = 1.0

func InitTracer() func() {

	bsp := honeycomb.NewBaggageSpanProcessor()
	// use honeycomb distro to setup OpenTelemetry SDK
	metricsKey := os.Getenv("HONEYCOMB_API_KEY")
	if metricsKey == "" {
		metricsKey = constants.HONEYCOMB_API_KEY
	}
	service_name := os.Getenv("OTEL_SERVICE_NAME")
	if service_name == "" {
		service_name = "brisk-cli"
	}
	// dataset_name := "brisk"
	shutdown, err := launcher.ConfigureOpenTelemetry(
		launcher.WithSpanProcessor(bsp), honeycomb.WithApiKey(metricsKey), launcher.WithServiceName(service_name), launcher.WithSampler(sdk_trace.TraceIDRatioBased(sampleRate)),
	)

	if err != nil {
		panic(err)
	}
	shutdownFunc = shutdown
	return shutdown

}
func SetSampleRate(sr float64) {
	sampleRate = sr
}

var shutdownFunc func()

func ShutdownTracer() {
	if shutdownFunc != nil {
		Logger(context.Background()).Info("Shutting down tracer")
		shutdownFunc()
	} else {
		panic("Shutdown called before init")
	}

}
