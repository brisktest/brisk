class CreateTestFileRuns < ActiveRecord::Migration[7.0]
  def change
    create_table :test_file_runs do |t|
      t.integer :test_file_id
      t.integer :ms_time_taken

      t.timestamps
    end
  end
end
