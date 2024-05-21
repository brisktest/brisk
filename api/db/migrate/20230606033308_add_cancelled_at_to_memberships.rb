class AddCancelledAtToMemberships < ActiveRecord::Migration[7.0]
  def change
    add_column :memberships, :cancelled_at, :datetime
  end
end
