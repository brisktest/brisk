# frozen_string_literal: true

require 'google/protobuf/well_known_types'

module InfraHelper
  def to_machine(machine)
    { id: machine.id, ip_address: machine.ip_address, state: machine.state, uid: machine.uid, image: machine.image,
      type: machine.type, cpus: machine.cpus, memory: machine.memory.to_s, disk: machine.disk.to_s }
  end
end
