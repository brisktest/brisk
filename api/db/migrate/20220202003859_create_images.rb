class CreateImages < ActiveRecord::Migration[6.1]
  def change
    create_table :images do |t|
      t.text :name
      t.text :version

      t.timestamps
    end
  end
end
