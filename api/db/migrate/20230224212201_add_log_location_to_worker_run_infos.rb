class AddLogLocationToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :log_location, :text
  end
end
