class AddMonthlyConcurrencyToPlans < ActiveRecord::Migration[7.0]
  def change
    add_column :plans, :monthly_concurrency, :integer
  end
end
