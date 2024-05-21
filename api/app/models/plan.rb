class Plan < ApplicationRecord
  has_many :subscriptions
  has_many :accounts, through: :subscriptions
  has_many :users, through: :subscriptions

  def self.default_plan
    Plan.where(name: 'Developer').first
  end
end
