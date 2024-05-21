class DataAddDefaultSchedule < ActiveRecord::Migration[7.0]
  def change
    Schedule.create!(
      org_id: nil,
      min_worker_percent: 0.9
    )
  end
end
