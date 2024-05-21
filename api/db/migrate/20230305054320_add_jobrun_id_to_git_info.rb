class AddJobrunIdToGitInfo < ActiveRecord::Migration[7.0]
  def change
    add_column :repo_infos, :jobrun_id, :integer
  end
end
