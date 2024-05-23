# Brisk
![logo-standard-color-rgb](https://github.com/brisktest/brisk/assets/405820/b15423d7-3704-4b3a-9b95-963d74e83a6e)



# Welcome

This repository contains the code required to run Brisk so you can host your own complete CI system. 

If you would prefer to use a hosted service instead of hosting your own CI system, check our hosted service at https://brisktest.com/.

Brisk is an extremely fast CI system, based around not rebuilding your environment on each test run. This allows us to really get the most from multiple workers. Instead of losing minutes rebuilding the environment on each run we instead can have the workers go straigth to work running tests. This dramatically shortens the total time a test run takes. With enough workers the speed of your longest test becomes the limit for how long your CI tests take. 

# Getting Started

The root of this repo contains a docker-compose.yml file which has a simple single worker deployment of Brisk. It is suitable for testing locally and can be used as a starting point for deploying to production. 

Brisk consists of several services that are designed to be run across hundreds to thousands of machines. The docker-compose.yml contains the simplest possible deployment, one of each of the services. 

To get started locally you can run 

```shell
docker compose up
```

# The CLI

In order to access Brisk from your project directory (where the source code you are trying to test lives) you'll need to use the CLI. You can download a prebuilt build from https://docs.brisktest.com/docs/installation or you can build the cli from this repo. 

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

We refer you to these docs for using the brisk client, for setting up and running the backend this repo is the main source of documentation. 

# Brisk High Level Architecture

Brisk consistes of a CLI program that talks to a an API and a dedicated supervisor (running in the cloud) which splits tests among many workers.
In production the CLI is the only componenet that runs locally (or in your CI Pipeline), everything else runs in the cloud. 

The following diagram shows a simple high level view of the architecture. In a production system we would expect more load balancers, bastion hosts, data stores and other infrastructure.

![brisk-arch-white drawio](https://github.com/brisktest/brisk/assets/405820/3ab5148b-49d6-4cb8-a48e-7eaae7174558)


# DEV mode

There are a few reasons why you would like to run brisk with different deployment requirements. 

First off you could be contributing code to the project.

Second, you could be trying out the project to see if it is suitable for your use case and 

Finally you could be running an instance of the project in production, perhaps in a private cloud or on some other infra.

The docker-compose deployment (which relies on DEV mode) is suitable for the first two objectives but is UNSUITABLE for deployment to production. 

To make the local setup easier many important security features are disabled in DEV mode (such as tls certs, isolation of different parts of the system, many safety mechanisms and fallbacks). Please do not run the system in DEV mode in production. 

In order to safely run Brisk in production you need to turn DEV mode off and implement relevant security measures. At a minimum configure TLS with certificates. 

