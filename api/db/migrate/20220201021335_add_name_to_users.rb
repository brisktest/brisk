# frozen_string_literal: true

class AddNameToUsers < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :name, :text
  end
end
