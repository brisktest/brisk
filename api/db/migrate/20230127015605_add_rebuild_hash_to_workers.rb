class AddRebuildHashToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :rebuild_hash, :text
  end
end
