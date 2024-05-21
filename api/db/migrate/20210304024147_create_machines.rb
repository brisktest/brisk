# frozen_string_literal: true

class CreateMachines < ActiveRecord::Migration[6.0]
  def change
    create_table :machines do |t|
      t.integer :project_id
      t.text :ip_address
      t.text :state
      t.text :instance_id
      t.text :instance_type
      t.integer :total_mem
      t.integer :total_cpu

      t.timestamps
    end
  end
end
