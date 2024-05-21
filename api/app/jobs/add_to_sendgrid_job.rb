class AddToSendgridJob < ApplicationJob
  def perform(user_id)
    user = User.find(user_id)
    return if user.add_to_sendgrid

    Rails.logger.info "Failed to add #{user.email} to sendgrid"
    raise 'Failed to add user to sendgrid'
  end
end
