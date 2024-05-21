class RemoveTestFilesRuntime < ActiveRecord::Migration[7.0]
  def change
    remove_column :test_files, :runtime
  end
end
