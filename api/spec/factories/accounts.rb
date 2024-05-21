# frozen_string_literal: true

FactoryBot.define do
  factory :account do
    after(:create) { |a| create(:subscription, account: a) }
  end
end
