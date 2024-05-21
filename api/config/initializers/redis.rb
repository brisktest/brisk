pool_size = ENV.fetch('MAX_REDIS_CONNECTIONS') { 1 }
url = ENV.fetch('REDIS_URL') { 'redis://127.0.0.1:6379/1' }

REDIS_POOL = ConnectionPool.new(size: pool_size) do
  Redis.new(url:)
end
