class FinishMachinesJob < ApplicationJob
  queue_as :default

  def perform(*_args)
    Machine.running.where('last_ping_at < ?', 30.minutes.ago).each do |m|
      Rails.logger.info "Finishing machine #{m.id} with uid #{m.uid}"
      m.finished_at = Time.now
      m.save!
    end
  end
end
