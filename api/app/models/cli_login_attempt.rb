class CliLoginAttempt < ApplicationRecord
  belongs_to :user, optional: true
  validates :nonce, presence: true, uniqueness: true, length: { minimum: 32 }
  validates :token, presence: true, uniqueness: true, length: { minimum: 32 }
  validates :valid_until, presence: true

  def self.create_from_nonce(nonce)
    token = SecureRandom.hex(32)
    valid_until = Time.now + 120.seconds
    CliLoginAttempt.create!(nonce:, token:, valid_until:)
  end

  def not_valid
    valid_until < Time.now
  end
end
