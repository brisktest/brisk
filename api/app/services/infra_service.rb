class InfraService
  def self.register(mesg)
    Machine.find_or_create_by(uid: mesg.uid) do |m|
      m.ip_address = mesg.ip_address
      m.host_ip = mesg.host_ip
      m.uid = mesg.uid
      m.os_info = mesg.os_info
      m.host_uid = mesg.host_uid
      m.image = mesg.image
      m.type = mesg.type
      m.cpus = mesg.cpus
      # Convert to MB
      m.memory = mesg.memory.to_i / 1024 / 1024
      m.disk = mesg.disk.to_i / 1024 / 1024
      m.json_data = mesg.json_data

      m.last_ping_at = Time.now
      m.save!
    end
  end

  def self.de_register(machine)
    machine.finished_at = Time.now
    machine.save!
  end
end
