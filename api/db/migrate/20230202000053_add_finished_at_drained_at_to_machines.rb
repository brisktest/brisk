class AddFinishedAtDrainedAtToMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :machines, :finished_at, :datetime
    add_column :machines, :drained_at, :datetime
  end
end
