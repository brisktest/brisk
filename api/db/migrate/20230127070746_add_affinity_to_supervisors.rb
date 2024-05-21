class AddAffinityToSupervisors < ActiveRecord::Migration[7.0]
  def change
    add_column :supervisors, :affinity, :text
  end
end
