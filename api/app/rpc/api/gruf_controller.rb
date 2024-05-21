# frozen_string_literal: true

module Api
  class GrufController < ::Gruf::Controllers::Base
    def append_info_to_payload(payload)
      super
      payload[:environment] = Rails.env
      payload[:trace_key] = request&.active_call&.metadata&.[]('trace-key')
    end
  end
end
