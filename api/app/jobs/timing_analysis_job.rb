class TimingAnalysisJob < ApplicationJob
  queue_as :default

  def perform(worker_info_id, delay = 30)
    worker_run_info = WorkerRunInfo.find(worker_info_id)

    return unless worker_run_info.succeeded? && worker_run_info.jobrun.completed?

    if worker_run_info.worker.machine.no_still_running_workers_from(worker_run_info.finished_at)
      Rails.logger.debug("TimingAnalysisJob: Timing - before compare_to #{worker_info_id}")

      worker_run_info.compare_to_previous_run
      worker_run_info.timing_processed = Time.now
      worker_run_info.save!
      if worker_run_info.jobrun.worker_run_infos.all?(&:timing_processed)
        PreSplitJob.perform_later(worker_run_info.jobrun_id)
      end
      Rails.logger.debug("TimingAnalysisJob: Timing - after compare_to #{worker_run_info}")
    else
      Rails.logger.info("Requeueing TimingAnalysisJob for #{worker_info_id} in #{delay} seconds because we still have running workers")
      TimingAnalysisJob.set(wait_until: delay.seconds.from_now).perform_later(worker_info_id)
    end
  end
end
