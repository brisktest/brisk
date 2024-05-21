class CreateMemberships < ActiveRecord::Migration[7.0]
  def change
    create_table :memberships do |t|
      t.text :invited_email
      t.integer :user_id
      t.datetime :accepted_at
      t.integer :invited_by
      t.text :role
      t.integer :org_id

      t.timestamps
    end
  end
end
