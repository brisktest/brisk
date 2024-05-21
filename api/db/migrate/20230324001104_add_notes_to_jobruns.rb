class AddNotesToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :notes, :text
  end
end
