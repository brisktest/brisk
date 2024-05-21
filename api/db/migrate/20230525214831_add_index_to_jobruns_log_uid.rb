class AddIndexToJobrunsLogUid < ActiveRecord::Migration[7.0]
  def change
    add_index :jobruns, :log_uid
  end
end
