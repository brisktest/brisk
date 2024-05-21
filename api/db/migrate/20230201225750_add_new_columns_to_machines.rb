class AddNewColumnsToMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :machines, :host_ip, :text
    add_column :machines, :uid, :text
    add_column :machines, :os_info, :text
    add_column :machines, :host_uid, :text
    add_column :machines, :image, :text
    add_column :machines, :type, :text
    add_column :machines, :cpus, :integer
    add_column :machines, :memory, :integer
    add_column :machines, :disk, :integer
    add_column :machines, :json_data, :jsonb
  end
end
