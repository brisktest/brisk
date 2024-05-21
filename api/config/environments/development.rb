# frozen_string_literal: true

Rails.application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  # In the development environment your application's code is reloaded on
  # every request. This slows down response time but is perfect for development
  # since you don't have to restart the web server when you make code changes.
  config.cache_classes = false
  config.active_storage.service = :amazon

  # Do not eager load code on boot.
  config.eager_load = false

  # Show full error reports.
  config.consider_all_requests_local = true

  # Enable/disable caching. By default caching is disabled.
  # Run rails dev:cache to toggle caching.
  if Rails.root.join('tmp/caching-dev.txt').exist?
    config.action_controller.perform_caching = true
    config.action_controller.enable_fragment_cache_logging = true

    config.cache_store = :memory_store
    config.public_file_server.headers = {
      'Cache-Control' => "public, max-age=#{2.days.to_i}"
    }
  else
    config.action_controller.perform_caching = false

    config.cache_store = :null_store
  end

  # Don't care if the mailer can't send.
  config.action_mailer.raise_delivery_errors = false

  config.action_mailer.perform_caching = false

  # Print deprecation notices to the Rails logger.
  config.active_support.deprecation = :log

  # Raise an error on page load if there are pending migrations.
  config.active_record.migration_error = :page_load

  # Highlight code that triggered database queries in logs.
  config.active_record.verbose_query_logs = true

  config.assets.debug = true

  config.assets.compile = true

  config.assets.quiet = true

  # Raises error for missing translations.
  # config.action_view.raise_on_missing_translations = true

  # Use an evented file watcher to asynchronously detect changes in source code,
  # routes, locales, etc. This feature depends on the listen gem.
  config.file_watcher = ActiveSupport::EventedFileUpdateChecker

  ENV['LOCKBOX_KEY'] = 'cea0b247d8870f8e6f8afba24fc89d4d71aaf6920c9ceb95e2f80482b367a855'
  config.action_mailer.default_url_options = { host: 'brisk-frontend.test', port: 80 }
  ENV['GRUF_HOST'] = '0.0.0.0'
  ENV['GRUF_PORT'] = '9001'

  ENV['LOG_REQUESTS'] ||= 'true'

  ENV['LOG_LEVEL'] ||= 'debug'

  config.hosts << 'brisk-frontend.test'
  config.hosts << 'brisktest.test'

  ENV['GRUF_HOST'] ||= '0.0.0.0'
  ENV['GRUF_PORT'] ||= '9001'

  $stdout.sync = true
  Rails.application.configure do
    config.lograge.enabled = true
  end
  Rails.application.routes.default_url_options[:host] = 'brisktest.test'

  Rails.application.config.content_security_policy_report_only = true
  ENV['GITHUB_APP_ID'] = '257a603287a4a7d6a21a'
  ENV['GITHUB_APP_SECRET'] = 'e69bf9946dd3a98f03f236d7913f89806217c359'
  ENV['GOOGLE_OAUTH_CLIENT_ID'] = 'google_oauth_client_id'
  ENV['GOOGLE_OAUTH_CLIENT_SECRET'] = 'google_oauth_client_secret'

  config.active_record.encryption.primary_key = ENV['ACTIVE_RECORD_ENCRYPTION_PRIMARY_KEY'] || 'key'
  config.active_record.encryption.deterministic_key = ENV['ACTIVE_RECORD_ENCRYPTION_DETERMINISTIC_KEY'] ||  'key'
  config.active_record.encryption.key_derivation_salt = ENV['ACTIVE_RECORD_ENCRYPTION_KEY_DERIVATION_SALT'] || 'salt'

  config.log_formatter = ::Logger::Formatter.new
  config.logger = Logger.new($stdout)
end
