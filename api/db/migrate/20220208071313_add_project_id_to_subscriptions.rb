class AddProjectIdToSubscriptions < ActiveRecord::Migration[7.0]
  def change
    add_column :subscriptions, :project_id, :integer
  end
end
