# frozen_string_literal: true

module Interceptors
  class TokenAuth < Gruf::Interceptors::ServerInterceptor
    include ApiHelper

    def insecure_mode?
      ENV['BRISK_INSECURE'] == 'true'
    end

    def call
      if insecure_mode? && !provided_credentials?
        Rails.logger.debug('Insecure mode enabled and no credentials provided')
        request.active_call.metadata[:current_user_id] = 1
        api_action = ApiAction.find_by_grpc_method_name request.method_name
        request.active_call.metadata[:authenticated_action] = api_action.grpc_method_name if api_action

      else
        unless bypass?
          get_user_from_creds

          fail!(:unauthenticated, :unauthenticated, 'Invalid credentials') unless valid?
        end
      end
      yield
    end

    private

    def provided_credentials?
      request_credentials && (request_credentials[:api_token].present? || request_credentials[:api_key].present?)
    end

    # get the user from the creds and add it to the metadata
    def get_user_from_creds
      user_creds = Credential.where(
        api_token: request_credentials[:api_token]
      ).first

      if user_creds && user_creds.credentialable_type == 'User' && user_creds.user && user_creds.api_key == request_credentials[:api_key]
        request.active_call.metadata[:current_user_id] = user_creds.user.id
      end
    end

    def validate_by_project(creds)
      p = Project.find_by_project_token(creds[:project_token])

      if p && p.users.any? { |u| u.credential.verify_api_key_and_token(creds[:api_key], creds[:api_token]) }
        Rails.logger.info("AUTHORIZED project_token #{creds[:project_token]} for #{creds[:api_token]}")
        request.active_call.metadata[:project] = p

        p
      else
        Rails.logger.debug("AUTH DENIED FOR #{creds}")
        fail!(:unauthenticated, :unauthorized, 'Not authorized for project')
        false
      end
    end

    def validate_by_api_action
      Rails.logger.debug "Validating by api action : #{request.method_name}"
      api_action = ApiAction.find_by_grpc_method_name request.method_name
      if api_action
        request.active_call.metadata[:authenticated_action] = api_action.grpc_method_name
        return api_action
      end

      if api_action.nil?
        fail!(:unauthenticated, :unauthenticated,
              "#{request.method_name} is not a valid API action. hint: have you logged in and/or are you using the correct credentials file?")
        return false
      end

      Rails.logger.debug("API action is #{api_action.grpc_method_name}")
      cred = api_action.credentials.not_expired.select do |cred|
        cred.verify_api_key_and_token(request_credentials[:api_key], request_credentials[:api_token])
      end

      if cred.blank?
        Rails.logger.info "AUTH FAILED FOR #{request_credentials} and #{api_action.inspect}"
        fail!(:unauthenticated, :unauthenticated, 'Bad credentials provided for route')
        false
      else
        request.active_call.metadata[:authenticated_action] = api_action.grpc_method_name
        api_action
      end
    end

    def valid?
      Rails.logger.debug('in valid?')
      creds = request_credentials
      Rails.logger.debug("the credentials are #{creds}")
      return false if creds.blank? && !insecure_mode?

      if creds.present? && creds[:project_token].present?
        validate_by_project(creds)
      else
        api_current_user || validate_by_api_action
      end
    end

    def request_credentials
      Rails.logger.debug('In request credentials')
      md = request.active_call&.metadata
      Rails.logger.debug("metadata is #{md}")

      creds = md['authorization']
      Rails.logger.debug("creds are #{creds}")
      if creds.blank?
        Rails.logger.debug('No credentials passed to call ')
        fail!(:unauthenticated, :unauthenticated, 'no credentials provided to api call') unless insecure_mode?
        return nil
      end

      # go sends these in in camel case
      creds = JSON.parse(creds)
      creds.transform_keys { |key| key.to_s.underscore }.with_indifferent_access
      # JSON.parse(Base64.decode64(creds.to_s))
    end

    def bypass?
      Rails.logger.debug("in bypass? method name is #{request.method_name} should #{if options.fetch(:excluded_methods,
                                                                                                     []).include?(request.method_name)
                                                                                      'bypass'
                                                                                    else
                                                                                      'not bypass'
                                                                                    end}")

      options.fetch(:excluded_methods, []).include?(request.method_name)
    end
  end
end
