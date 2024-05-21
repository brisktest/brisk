# frozen_string_literal: true

class WorkerService
  def self.register(host_ip, ip_address, port, uid, worker_image, host_uid, sync_port)
    Rails.logger.debug("self.register worker #{uid} #{host_ip} #{ip_address} #{port} #{worker_image} #{host_uid}")
    raise 'RegisterWorker: need a uid ' if uid.blank?
    raise 'RegisterWorker: need a host_ip ' if host_ip.blank?
    raise 'RegisterWorker: need an ip_address ' if ip_address.blank?
    raise 'RegisterWorker: need a port ' if ip_address.blank?
    raise 'RegisterWorker: need a worker_image ' if worker_image.blank?
    raise 'RegisterWorker: need a host_uid ' if host_uid.blank?
    raise 'RegisterWorker: need a sync_port ' if sync_port.blank?

    # TODO: move to the UID of the machine
    m = Machine.where(uid: host_uid).last
    raise "no machine found in worker register uid: #{host_uid}" unless m

    # we were probably doing this for a reason
    # return nil unless m
    Rails.logger.debug("Found machine with id #{m.id}")

    image = Image.where(name: worker_image).last
    raise "No image found with name #{worker_image}" unless image

    worker = m.workers.new(host_ip:, ip_address:, port:, machine: m,
                           endpoint: "#{ip_address}:#{port}", uid:, worker_image:, image_id: image.id, sync_port:)

    Rails.logger.debug { "Worker is #{worker}" }

    begin
      worker.save!
    rescue ActiveRecord::RecordInvalid => e
      # this happens if we have two registers close together
      raise e unless e.message.include? 'Uid has already been taken'

      Rails.logger.debug("Got subsequent register for worker #{worker.uid}")
      worker = Worker.find_by_uid worker.uid
      worker.last_checked_at = Time.now
      worker.save!
      return worker
    rescue ActiveRecord::RecordNotUnique => e
      # this happens if we have two registers close together

      Rails.logger.debug("Got subsequent register for worker #{worker.uid}")
      worker = Worker.find_by_uid worker.uid
      worker.last_checked_at = Time.now
      worker.save!
      return worker
    end

    worker
  end

  def self.de_register(worker_uid)
    Rails.logger.debug("self.de_register worker #{worker_uid}")
    worker = Worker.find_by_uid worker_uid
    raise ActiveRecord::RecordNotFound unless worker

    Rails.logger.debug("Worker we de_register is id=#{worker.id}")
    worker.de_register! unless worker.finished?
    worker
  end

  def self.set_build_commands_run(worker_id, project)
    worker = Worker.find worker_id
    raise ActiveRecord::RecordNotFound unless worker

    if worker.project != project
      Rails.logger.Info("Attempting to mark build commands for the wrong project project id: #{project.id} , worker id : #{worker.id}")
      throw :unauthorized
    end
    worker.build_commands_run_at = Time.now
    worker.save!
    worker
  end
end
