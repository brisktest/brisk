# frozen_string_literal: true

class AddHostIpToSupervisors < ActiveRecord::Migration[6.1]
  def change
    add_column :supervisors, :host_ip, :string
  end
end
