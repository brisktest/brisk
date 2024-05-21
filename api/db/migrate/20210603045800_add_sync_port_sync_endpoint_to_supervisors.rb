# frozen_string_literal: true

class AddSyncPortSyncEndpointToSupervisors < ActiveRecord::Migration[6.1]
  def change
    add_column :supervisors, :sync_port, :string
    add_column :supervisors, :sync_endpoint, :string
  end
end
