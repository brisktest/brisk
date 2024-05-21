require 'opentelemetry/sdk'
require 'opentelemetry/exporter/otlp'
require 'opentelemetry/instrumentation/all'

ENV['HONEYCOMB_API_KEY'] ||= Rails.application.credentials.honeycomb_key
ENV['BUGSNAG_API_KEY'] ||= Rails.application.credentials.bugsnag_api_key
# Settings specified here will take precedence over those in config/application.rb.

ENV['OTEL_EXPORTER_OTLP_ENDPOINT'] ||= 'https://api.honeycomb.io'
ENV['OTEL_EXPORTER_OTLP_HEADERS'] ||= "x-honeycomb-team=#{ENV['HONEYCOMB_API_KEY']}"
ENV['OTEL_SERVICE_NAME'] ||= 'rails-app'
ENV['OTEL_EXPORTER_OTLP_TRACES_ENDPOINT'] ||= 'https://api.honeycomb.io/v1/traces'
ENV['OTEL_EXPORTER_OTLP_METRICS_ENDPOINT'] ||= 'https://api.honeycomb.io/v1/metrics'
MyAppTracer = OpenTelemetry.tracer_provider.tracer('brisk-frontend')

OpenTelemetry::SDK.configure do |c|
  # c.service_name = "brisk-frontend"
  # c.use_all unless Rails.env.test?
end
