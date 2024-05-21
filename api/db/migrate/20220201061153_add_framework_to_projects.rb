class AddFrameworkToProjects < ActiveRecord::Migration[6.1]
  def change
    add_column :projects, :framework, :text
    Project.update_all framework: 'Rails'
  end
end
