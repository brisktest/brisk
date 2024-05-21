# frozen_string_literal: true

class ApplicationMailer < ActionMailer::Base
  default from: 'support@brisktest.com'
  layout 'mailer'
end
