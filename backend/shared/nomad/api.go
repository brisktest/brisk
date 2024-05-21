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

package nomad

import (
	"brisk-supervisor/api"
	"brisk-supervisor/shared/constants"
	. "brisk-supervisor/shared/logger"
	"context"
	"errors"
	"os"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
	"go.opentelemetry.io/otel"
)

func Client(ctx context.Context) (*nomadapi.Client, error) {
	config := nomadapi.DefaultConfig()

	config.Address = os.Getenv("NOMAD_ADDR")
	config.Region = "global"
	config.WaitTime = time.Duration(2) * time.Second
	Logger(ctx).Debugf("NOMAD_ADDR is %s \n", os.Getenv("NOMAD_ADDR"))
	return nomadapi.NewClient(config)

}

func GetNodes(ctx context.Context, client *nomadapi.Client) ([]*nomadapi.Node, error) {
	ctx, span := otel.Tracer("nomad").Start(ctx, "GetNodes")
	defer span.End()
	nodes, _, err := client.Nodes().List(&nomadapi.QueryOptions{AllowStale: true, AuthToken: os.Getenv("NOMAD_TOKEN")})

	if err != nil {
		Logger(ctx).Errorf("Error in GetNodes: %v", err)
		return nil, err
	}

	var actual_nodes []*nomadapi.Node
	for _, node := range nodes {
		Logger(ctx).Debugf("Node is %+v \n", node)
		if node.Status == "ready" {
			// get the node from the NodeStub
			n, _, err := client.Nodes().Info(node.ID, &nomadapi.QueryOptions{AllowStale: true, AuthToken: os.Getenv("NOMAD_TOKEN")})
			if err != nil {
				Logger(ctx).Errorf("Error in GetNodes inner: %v", err)
				return nil, err
			}
			Logger(ctx).Debugf("Node to be registered is %+v \n", n)
			actual_nodes = append(actual_nodes, n)
		}
	}

	return actual_nodes, err
}

var NotFound = errors.New("Unexpected response code: 404 (alloc not found)")

func UpdateWorkerRegistry(ctx context.Context, client *nomadapi.Client, deRegisterWorkerFunc func(ctx context.Context, w *api.Worker) error) error {
	ctx, span := otel.Tracer("nomad").Start(ctx, "UpdateWorkerRegistry")
	defer span.End()
	params := map[string]string{"resources": "true"}
	a, meta, err := client.Allocations().List(&nomadapi.QueryOptions{AllowStale: true, Params: params, AuthToken: os.Getenv("NOMAD_TOKEN")})
	Logger(ctx).Debugf("Meta is %+v \n", meta)
	if err != nil {
		Logger(ctx).Errorf("Error in UpdateWorkerRegistry: %v", err)
		return err
	}
	Logger(ctx).Debugf("Allocations are %+v \n", a)
	for _, a := range a {
		if a == nil {
			continue
		}
		switch a.JobID {
		case constants.NOMAD_DEPLOY_WORKER, constants.NOMAD_DEPLOY_WORKER_NODE, constants.NOMAD_DEPLOY_WORKER_RAILS, constants.NOMAD_DEPLOY_WORKER_PYTHON, constants.NOMAD_DEPLOY_WORKER_RAW:
			if a == nil {
				Logger(ctx).Debugf("SOMEHOW a is nil")
			}
			if a != nil {
				Logger(ctx).Debugf("Job  is ", a.JobID)
			}

			if a != nil && a.DeploymentStatus != nil && a.ClientStatus == "failed" {
				Logger(ctx).Debugf("The deployment status is %+v", *a.DeploymentStatus)
				deRegisterWorkerFunc(ctx, &api.Worker{Uid: a.ID})
			}
			if a != nil && a.ClientStatus == "complete" {
				deRegisterWorkerFunc(ctx, &api.Worker{Uid: a.ID})
			}

		default:
			Logger(ctx).Debug("NOP")

		}
	}
	return err
}

func CheckWorkerNomad(ctx context.Context, worker api.Worker) (bool, error) {
	ctx, span := otel.Tracer("nomad").Start(ctx, "CheckWorkerNomad")
	defer span.End()
	client, err := Client(ctx)
	if err != nil {
		return false, err
	}

	a, _, err := client.Allocations().Info(worker.Uid, &nomadapi.QueryOptions{AllowStale: true, AuthToken: os.Getenv("NOMAD_TOKEN")})
	if err == NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if a.DeploymentStatus == nil {
		return false, errors.New("no deployment status found")
	}
	return *a.DeploymentStatus.Healthy, err
}

func CheckSuperNomad(ctx context.Context, super api.Super) (bool, error) {
	ctx, span := otel.Tracer("nomad").Start(ctx, "CheckSuperNomad")
	defer span.End()
	client, err := Client(ctx)
	if err != nil {
		return false, err
	}

	a, _, err := client.Allocations().Info(super.Uid, &nomadapi.QueryOptions{AllowStale: false, AuthToken: os.Getenv("NOMAD_TOKEN")})
	if err == NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return *a.DeploymentStatus.Healthy, err
}

func StopAllocation(ctx context.Context, worker api.Worker) error {
	ctx, span := otel.Tracer("nomad").Start(ctx, "StopAllocation")
	defer span.End()
	client, err := Client(ctx)
	if err != nil {
		return err
	}
	alloc := nomadapi.Allocation{ID: worker.Uid}
	a, err := client.Allocations().Stop(&alloc, &nomadapi.QueryOptions{AllowStale: true, AuthToken: os.Getenv("NOMAD_TOKEN")})
	if err == NotFound {
		return err
	}
	if err != nil {
		return err
	}

	Logger(ctx).Debug("Stop Allocation Response is %+v", a)
	return nil
}

