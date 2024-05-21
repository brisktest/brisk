# frozen_string_literal: true

module SupersHelper
  def to_super(sup)
    { id: sup.id, ip_address: sup.ip_address, port: sup.port, state: sup.state, endpoint: sup.endpoint,
      external_endpoint: sup.external_endpoint, sync_endpoint: sup.sync_endpoint, sync_port: sup.sync_port }
  end
end
