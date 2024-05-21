#!/bin/sh

# WARNING! Parts of this script are destructive- make sure you check the bits that delete stuff...
# Heavily stolen from 
# https://scriptingosx.com/2021/07/notarize-a-command-line-tool-with-notarytool/
# With thanks to Armin Briegel

# loads of unneccesary checks included, strip these out when ready...

# Exit script if anything fails. This can be commented out when working properly
set -e

# Specify your variables here- sample text included for reference
# this script assumes you are installing a command line tool in /usr/local edit line 140 if this is not true
version=$CLI_RELEASE_VERSION
author="Brisk"
project="Brisk CLI"
identifier="com.brisk.cli"
productname="brisk"
rawbinary="brisk-temp"
icon="assets/icon-standard-color-rgb.jpg"

path_to_executable=$1
latest_path=$2
executable=$path_to_executable/brisk
echo "notarizing $executable"

# Apple Developer account email address
dev_account="apple@brisktest.com"
# Name of keychain item containing app password for signing
dev_keychain_label="gon"
# Name of file containing security entitlements- 
# this file must be in the build folder, signing the binary also attaches these entitlements
entitlements="myproject.entitlements"
# Signature to use for building installer package
signature="Developer ID Installer: Sean Reilly (927A7F8X2M)"
Developer_ID_Application="Developer ID Application: Sean Reilly (927A7F8X2M)"

# Main project directory- the location of this script!
projectdir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
echo $projectdir
#Location where the raw download, icon file, info.plist and entitlements file must exist                        
builddir="$projectdir/codesign"

echo "cleaning up previous runs"

rm -rf $builddir
mkdir -pv $builddir
# build root- this is location of signed binary to build into pkg

# Location where our .pkg will be saved for notarisation
pkgpath="$builddir/brisk-$version.pkg"


# let's make sure we are in the correct directory
cd $builddir
echo we are in $builddir

# Cleaning out from previous runs
echo Cleaning out from previous runs- WARNING- destructive!

rm -rf codesign


echo "copying $executable to $rawbinary"
cp ../$executable ./brisk


echo "codesigning brisk"
# Codesign Binary
# make sure entitlements file is in the raw binary folder

echo "codesign --deep --force --options=runtime --sign \"$Developer_ID_Application\" --timestamp ./brisk"

codesign --deep --force --options=runtime --sign "$Developer_ID_Application" --timestamp ./brisk
if  [ $? -eq 0 ]
then
     echo signing $binary Package Succeeded
else
     echo signing $binary Package Failed
fi

echo "Finished codesigning $productname"

# copy the signed binary into the folder structure to mirror install location

mkdir -pv "$projectdir/codesign"

 
# mv "$projectdir/codesign/brisk" "$projectdir/codesign/brisk"

find "$projectdir/codesign/" -name '*.DS_Store' -type f -delete
chmod -R +x "$projectdir/codesign/"
chmod -R +x "$projectdir/codesign/brisk"
chmod 555 "$projectdir/codesign/brisk"

#pkgbuild --root 
# "/Users/sean/Programming/brisk-supervisor/brisk-cli/codesign/"  
# --identifier "com.brisk.cli" 
# --install-location "/usr/local/bin/" 
#  --sign "$signature" 
#  /Users/sean/Programming/brisk-supervisor/brisk-cli/codesign/brisk-22.pkg

pkgbuild --root "$projectdir/codesign/" \
         --identifier "$identifier" \
         --version "$version" \
         --install-location "/usr/local/bin/" \
         --sign "$signature" \
        $projectdir/codesign/brisk-$version.pkg

# Build into a .pkg and Sign the bundle 
if  [ $? -eq 0 ]
then
     echo pkgbuild Succeeded
else
     echo pkgbuild Failed
fi

echo "attemping to notarize $builddir"/brisk-$version.pkg
# upload for notarization
#notarizefile "$pkgpath" "$identifier"
xcrun notarytool submit $builddir"/brisk-$version.pkg" \
                   --keychain-profile "Dev Installer"\
                   --wait


# staple result
echo "## Stapling $pkgpath"
xcrun stapler staple "$pkgpath"

echo '## Done!'
echo lets check our work 
spctl --assess -vv --type install $pkgpath

cp $pkgpath $projectdir/$path_to_executable/

# we copy it to the latest path so that the latest path is always the latest version
cp $pkgpath $projectdir/$latest_path
# show the pkg in Finder
open -R "$pkgpath"

exit 0

fi