class CreateTestFiles < ActiveRecord::Migration[7.0]
  def change
    create_table :test_files do |t|
      t.string :filename
      t.string :runtime
      t.string :version
      t.string :language

      t.timestamps
    end
  end
end
