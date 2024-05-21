class AddWorkerRunInfoIdToExecutionInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :execution_infos, :worker_run_info_id, :integer
  end
end
