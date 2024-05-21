# frozen_string_literal: true

class AddAccountIdToOrg < ActiveRecord::Migration[6.1]
  def change
    add_column :orgs, :account_id, :integer
  end
end
