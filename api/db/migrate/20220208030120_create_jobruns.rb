class CreateJobruns < ActiveRecord::Migration[7.0]
  def change
    create_table :jobruns do |t|
      t.integer :project_id
      t.integer :user_id
      t.text :state

      t.timestamps
    end
  end
end
