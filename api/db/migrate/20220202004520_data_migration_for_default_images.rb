class DataMigrationForDefaultImages < ActiveRecord::Migration[6.1]
  def change
    Image.create! name: 'node-lts', version: '1'
    Image.create! name: 'rails', version: '1'
  end
end
