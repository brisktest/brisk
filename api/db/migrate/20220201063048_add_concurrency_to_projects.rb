class AddConcurrencyToProjects < ActiveRecord::Migration[6.1]
  def change
    add_column :projects, :worker_concurrency, :integer, default: 1
  end
end
