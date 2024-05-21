class Schedule < ApplicationRecord
  # there is also a check in the super for the number of workers returned
  # we have it hardwired to 20% at the moment.
  # in super.main func getWorkingWorkers

  # we allow nil but only for our default so exacltly one nil
  validates :org_id, uniqueness: true

  # this is really just a stub for now
  # if we need complicated scheulduling logic we can add it here
  # but for now we just need a default
  # using postgres timerange seems the way to go something like
  # https://dba.stackexchange.com/questions/265841/storing-companies-working-hours-in-postgres

  def self.default_schedule
    Schedule.new min_worker_percent: 0.4
  end

  def get_current_min_worker_percent
    if Time.now.in_time_zone('Pacific Time (US & Canada)').hour > 8 &&
       Time.now.in_time_zone('Pacific Time (US & Canada)').hour < 20 &&
       Time.now.in_time_zone('Pacific Time (US & Canada)').on_weekday?
      ENV['MIN_WORKER_PERCENT_DAY'] || 0.9
    else
      Rails.logger.debug 'Using night time min worker percent'
      min_worker_percent
    end
  end
end
