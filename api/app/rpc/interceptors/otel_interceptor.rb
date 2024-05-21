require 'opentelemetry/sdk'

module Interceptors
  class OtelInterceptor < Gruf::Interceptors::ServerInterceptor
    def call
      active_call = request.active_call

      Rails.logger.debug "OTEL INTERCEPTOR: #{active_call.inspect}"
      Rails.logger.debug "OTEL INTERCEPTOR: #{active_call.metadata.inspect}"
      # Rails.logger.debug "OTEL INTERCEPTOR: #{active_call.peer.inspect}"
      context = OpenTelemetry.propagation.extract(active_call.metadata)

      attrs = {
        'rpc.system' => 'grpc',
        'rpc.service' => request.service_key.to_s,
        'rpc.method' => request.method_key.to_s,
        'net.peer.ip' => active_call.metadata['x-forwarded-for'] || 'none',
        'net.peer.port' => active_call.metadata['x-forwarded-port'] || 'none',
        'net.peer.name' => active_call.metadata['x-forwarded-host'] || 'none'

      }
      tracer = OpenTelemetry.tracer_provider.tracer('brisk-grpc')
      span = tracer.start_span(
        "#{request.service_key}/#{request.method_key}",
        with_parent: context,
        kind: :server,
        attributes: attrs
      )
      yield
    ensure
      span.finish
    end
  end
end
