class AddProjectIdToTestFiles < ActiveRecord::Migration[7.0]
  def change
    add_column :test_files, :project_id, :integer
  end
end
