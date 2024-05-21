module Api
  include Rails.application.routes.url_helpers
  class UsersController < Api::GrufController
    bind ::Api::Users::Service
    def login
      return enum_for(:login) unless block_given?

      Rails.logger.debug("the nonce is #{request.message}")
      nonce = request.message.nonce
      token = UserService.login(nonce)
      url = Rails.application.routes.url_helpers.cli_auth_url(token)

      yield Api::LoginResponse.new(url:, description: 'description')
      start_time = Time.now
      while Time.now - start_time < 120.seconds
        Rails.logger.debug('waiting for 120 seconds')
        cla = CliLoginAttempt.find_by_token token
        if cla.user.nil?
          Rails.logger.debug("User hasn't logged in yet")
          sleep 1
        else
          yield Api::LoginResponse.new(url:, description: 'description', status: 'SUCCESS',
                                       credentials: { api_token: cla.user.credential.api_token, api_key: cla.user.credential.api_key })
          return
        end
      end
      Rails.logger.debug('120 seconds is up no login attempt found')
      yield Api::LoginResponse.new(url:, description: 'timeout', status: 'FAILURE')
    end
  end
end
