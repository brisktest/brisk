# frozen_string_literal: true

class CreateSupervisors < ActiveRecord::Migration[6.1]
  def change
    create_table :supervisors do |t|
      t.integer :project_id
      t.string :ip_address
      t.string :port
      t.string :state
      t.integer :machine_id

      t.timestamps
    end
  end
end
