class DoTheTestCountAgain < ActiveRecord::Migration[7.0]
  def change
    0.upto(300) do |i|
      WorkerRunInfo.where(test_count: nil).with_number_of_test_files_db(i).update_all test_count: i
    end
  end
end
