# frozen_string_literal: true

class RemoveUserIdFromCredentials < ActiveRecord::Migration[6.1]
  def change
    remove_column :credentials, :user_id
  end
end
