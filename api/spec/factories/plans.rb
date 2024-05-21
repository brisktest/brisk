# frozen_string_literal: true

FactoryBot.define do
  factory :plan do
    name { 'Developer' }
    trial_period { '14 days' }
    monthly_concurrency { 1200 }
  end
end
