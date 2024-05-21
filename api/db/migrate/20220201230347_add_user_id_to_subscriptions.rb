class AddUserIdToSubscriptions < ActiveRecord::Migration[6.1]
  def change
    add_column :subscriptions, :user_id, :integer
  end
end
