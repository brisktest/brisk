class AddProjectIdAndSupervisorIdToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :project_id, :integer
    add_column :worker_run_infos, :supervisor_id, :integer
  end
end
