# frozen_string_literal: true

module Interceptors
  class BugsnagInterceptor < Gruf::Interceptors::ServerInterceptor
    def call
      yield
    rescue StandardError => e
      Rails.logger.error "in sentry/bugsnag interceptor: #{e.message} - #{e.backtrace}"
      # Bugsnag.notify(e)
      Sentry.capture_exception(e)
      current_span = OpenTelemetry::Trace.current_span
      current_span.status = OpenTelemetry::Trace::Status.error(e.message)
      current_span.record_exception(e)
      raise e
    end
  end
end
