require 'rails_helper'

RSpec.describe Membership, type: :model do
  let(:org) { create(:org) }
  let(:user) { create(:user) }
  let(:accepted_user) { create(:user) }
  let(:inviter) { create(:user) }
  let(:membership) { create(:membership, org:, user:, inviter:) }

  describe 'scopes' do
    let!(:accepted_membership) { create(:membership, accepted_at: Time.now, user: accepted_user) }
    let!(:cancelled_membership) { create(:membership, cancelled_at: Time.now) }
    let!(:pending_membership) { create(:membership) }

    it 'returns accepted memberships' do
      expect(Membership.accepted).to include(accepted_membership)
      expect(Membership.accepted).not_to include(cancelled_membership, pending_membership)
    end

    it 'returns cancelled memberships' do
      expect(Membership.cancelled).to include(cancelled_membership)
      expect(Membership.cancelled).not_to include(accepted_membership, pending_membership)
    end

    it 'returns pending memberships' do
      expect(Membership.pending).to include(pending_membership)
      expect(Membership.pending).not_to include(accepted_membership, cancelled_membership)
    end

    it 'returns active memberships' do
      expect(Membership.active).to include(accepted_membership)
      expect(Membership.active).not_to include(cancelled_membership, pending_membership)
    end
  end

  describe '#state' do
    it 'returns "accepted" if accepted_at is present and cancelled_at is not' do
      membership.accepted_at = Time.now
      membership.cancelled_at = nil
      expect(membership.state).to eq('accepted')
    end

    it 'returns "pending" if accepted_at and cancelled_at are both nil' do
      membership.accepted_at = nil
      membership.cancelled_at = nil
      expect(membership.state).to eq('pending')
    end

    it 'returns "cancelled" if cancelled_at is present' do
      membership.accepted_at = nil
      membership.cancelled_at = Time.now
      expect(membership.state).to eq('cancelled')
    end
  end

  describe '#set_default_role' do
    it 'sets the default role to "member" if role is not set' do
      membership.role = nil
      membership.set_default_role
      expect(membership.role).to eq('member')
    end

    it 'does not override the role if it is already set' do
      membership.role = 'admin'
      membership.set_default_role
      expect(membership.role).to eq('admin')
    end
  end

  describe '#set_token' do
    it 'sets a unique token if token is not already set' do
      membership.token = nil
      membership.set_token
      expect(membership.token).not_to be_nil
    end

    it 'does not override the token if it is already set' do
      membership.token = 'existing_token'
      membership.set_token
      expect(membership.token).to eq('existing_token')
    end
  end

  describe '#accept!' do
    it 'assigns the user, sets accepted_at, and saves the membership' do
      new_user = create(:user)
      membership.accept!(new_user)
      expect(membership.user).to eq(new_user)
      expect(membership.accepted_at).not_to be_nil
      expect(membership).to be_persisted
    end
  end

  describe '#accepted?' do
    it 'returns true if accepted_at is present' do
      membership.accepted_at = Time.now
      expect(membership.accepted?).to be_truthy
    end

    it 'returns false if accepted_at is not present' do
      membership.accepted_at = nil
      expect(membership.accepted?).to be_falsy
    end
  end

  describe '#cancel!' do
    it 'sets cancelled_at and saves the membership' do
      membership.cancel!
      expect(membership.cancelled_at).not_to be_nil
      expect(membership).to be_persisted
    end
  end

  describe '#is_expired?' do
    it 'returns true if created_at is more than 1 week ago' do
      membership.created_at = 2.weeks.ago
      expect(membership.is_expired?).to be_truthy
    end

    it 'returns false if created_at is within 1 week' do
      membership.created_at = 3.days.ago
      expect(membership.is_expired?).to be_falsy
    end
  end

  describe '#change_role' do
    it 'changes the role and saves the membership' do
      membership.change_role('admin')
      expect(membership.role).to eq('admin')
      expect(membership).to be_persisted
    end
  end

  describe '#admin?' do
    it 'returns true if the role is "admin"' do
      membership.role = 'admin'
      expect(membership.admin?).to be_truthy
    end

    it 'returns false if the role is not "admin"' do
      membership.role = 'member'
      expect(membership.admin?).to be_falsy
    end
  end
end
