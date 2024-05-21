class Membership < ApplicationRecord
  belongs_to :user, optional: true # before they are claimed there is no user
  belongs_to :org
  belongs_to :inviter, class_name: 'User', optional: true # the first person is not invited
  validates :user_id, uniqueness: { scope: :org_id, conditions: lambda {
                                                                  active
                                                                }, message: 'a user can only be a member of an org one time' }
  validates :org_id, presence: true
  validates :user_id,
            presence: { message: 'once a membership is accepted it must have an active user', if: :accepted_at }
  validates :role, presence: true
  validates :invited_email, presence: true, unless: :user_id
  validates :invited_email, uniqueness: { scope: :org_id, conditions: -> { active } }, unless: :user_id
  validates :token, uniqueness: true
  validates :token, presence: true
  validates :invited_email, format: { with: URI::MailTo::EMAIL_REGEXP }, unless: :user_id
  ROLES = %w[member admin]
  validates :role, inclusion: { in: Membership::ROLES }

  before_validation :set_default_role, on: :create
  before_validation :set_token, on: :create

  scope :accepted, -> { where.not(accepted_at: nil) }
  scope :cancelled, -> { where.not(cancelled_at: nil) }
  scope :pending, -> { where(accepted_at: nil).where(cancelled_at: nil) }

  scope :active, -> { where(cancelled_at: nil).where.not(accepted_at: nil) }

  def state
    if accepted_at && !cancelled_at
      'accepted'
    elsif !accepted_at && !cancelled_at
      'pending'
    else
      'cancelled'
    end
  end

  def set_default_role
    self.role ||= :member
  end

  def set_token
    self.token ||= SecureRandom.hex(10)
  end

  def accept!(user)
    self.user = user
    self.accepted_at = Time.now
    save!
  end

  def accepted?
    accepted_at.present?
  end

  def cancel!
    self.cancelled_at = Time.now
    save!
  end

  def is_expired?
    created_at < 1.week.ago
  end

  def change_role(new_role)
    self.role = new_role
    save!
  end

  def admin?
    self.role == 'admin'
  end
end
