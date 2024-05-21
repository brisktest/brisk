class AddInviterToMemberships < ActiveRecord::Migration[7.0]
  def change
    add_column :memberships, :inviter_id, :integer
  end
end
