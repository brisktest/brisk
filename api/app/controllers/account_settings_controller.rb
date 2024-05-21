class AccountSettingsController < ApplicationController
  before_action :authenticate_user!
  # this is where we allow the user to download their credentials
  def settings; end
end
