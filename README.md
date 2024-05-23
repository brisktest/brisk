# Brisk

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1262.62 325.05"><defs><style>.cls-1{fill:#5389e3;}</style></defs><g id="Layer_2" data-name="Layer 2"><g id="Layer_1-2" data-name="Layer 1"><path d="M318.32,70.85h94.84c20.89,0,33.53,3,43.42,9.9C468.13,89,474.73,103,474.73,118.41c0,19.51-9.9,35.18-27.77,42.88,19.52,7.42,29.14,21.44,29.14,42.33,0,19-8.24,34.63-22.26,42.61-9.62,5.49-23.64,8-45.91,8H318.32Zm89.61,76.69c11.55,0,16.77-1.1,21.72-4.94,5.77-4.68,9.07-12.1,9.07-20.34,0-7.42-3-14.57-7.7-18.7-5.22-4.12-9.89-5.22-22.81-5.22H351.86v49.2Zm-1.37,79.17c22.27,0,33-8.79,33-26.94,0-17.32-9.34-24.74-31.33-24.74H351.86v51.68Z"/><path d="M577.84,254.2H544.3V70.85h86c25.84,0,41.78,3,51.95,9.9,13.2,8.79,20.89,23.91,20.89,41.23,0,15.4-5.49,29.41-14.29,38.21-6.33,5.77-11.55,8.79-22.27,12.09,22.82,8,31.89,22.82,33.53,53.88,1.38,19.51,1.66,21.17,4.68,28H668.55a37.37,37.37,0,0,1-2.2-8c0-1.65-.55-4.95-.83-9.62l-.82-8.8c-3.3-36-17.59-48.1-56.62-47.83H577.84Zm53.32-101.71c13.2,0,20.89-1.37,25.57-5,6.59-4.67,9.89-12.36,9.89-22.54,0-11.26-4.39-19.79-12.37-23.64-4.12-2.19-10.44-3-23.09-3H577.84v54.15Z"/><path d="M772.47,70.85H806V254.2H772.47Z"/><path d="M881.06,226.71h81.09c14,0,21.44-1.1,26.66-4.4,7.15-4.12,11-11.27,11-20.89,0-11-4.95-19.52-13.74-23.36-5-1.93-11.82-2.75-24.75-2.75H937.68c-21.71,0-35.73-3-45.08-9.9-12.09-8.52-19-23.09-19-39.85,0-21.17,10.45-39.59,27.21-48.11,9.9-4.67,22-6.6,43.71-6.6H1027V98.34H948.68c-14,0-21.44,1.1-27.21,3.58-6.88,3.29-11.27,11.26-11.27,20.88,0,16.77,10.17,23.37,36.83,23.37h19.52c28.31,0,42.33,3,52.78,11.82,10.71,8.8,17,25,17,42.88,0,20.34-8.8,37.11-23.92,45.63-9.89,6.05-21.16,7.7-48.37,7.7h-83Z"/><path d="M1137.55,254.2H1104V70.85h33.53Zm81.09-183.35h42.61l-85.49,89.34,86.86,94h-44.8L1138.38,161Z"/><polygon class="cls-1" points="199.2 140.88 70.91 140.88 167.23 0 0 0 0 164.41 86.25 164.41 86.25 325.05 199.2 140.88"/></g></g></svg>![logo-standard-color-rgb](https://github.com/brisktest/brisk/assets/405820/df3828eb-ac25-4ee4-8de9-c1477c468d84)


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

