Bugsnag.configure do |config|
  config.api_key = 'e35efbe05552b7db723a75134c69b36b'
  config.notify_release_stages = ['production']
  config.release_stage = ENV['RAILS_ENV']

  config.ignore_classes << 'Net::SMTPAuthenticationError'
  config.ignore_classes << 'Net::SMTPServerBusy'
end
