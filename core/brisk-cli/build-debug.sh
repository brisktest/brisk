set -u
CGO_ENABLED=0 sh -c 'go build -o brisk-cli main.go' # should just build whatever works for this machine