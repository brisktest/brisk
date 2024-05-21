# frozen_string_literal: true

module Api
  class SupersController < Api::GrufController
    bind ::Api::Supers::Service
    include SupersHelper

    def record_setup
      project = request.project
      id = request.id
      sup = SuperService.record_setup(id, project)
      to_response(sup)
    end

    def register
      # api.supers.register method_name

      if ENV['DEV'] == 'true'
        # we only can route to one super in dev usually
        Supervisor.delete_all
      end

      mesg = request.message
      sup = SuperService.register(mesg.host_ip, mesg.ip_address, mesg.port, mesg.sync_port, mesg.external_endpoint, mesg.uid,
                                  mesg.sync_endpoint, mesg.host_uid)
      to_response(sup)
    end

    def mark_super_as_unreachable
      sup = SuperService.mark_super_as_unreachable(request.message.super.id, request.metadata[:project])
      Api::UnreachableResp.new(super: to_super(sup))
    end

    def de_register
      # api.supers.de_register
      sup = SuperService.de_register(request.message.id)
      Rails.logger.debug("Super we deregistered is id=#{sup.id}")
      to_response(sup)
    end

    def to_response(sup)
      ::Api::SuperResponse.new(super: to_super(sup))
    end
  end
end
