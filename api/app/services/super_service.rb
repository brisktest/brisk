# frozen_string_literal: true

class SuperService
  def self.record_setup(id, project)
    sup = project.supervisors.find_by_id(id)
    sup.setup_run_at = Time.now unless sup.setup_run_at
    sup.save!
    sup
  end

  def self.register(host_ip, ip_address, port, sync_port, external_endpoint, uid, sync_endpoint, host_uid)
    m = Machine.find_by(uid: host_uid)
    raise "No machine found uid: #{host_uid}" unless m

    Rails.logger.debug("Found machine with id #{m.id}")
    sup = Supervisor.find_by(uid:)
    if sup && !sup.finished?
      Rails.logger.info("Supervisor already  registered  and not finished for uid #{uid}")
      return sup
    end

    if sup && sup.finished?
      # Rails.logger.info("reviving super")
      # sup.revive!
      # sup.save!
      # return sup
      raise "Supervisor already registered and finished for uid #{uid}"
    end

    sup = m.supervisors.new(host_ip:, ip_address:, port:, machine: m, sync_port:,
                            external_endpoint:, sync_endpoint:, uid:)

    Rails.logger.debug("Super is #{sup}")

    sup.save!
    sup
  end

  def self.de_register(super_id)
    sup = Supervisor.find super_id
    raise ActiveRecord::RecordNotFound unless sup

    Rails.logger.debug("Super we want to de_register is id=#{sup.id}")
    sup.de_register!
    sup
  end

  def self.mark_super_as_unreachable(super_id, project)
    # later we can add more code here to deal with any dodgy actors on the front end
    # for example we can add a count and have to have a certain number of votes before we quit
    # or we can try and connect from here, or potentially only allow a cli to cancel a certain amount
    # cause we don't want to share supers
    sup = Supervisor.find super_id
    raise ActiveRecord::RecordNotFound unless sup

    if sup.project != project
      Rails.logger.error("Unauthrized attempt to mark super as unreachable project: #{project.id} tried to mark super: #{sup.id}")
      throw :unauthorized
    end
    Rails.logger.info("Marking super as unreachable #{sup.id} for projet #{project.id}")
    sup.de_register!
    sup
  end
end
