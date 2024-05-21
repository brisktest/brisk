class AddInUseToSupervisors < ActiveRecord::Migration[7.0]
  def change
    add_column :supervisors, :in_use, :datetime
  end
end
