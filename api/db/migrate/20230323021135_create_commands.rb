class CreateCommands < ActiveRecord::Migration[7.0]
  def change
    create_table :commands do |t|
      t.text :commandline
      t.text :args
      t.text :work_directory
      t.json :environment
      t.boolean :is_test_run
      t.integer :sequence_number
      t.integer :worker_number
      t.text :stage
      t.boolean :is_list_test
      t.text :test_framework
      t.boolean :background
      t.integer :total_worker_count

      t.timestamps
    end
  end
end
