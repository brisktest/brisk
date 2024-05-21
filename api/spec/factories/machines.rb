# frozen_string_literal: true

FactoryBot.define do
  factory :machine do
    ip_address { Faker::Internet.ip_v4_address }
    uid { Faker::Alphanumeric.alpha(number: 10) }
    memory { 8900 }
    cpus { 2 }
  end
end
