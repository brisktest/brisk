class AddIndexesToTestFileRuns < ActiveRecord::Migration[7.0]
  def change
    add_index :test_file_runs, :test_file_id
    add_index :test_file_runs, :worker_run_info_id
    add_index :test_file_runs, %i[test_file_id worker_run_info_id]
  end
end
