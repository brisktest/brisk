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

package utilities

import (
	constants "brisk-supervisor/shared/constants"
	. "brisk-supervisor/shared/logger"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"go.opentelemetry.io/otel"
)

var forceUpdater = &selfupdate.Updater{
	CurrentVersion: constants.VERSION,
	ApiURL:         constants.UPDATE_SERVER_URL,
	BinURL:         constants.UPDATE_SERVER_URL,
	DiffURL:        constants.UPDATE_SERVER_URL,
	Dir:            "/tmp/updates/",
	CmdName:        "brisk", // app name
	ForceCheck:     true,
	Requester:      LoggingHttpRequester{},
}

var updater = &selfupdate.Updater{
	CurrentVersion: constants.VERSION,
	ApiURL:         constants.UPDATE_SERVER_URL,
	BinURL:         constants.UPDATE_SERVER_URL,
	DiffURL:        constants.UPDATE_SERVER_URL,
	Dir:            "/tmp/updates/",
	CmdName:        "brisk", // app name
	ForceCheck:     false,
	Requester:      LoggingHttpRequester{},
}

type LoggingHttpRequester struct {
}

func (r LoggingHttpRequester) Fetch(url string) (io.ReadCloser, error) {
	ctx := context.Background()
	ctx, span := otel.Tracer("updating").Start(ctx, "Fetch")
	defer span.End()
	fmt.Printf("Fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}
	b, err := httputil.DumpResponse(resp, false)
	if err != nil {
		Logger(ctx).Fatal(err)
	}
	Logger(ctx).Debugf("Response Status: %s", string(resp.Status))
	Logger(ctx).Debugf("Fetched %s", b)

	return resp.Body, nil

}

func DoUpdateNow(ctx context.Context) error {

	if forceUpdater != nil {

		fmt.Println("updating brisk")
		fmt.Printf("current version is %v \n", constants.VERSION)

		err := forceUpdater.BackgroundRun()
		if err != nil {
			fmt.Printf("Error is %v", err)
		}

	}
	return nil
}
func DoBackgroundUpdate(ctx context.Context) error {

	if updater != nil {

		Logger(ctx).Debug("updater initialized in background ", "current version: ", constants.VERSION)

		err := updater.BackgroundRun()
		if err != nil {
			fmt.Printf("Error is %v", err)
		}

	}
	return nil
}
