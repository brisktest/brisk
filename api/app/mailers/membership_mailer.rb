class MembershipMailer < ApplicationMailer
  def invite(membership)
    @membership = membership
    @org = membership.org
    @inviter = membership.inviter
    @token = membership.token
    mail to: membership.invited_email, subject: "Invitation to join #{membership.org.name} on Brisk"
  end

  def send_request_email(org, user)
    @org = org
    @user = user
    mail to: org.admins.pluck(:email),
         subject: "#{user.name.presence || user.email} Request to join #{org.name} on Brisk"
  end
end
