class AddIndexForUids < ActiveRecord::Migration[7.0]
  def change
    add_index :jobruns, :uid, unique: true
    add_index :worker_run_infos, :uid, unique: true
  end
end
