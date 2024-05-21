class ChangeColumnSizeForMachines < ActiveRecord::Migration[7.0]
  def change
    change_column :machines, :memory, :integer, limit: 8
    change_column :machines, :disk, :integer, limit: 8
  end
end
