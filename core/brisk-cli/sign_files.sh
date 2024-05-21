set -e
export BRISK_BUILT_FILE_WITH_PATH=$1
export BRISK_ARCH=$2
# envsubst < gon-sign.hcl > ./gon-sign.hcl.tmp.hcl

# gon -log-level=debug  ./gon-sign.hcl.tmp.hcl
# rm gon-sign.hcl.tmp.hcl
# echo "done signing files"

BUNDLE_ID=com.brisk.cli
DEVELOPER_ID="Developer ID Application: Sean Reilly (927A7F8X2M)"
#--entitlements "<path/to/file.entitlements>"
codesign -s "$DEVELOPER_ID" -f --timestamp -o runtime -i "<$BUNDLE_ID>"  $BRISK_BUILT_FILE_WITH_PATH
xcrun notarytool submit "$BRISK_BUILT_FILE_WITH_PATH" --keychain-profile "gon" --wait

