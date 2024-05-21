class ReleaseOldWorkersJob < ApplicationJob
  queue_as :default

  def perform(*_args)
    Project.all.shuffle.each do |project|
      Rails.logger.info "Releasing old workers for project #{project.id}"
      project.release_old_workers 15.minutes
    end
  end
end
