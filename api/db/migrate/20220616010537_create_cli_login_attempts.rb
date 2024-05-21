class CreateCliLoginAttempts < ActiveRecord::Migration[7.0]
  def change
    create_table :cli_login_attempts do |t|
      t.string :nonce
      t.string :token
      t.integer :user_id
      t.datetime :valid_until

      t.timestamps
    end
  end
end
