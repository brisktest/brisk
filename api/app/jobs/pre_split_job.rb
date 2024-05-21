class PreSplitJob < ApplicationJob
  queue_as :test_splitter

  def perform(jobrun_id)
    Rails.logger.info "PreSplitJob: #{jobrun_id} ++"
    jobrun = Jobrun.find(jobrun_id)

    if jobrun.test_files.count > 0
      tss = TestSplitterService.new(jobrun.project, jobrun.assigned_concurrency, jobrun.test_files.map(&:filename))
      tss.store_split
    else
      Rails.logger.info "PreSplitJob: #{jobrun_id} -- No test files found so not storing a split"
    end

    Rails.logger.info "PreSplitJob: #{jobrun_id} -- "
  end
end
