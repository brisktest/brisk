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

package mixpanel

import (
	"brisk-supervisor/shared/constants"
	"context"
	"errors"
	"os"
	"runtime"

	"golang.org/x/exp/maps"
	"vizzlo.com/mixpanel"
)

var client *mixpanel.Client

// InitMixpanelClient initializes the mixpanel client
func InitMixpanelClient() {
	client = mixpanel.New("df244eb3f4976afb3ba776c9e0955d3c")
}

func track(event string, second string, properties map[string]interface{}) error {
	if client == nil {
		return errors.New("mixpanel client is not initialized")
	}

	err := client.Track(event, second, properties)
	return err
}

func TrackIfEnabled(ctx context.Context, event string, second string, properties map[string]interface{}) error {
	if client == nil {
		return nil
	}
	maps.Copy(properties, getAllProperties(ctx))

	err := track(event, second, properties)
	return err
}

// get the current OS
func getOS(ctx context.Context) string {
	return runtime.GOOS
}

func getArch(ctx context.Context) string {
	return runtime.GOARCH
}

func getAllProperties(ctx context.Context) map[string]interface{} {

	properties := make(map[string]interface{})

	properties["os"] = getOS(ctx)
	properties["arch"] = getArch(ctx)
	properties["version"] = constants.VERSION
	properties["trace-key"] = ctx.Value("trace-key")
	properties["args"] = os.Args
	return properties
}
