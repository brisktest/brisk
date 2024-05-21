# export RELEASE_VERSION=0.0.1
# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o release/mac/brisk$RELEASE_VERSION brisk-cli/main.go

# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o release/mac-arm/brisk$RELEASE_VERSION brisk-cli/main.go

# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o release/linux/brisk$RELEASE_VERSION brisk-cli/main.go

echo "not sure why we are using this"
error_exit()
{
    echo "$1" 1>&2
    exit 1
}