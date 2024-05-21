class Users::OmniauthCallbacksController < Devise::OmniauthCallbacksController
  # See https://github.com/omniauth/omniauth/wiki/FAQ#rails-session-is-clobbered-after-callback-on-developer-strategy
  skip_before_action :verify_authenticity_token, only: :github
  skip_before_action :authenticate_user!, only: :github

  def github
    # You need to implement the method below in your model (e.g. app/models/user.rb)
    @user = User.from_omniauth(request.env['omniauth.auth'])
    Rails.logger.debug { '----------------------------------------    ----------------' }
    Rails.logger.debug { "Omniauth auth is #{request.env['omniauth.auth'].inspect}" }
    Rails.logger.debug { "User is #{@user.inspect}" }

    if @user.persisted?
      sign_in_and_redirect @user, event: :authentication # this will throw if @user is not activated
      set_flash_message(:notice, :success, kind: 'Github') if is_navigational_format?
    else
      session['devise.github_data'] = request.env['omniauth.auth'].except(:extra) # Removing extra as it can overflow some session stores
      redirect_to new_user_registration_url
    end
  end

  def google_oauth2
    user = User.from_omniauth(auth)
    if user.present?
      sign_out_all_scopes
      flash[:success] = t 'devise.omniauth_callbacks.success', kind: 'Google'
      sign_in_and_redirect user, event: :authentication
    else
      flash[:alert] =
        t 'devise.omniauth_callbacks.failure', kind: 'Google', reason: "#{auth.info.email} is not authorized."
      redirect_to new_user_session_path
    end
  end

  def failure
    redirect_to root_path
  end

  private

  def auth
    @auth ||= request.env['omniauth.auth']
  end
end
