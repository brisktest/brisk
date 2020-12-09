# Brisk

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

The path to the remote workspace, this is where we will map the workspace to on the remote hosts.

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
