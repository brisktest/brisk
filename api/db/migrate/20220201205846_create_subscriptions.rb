class CreateSubscriptions < ActiveRecord::Migration[6.1]
  def change
    create_table :subscriptions do |t|
      t.integer :plan_id
      t.integer :account_id
      t.integer :address_id
      t.text :status

      t.timestamps
    end
  end
end
