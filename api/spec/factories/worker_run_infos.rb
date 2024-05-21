FactoryBot.define do
  factory :worker_run_info do
    association :worker
    association :project
    association :supervisor
    association :jobrun
  end
end
