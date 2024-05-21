class AddUidToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :uid, :text
  end
end
