Brisk
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
===

<img alt="Brisk Logo" src="https://github.com/brisktest/brisk/assets/405820/73e18bb2-0e11-4a8e-8465-b9a7fc9cbabc" width="100%" />


# Welcome

This repository contains the code required to run Brisk so you can host your own complete CI system. 

If you would prefer to use a hosted service instead of hosting your own CI system, check our hosted service at https://brisktest.com/.

Brisk is an extremely fast CI system, based around not rebuilding your environment on each test run. This allows us to really get the most from multiple workers. Instead of losing minutes rebuilding the environment on each run we instead can have the workers go straight to work running tests. This dramatically shortens the total time a test run takes. With enough workers the speed of your longest test becomes the limit for how long your CI tests take. 

# Getting Started

The root of this repository contains a docker-compose.yml file which has a simple single worker deployment of Brisk. It is suitable for testing locally and can be used as a starting point for deploying to production. 

Brisk consists of several services that are designed to be run across hundreds to thousands of machines. The docker-compose.yml contains the simplest possible deployment, one of each of the services. It is good for development and for testing locally.

To get started locally you can run 

```shell
docker compose up
```
# Deployment to Production

We recommend using a container orchestrator such as Kubernetes or Nomad to orchestrate your production deployment. We use Nomad internally but we also have examples which we can share for running k8s. Please contact us at support support@brisktest.com for more info and specific advice on your deployment. 


# The CLI

In order to access Brisk from your project directory (where the source code you are trying to test lives) you'll need to use the CLI. You can download a pre-built build from https://docs.brisktest.com/docs/installation or you can build the cli from this repository. 

In order to build the CLI you cd to 

```
cd core/brisk-cli
```

and execute ./build-debug to build a version for your system.

# Getting Started Using Brisk with your CI

Complete guides with information for setting up Brisk to work with your CI system and workflow are available at

https://docs.brisktest.com/

Examples include Github Actions, CircleCI, Bitbucket, AWS etc.

These docs cover using the CLI as a client and include everything about configuring brisk.json, using the CLI and setting up your build. 

We refer you to these docs for using the brisk client, for setting up and running the backend this repository is the main source of documentation. 

# Getting started contributing

- Checkout this repo locally
- run 
```
docker compose up
```
to start your local instance.
- run the following commands to seed the local db and create the dev user
```
docker exec -it brisk-api-1 rails db:prepare db:seed
```
- either install the brisk cli or build your own by running local version (see above)
- in a separate directory checkout out https://github.com/brisktest/react this will be your demo project
- cd to the react project we just checked out
- run 
```
BRISK_NO_BASTION=true BRISK_CONFIG_WARNINGS=true BRISK_APITOKEN=AfzWBMS8oy BRISK_APIKEY=dYho0h93lNfD/u/P  BRISK_DEV=true BRISK_APIENDPOINT=localhost:9001  brisk project init node
```
To create a node project in this directory
- run
```
BRISK_NO_BASTION=true BRISK_CONFIG_WARNINGS=true BRISK_APITOKEN=AfzWBMS8oy BRISK_APIKEY=dYho0h93lNfD/u/P  BRISK_DEV=true BRISK_APIENDPOINT=localhost:9001  brisk
```
to run your first test suite in brisk on your local machine.




# Brisk High Level Architecture

Brisk consists of a CLI program that talks to a an API and a dedicated supervisor (running in the cloud) which splits tests among many workers.
In production the CLI is the only component that runs locally (or in your CI Pipeline), everything else runs in the cloud. 




# DEV mode

There are a few reasons why you would like to run brisk with different deployment requirements. 

First off you could be contributing code to the project.

Second, you could be trying out the project to see if it is suitable for your use case and 

Finally you could be running an instance of the project in production, perhaps in a private cloud or on some other infra.

The docker-compose deployment (which relies on DEV mode) is suitable for the first two objectives but is UNSUITABLE for deployment to production. 

To make the local setup easier many important security features are disabled in DEV mode (such as tls certs, isolation of different parts of the system, many safety mechanisms and fallbacks). Please do not run the system in DEV mode in production. 

In order to safely run Brisk in production you need to turn DEV mode off and implement relevant security measures. At a minimum configure TLS with certificates. 

