class ChangeTestFilesRuntimeToInt < ActiveRecord::Migration[7.0]
  def down
    change_column :test_files, :runtime, :string
  end

  def up; end
end
