class CliController < ApplicationController
  before_action :set_sign_in_path, only: [:auth]
  before_action :authenticate_user!

  # should I have a confirm here? It's possible to send this link to someone else and have them open it
  # which would result in them being logged in
  # could add a confirm here to make sure they really want to do this
  def auth
    @token = params[:token]
    cla = CliLoginAttempt.find_by(token: @token, user: nil)
    if cla.nil?
      @error = 'Token not found'
    elsif cla.not_valid
      @error = 'Token expired'
    end
  end

  def confirm_auth
    token = params[:token]
    cla = CliLoginAttempt.find_by(token:)
    if cla.nil?
      @error = 'Token not found'
    elsif cla.not_valid
      @error = 'Token expired'
    end
  end

  def do_confirm_auth
    token = params[:token]
    cla = CliLoginAttempt.find_by(token:, user: nil)
    if cla.nil?
      redirect_to :confirm_auth
    else
      @error = nil
      cla.user = current_user
      cla.save!
      redirect_to :confirm_auth
    end
  end

  private

  def set_sign_in_path
    store_location_for(:user, request.fullpath)
  end
end
