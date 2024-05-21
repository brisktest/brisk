class CredentialsController < ApplicationController
  before_action :authenticate_user!

  # delete a credential from the user's account by JS
  def destroy
    @credential = current_user.credential
    redirect_to account_settings_path, notice: 'Credential not found' and return if @credential.nil?

    @credential.destroy

    current_user.generate_credentials

    redirect_to account_settings_path, notice: 'Credentials Recreated' and return
  end

  # show the credential if it hasn't been shown before
  def one_time_show
    @credential = current_user.credential
    if @credential.nil?
      redirect_to account_settings_path, notice: 'Credential not found' and return
    elsif @credential.shown_to_user_at?
      redirect_to account_settings_path, notice: 'Credential already shown' and return
    else
      @credential.update(shown_to_user_at: Time.now)
      @credential_show = true
      render 'account_settings/settings' and return
    end
  end
end
