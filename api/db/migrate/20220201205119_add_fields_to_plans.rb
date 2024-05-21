class AddFieldsToPlans < ActiveRecord::Migration[6.1]
  def change
    add_column :plans, :name, :text
    add_column :plans, :amount_cents, :integer
    add_column :plans, :currency, :text
    add_column :plans, :description, :text
    add_column :plans, :period, :text
    add_column :plans, :trial_period, :text
    add_column :plans, :status, :text
  end
end
