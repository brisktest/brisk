# frozen_string_literal: true

class ChangeAccountsCompanyIdToUserId < ActiveRecord::Migration[6.1]
  def change
    rename_column :accounts, :company_id, :user_id
  end
end
