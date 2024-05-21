# frozen_string_literal: true

class AddHostIdToWorkers < ActiveRecord::Migration[6.0]
  def change
    add_column :workers, :host_ip, :text
  end
end
