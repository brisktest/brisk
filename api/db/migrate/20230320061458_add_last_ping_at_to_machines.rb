class AddLastPingAtToMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :machines, :last_ping_at, :datetime
  end
end
