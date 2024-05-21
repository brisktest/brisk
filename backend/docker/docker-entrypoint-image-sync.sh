#!/bin/sh
set -e
aws ecr get-login-password --region us-east-1 |  docker login --username AWS --password-stdin 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk
docker login --username "brisktest" --password $DOCKER_HUB_PASSWD

docker pull brisktest/brisk-worker-rails:latest
docker pull brisktest/brisk-worker-node-lts:latest
docker pull brisktest/brisk-super:latest
docker pull brisktest/brisk-ssh:latest
docker pull brisktest/brisk-monitor:latest
docker pull brisktest/dockerdind:latest


docker tag    brisktest/brisk-worker-rails:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker-rails:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker-rails:latest

docker tag    brisktest/brisk-worker-node-lts:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker-node-lts:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker-node-lts:latest

docker tag    brisktest/brisk-super:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-super:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-super:latest

docker tag    brisktest/brisk-ssh:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-ssh:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-ssh:latest

docker tag    brisktest/brisk-monitor:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-monitor:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-monitor:latest



docker tag    brisktest/dockerdind:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/dockerdind:latest
docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/dockerdind:latest