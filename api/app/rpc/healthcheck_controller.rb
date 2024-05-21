# frozen_string_literal: true

require 'grpc/health/v1/health_pb'
require 'grpc/health/v1/health_services_pb'

class HealthcheckController < ::Gruf::Controllers::Base
  bind Grpc::Health::V1::Health::Service
  # Implement health service.

  def check
    Rails.logger.debug { 'Running healthcheck' }
    checker = Grpc::Health::Checker.new
    Grpc::Health::V1::HealthCheckResponse.new(status: Grpc::Health::V1::HealthCheckResponse::ServingStatus::SERVING)
  end
end
