class TestFile < ApplicationRecord
  belongs_to :project

  has_many :test_file_runs
  has_many :worker_run_infos, through: :test_file_runs
  validates :filename, uniqueness: { scope: :project_id }, presence: true
end
