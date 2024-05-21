class AddSupervisorIdToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :supervisor_id, :integer
  end
end
