class Jobrun < ApplicationRecord
  include AASM
  paginates_per 50

  belongs_to :supervisor
  has_one :project, through: :supervisor
  has_one :repo_info
  scope :started, -> { where('state = ?', :starting) }
  scope :failed, -> { where('state = ?', :failed) }
  scope :completed, -> { where('state = ?', :completed) }
  scope :running, -> { where('state = ?', :running) }

  scope :finished, -> { where('state = ? OR state = ?', :completed, :failed) }

  validates :state, presence: true, inclusion: { in: %w[completed failed starting running unfulfilled] }

  before_validation :set_uid, on: :create

  validates :uid, presence: true, uniqueness: true

  has_one :user, through: :project
  has_many :worker_run_infos, dependent: :destroy
  has_many :test_files, through: :worker_run_infos
  has_many :workers
  aasm column: 'state' do
    state :starting, initial: true
    state :running
    state :failed
    state :completed
    state :unfulfilled

    event :finish do
      transitions from: %i[starting running], to: :completed
    end

    event :fail do
      transitions to: :failed, after: :cleanup_failed
    end

    event :unfulfill do
      transitions from: :starting, to: :unfulfilled, after: :cleanup_unfulfilled # this is for when we have no workers
    end

    # should I make sure to release workers here now that we associate?
  end

  def finished?
    completed? || failed?
  end

  def cleanup_failed
    Rails.logger.debug "Cleaning up failed jobrun #{id}"
    update(finished_at: Time.now)
    free_resources
  end

  def cleanup_unfulfilled
    Rails.logger.debug "Cleaning up unfulfilled jobrun #{id}"
    update(finished_at: Time.now)
    free_resources
  end

  def free_resources
    Rails.logger.debug "Freeing resources for jobrun #{id}"
    workers.each do |w|
      w.free_from_super
    end
    supervisor.release
  end

  def duration
    finished = finished_at || Time.now
    (finished - created_at).to_i
  end

  def set_uid
    self.uid = SecureRandom.uuid unless uid
  end

  def previous_jobrun
    project.jobruns.succeeded.where('created_at < ?', created_at).order(created_at: :desc).first
  end

  def parse_trace_key
    return nil unless trace_key

    begin
      JSON.parse(trace_key).first
    rescue StandardError
      trace_key
    end
  end

  # going to add a schedule in here
  # I think at night we are not going to want to have as many workers so we can accept fewer workers
  # and then we can have more workers during the day
  def get_schedule
    project.org.schedule || Schedule.default_schedule
  end

  def not_enough_workers?
    min_worker_percent = get_schedule.get_current_min_worker_percent
    raise 'min_worker_percent is nil' unless min_worker_percent

    (assigned_concurrency || 0) < min_worker_percent * concurrency.to_f
  end

  def add_note(note)
    self.notes = notes || '' + note
    save!
  end

  def debug_worker_info
    "#WORKER_INFO_STATS self.project.workers.assigned.count  = #{project.workers.assigned.count}
    self.project.workers.not_stale.count = #{project.workers.not_stale.count}
    self.project.workers.not_stale.busy = #{project.workers.not_stale.busy.count}
    self.project.workers.not_stale.free = #{project.workers.not_stale.free.count}
    total free workers for this workers image #{project.image} = #{Worker.where(image: project.image).not_stale.ready.free.count}
    max workers for this project = #{project.max_workers}
    "
  end

  # this measures the ms time of all the workers
  def total_worker_time
    worker_run_infos.to_a.sum(&:duration)
  end
end
