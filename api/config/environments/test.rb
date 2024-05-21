# frozen_string_literal: true

# The test environment is used exclusively to run your application's
# test suite. You never need to work with it otherwise. Remember that
# your test database is "scratch space" for the test suite and is wiped
# and recreated between test runs. Don't rely on the data there!

Rails.application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  config.log_level = :debug

  config.action_view.cache_template_loading = true
  config.active_storage.service = :test

  # Do not eager load code on boot. This avoids loading your whole application
  # just for the purpose of running a single test. If you are using a tool that
  # preloads Rails for running tests, you may have to set it to true.
  config.eager_load = false

  # Configure public file server for tests with Cache-Control for performance.
  config.public_file_server.enabled = true
  config.public_file_server.headers = {
    'Cache-Control' => "public, max-age=#{1.hour.to_i}"
  }

  # Show full error reports and disable caching.
  config.consider_all_requests_local = true
  config.action_controller.perform_caching = false
  config.cache_store = :null_store

  # Raise exceptions instead of rendering exception templates.
  config.action_dispatch.show_exceptions = false

  # Disable request forgery protection in test environment.
  config.action_controller.allow_forgery_protection = false

  # Store uploaded files on the local file system in a temporary directory.
  config.active_storage.service = :test

  config.action_mailer.perform_caching = false

  # Tell Action Mailer not to deliver emails to the real world.
  # The :test delivery method accumulates sent emails in the
  # ActionMailer::Base.deliveries array.
  config.action_mailer.delivery_method = :test

  # Print deprecation notices to the stderr.
  config.active_support.deprecation = :stderr
  # config.webpacker.check_yarn_integrity = false
  # Raises error for missing translations.
  # config.action_view.raise_on_missing_translations = true
  ENV['LOCKBOX_KEY'] = 'cea0b247d8870f8e6f8afba24fc89d4d71aaf6920c9ceb95e2f80482b367a855'
  config.action_mailer.default_url_options = { host: 'brisk-frontend.test', port: 80 }

  config.active_record.encryption.primary_key = 'primary-key'
  config.active_record.encryption.deterministic_key = 'deterministic-key'
  config.active_record.encryption.key_derivation_salt = 'salt'
  config.active_record.logger = nil
  ActiveRecord::Base.logger = nil

  config.logger = Logger.new(STDOUT)

  Rails.application.routes.default_url_options[:host] = 'brisktest-test.test'
  ENV['GITHUB_APP_ID'] = 'APP_ID'
  ENV['GITHUB_APP_SECRET'] = 'APP_SECRET'
  ENV['REDIS_URL'] = 'redis://127.0.0.1:6379/1'

  config.assets.css_compressor = nil
end
