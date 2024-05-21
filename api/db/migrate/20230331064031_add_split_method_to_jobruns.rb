class AddSplitMethodToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :split_method, :text
  end
end
