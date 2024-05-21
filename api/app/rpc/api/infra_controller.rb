module Api
  class InfraController < Api::GrufController
    bind ::Api::Infra::Service
    include InfraHelper

    def insecure_mode?
      ENV['BRISK_INSECURE'] == 'true'
    end

    def auth_action_with(authenticated_action, this_action)
      return true if insecure_mode?

      return unless authenticated_action != this_action

      Rails.logger.error("not authorized for route #{authenticated_action} != #{this_action}")
      fail!(:unauthenticated, :unauthorized, 'Not authorized for this route')
    end

    def register_machine
      # api.infra.register_machine
      mesg = request.message
      Rails.logger.debug { "register_machine: #{mesg.inspect}" }

      auth_action_with(request.metadata[:authenticated_action], 'api.infra.register_machine')
      Rails.logger.debug { mesg.inspect }
      m = ::Machine.find_by(uid: mesg.machine.uid)
      if m
        if m.finished?
          Rails.logger.error("Got register for finished machine #{m.id}")
          fail!(:internal, :error, 'Failed to register finished machine')
        else
          m.last_ping_at = Time.now
          m.save!
          to_response(m)
        end
      else
        machine = InfraService.register(mesg.machine)
        to_response(machine)
      end
    end

    def de_register_machine
      # api.infra.de_register_machine
      mesg = request.message
      auth_action_with(request.metadata[:authenticated_action], 'api.infra.de_register_machine')
      Rails.logger.debug { mesg.inspect }
      m = ::Machine.find_by_uid(mesg.machine.uid)
      if m
        if !m.finished_at?
          Rails.logger.debug { "Got de-register for machine #{m.id}" }
          InfraService.de_register(m)
          to_response(m)
        else
          Rails.logger.error("Got de-register for finished machine #{m.id}")
          fail!(:internal, :error, 'Failed to de-register finished machine')
        end
      else
        Rails.logger.error("Got de-register for non-existent machine #{mesg.machine.uid}")
        fail!(:internal, :error, 'Failed to de-register non-existent machine')
      end
    end

    def drain_machine
      # api.infra.drain_machine
      mesg = request.message
      auth_action_with(request.metadata[:authenticated_action], 'api.infra.drain_machine')
      Rails.logger.debug { mesg.inspect }
      m = ::Machine.find_by_uid(mesg.machine.uid)
      if m
        if !m.drained_at
          Rails.logger.debug { "Got drain for machine #{m.id}" }
          m.update_attribute(:drained_at, Time.now)
        else
          Rails.logger.error("Got drain for already drained machine #{m.id}")
        end
        to_response(m)
      else
        Rails.logger.error("Got drain for non-existent machine #{mesg.machine.uid}")
        raise(:internal, :error, 'Failed to drain non-existent machine')
      end
    end

    def to_response(machine)
      ::Api::MachineResponse.new(machine: to_machine(machine))
    end
  end
end
