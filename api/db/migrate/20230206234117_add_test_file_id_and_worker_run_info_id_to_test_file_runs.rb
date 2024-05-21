class AddTestFileIdAndWorkerRunInfoIdToTestFileRuns < ActiveRecord::Migration[7.0]
  def change
    add_column :test_file_runs, :worker_run_info_id, :integer
  end
end
