class AddExecutionInfoIdToCommands < ActiveRecord::Migration[7.0]
  def change
    add_column :commands, :execution_info_id, :integer
  end
end
