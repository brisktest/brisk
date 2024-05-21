# frozen_string_literal: true

require 'gruf'

require 'grpc/health/checker'
require 'grpc/health/v1/health_pb'
require 'grpc/health/v1/health_services_pb'
# require 'rpc/interceptors/token_auth'
# require Rails.root.join('app/rpc/interceptors/token_auth.rb').to_s

$stdout.sync = true
Rails.application.reloader.to_prepare do
  Gruf.configure do |c|
    # c.interceptors.use(
    #   OpenTelemetry::Instrumentation::GRPC::ServerInterceptor
    # )

    c.interceptors.use(
      Interceptors::OtelInterceptor, excluded_methods: ['grpc.health.v1.health.check']
    )

    c.interceptors.use(

      # Autoload classes and modules needed at boot time here.
      Interceptors::TokenAuth, excluded_methods: ['grpc.health.v1.health.check', 'api.users.login']
    )

    c.interceptors.use(
      Interceptors::BugsnagInterceptor
    )

    c.server_binding_url = "#{ENV['GRUF_HOST']}:#{ENV['GRUF_PORT']}"
    # c.server_binding_url = '0.0.0.0:9001'
    c.rpc_server_options = c.rpc_server_options.merge(pool_size: (ENV['GRUF_POOL_SIZE'] || 10).to_i)
    c.rpc_server_options = c.rpc_server_options.merge(poll_period: 5)
    c.backtrace_on_error = true
    c.use_exception_message = true
  end
end
Gruf.configure do |c|
  c.default_client_host = "#{ENV['GRUF_HOST']}:#{ENV['GRUF_PORT']}"
end
Rails.logger.info "Gruf init listening on #{ENV['GRUF_HOST']}:#{ENV['GRUF_PORT']}"
