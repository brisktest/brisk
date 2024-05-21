#!/bin/sh
set -e
# Create user
#useradd myservice

# Setup config
#sed -i "s/PASSWD/${SERVICE_PASSWORD}/g" /etc/myservice.conf
#sudo service ssh start
# Start service
exec "$@"