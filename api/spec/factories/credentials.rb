# spec/factories/credentials.rb

FactoryBot.define do
  factory :credential do
    # Generate a unique API token and API key for each created credential
    sequence(:api_token) { |n| "api_token_#{n}" }
    sequence(:api_key) { |n| "api_key_#{n}" }

    # Associate the credential with a user (assuming you have a User factory)
    association :credentialable, factory: :user
  end
end
