# frozen_string_literal: true

class Worker < ApplicationRecord
  include AASM
  belongs_to :project, optional: true
  belongs_to :machine
  belongs_to :image
  belongs_to :supervisor, optional: true
  belongs_to :jobrun, optional: true
  has_many :worker_run_infos
  validates :uid, uniqueness: true

  scope :with_no_project, -> { active.where(project_id: nil) }

  scope :busy, -> { where(freed_at: nil) }
  scope :free, -> { where.not(freed_at: nil) }
  scope :free_workers, -> { free.not_stale }
  scope :machine_not_draining, -> { joins(:machine).where('machines.drained_at IS NULL') }

  # scope :with_memory_space, ->(ram) { joins(:machine).where("sum(workers.assigned_ram) + ? < machines.ram", ram) }
  # scope :on_machine_with_free_mem, ->(ram) { joins(:machine).where("sum(workers.assigned_ram) + ? < machines.ram", ram) }

  scope :with_distinct_host_ip, -> { select('DISTINCT ON (host_ip) *') }
  scope :active, -> { where(state: :active) }
  # TODO: - refactor this is confusing
  scope :in_use, -> { where("workers.state != 'finished'") }
  # stale_timeout = ENV["STALE_WORKER_CUT_OFF"] || 520
  scope :not_stale, lambda {
                      ENV['DEV'] ? where(state: %i[active assigned]) : where('last_checked_at > ?', 120.seconds.ago)
                    }

  # nothing is stale when we are in dev
  scope :stale, -> { !ENV['DEV'] ? where('last_checked_at < ?', 120.seconds.ago) : where('id < 0') }

  scope :assigned, -> { where(state: :assigned) }
  scope :hash_or_empty, lambda { |rebuild_hash|
                          where("rebuild_hash = ? OR rebuild_hash = '' OR rebuild_hash is NULL ", rebuild_hash)
                        }
  scope :with_rebuild_hash, ->(rebuild_hash) { where(rebuild_hash:) }

  scope :my_super_id_or_none, ->(super_id) { where('supervisor_id = ? OR supervisor_id is NULL ', super_id) }
  scope :with_build_commands, -> { where.not(build_commands_run_at: nil) }
  scope :without_build_commands, -> { where(build_commands_run_at: nil) }
  scope :finished, -> { where(state: :finished) }

  scope :without_host_ips, ->(host_ips) { where.not(host_ip: host_ips) }
  scope :without_project, -> { where(project_id: nil) }

  scope :with_worker_image, ->(worker_image) { (where worker_image:) }

  aasm column: 'state' do
    state :active, initial: true
    state :ready
    state :assigned
    state :finished

    event :reserve do
      transitions from: %i[ready active], to: :assigned
    end

    event :de_register do
      transitions to: :finished
    end
  end

  before_validation :set_freed_at, on: :create

  def set_freed_at
    self.freed_at = Time.now if freed_at.nil?
  end

  def free?
    freed_at?
  end

  def self.safe_assign(workers, project, num_to_assign, supervisor_id, jobrun_id)
    return [] if num_to_assign <= 0

    Rails.logger.debug "Trying to assign #{num_to_assign} workers to supervisor #{supervisor_id} - workers: #{workers.size}"
    Rails.logger.debug "Workers: #{workers.map(&:inspect)}"
    new_workers = []
    supervisor = Supervisor.find(supervisor_id)
    workers.each do |w|
      w.with_lock do
        next unless w.assigned? || w.active?

        if w.assigned? && w.project_id != project.id
          Rails.logger.info "Worker #{w.id} already assigned to project #{w.project_id} and maybe supervisor #{w.supervisor_id}"
          next
        end

        if w.machine.memory_used + project.memory_requirement > w.machine.memory
          Rails.logger.info "Cannot assign worker: Worker #{w.id} on machine #{w.machine.id} has #{w.machine.memory_used} used and #{project.memory_requirement} required"
          next
        end
        Rails.logger.debug "Assigning worker #{w.id} to project #{project.id} and supervisor #{supervisor_id}"
        w.project_id = project.id
        w.supervisor_id = supervisor_id
        w.freed_at = nil
        w.assigned_ram = project.memory_requirement
        w.reserved_at = Time.now
        w.jobrun_id = jobrun_id
        w.save!
        w.reserve! unless w.assigned?
        Rails.logger.debug "Assigned worker #{w.id} to project #{project.id} and supervisor #{supervisor_id}"
        new_workers << w
        Rails.logger.debug "new_workers is now: #{new_workers.size} - #{new_workers.map(&:inspect)}"
      end

      return new_workers if new_workers.size >= num_to_assign
    end
    Rails.logger.info "Not sure why we don't have enough new_workers but returning #{new_workers.size} workers now"
    new_workers
  end

  # TODO: create a worker run to track individual running
  # could even group update it at the finish_run level

  def safe_release
    with_lock do |_s|
      # raise "Not free" unless self.freed_at?
      return false unless freed_at?

      de_register!
      save!
    end
  end

  def log_run(wi, s, no_test_files = false)
    MyAppTracer.in_span('do_work') do |_span|
      Rails.logger.debug("Log run for #{wi.inspect} and #{s.inspect} ")
      unless wi.finished_at
        Rails.logger.error("No finished_at for #{wi.inspect} setting it to now")

        @finished_at = Time.now
      end
      unless wi.started_at
        Rails.logger.error("No started_at for #{wi.inspect} setting it to start of jobrun")
        @started_at = jobrun ? jobrun.created_at : Time.now
      end

      worker_info = worker_run_infos.create!({ supervisor_id: s.id, log_location: wi.log_location, uid: wi.uid, jobrun_id: wi.jobrun_id,
                                               rebuild_hash: wi.rebuild_hash.delete("\u0000"), exit_code: wi.exit_code.delete("\u0000"),
                                               output: wi.output.delete("\u0000"), error: wi.error.delete("\u0000"),
                                               project_id:, finished_at: @finished_at || wi.finished_at.to_time,
                                               started_at: @started_at || wi.started_at.to_time, ms_time_taken: ((@finished_at || wi.finished_at.to_time) - (@started_at || wi.started_at.to_time)) * 1000 })

      if no_test_files
        Rails.logger.debug("LogRun: Timing - after compare_to #{id}")
        self.rebuild_hash = wi.rebuild_hash if wi.rebuild_hash.present?
        save!
        return worker_info
      end

      execution_infos = wi.execution_infos
      execution_infos.each do |execution_info|
        next unless execution_info.started

        worker_info.execution_infos.create!({
                                              started: execution_info.started.to_time,
                                              finished: execution_info&.finished&.to_time || Time.now,
                                              exit_code: execution_info.exit_code,
                                              rebuild_hash: execution_info.rebuild_hash,
                                              command: Command.new_from_proto(execution_info.command),
                                              output: execution_info.output
                                            })
      end

      worker_info.save!
      worker_info.record_test_file_run wi
      worker_info.set_test_file_count
      Rails.logger.debug("LogRun: Timing - before compare_to #{id}")

      begin
        # we wait so any other workers can finish and the contention numbers will be accurate
        TimingAnalysisJob.perform_later worker_info.id
      rescue StandardError => e
        Rails.logger.error("Error in log_run #{e} - proceeding")
        Sentry.capture_exception(e)
      end
      Rails.logger.debug("LogRun: Timing - after compare_to #{id}")
      self.rebuild_hash = wi.rebuild_hash if wi.rebuild_hash.present?
      save!
      return worker_info
    end
  end

  def free_from_super
    Rails.logger.debug do
      "Freeing worker #{id} from supervisor #{supervisor_id} and jobrun #{jobrun_id}"
    end
    with_lock do |_s|
      Rails.logger.debug { "acquired lock for #{id}" }
      unless assigned?
        Rails.logger.debug { "Freeing worker: worker #{id} not assigned" }
        return nil
      end
      unless supervisor_id?
        Rails.logger.debug { "Freeing worker: worker #{id} not assigned to supervisor" }
        return nil
      end
      unless freed_at.nil?
        Rails.logger.debug { "Freeing worker: worker #{id} already freed" }
        return nil
      end

      self.supervisor_id = nil
      self.freed_at = Time.now
      self.assigned_ram = 0
      self.reserved_at = nil
      self.jobrun_id = nil

      save!

      Rails.logger.debug { "saved #{id} #{inspect}" }
    end
  end
end
