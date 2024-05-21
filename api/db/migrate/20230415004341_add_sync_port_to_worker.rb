class AddSyncPortToWorker < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :sync_port, :text
  end
end
