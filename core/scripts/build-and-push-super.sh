sudo echo 'building'
set -e
sh ./build.sh
 docker build ./ -f docker/super.Dockerfile -t brisktest/brisk-super:latest 
 docker push brisktest/brisk-super:latest

#source /Users/sean/Programming/nomad/terraform/aws/env/us-east/second-time
export NOMAD_VAR_SUPER_IMAGE="brisktest/brisk-super:latest"
echo "DEV DEV DEV DEV"
source ~/.brisktesting-dev-env

nomad run ~/Programming/brisk-supervisor/nomad/deploy_super_task.nomad.hcl