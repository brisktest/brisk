FactoryBot.define do
  factory :execution_info do
    association :command
    association :worker_run_info
    started { Time.now }
    finished { Time.now + 1 }
    duration { 1 }
    status { 0 }
    output { 'output' }
  end
end
