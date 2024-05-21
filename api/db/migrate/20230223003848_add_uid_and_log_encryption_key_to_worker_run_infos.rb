class AddUidAndLogEncryptionKeyToWorkerRunInfos < ActiveRecord::Migration[7.0]
  def change
    add_column :worker_run_infos, :uid, :text
    add_column :worker_run_infos, :log_encryption_key, :text
  end
end
