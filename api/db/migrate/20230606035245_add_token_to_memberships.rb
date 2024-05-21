class AddTokenToMemberships < ActiveRecord::Migration[7.0]
  def change
    add_column :memberships, :token, :text
  end
end
