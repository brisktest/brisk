# frozen_string_literal: true

class CreateProjects < ActiveRecord::Migration[6.0]
  def change
    create_table :users, &:timestamps

    create_table :credentials do |t|
      t.references :user, null: false, foreign_key: true
      t.text :api_token

      t.timestamps
    end

    create_table :projects do |t|
      t.references :org, null: false, foreign_key: true
      t.string :name
      t.references :user, null: false, foreign_key: true

      t.timestamps
    end
  end
end
