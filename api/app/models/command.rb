class Command < ApplicationRecord
  belongs_to :execution_info

  # message Command {
  #     string commandline = 1;
  #     repeated string args = 2;
  #     string workDirectory=3;
  #     map<string, string> environment = 4;
  #     bool isTestRun  = 5;
  #     bool lastCommand = 6;
  #     int32 sequenceNumber = 7;
  #     int32 workerNumber = 8;
  #     string stage = 9 ;
  #     // bool recalcFiles = 10;
  #     bool isListTest  = 11;
  #     string testFramework = 12;
  #     bool background = 13;
  #     int32 totalWorkerCount = 14;

  #   }

  def self.new_from_proto(command)
    Command.new(
      commandline: command.commandline,
      args: command.args,
      work_directory: command.workDirectory,
      environment: command.environment,
      is_test_run: command.isTestRun,
      sequence_number: command.sequenceNumber,
      worker_number: command.workerNumber,
      stage: command.stage,
      is_list_test: command.isListTest,
      test_framework: command.testFramework,
      background: command.background,
      total_worker_count: command.totalWorkerCount,
      no_test_files: command.noTestFiles
    )
  end
end
