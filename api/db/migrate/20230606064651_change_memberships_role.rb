class ChangeMembershipsRole < ActiveRecord::Migration[7.0]
  def change
    change_column :memberships, :role, :string, default: 'member'
  end
end
