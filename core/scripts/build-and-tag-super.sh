#!/bin/bash
TAG_BIT="super-dev-build-$(echo $RANDOM | md5sum |head -c 6)"
export TAG="brisktest/brisk-super:$TAG_BIT"
sudo echo "building and tagging with $TAG"
set -e
sh ./build.sh
 docker build ./ -f docker/super.Dockerfile -t $TAG
 docker push $TAG

#source /Users/sean/Programming/nomad/terraform/aws/env/us-east/second-time
export NOMAD_VAR_SUPER_IMAGE=$TAG
echo "NOMAD_VAR_SUPER_IMAGE=$TAG"
echo "DEV DEV DEV DEV"
source ~/.brisktesting-dev-env

nomad run ~/Programming/brisk-supervisor/nomad/deploy_super_task.nomad.hcl
