class AddReservedAtToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :reserved_at, :datetime
  end
end
