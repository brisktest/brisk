class AddMinWorkerPercentToScheules < ActiveRecord::Migration[7.0]
  def change
    add_column :schedules, :min_worker_percent, :decimal
  end
end
