# frozen_string_literal: true

class CreateWorkers < ActiveRecord::Migration[6.0]
  def change
    create_table :workers do |t|
      t.integer :project_id
      t.text :ip_address
      t.text :port
      t.text :state
      t.text :image_id
      t.text :endpoint

      t.timestamps
    end
  end
end
