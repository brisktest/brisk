class AddIndexOnTestFilesFilename < ActiveRecord::Migration[7.0]
  def change
    add_index :test_files, :filename
  end
end
