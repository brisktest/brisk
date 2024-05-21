# frozen_string_literal: true

class Project < ApplicationRecord
  FRAMEWORKS = %w[Node Jest Rspec Rails Cypress Raw Python]
  VALID_GIT_HOSTING_PROVIDERS = %w[github gitlab bitbucket]
  paginates_per 2
  belongs_to :org
  belongs_to :user # the user who created the project
  has_many :workers
  has_many :machines, through: :workers
  has_many :running_machines, -> { merge(Worker.in_use).merge(Worker.not_stale) }, through: :workers, source: :machine
  has_many :supervisors
  has_many :jobruns
  has_many :worker_run_infos
  has_many :test_files
  before_validation :generate_token, :generate_username, on: :create
  validates :project_token, :username, presence: true, uniqueness: true
  validates :name, uniqueness: { scope: :user_id }
  validates :name, presence: true
  validates :framework, presence: true, inclusion: { in: FRAMEWORKS }
  belongs_to :image
  before_validation :assign_default_image, on: :create
  before_validation :set_default_memory_requirement, on: :create

  has_one :account, through: :org
  has_one :subscription, through: :account
  has_one :plan, through: :subscription
  has_many :users, through: :org

  def assigned_workers
    workers.assigned.not_stale
  end

  def set_default_memory_requirement
    self.memory_requirement = 2600 if memory_requirement.nil?
  end

  def self.valid_frameworks
    %w[Jest Rails Rspec Cypress Python Raw]
  end

  def to_param
    project_token
  end

  def allowed_concurrency
    # we can do some accounting here too later - max runs etc
    worker_concurrency
  end

  def generate_token
    self.project_token = SecureRandom.alphanumeric(10)
  end

  def generate_username
    self.username = SecureRandom.alphanumeric(20)
  end

  def assigned_supervisor(affinity)
    supervisors.assigned.where(affinity:).first
  end

  def assigned_supervisors
    supervisors.assigned
  end

  def free_assigned_supervisor
    # this should sort us so we leave the affinity until last
    supervisors.assigned.not_in_use.order('affinity asc').first
  end

  def can_assign_supervisor?
    assigned_supervisors.count < max_supervisors
  end

  # def assign_supervisor
  #   free_assigned_supervisor.with_lock do |s|
  #     raise "Should not be in use" if s.in_use?
  #     raise "Should be assigned" unless s.assigned?
  #     s.project = self if s.project_id == nil
  #     raise "Should be this project" unless s.project_id == self.id
  #     s.save!
  #   end
  #   return s
  # end

  def assign_default_image
    return image if image

    i = case framework
        when 'Jest'
          Image.where(name: 'node-lts').last
        when 'Rails'
          Image.where(name: 'rails').last
        when 'Rspec'
          Image.where(name: 'rails').last
        when 'Raw'
          Image.where(name: 'node-lts').last
        when 'Python'
          Image.where(name: 'python').last
        when 'Cypress'
          Image.where(name: 'rails').last
        else
          Image.where(name: 'rails').last
        end
    self.image = i
  end

  def used_concurrency
    jobruns.where('created_at > ?', 1.month.ago).where('assigned_concurrency > ?', 1).sum(:assigned_concurrency) || 0
  end

  def get_capacity
    get_remaining_capacity > plan.minimum_capacity ? get_remaining_capacity : plan.minimum_capacity
  end

  def get_remaining_capacity
    plan.monthly_concurrency - used_concurrency
  end

  def max_workers
    # max workers is the maximum number of workers we can have
    # this is the concurrency * the number of supervisors
    # we want to have a couple of extra workers for when we miss a few when syncing

    # I'm trying to avoid a situation where we always are one short
    # and we are constantly having to build a new worker
    # perhaps we should make sure and keep this tighter and remove the 2
    # because we really shouldn't have other parts failing and leaving
    # the workers hanging - essentially it means we need 2X the number of workers
    (worker_concurrency * max_supervisors * 2 * 1.1).ceil

    # we have a situation where supervisors crash (usually OOM and we end up with not enough workers cause they are stuck)
    # we can lean on the supervisor to limit the number of workers - so perhaps have 2X the supervisors allotment
  end

  def balance_workers
    # we want to free up workers that are reserved for this project
    # going to find the one that was freed the longest ago

    while workers.in_use.assigned.not_stale.free.count > max_workers
      # I'm trying to continually get the workers that is most congested
      machine_count = workers.in_use.assigned.not_stale.group_by(&:machine_id)
      w = workers.in_use.assigned.not_stale.free.sort_by { |w| machine_count[w.machine_id].size }.last
      w.with_lock do
        Rails.logger.debug("Trying to free worker #{w.id} from project #{w.project_id}")
        next unless w && w.assigned? && w.free?

        Rails.logger.debug("calling de_register! on worker #{w.id}")
        w.de_register!
      end
    end
    Rails.logger.debug("workers <= max_workers is #{workers.in_use.assigned.not_stale.free.count <= max_workers}")
    Rails.logger.debug("At end of balance workers which suggests we may not have freed enough workers for project #{id}
       assigned workers: #{workers.in_use.assigned.not_stale.free.count} max workers: #{max_workers}")
    workers.in_use.assigned.not_stale.free.count <= max_workers
  end

  def split_files(num_buckets, filenames)
    Rails.logger.debug("Splitting files for project #{id} with #{num_buckets} buckets and filenames #{filenames}")
    tss = TestSplitterService.new(self, num_buckets, filenames)
    split = tss.retrieve_split
    return split, 'partition-pre-split' if split

    [tss.default_split, 'default']
  end

  # we can use this to reduce the number of workers we have
  # assigned without clearing them all
  def release_old_workers(time = 1.hour)
    workers.assigned.where('reserved_at < ?', time.ago).each do |w|
      Rails.logger.info "Releasing worker #{w.id} for project #{id}"
      w.safe_release
    end
  end

  def has_git_repo_info?
    git_hosting_provider.present? && git_org.present? && git_repo_name.present?
  end

  def is_github_repo?
    git_hosting_provider == 'github' && git_org.present? && git_repo_name.present?
  end

  def is_gitlab_repo?
    git_hosting_provider == 'gitlab' && git_org.present? && git_repo_name.present?
  end
end