func SubscribeToClientEventStream(ctx context.Context, client *nomadapi.Client, registerClientFunc func(context.Context, *nomadapi.Node) error, deRegisterClientFunc func(ctx context.Context, m *nomadapi.Node) error, drainClientFunc func(ctx context.Context, m *nomadapi.Node) error) error {
	ctx, span := otel.Tracer("nomad").Start(ctx, "SubscribeToClientEventStream")
	defer span.End()
	Logger(ctx).Debug("Subscribing to Nomad Client Event Stream")
	topics := map[nomadapi.Topic][]string{
		// nomadapi.TopicNode: {"*"},
		nomadapi.TopicNode: {"*"},
	}

	es := client.EventStream()

	eventChan, err := es.Stream(ctx, topics, 0, &nomadapi.QueryOptions{AllowStale: false, AuthToken: os.Getenv("NOMAD_TOKEN")})

	if err != nil {
		Logger(ctx).Errorf("Error in SubscribeToClientEventStream: %v", err)
		return err
	}
	Logger(ctx).Debug("About to start dealing with event chan ")
	for events := range eventChan {
		// Logger(ctx).Debug("Got events %+v", events)
		for _, e := range events.Events {
			Logger(ctx).Debugf("Event is %+v", e)
			m, nodeErr := e.Node()
			if nodeErr != nil {
				Logger(ctx).Error(nodeErr)
				Logger(ctx).Error("breaking")
				break
			}
			switch e.Type {

			case "NodeRegistration":
				Logger(ctx).Debugf("NodeRegistration is %+v", e.Payload)
				Logger(ctx).Debugf("Just registered node is %+v", m)

				registerClientFunc(ctx, m)
			case "NodeDeregistration":
				Logger(ctx).Debugf("NodeDeregistration is %+v", e.Payload)
				Logger(ctx).Debugf("Just deregistered node is %+v", m)
				deRegisterClientFunc(ctx, m)
			case "NodeDrain":
				Logger(ctx).Debugf("NodeDrain is %+v", e.Payload)
				drainClientFunc(ctx, m)
			case "NodeEvent":
				Logger(ctx).Debugf("NodeEvent is %+v", e.Payload)
				Logger(ctx).Debugf("Noop - Just evented node is %+v", m)
			case "NodeEligibility":
				Logger(ctx).Debugf("NodeEligibility is %+v", e.Payload)
				Logger(ctx).Debugf("Noop - Just evented node is %+v", m)
			default:
				// Logger(ctx).Debugf("NOP for %v  is %+v", e.Type, e.Payload)
				Logger(ctx).Debugf("NOP for %v  ", e.Type)
			}
		}
	}
	return err
}

// add a func for register deregister and do it for both jobs?
func SubscribeToEventStream(ctx context.Context, client *nomadapi.Client, deRegisterWorkerFunc func(ctx context.Context, w *api.Worker) error) error {
	ctx, span := otel.Tracer("nomad").Start(ctx, "SubscribeToEventStream")
	defer span.End()
	topics := map[nomadapi.Topic][]string{
		//	nomadapi.TopicDeployment: {"Job:deploy_worker"}, nomadapi.TopicAllocation: {"Job:deploy_worker"}
		nomadapi.TopicAllocation: {constants.NOMAD_DEPLOY_WORKER,
			constants.NOMAD_DEPLOY_WORKER_NODE, constants.NOMAD_DEPLOY_WORKER_RAILS,
			constants.NOMAD_DEPLOY_WORKER_PYTHON, constants.NOMAD_DEPLOY_WORKER_RAW},
	}
	es := client.EventStream()

	eventChan, err := es.Stream(ctx, topics, 0, &nomadapi.QueryOptions{AllowStale: false, AuthToken: os.Getenv("NOMAD_TOKEN")})
	if err != nil {
		return err
	}
	for events := range eventChan {

		for _, e := range events.Events {
			alloc, err := e.Allocation()
			if err != nil {
				Logger(ctx).Debug("Expected an allocations got %+v", e)
			} else {
				Logger(ctx).Debug("-------------------------")

				Logger(ctx).Debugf("Alloc is %+v", alloc)
				Logger(ctx).Debugf("TaskGroup is %+v", alloc.TaskGroup)
				Logger(ctx).Debugf("TaskResources is %+v", alloc.TaskResources)
				Logger(ctx).Debugf("Resources is %+v", alloc.Resources)
				Logger(ctx).Debugf("Resources is %+v", alloc.Services)
				Logger(ctx).Debugf("TaskResources worker %+v", alloc.TaskResources["worker"])
				//	spew.Dump(alloc)
				switch alloc.ClientStatus {
				case nomadapi.AllocClientStatusRunning:
					Logger(ctx).Debugf("Alloc %v is running", alloc.ID)
					Logger(ctx).Debugf("Alloc %+v \n", alloc)
				//	registerWorkerFunc(types.WORKER_PORT, alloc.ID)
				//register
				case nomadapi.AllocClientStatusComplete:
					Logger(ctx).Debugf("Alloc %v is complete", alloc.ID)
					deRegisterWorkerFunc(ctx, &api.Worker{Uid: alloc.ID})
					//deregister
				case nomadapi.AllocClientStatusFailed:
					Logger(ctx).Debug("hey")
					Logger(ctx).Debugf("Alloc %v is failed", alloc.ID)
					deRegisterWorkerFunc(ctx, &api.Worker{Uid: alloc.ID})

				//deregister
				case nomadapi.AllocClientStatusPending:
					Logger(ctx).Debugf("Alloc %v is pending", alloc.ID)

				default:
					Logger(ctx).Debugf("Don't know this status : %+v", alloc.ClientStatus)
				}

			}
		}
	}

	return err
}
