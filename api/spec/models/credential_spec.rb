# spec/models/credential_spec.rb

require 'rails_helper'

RSpec.describe Credential, type: :model do
  describe 'scopes' do
    describe '.not_expired' do
      it 'includes credentials with a valid_through greater than the current time' do
        valid_credential = create(:credential, valid_through: 1.hour.from_now)
        expect(Credential.not_expired).to include(valid_credential)
      end

      it 'includes credentials with a nil valid_through' do
        credential_without_expiry = create(:credential, valid_through: nil)
        expect(Credential.not_expired).to include(credential_without_expiry)
      end

      it 'excludes credentials with a valid_through in the past' do
        expired_credential = create(:credential, valid_through: 1.hour.ago)
        expect(Credential.not_expired).not_to include(expired_credential)
      end
    end
  end

  describe 'methods' do
    describe '#generate_api_key_and_token' do
      it 'generates api_key and api_token if they are blank' do
        credential = build(:credential, api_key: nil, api_token: nil)
        credential.generate_api_key_and_token
        expect(credential.api_key).not_to be_nil
        expect(credential.api_token).not_to be_nil
      end

      it 'does not overwrite api_key and api_token if they are already present' do
        api_key = 'existing_api_key'
        api_token = 'existing_api_token'
        credential = build(:credential, api_key:, api_token:)
        credential.generate_api_key_and_token
        expect(credential.api_key).to eq(api_key)
        expect(credential.api_token).to eq(api_token)
      end
    end

    describe '#verify_api_key_and_token' do
      it 'returns true when api_key and api_token match' do
        credential = create(:credential)
        expect(credential.verify_api_key_and_token(credential.api_key, credential.api_token)).to be true
      end

      it 'returns false when api_key and api_token do not match' do
        credential = create(:credential)
        expect(credential.verify_api_key_and_token('wrong_key', 'wrong_token')).to be false
      end
    end

    describe '#user' do
      it 'returns the associated user if credentialable is a User' do
        user = create(:user)
        credential = create(:credential, credentialable: user)
        expect(credential.user).to eq(user)
      end
    end
  end
end
