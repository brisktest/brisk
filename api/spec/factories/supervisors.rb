# frozen_string_literal: true

FactoryBot.define do
  factory :supervisor do
    project
    machine
    in_use { nil }
    ip_address { '122.232.122.1' }
    host_ip { '232.232.121.12' }
    port { '50050' }
    uid { SecureRandom.uuid }
  end
end
