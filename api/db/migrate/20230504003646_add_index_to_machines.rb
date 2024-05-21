class AddIndexToMachines < ActiveRecord::Migration[7.0]
  def change
    add_index :machines, :uid, unique: true
  end
end
