module LogrageControllerOverride
  def append_info_to_payload(payload)
    super
    payload[:trace_key] = request&.active_call&.metadata&.[]('trace-key')
  end
end

# should probably do this only if Rails.env is production
::Gruf::Controllers::Base.prepend LogrageControllerOverride

Rails.application.configure do
  # Sane prod log formatting with lograge
  config.lograge.enabled = true
  # OPTIONAL to use JSON formatted logging
  config.lograge.formatter = Lograge::Formatters::Logstash.new
  config.lograge.base_controller_class = %w[GrufController ApplicationController]
  config.lograge.custom_options = lambda do |event|
    {
      exception: event.payload[:exception],
      exception_object: event.payload[:exception_object],
      # Convert backtrace into one string - \n is OK for JSON, but a different
      # joiner would be needed if we didn't use the Logstash/JSON format.
      backtrace: begin
        event.payload[:exception_object].backtrace.join("\n")
      rescue StandardError
        nil
      end,

      trace_key: event.payload[:trace_key],
      environment: event.payload[:environment],
      host: event.payload[:host],
      remote_ip: event.payload[:remote_ip],
      ip: event.payload[:ip]
    }
  end
end
