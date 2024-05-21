class AddUniqueInstanceIdToSupervisors < ActiveRecord::Migration[7.0]
  def change
    add_column :supervisors, :unique_instance_id, :text
  end
end
