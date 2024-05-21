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

	"github.com/spf13/viper"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthChecker struct{}

func (s *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if viper.GetBool("LOG_HEALTH_CHECKS") {
		Logger(ctx).Debug("Serving the Check request for health check")
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	ctx := server.Context()
	if viper.GetBool("LOG_HEALTH_CHECKS") {

		Logger(ctx).Debug("Serving the Watch request for health check")
	}
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

func (s *HealthChecker) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
