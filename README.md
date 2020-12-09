# Brisk

# Installing

npm i @brisk-test/brisk
or
yarn add @brisk-test/brisk

# Running

(you may need to prefix these commands with ./node_modules/.bin/ depending on where you have installed brisk)

### brisk-setup-keys

Run this command to add ssh keys for the hosts in brisk.json. If you already have connected to these hosts over ssh you can skip this step

### brisk-sync

This command syncs your filesystem with the remote hosts, leave this running and it will watch your directory and sync updates. Run in the project root.

### SETUP_REMOTE=true brisk

Run this the first time you use a new host to set things up

### VERBOSE_OUTPUT=true brisk

Each time you want to run your test suite, super fast.

# Configuration

Place a file called brisk.json in root of project. Here is an example.

```json
{
    "hosts": {
        "ssh://ubuntu@hostname": {},
        "ssh://ubuntu@hostname": {}
    },
    "pathToDockerFile": ".devcontainer/",
    "dockerContainer": "name-of-container",
    "dockerCommand": "docker-compose",
    "remoteWorkspace": "/home/ubuntu/name-of-directory",
    "testPattern": "**/__tests__/*.js",
    "testPathIgnorePatterns": "node_modules/**",
    "localPath": "./",
    "remoteEnv": [],
    "cmdArgs": ["npm run test "]
}
```

# Keys

### hosts

A map of hostnames to host options.

### pathToDockerFile

The path to the folder containing the docker file on the local machine.

### dockerContainer

Name of the docker container

### dockerCommand

The docker command to run, usually docker-compose

### remoteWorkspace

The path to the remote workspace, this is where we will map the workspace to on the remote hosts. This must match the left hand side of the volumes command in your docker-compose.yml

### testPattern

The glob pattern to match your tests

### testPathIgnorePatterns

A glob pattern to be ignored when matching tests

### localPath

### remoteEnv

Environment variables that should be set on the remote host

### cmdArgs

The test command to run

## Servers

Your hosts need to be running docker. Also you should have write priveleges to the remoteWorkspace so you can sync your filesystem.

## Sample docker-compose.yml

### From the React project

```yml
version: '3'
services:
    react:
        # Uncomment the next line to use a non-root user for all processes.
        # See https://aka.ms/vscode-remote/containers/non-root for details.
        # user: node

        build:
            context: .
            dockerfile: Dockerfile
            args:
                # On Linux, you may need to update USER_UID and USER_GID below if not your local UID is not 1000.
                USER_UID: 1000
                USER_GID: 1000

        volumes:
            - /home/ubuntu/react:/workspace:cached

        # Overrides default command so things don't shut down after the process ends.
        command: sleep infinity
```

## Sample Dockerfile

```dockerfile
FROM node:14.5.0-stretch

# Update args in docker-compose.yaml to set the UID/GID of the "node" user.
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN if [ "$USER_GID" != "1000" ] || [ "$USER_UID" != "1000" ]; then \
        groupmod --gid $USER_GID node \
        && usermod --uid $USER_UID --gid $USER_GID node \
        && chmod -R $USER_UID:$USER_GID /home/node \
        && chmod -R $USER_UID:root /usr/local/share/nvm /usr/local/share/npm-global; \
    fi

# [Optional] Uncomment this section to install additional OS packages.

RUN apt-get update && apt-get -y install mongodb
# [Optional] Uncomment if you want to install an additional version of node using nvm
# ARG EXTRA_NODE_VERSION=10
# RUN su node -c "source /usr/local/share/nvm/nvm.sh && nvm install ${EXTRA_NODE_VERSION}"

# [Optional] Uncomment if you want to install more global node packages
# RUN sudo -u node npm install -g <your-package-list-here>


```
