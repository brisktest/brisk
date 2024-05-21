#!/bin/sh
set -e
# Create user
#useradd myservice

# Setup config
#sed -i "s/PASSWD/${SERVICE_PASSWORD}/g" /etc/myservice.conf
echo "Entrypointing..."
/usr/sbin/sshd  -D -e 
# Start service
#exec "$@"