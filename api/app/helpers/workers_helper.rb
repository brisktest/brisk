# frozen_string_literal: true

require 'google/protobuf/well_known_types'

module WorkersHelper
  def to_worker(worker)
    bcr_at = worker.build_commands_run_at
    timestamp = (Google::Protobuf::Timestamp.new(seconds: bcr_at.to_i, nanos: bcr_at.nsec) if bcr_at)
    { id: worker.id, ip_address: worker.ip_address, build_commands_run_at: timestamp, port: worker.port,
      state: worker.state, endpoint: worker.endpoint, uid: worker.uid, worker_image: worker.worker_image, sync_port: worker.sync_port }
  end
end
