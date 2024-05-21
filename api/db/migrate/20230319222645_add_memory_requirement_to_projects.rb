class AddMemoryRequirementToProjects < ActiveRecord::Migration[7.0]
  def change
    add_column :projects, :memory_requirement, :integer
  end
end
