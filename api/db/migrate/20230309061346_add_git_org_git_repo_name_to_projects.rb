class AddGitOrgGitRepoNameToProjects < ActiveRecord::Migration[7.0]
  def change
    add_column :projects, :git_org, :text
    add_column :projects, :git_repo_name, :text
  end
end
