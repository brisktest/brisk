class AddIndexOnCommandsExecutionInfoId < ActiveRecord::Migration[7.0]
  def change
    add_index :commands, :execution_info_id
  end
end
