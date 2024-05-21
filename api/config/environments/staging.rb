# frozen_string_literal: true

Rails.application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  # Code is not reloaded between requests.
  config.cache_classes = true
  config.active_storage.service = :amazon

  # Eager load code on boot. This eager loads most of Rails and
  # your application in memory, allowing both threaded web servers
  # and those relying on copy on write to perform better.
  # Rake tasks automatically ignore this option for performance.
  config.eager_load = true

  # Full error reports are disabled and caching is turned on.
  config.consider_all_requests_local = true
  config.action_controller.perform_caching = true

  # Ensures that a master key has been made available in either ENV["RAILS_MASTER_KEY"]
  # or in config/master.key. This key is used to decrypt credentials (and other encrypted files).
  # config.require_master_key = true

  # Disable serving static files from the `/public` folder by default since
  # Apache or NGINX already handles this.
  config.public_file_server.enabled = true # ENV['RAILS_SERVE_STATIC_FILES'].present?

  # Compress CSS using a preprocessor.
  # config.assets.css_compressor = :sass

  # Do not fallback to assets pipeline if a precompiled asset is missed.
  config.assets.compile = true
  config.public_file_server.enabled = true

  # Enable serving of images, stylesheets, and JavaScripts from an asset server.
  # config.action_controller.asset_host = 'http://assets.example.com'

  # Specifies the header that your server uses for sending files.
  # config.action_dispatch.x_sendfile_header = 'X-Sendfile' # for Apache
  # config.action_dispatch.x_sendfile_header = 'X-Accel-Redirect' # for NGINX

  # Mount Action Cable outside main process or domain.
  # config.action_cable.mount_path = nil
  # config.action_cable.url = 'wss://example.com/cable'
  # config.action_cable.allowed_request_origins = [ 'http://example.com', /http:\/\/example.*/ ]

  # Force all access to the app over SSL, use Strict-Transport-Security, and use secure cookies.
  # config.force_ssl = true
  # config.ssl_options = { redirect: { exclude: -> request { request.path =~ /health/ } } }

  # Use the lowest log level to ensure availability of diagnostic information
  # when problems arise.
  config.log_level = ENV['LOG_LEVEL'] || :debug

  # Prepend all log lines with the following tags.
  config.log_tags = [:request_id]

  # Use a different cache store in production.
  # config.cache_store = :mem_cache_store

  # Use a real queuing backend for Active Job (and separate queues per environment).
  # config.active_job.queue_adapter     = :resque
  # config.active_job.queue_name_prefix = "brisk_frontend_production"

  config.action_mailer.perform_caching = false

  # Ignore bad email addresses and do not raise email delivery errors.
  # Set this to true and configure the email server for immediate delivery to raise delivery errors.
  # config.action_mailer.raise_delivery_errors = false

  # Enable locale fallbacks for I18n (makes lookups for any locale fall back to
  # the I18n.default_locale when a translation cannot be found).
  config.i18n.fallbacks = true

  # Send deprecation notices to registered listeners.
  config.active_support.deprecation = :notify

  # Use default logging formatter so that PID and timestamp are not suppressed.

  # Use a different logger for distributed setups.
  # require 'syslog/logger'
  # config.logger = ActiveSupport::TaggedLogging.new(Syslog::Logger.new 'app-name')

  # Do not dump schema after migrations.
  config.active_record.dump_schema_after_migration = false

  # Inserts middleware to perform automatic connection switching.
  # The `database_selector` hash is used to pass options to the DatabaseSelector
  # middleware. The `delay` is used to determine how long to wait after a write
  # to send a subsequent read to the primary.
  #
  # The `database_resolver` class is used by the middleware to determine which
  # database is appropriate to use based on the time delay.
  #
  # The `database_resolver_context` class is used by the middleware to set
  # timestamps for the last write to the primary. The resolver uses the context
  # class timestamps to determine how long to wait before reading from the
  # replica.
  #
  # By default Rails will store a last write timestamp in the session. The
  # DatabaseSelector middleware is designed as such you can define your own
  # strategy for connection switching and pass that into the middleware through
  # these configuration options.
  # config.active_record.database_selector = { delay: 2.seconds }
  # config.active_record.database_resolver = ActiveRecord::Middleware::DatabaseSelector::Resolver
  # config.active_record.database_resolver_context = ActiveRecord::Middleware::DatabaseSelector::Resolver::Session
  config.assets.digest = true

  config.action_mailer.delivery_method = :smtp

  config.x.mail_from = %(Brisk <info@brisktest.com>)
  config.action_mailer.default_url_options = { host: 'development-www.briskci.com' }
  config.action_mailer.smtp_settings = {
    user_name: 'apikey', # This is the string literal 'apikey', NOT the ID of your API key
    password: ENV['SENDGRID_API_KEY'], # This is the secret sendgrid API key which was issued during API key creation
    domain: 'brisktest.com',
    address: 'smtp.sendgrid.net',
    port: 587,
    authentication: :plain,
    enable_starttls_auto: true
  }

  config.action_mailer.raise_delivery_errors = true
  ENV['GRUF_HOST'] ||= '0.0.0.0'
  ENV['GRUF_PORT'] ||= '9001'

  config.active_record.encryption.primary_key = ENV['ACTIVE_RECORD_ENCRYPTION_PRIMARY_KEY'] || 'cea0b247dfba24fc89d4d71aaf69f80482b367a855'
  config.active_record.encryption.deterministic_key = ENV['ACTIVE_RECORD_ENCRYPTION_DETERMINISTIC_KEY'] || '8870f8e6f8afba24fc8b95e2f80482b367a855'
  config.active_record.encryption.key_derivation_salt = ENV['ACTIVE_RECORD_ENCRYPTION_KEY_DERIVATION_SALT'] || 'cea0b247d83434afba24fc89d4d71aaf5'
  config.log_formatter = ::Logger::Formatter.new
  config.logger = Logger.new($stdout)

  config.log_formatter = ::Logger::Formatter.new
  config.logger = Logger.new($stdout)

  Rails.application.configure do
    config.lograge.enabled = true
  end
  Rails.application.routes.default_url_options[:host] = 'development-nomad.brisktesting.com'
  ENV['GITHUB_APP_ID'] = 'APP_ID'
  ENV['GITHUB_APP_SECRET'] = 'APP_SECRET'
end
