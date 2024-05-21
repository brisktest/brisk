# frozen_string_literal: true

class Supervisor < ApplicationRecord
  include AASM
  belongs_to :machine
  belongs_to :project, optional: true
  has_many :workers

  # validates_uniqueness_of :project, scope: :state, conditions: -> { where(state: 'assigned') }
  validates :ip_address, presence: true, if: :assigned?
  validates :port, presence: true, if: :assigned?
  validates :uid, uniqueness: true
  scope :assigned, -> { where(state: :assigned) }
  scope :ready, -> { where(state: :ready) }
  scope :ready_and_assigned, -> { where(state: %i[ready assigned]) }

  scope :in_use, -> { where.not(in_use: nil) }
  scope :not_in_use, -> { where(in_use: nil) }

  aasm column: 'state' do
    state :ready, initial: true
    state :assigned
    state :finished

    event :revive_assigned do
      transitions from: :finished, to: :assigned
    end

    event :revive_ready do
      transitions from: :finished, to: :ready
    end

    event :reserve do
      transitions from: :ready, to: :assigned
    end

    event :de_register do
      transitions to: :finished, after: :cleanup_workers
    end
  end

  def cleanup_workers
    if workers.count > 0
      Rails.logger.debug "Cleaning up workers for supervisor #{id}"
      workers.busy.each do |w|
        Rails.logger.info "Cleaning up worker #{w.id}"
        w.free_from_super
        w.de_register!
        w.save!
      end
    else
      Rails.logger.debug "No workers to clean up for supervisor #{id}"
    end
  end

  # def revive!
  #   if self.project_id
  #     revive_assigned!
  #   else
  #     revive_ready!
  #   end
  #   save!
  # end

  def endpoint
    "#{ip_address}:#{port}"
  end

  # def free_workers(workers_info, jobrun)
  #   Rails.logger.info "Freeing workers for supervisor #{id}, we have #{workers.count} assigned workers"
  #   Rails.logger.debug "Workers info: #{workers_info.inspect}"
  #   workers_info.each do |wi|
  #     worker = self.workers.find(wi.worker_id)

  #     #   message RunInfo {
  #     #     uint32 worker_id =1;;
  #     #     string rebuild_hash = 2;
  #     #     string exit_code = 3;
  #     #     string output = 4;
  #     #     google.protobuf.Timestamp finished_at = 5;
  #     #     google.protobuf.Timestamp started_at = 6;
  #     #     string error = 7;
  #     # }
  #     #WorkerRunInfo(id: integer, rebuild_hash: text, exit_code: text, output: text, finished_at: datetime, started_at: datetime, error: text, worker_id: integer, created_at: datetime, updated_at: datetime, project_id: integer, supervisor_id: integer)

  #     worker_info = worker.worker_run_infos.create!({ log_location: wi.log_location, uid: wi.uid, log_encryption_key: wi.log_encryption_key, jobrun_id: jobrun.id, rebuild_hash: wi.rebuild_hash.delete("\u0000"), exit_code: wi.exit_code.delete("\u0000"), output: wi.output.delete("\u0000"), error: wi.error.delete("\u0000"), supervisor_id: self.id, project_id: self.project_id, finished_at: wi.finished_at.to_time, started_at: wi.started_at.to_time, ms_time_taken: (wi.finished_at.to_time - wi.started_at.to_time) * 1000 })
  #     if worker_info.succeeded?
  #       wi.files.each do |filename|
  #         test_file = worker_info.project.test_files.find_or_create_by! filename: filename
  #         # we have an adjustment to make here because we don't want to include the startup time
  #         # this will vary per project so we allow it to be configurable
  #         adjusted_time_taken = ((worker_info.ms_time_taken - (project.startup_time_in_ms || 1750)))
  #         adjusted_time_taken = 100 if adjusted_time_taken < 100
  #         adjusted_time_taken = adjusted_time_taken / wi.files.size
  #         worker_info.test_file_runs.create! test_file: test_file, worker_run_info: worker_info, ms_time_taken: adjusted_time_taken
  #         if wi.files.size == 1
  #           test_file.runtime = adjusted_time_taken
  #           test_file.timing_confidence = 1
  #           test_file.save!
  #         elsif test_file.runtime.nil?
  #           test_file.runtime = (adjusted_time_taken / wi.files.size.to_f).to_i
  #           # test_file.timing_confidence = 1 / wi.files.size.to_f
  #           test_file.runtime = adjusted_time_taken / wi.files.size.to_f

  #           test_file.save!
  #         elsif (test_file.timing_confidence || 0) < 1
  #           test_file.runtime = adjusted_time_taken / wi.files.size.to_f
  #           #test_file.runtime = (test_file.runtime * test_file.timing_confidence + adjusted_time_taken / wi.files.size.to_f) / ((test_file.timing_confidence + 1).to_f)
  #           #test_file.timing_confidence = (test_file.timing_confidence + 1) / 2.0
  #           test_file.save!

  #           # we hit stalemates quickly (like if we have a big and small file)
  #           # what if we randomly assigned info to the test files to shake them up out of local max/min
  #           # we could assign the value with a variable "confidence/probabliity" which would force all the tests away from each other
  #           # then, we could differentiate between files quicker
  #         end
  #       end
  #     end
  #   end
  #   workers.update_all supervisor_id: nil, freed_at: Time.now
  # end

  def assign_to_project(project)
    raise 'Supervisor is not ready' unless ready?

    self.project = project
    # sup.unique_instance_id = unique_instance_id
    reserve!
    save!
  end

  def release
    in_use = nil
    save!
  end
end
