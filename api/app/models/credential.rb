# frozen_string_literal: true

class Credential < ApplicationRecord
  belongs_to :user, optional: true
  belongs_to :credentialable, polymorphic: true, optional: true
  belongs_to :api_actions_credential, optional: true
  encrypts :api_key
  before_validation :generate_api_key_and_token, on: :create
  validates :api_token, presence: true, uniqueness: true
  validates :api_key, presence: true, uniqueness: true

  scope :not_expired, ->(date = Time.now) { where('valid_through > ?', date).or(where(valid_through: nil)) }

  def generate_api_key_and_token
    if api_key.blank? || api_token.blank?
      self.api_key = SecureRandom.base64(12)
      self.api_token = SecureRandom.alphanumeric(10)
    else
      puts 'Not overwriting creds'
    end
  end

  def verify_api_key_and_token(key, token)
    api_token == token && api_key == key
  end

  def user
    return credentialable if credentialable.is_a?(User)

    nil
  end
end
