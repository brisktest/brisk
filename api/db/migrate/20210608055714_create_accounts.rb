# frozen_string_literal: true

class CreateAccounts < ActiveRecord::Migration[6.1]
  def change
    create_table :accounts do |t|
      t.integer :company_id

      t.timestamps
    end
  end
end
