class AddAssignedConcurrencyToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :assigned_concurrency, :integer
  end
end
