class AddModeratorToUsers < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :moderator, :boolean
  end
end
