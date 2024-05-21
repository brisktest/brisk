class AddIndexOnWorkerRunInfosTestCount < ActiveRecord::Migration[7.0]
  def change
    add_index :worker_run_infos, :test_count
    add_index :execution_infos, :finished
    add_index :execution_infos, :started
  end
end
