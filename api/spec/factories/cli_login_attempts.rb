FactoryBot.define do
  factory :cli_login_attempt do
    association :user
    token {  SecureRandom.hex(32) }
    nonce {  SecureRandom.hex(32) }
    valid_until { Time.now + 120.seconds }
  end
end
