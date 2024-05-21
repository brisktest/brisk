class CleanupJobrunsJob < ApplicationJob
  queue_as :default

  def perform(*_args)
    Jobrun.started.where('created_at < ?', 14.minutes.ago).shuffle.each do |jr|
      Rails.logger.info "Cleaning up jobrun #{jr.id} for project #{jr.project.id}
      Jobrun is a #{jr.project.image.name} jobrun "

      jr.fail!
      jr.workers.each { |w| w.free_from_super }
    end
  end
end
