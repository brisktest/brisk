# frozen_string_literal: true

FactoryBot.define do
  factory :worker do
    association :machine
    ip_address { Faker::Internet.ip_v4_address }
    host_ip { Faker::Internet.ip_v4_address }
    uid { SecureRandom.uuid }
    last_checked_at { Time.now }
    worker_image { 'rails' }
    image { FactoryBot.create(:image, name: 'rails') }
    freed_at { Time.now }
  end

  factory :busy_worker, class: 'Worker' do
    association :machine
    ip_address { Faker::Internet.ip_v4_address }
    host_ip { Faker::Internet.ip_v4_address }
    uid { SecureRandom.uuid }
    last_checked_at { Time.now }
    worker_image { 'rails' }
    image { FactoryBot.create(:image, name: 'rails') }
    # this doesn't work cause we set the freed at in the before validation
    freed_at { nil }
    reserved_at { 1.minute.ago }
  end
end
