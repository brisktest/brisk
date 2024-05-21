require 'rails_helper'

RSpec.describe MembershipMailer, type: :mailer do
  let(:membership) { create(:membership) }
  let(:org) { membership.org }
  let(:user) { create(:user) }

  describe '#invite' do
    let(:mail) { MembershipMailer.invite(membership) }

    it 'renders the headers correctly' do
      expect(mail.subject).to eq("Invitation to join #{org.name} on Brisk")
      expect(mail.to).to eq([membership.invited_email])
      expect(mail.from).to eq(['support@brisktest.com']) # Replace with your own email
    end

    it 'renders the body correctly' do
      expect(mail.body.encoded).to include(membership.org.name)
      expect(mail.body.encoded).to include(membership.inviter.name)
      expect(mail.body.encoded).to include(membership.token)
    end
  end

  describe '#send_request_email' do
    let(:mail) { MembershipMailer.send_request_email(org, user) }

    it 'renders the headers correctly' do
      expect(mail.subject).to eq("#{user.name.presence || user.email} Request to join #{org.name} on Brisk")
      expect(mail.to).to eq(org.admins.pluck(:email))
      expect(mail.from).to eq(['support@brisktest.com']) # Replace with your own email
    end

    it 'renders the body correctly' do
      expect(mail.body.encoded).to include((user.name.presence || user.email))
      expect(mail.body.encoded).to include(org.name)
    end
  end
end
