set -u
set -e
if [[ -z "$CLI_RELEASE_VERSION" ]]; then
    echo "Must provide CLI_RELEASE_VERSION in environment" 1>&2
    exit 1
fi
mkdir -p public/latest/linux-amd64/ 
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 sh -c 'go build -o public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH main.go'
# GOOS=darwin GOARCH=amd64  sh -c 'cp public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH public/latest/$GOOS-$GOARCH/brisk'

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 sh -c 'go build -o public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH main.go'
# GOOS=darwin GOARCH=arm64 sh -c 'cp public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH public/latest/$GOOS-$GOARCH/brisk'
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 sh -c 'go build -o public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH main.go'
#only need to copy the linux version to the latest because the signing process outputs the other ones
GOOS=linux GOARCH=amd64 sh -c 'cp public/$CLI_RELEASE_VERSION/$GOOS-$GOARCH public/latest/$GOOS-$GOARCH/brisk'

echo "copied linux version to latest"
# ls -lrt  public/latest/linux-amd64/brisk
mkdir -p public/$CLI_RELEASE_VERSION-amd64
mkdir -p public/$CLI_RELEASE_VERSION-arm64
cp public/$CLI_RELEASE_VERSION/darwin-amd64 public/$CLI_RELEASE_VERSION-amd64/brisk
cp public/$CLI_RELEASE_VERSION/darwin-arm64 public/$CLI_RELEASE_VERSION-arm64/brisk
