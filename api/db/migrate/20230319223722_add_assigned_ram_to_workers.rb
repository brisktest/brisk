class AddAssignedRamToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :assigned_ram, :integer
  end
end
