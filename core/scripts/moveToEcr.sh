set -e
aws ecr get-login-password --region us-east-1 | sudo docker login --username AWS --password-stdin 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk

docker login --username "brisktest" --password $DOCKER_HUB_PASSWD
sudo docker pull brisktest/brisk-worker:latest
sudo docker pull brisktest/brisk-super:latest
sudo docker pull brisktest/brisk-ssh:latest
sudo docker pull brisktest/brisk-monitor:latest

sudo docker tag    brisktest/brisk-worker:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker:latest
sudo docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-worker:latest

sudo docker tag    brisktest/brisk-super:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-super:latest
sudo docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-super:latest

sudo docker tag    brisktest/brisk-ssh:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-ssh:latest
sudo docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-ssh:latest

sudo docker tag    brisktest/brisk-monitor:latest 561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-monitor:latest
sudo docker push  561398244478.dkr.ecr.us-east-1.amazonaws.com/brisk/brisk-monitor:latest


