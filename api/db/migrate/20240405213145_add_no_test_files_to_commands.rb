class AddNoTestFilesToCommands < ActiveRecord::Migration[7.0]
  def change
    add_column :commands, :no_test_files, :boolean
    add_column :commands, :command_id, :string
    add_column :commands, :command_concurrency, :integer
  end
end
