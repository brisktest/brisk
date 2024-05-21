class LockManager
  def self.redis_lock(lock_name, timeout = 10)
    REDIS_POOL.with do |conn|
      conn.lock_for_update(lock_name, timeout) do
        Rails.logger.info "Acquired redis lock #{lock_name}"
        yield
      end
    end
  end

  # watch out with these transaction locks - if we have long transactions it fucks everything up

  def self.with_transaction_lock(lock_name, timeout = 30)
    lock_id = Zlib.crc32(lock_name.to_s)
    # having this txn here breaks shit in the caller - I assume we have a txn somewhere...
    ActiveRecord::Base.transaction do
      ActiveRecord::Base.connection.execute("SET LOCAL lock_timeout = '#{timeout}s'")
      ActiveRecord::Base.connection.execute("SELECT pg_advisory_xact_lock(#{lock_id})")
      Rails.logger.info "Acquired transaction lock #{lock_name}"
      yield
    end
    Rails.logger.info "Released transaction lock #{lock_name}"
  end

  def self.with_session_lock(lock_name)
    lock_id = Zlib.crc32(lock_name.to_s)
    begin
      ActiveRecord::Base.connection.execute("SELECT pg_advisory_lock(#{lock_id})")
      Rails.logger.info "Acquired session lock #{lock_name}"
      yield
    ensure
      ActiveRecord::Base.connection.execute("SELECT pg_advisory_unlock(#{lock_id})")
      Rails.logger.info "Released session lock #{lock_name}"
    end
  end
end
