# frozen_string_literal: true

module Interceptors
  class Logging < Gruf::Interceptors::ServerInterceptor
    def call
      yield
    end

    # rescue StandardError => e
    #   Rails.logger.error "in logging interceptor: #{e.message}"
    #   Bugsnag.notify(e)
    #   raise e
    # end
  end
end
