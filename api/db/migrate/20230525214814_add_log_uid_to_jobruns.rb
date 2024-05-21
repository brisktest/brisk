class AddLogUidToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :log_uid, :uuid
  end
end
