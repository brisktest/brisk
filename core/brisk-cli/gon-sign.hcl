source    = ["$BRISK_BUILT_FILE_WITH_PATH"]
bundle_id = "com.brisk.cli"

apple_id {
  username = "apple@brisktest.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Sean Reilly (927A7F8X2M)"
}

dmg {
  output_path = "public/latest/brisk-$BRISK_ARCH.dmg"
  volume_name = "Brisk"
}

// zip {
//   output_path = "public/$RELEASE_VERSION/brisk-$BRISK_ARCH.zip"
// }