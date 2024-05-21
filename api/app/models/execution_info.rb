class ExecutionInfo < ApplicationRecord
  # message ExecutionInfo {
  #     google.protobuf.Timestamp started = 1;
  #     google.protobuf.Timestamp finished = 2;
  #     int32 exit_code = 3;
  #     string rebuildHash = 4;
  #     Command command = 5;
  #     string output = 6;
  #   }
  belongs_to :worker_run_info
  has_one :command, dependent: :destroy

  def duration
    return -1 unless finished && started

    (finished - started) * 1000
  end
end
