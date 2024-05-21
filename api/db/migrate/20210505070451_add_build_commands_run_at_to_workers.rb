# frozen_string_literal: true

class AddBuildCommandsRunAtToWorkers < ActiveRecord::Migration[6.0]
  def change
    add_column :workers, :build_commands_run_at, :datetime
  end
end
