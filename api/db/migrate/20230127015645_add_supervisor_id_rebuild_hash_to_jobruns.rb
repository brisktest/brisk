class AddSupervisorIdRebuildHashToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :supervisor_id, :integer
    add_column :jobruns, :rebuild_hash, :text
  end
end
