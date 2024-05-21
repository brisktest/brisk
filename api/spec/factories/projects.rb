# frozen_string_literal: true

FactoryBot.define do
  factory :project do
    worker_concurrency { 2 }
    max_supervisors { 1 }
    project_token { 'my-project-token' }
    sequence(:name) { |n| "my-project-#{n}" }
    framework { 'Jest' }
    user
    image
    org
  end
end
