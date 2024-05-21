class AddIndexesOnVariousTables < ActiveRecord::Migration[7.0]
  def change
    add_index :workers, :freed_at
    add_index :workers, :last_checked_at
    add_index :workers, :rebuild_hash
    add_index :workers, :worker_image
    add_index :workers, :host_ip
    add_index :workers, :machine_id
    add_index :workers, :state
    add_index :workers, :supervisor_id
    add_index :workers, :project_id
    add_index :workers, :image_id
    add_index :workers, :created_at
    add_index :workers, :updated_at

    add_index :supervisors, :project_id
    add_index :supervisors, :affinity
    add_index :supervisors, :state
    add_index :supervisors, :machine_id
    add_index :supervisors, :created_at

    add_index :worker_run_infos, :worker_id
    add_index :worker_run_infos, :jobrun_id
    add_index :worker_run_infos, :started_at
    add_index :worker_run_infos, :finished_at
    add_index :worker_run_infos, :supervisor_id
    add_index :worker_run_infos, :project_id
    add_index :worker_run_infos, :created_at
  end
end
