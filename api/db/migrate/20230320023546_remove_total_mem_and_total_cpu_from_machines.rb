class RemoveTotalMemAndTotalCpuFromMachines < ActiveRecord::Migration[7.0]
  def change
    remove_column :machines, :total_mem
    remove_column :machines, :total_cpu
  end
end
