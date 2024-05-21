class AddMinimumCapacityToPlans < ActiveRecord::Migration[7.0]
  def change
    add_column :plans, :minimum_capacity, :integer, default: 5
  end
end
