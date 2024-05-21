# frozen_string_literal: true

# account is for billing purposes
# org is for managing projects and users who are in the same org, mostly for
# access and billing purposes - one account for an org
# a company can have multiple orgs and multiple accounts in those orgs

class Account < ApplicationRecord
  has_one :org
  has_one :subscription

  before_validation :ensure_subscription, on: :create

  def ensure_subscription
    return unless subscription.nil?
    raise 'No default plan found - please make sure to seed the database with a default plan' if Plan.default_plan.nil?

    build_subscription plan: Plan.default_plan
    subscription.save!
  end
end
