set -e
# sh ./build.sh

#aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 054642462083.dkr.ecr.us-east-1.amazonaws.com


 docker build ./ -f docker/super.Dockerfile -t brisktest/brisk-super:latest &
# sudo docker build ./ -f docker/worker-rails.Dockerfile -t brisktest/brisk-worker-rails:latest & 
 docker build ./ -f docker/worker-node-lts.Dockerfile -t brisktest/brisk-worker-node-lts:latest & 
# sudo docker build ./ -f docker/monitor.Dockerfile -t brisktest/brisk-monitor:latest & 
# sudo docker build ./ -f docker/ssh.Dockerfile -t brisktest/brisk-ssh:latest & 
# sudo docker build ./ -f docker/ssh.Dockerfile -t brisktest/brisk-ssh:latest & 
# sudo docker build ./ -f docker/image-sync.Dockerfile -t brisktest/image-sync:latest &
wait
# sudo docker push brisktest/image-sync:latest &
 docker push brisktest/brisk-super:latest &
 docker push brisktest/brisk-worker-node-lts:latest & 
# sudo docker push brisktest/brisk-worker-rails:latest & 
# sudo docker push brisktest/brisk-monitor:latest & 
# sudo docker push brisktest/brisk-ssh:latest &
wait
# sh /Users/sean/Programming/nomad/terraform/aws/env/us-east/redeploy_jobs.sh
