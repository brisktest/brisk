class UpdateDataForWorkerRunInfosTestCount < ActiveRecord::Migration[7.0]
  def change
    # WorkerRunInfo.where(test_count: nil).in_batches do |wris|
    #   wris.each do |wri|
    #     wri.test_count = wri.test_files.count
    #     wri.save! rescue next
    #   end
    # end
    0.upto(300) do |i|
      WorkerRunInfo.with_number_of_test_files(i).update_all test_count: i
    end
  end
end
