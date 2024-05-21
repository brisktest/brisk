FactoryBot.define do
  factory :jobrun do
    state { 'starting' }
    supervisor
    assigned_concurrency { 1 }
  end
end
