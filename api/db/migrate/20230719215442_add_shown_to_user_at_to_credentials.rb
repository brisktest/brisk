class AddShownToUserAtToCredentials < ActiveRecord::Migration[7.0]
  def change
    add_column :credentials, :shown_to_user_at, :datetime
  end
end
