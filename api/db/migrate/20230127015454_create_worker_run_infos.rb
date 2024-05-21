class CreateWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    create_table :worker_run_infos do |t|
      t.text :rebuild_hash
      t.text :exit_code
      t.text :output
      t.datetime :finished_at
      t.datetime :started_at
      t.text :error
      t.integer :worker_id

      t.timestamps
    end
  end
end
