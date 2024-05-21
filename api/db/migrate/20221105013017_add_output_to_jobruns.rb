class AddOutputToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :output, :text
  end
end
