class AddIndexOnExecutionInfosWorkerRunInfoId < ActiveRecord::Migration[7.0]
  def change
    add_index :execution_infos, :worker_run_info_id
  end
end
