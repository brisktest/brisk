# frozen_string_literal: true

class ApiAction < ApplicationRecord
  has_many :api_actions_credentials, dependent: :destroy
  has_many :credentials, through: :api_actions_credentials

  after_create :generate_new_credential

  def generate_new_credential
    api_actions_credentials.create!
  end
end
