# Welcome

This repository contains the code required to run Brisk so you can host your own complete CI system.

There are a few reasons why you would like to do that with different deployment requirements. 

First off you could be contributing code to the project
Second, you could be trying out the project to see if it is suitable for your use case
and finally you could be running an instance of the project in production, perhaps in a private cloud or on some other infra.

The docker-compose deployment is suitable for the first two objectives but is UNSUITABLE for deployment to production. 

To make the local setup easier many important security features are disabled in DEV mode (such as tls certs, isolation of different parts of the system, many safety mechanisms and fallbacks). Please do not run the system in DEV mode in production. 


