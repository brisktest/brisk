# frozen_string_literal: true

FactoryBot.define do
  factory :org do
    sequence(:name, 'a') { |n| 'orgname-' + n }

    association :account
    association :owner, factory: :user
  end
end
