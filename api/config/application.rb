# frozen_string_literal: true

require_relative 'boot'

require 'rails/all'

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module BriskFrontend
  class Application < Rails::Application
    ENV['LOG_LEVEL'] ||= 'info'
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 7.0
    config.active_record.async_query_executor = :global_thread_pool
    config.active_record.global_executor_concurrency = 4
    config.active_record.logger = nil

    # config.assets.paths << Rails.root.join('app', 'assets', 'fonts')
    # Settings in config/environments/* take precedence over those specified here.
    # Application configuration can go into files in config/initializers
    # -- all .rb files in that directory are automatically loaded after loading
    # the framework and any gems in your application.
    config.active_job.queue_adapter = :sidekiq
    config.action_mailer.deliver_later_queue_name = nil # defaults to "mailers"
    config.action_mailbox.queues.routing = nil # defaults to "action_mailbox_routing"
    config.active_storage.queues.analysis = nil # defaults to "active_storage_analysis"
    config.active_storage.queues.purge = nil # defaults to "active_storage_purge"
    config.active_storage.queues.mirror = nil # defaults to "active_storage_mirror"
    config.time_zone = 'Pacific Time (US & Canada)'

    $stdout.sync = true

    config.generators.test_framework :rspec
  end

  Rails.application.config.assets.configure do |env|
    env.export_concurrent = false
  end

  if ENV['BRISK_INSECURE'] == 'true'
    puts 'INSECURE MODE ENABLED - NO AUTHENTICATION CHECKS WILL BE PERFORMED ON API CALLS - DO NOT USE IN MULTI-TENANT ENVIRONMENTS OR WHEN SERVICE IS EXPOSED TO THE INTERNET'
  end
end
