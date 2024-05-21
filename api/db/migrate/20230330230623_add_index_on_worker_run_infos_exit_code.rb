class AddIndexOnWorkerRunInfosExitCode < ActiveRecord::Migration[7.0]
  def change
    add_index :worker_run_infos, :exit_code
  end
end
