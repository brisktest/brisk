class AddTimingProcessedToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :timing_processed, :datetime
  end
end
