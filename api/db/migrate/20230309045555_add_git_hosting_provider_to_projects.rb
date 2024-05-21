class AddGitHostingProviderToProjects < ActiveRecord::Migration[7.0]
  def change
    add_column :projects, :git_hosting_provider, :text
  end
end
