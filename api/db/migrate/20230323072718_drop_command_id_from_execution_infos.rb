class DropCommandIdFromExecutionInfos < ActiveRecord::Migration[7.0]
  def change
    remove_column :execution_infos, :command_id, :integer
  end
end
