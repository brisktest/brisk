# spec/factories/memberships.rb
FactoryBot.define do
  factory :membership do
    invited_email { Faker::Internet.email }
    org { association(:org) }
    inviter { association(:user) }
    token { SecureRandom.hex(8) }
  end
end
