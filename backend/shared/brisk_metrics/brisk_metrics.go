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

package brisk_metrics

import (
	. "brisk-supervisor/shared/logger"
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

func StartPrometheusServer(ctx context.Context) {

	Logger(ctx).Info("Starting Prometheus metrics server")
	http.Handle("/metrics", promhttp.Handler())
	Logger(ctx).Info("Health check initialized")
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if viper.GetBool("LOG_HEALTH_CHECKS") {
			Logger(ctx).Debug("Responding to health check")
		}
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		http.ListenAndServe(":2112", nil)
	}()

}
