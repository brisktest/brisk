FactoryBot.define do
  factory :subscription do
    association :plan
    association :account
  end
end
