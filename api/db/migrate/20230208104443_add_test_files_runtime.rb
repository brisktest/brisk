class AddTestFilesRuntime < ActiveRecord::Migration[7.0]
  def change
    add_column :test_files, :runtime, :integer
  end
end
