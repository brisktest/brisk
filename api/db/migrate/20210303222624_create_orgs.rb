# frozen_string_literal: true

class CreateOrgs < ActiveRecord::Migration[6.0]
  def change
    create_table :orgs do |t|
      t.integer :owner_id
      t.string :name

      t.timestamps
    end
  end
end
