class TestFileRun < ApplicationRecord
  belongs_to :test_file
  belongs_to :worker_run_info
end
