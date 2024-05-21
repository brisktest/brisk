# frozen_string_literal: true

class Org < ApplicationRecord
  has_many :projects
  belongs_to :account
  belongs_to :owner, class_name: 'User'
  belongs_to :schedule, optional: true
  has_many :memberships
  has_many :admins, lambda {
                      where("memberships.cancelled_at": nil).where.not("memberships.accepted_at": nil).where({ 'memberships.role' => 'admin' })
                    }, through: :memberships, source: :user
  validates :name, presence: true, uniqueness: true, length: { minimum: 3, maximum: 50 },
                   format: { with: /\A[a-zA-Z0-9-]+\Z/ }
  after_create :create_membership_for_owner
  before_validation :ensure_account, on: :create
  has_many :users, lambda {
                     where.not('memberships.accepted_at' => nil).where('memberships.cancelled_at' => nil)
                   }, through: 'memberships', source: :user

  def ensure_account
    return unless account.nil?

    self.account = Account.create!
  end

  def create_membership_for_owner
    Membership.create!(org: self, user: owner, role: :admin, accepted_at: Time.now)
  end

  def to_param
    name
  end

  def is_manager?(user)
    memberships.active.where(user_id: user.id, role: :admin).any? || user.admin?
  end

  def is_member?(user)
    memberships.active.where(user_id: user.id, role: :member).any? || is_manager?(user)
  end

  def is_invited?(user)
    memberships.where(user_id: user.id, accepted_at: nil).any?
  end

  def status_for_user(u)
    if is_manager?(u)
      :admin
    elsif is_member?(u)
      :member
    elsif is_invited?(u)
      :invited
    else
      :none
    end
  end
end
