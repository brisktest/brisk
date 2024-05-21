class AddMaxSupervisorsToProjects < ActiveRecord::Migration[7.0]
  def change
    add_column :projects, :max_supervisors, :integer, default: 1
  end
end
