# frozen_string_literal: true

module Api
  class WorkersController < Api::GrufController
    bind ::Api::Workers::Service
    include WorkersHelper

    def insecure_mode?
      ENV['BRISK_INSECURE'] == 'true'
    end

    def auth_action_with(authenticated_action, this_action)
      return true if insecure_mode?

      return unless authenticated_action != this_action

      Rails.logger.error("not authorized for route #{authenticated_action} != #{this_action}")
      fail!(:unauthenticated, :unauthorized, 'Not authorized for this route')
    end

    def register
      # api.workers.register
      # we can currently only route to one worker in DEV
      if ENV['DEV'] == 'true'
        # little hack so we can have more than one worker in dev mode but we don't
        # have to deal with expiring and checking the workers.
        ::Worker.active.where("created_at > ? ", 1.minute.ago).each(&:de_register!)
        ::Worker.delete_all
        Rails.logger.info('DEV mode, deregistering all workers before we add the new one')
      end

      mesg = request.message
      Rails.logger.debug("Got register for worker with mesg : #{mesg.inspect}")
      auth_action_with(request.metadata[:authenticated_action], 'api.workers.register')

      w = ::Worker.where(uid: mesg.uid).last
      if w
        if !w.finished?
          Rails.logger.debug("Got subsequent register for worker #{w.id}")
          w.last_checked_at = Time.now
          w.save!
          to_response(w)
        else
          Rails.logger.error("Got register for finished worker #{w.id}")
          to_response(w)
          # fail!(:internal, :error, "Failed to register finished worker")
        end
      else
        worker = WorkerService.register(mesg.host_ip, mesg.ip_address, mesg.port, mesg.uid, mesg.worker_image,
                                        mesg.host_uid, mesg.sync_port)
        to_response(worker)
      end
    end

    def de_register
      # api.workers.de_register
      auth_action_with(request.metadata[:authenticated_action], 'api.workers.de_register')

      worker = WorkerService.de_register(request.message.uid)
      Rails.logger.debug("Worker we deregistered is id=#{worker.id}")
      to_response(worker)
    rescue ActiveRecord::RecordNotFound
      # fail!(:not_found, :worker_not_found, "Failed to deregister worker: #{request.message}")
      to_response(nil)
    end

    def get_recently_deregistered
      auth_action_with(request.metadata[:authenticated_action], 'api.workers.get_recently_deregistered')

      workers = ::Worker.finished.where('updated_at > ? ', 10.minutes.ago)
      ::Api::WorkersResp.new({ workers: workers.map { |w| { uid: w.uid } } })
    end

    def build_commands_run
      Rails.logger.debug("build_commands_run for worker #{request.message} project #{request.metadata[:project]}")
      worker = WorkerService.set_build_commands_run(request.message.id, request.metadata[:project])
      to_response(worker)
      # rescue ActiveRecord::RecordNotFound
      #   fail!(:not_found, :worker_not_found, "Failed to save build commands: #{request.message.id}")
    end

    def to_response(worker)
      if worker
        ::Api::WorkerResponse.new(worker: to_worker(worker))
      else
        Rails.logger.error("No worker sent in response  #{request.message}")
        ::Api::WorkerResponse.new
      end
    end
  end
end
