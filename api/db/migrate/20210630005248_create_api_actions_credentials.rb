# frozen_string_literal: true

class CreateApiActionsCredentials < ActiveRecord::Migration[6.1]
  def change
    create_table :api_actions_credentials do |t|
      t.integer :credential_id
      t.integer :api_action_id
      t.timestamps
    end
  end
end
