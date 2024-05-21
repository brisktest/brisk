class CreateExecutionInfos < ActiveRecord::Migration[7.0]
  def change
    create_table :execution_infos do |t|
      t.datetime :started
      t.datetime :finished
      t.integer :exit_code
      t.text :rebuild_hash
      t.integer :command_id
      t.string :output
      t.string :text

      t.timestamps
    end
  end
end
