class Admin::AdminController < ApplicationController
  before_action :authenticate_user!
  before_action :authenticate_admin!

  private

  def authenticate_admin!
    redirect_to root_path, alert: 'You must be an admin to do that.' unless current_user && current_user.admin?
  end
end
