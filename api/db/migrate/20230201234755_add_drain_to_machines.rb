class AddDrainToMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :machines, :drain, :datetime
  end
end
