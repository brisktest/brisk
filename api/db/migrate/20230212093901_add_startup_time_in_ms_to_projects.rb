class AddStartupTimeInMsToProjects < ActiveRecord::Migration[7.0]
  def change
    add_column :projects, :startup_time_in_ms, :integer
  end
end
