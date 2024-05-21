class AddConcurrencyToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :concurrency, :integer
  end
end
