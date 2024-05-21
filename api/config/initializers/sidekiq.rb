require 'sidekiq-scheduler'

Rails.logger.info "Starting #{Rails.env} environment ininitializer"
Rails.logger.info "Keys from credentials: #{Rails.application.credentials.keys}"
# REDIS_URL = Rails.application.credentials.redis_url
