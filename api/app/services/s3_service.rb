# dealing with s3
class S3Service
  def initialize
    @s3 = Aws::S3::Resource.new(region: 'us-east-1',
                                credentials: Aws::Credentials.new(
                                  ENV['S3_AWS_ACCESS_KEY_ID'], ENV['S3_AWS_SECRET_ACCESS_KEY']
                                ))
  end

  def get_bucket(bucket_name)
    @s3.bucket(bucket_name)
  end

  def get_object(bucket_name, key)
    bucket = get_bucket(bucket_name)
    bucket.object(key)
  end

  def get_object_content(bucket_name, key)
    obj = get_object(bucket_name, key)
    obj.get.body.read
  end

  def get_presigned_url(bucket_name, key, expires_in = 1200)
    obj = get_object(bucket_name, key)

    # url = obj.presigned_url(:get, expires_in: expires_in, whitelist_headers: ["x-amz-server-side-encryption-customer-algorithm", "x-amz-server-side-encryption-customer-key"])
    # url += "&x-amz-server-side-encryption-customer-algorithm=AES-256&x-amz-server-side-encryption-customer-key=#{encryption_key}&x-amz-server-side-encryption-customer-key-MD5=#{md5_hash}"
    url, headers = obj.presigned_request(:get, expires_in:)
  end

  # curl https://brisktest-logs.s3.amazonaws.com/64e6280d-35db-4fd1-b38e-a7e07327759c\?X-Amz-Algorithm\=AWS4-HMAC-SHA256\&X-Amz-Credential\=AKIAVVUBO4WGYZ2QBIUE%2F20230225%2Fus-east-1%2Fs3%2Faws4_request\&X-Amz-Date\=20230225T005826Z\&X-Amz-Expires\=1200\&X-Amz-SignedHeaders\=host\&X-Amz-Signature\=79f400fdc5336273e63d0fd49f11cec5db656df57ca937c84bdca4e287161fef\&x-amz-server-side-encryption-customer-algorithm\=AES-256\&\x-amz-server-side-encryption-customer-key\=test
end
