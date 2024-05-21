class CreateSchedules < ActiveRecord::Migration[7.0]
  def change
    create_table :schedules do |t|
      t.string :org_id

      t.timestamps
    end
  end
end
