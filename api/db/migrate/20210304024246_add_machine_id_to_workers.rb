# frozen_string_literal: true

class AddMachineIdToWorkers < ActiveRecord::Migration[6.0]
  def change
    add_column :workers, :machine_id, :integer
  end
end
