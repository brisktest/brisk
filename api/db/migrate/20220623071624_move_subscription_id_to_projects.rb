class MoveSubscriptionIdToProjects < ActiveRecord::Migration[7.0]
  def change
    remove_column :subscriptions, :user_id
  end
end
