# frozen_string_literal: true

class ApiActionsCredential < ApplicationRecord
  belongs_to :credential
  belongs_to :api_action

  before_validation :generate_credential, on: :create

  def generate_credential
    return unless credential.nil?

    self.credential = Credential.new
    credential.save!
    save!
  end
end
