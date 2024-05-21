# frozen_string_literal: true

class Machine < ApplicationRecord
  has_many :workers
  has_many :worker_run_infos, through: :workers
  has_many :supervisors

  scope :running, -> { where(finished_at: nil) }
  scope :not_draining, -> { where(drained_at: nil) }
  scope :finished, -> { where.not(finished_at: nil) }

  scope :running_workers, -> { joins(:workers).merge(Worker.not_stale.busy) }
  scope :not_stale, -> { where('last_ping_at > ?', 5.minutes.ago) }
  validates :uid, uniqueness: true

  def memory_oversubscribed?
    workers.not_stale.busy.sum(:assigned_ram) > memory
  end

  def finished?
    finished_at?
  end

  def running?
    !finished?
  end

  def memory_used
    workers.not_stale.busy.sum(:assigned_ram)
  end

  def workers_running
    workers.not_stale.busy.count
  end

  # 2 seems to be the minimum vpcu's for a worker on aws
  def free_cpu
    (cpus || 2) - workers_running
  end

  def overlapping_worker_run_infos(wri)
    last_execution_info = wri.execution_infos.order('started DESC').first
    return [] unless last_execution_info

    worker_run_infos.where(
      'worker_run_infos.id != ? AND ((started_at < ? AND finished_at > ?) OR (started_at < ? AND finished_at > ?))', wri.id, last_execution_info.finished, last_execution_info.started, last_execution_info.finished, last_execution_info.started
    ).uniq
  end

  def no_still_running_workers_from(t)
    workers.busy.assigned.where('reserved_at < ?', t).count == 0
  end
end
