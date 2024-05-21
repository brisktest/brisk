class CreateRepoInfos < ActiveRecord::Migration[7.0]
  def change
    create_table :repo_infos do |t|
      t.text :commit_hash
      t.text :repo
      t.text :branch
      t.text :tag
      t.text :commit_message
      t.text :commit_author
      t.text :commit_author_email

      t.timestamps
    end
  end
end
