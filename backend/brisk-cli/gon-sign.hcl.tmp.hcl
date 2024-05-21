source    = ["public/0.1.68-amd64/brisk"]
bundle_id = "com.brisk.cli"

apple_id {
  username = "apple@brisktest.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Sean Reilly (927A7F8X2M)"
}

dmg {
  output_path = "public/latest/brisk-amd64.dmg"
  volume_name = "Brisk"
}

// zip {
//   output_path = "public//brisk-amd64.zip"
// }