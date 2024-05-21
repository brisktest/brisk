class RepoInfo < ApplicationRecord
  belongs_to :jobrun

  def branch_name
    if branch && branch.match?(%r{^refs/heads/})
      branch.split('/').last
    else
      branch
    end
  end
end
