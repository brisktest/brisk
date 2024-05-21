class AddExitCodeToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :exit_code, :integer
    add_column :jobruns, :error, :text
  end
end
