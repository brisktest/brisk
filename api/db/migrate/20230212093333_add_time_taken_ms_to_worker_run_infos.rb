class AddTimeTakenMsToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :ms_time_taken, :integer
  end
end
