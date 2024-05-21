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

package consul

import (
	brisk "brisk-supervisor/api"
	. "brisk-supervisor/shared/logger"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

// This package deals with the consul api, used initially to register workers because it seems impossible to get the CNI weave details from
// nomad but the are available in consul.

var clientConn *api.Client

func client() (*api.Client, error) {
	if clientConn != nil {
		return clientConn, nil
	}

	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		clientConn = nil
		return nil, err
	}
	clientConn = cli
	return clientConn, nil
}

func GetRunningWorkers() ([]brisk.Worker, error) {

	startTime := time.Now()
	Logger(context.Background()).Info("Getting running workers from consul")
	c, err := client()
	if err != nil {
		return nil, err
	}
	service, _, err := c.Catalog().Service("worker", "", &api.QueryOptions{Token: os.Getenv("CONSUL_TOKEN")})
	if err != nil {
		return nil, err
	}
	var workers []brisk.Worker
	for _, v := range service {
		Logger(context.Background()).Debugf("The Worker service tags data from consul is %+v", v.ServiceMeta)
		Logger(context.Background()).Debugf("The Worker  data from consul is %+v", v)

		workers = append(workers, brisk.Worker{HostUid: v.ServiceMeta["host-uid"], IpAddress: v.ServiceAddress, Port: fmt.Sprintf("%v", v.ServicePort), Uid: v.ServiceMeta["alloc-id"], HostIp: v.TaggedAddresses["lan"], WorkerImage: v.ServiceMeta["worker-image"], SyncPort: v.ServiceMeta["sync-port"]})

	}
	Logger(context.Background()).Infof("Getting running workers from consul took %v", time.Since(startTime))
	return workers, nil
}
