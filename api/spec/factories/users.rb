# frozen_string_literal: true

FactoryBot.define do
  factory :user do
    sequence(:email) { |_n| "user#{rand}@factory.com" }
    password { 'sfsdfsdfsdfsdfs' }
    sequence(:name) { |n| "Me#{n}" }
  end
end
