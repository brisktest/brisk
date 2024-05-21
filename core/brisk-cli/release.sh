set -e

source ~/.brisktesting-prod-env-language
# need the different AWS credentials for for S3 and Cloudfront
source ~/.brisk-env

# export BRISK_BUILT_FILE_WITH_PATH="public/$RELEASE_VERSION/brisk"
source ./cli-release-version.sh
echo "building release $CLI_RELEASE_VERSION"
./build-release.sh 
echo "done building release $CLI_RELEASE_VERSION"
echo "running selfupdate"
go-selfupdate public/$CLI_RELEASE_VERSION $CLI_RELEASE_VERSION

chmod 555 public/$CLI_RELEASE_VERSION-amd64/brisk*
chmod 555 public/$CLI_RELEASE_VERSION-arm64/brisk*
echo "notarizing release $CLI_RELEASE_VERSION"
sh notarize_files.sh public/$CLI_RELEASE_VERSION-amd64/ public/latest/brisk-amd64.pkg
sh notarize_files.sh public/$CLI_RELEASE_VERSION-arm64/ public/latest/brisk-arm64.pkg


echo "uploading release $CLI_RELEASE_VERSION to s3"

aws s3 sync public/ s3://brisk-releases/updates/brisk --acl public-read --exclude "*" --include "*.json"
aws s3 sync public/ s3://brisk-releases/updates/brisk --acl public-read --size-only

aws cloudfront create-invalidation \
    --distribution-id EU9CJ8A3EX8XN \
    --paths "/brisk/latest/*"
