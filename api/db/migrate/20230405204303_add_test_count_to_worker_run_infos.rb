class AddTestCountToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :test_count, :integer
  end
end
