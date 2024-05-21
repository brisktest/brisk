# frozen_string_literal: true

class ApplicationController < ActionController::Base
  before_action :store_user_location!, if: :storable_location?

  before_action :set_csrf_cookie
  protect_from_forgery prepend: true, with: :exception
  before_action :authenticate_user!
  around_action :log_everything, only: :index, if: :log_requests?
  around_action :log_everything, if: -> { Rails.env.development? || log_requests? }
  before_action :create_breadcrumbs

  around_action :set_time_zone

  unless Rails.env.development?
    rescue_from Exception do |exception|
      Rails.logger.error "Exception caught and released: #{exception.message} " + exception.backtrace.join("\n")
      Rails.logger.error
      raise
    end
  end

  def storable_location?
    request.get? && is_navigational_format? && !devise_controller? && !request.xhr?
  end

  def store_user_location!
    # :user is the scope we are authenticating
    Rails.logger.info { "Storing user location for :user  to #{request.fullpath}" }
    store_location_for(:user, request.fullpath)
  end

  # before_action :render_flash
  # def render_flash
  #   render turbo_stream: turbo_stream.update("flash", partial: "layouts/shared/flash")
  # end

  def render_turbo_flash
    render(turbo_stream: turbo_stream.update('flash', partial: 'layouts/shared/flash'))
  end

  def set_time_zone(&block)
    if current_user && current_user.time_zone
      Time.use_zone(current_user.time_zone, &block)
    else
      yield
    end
  end

  unless Rails.env.development?
    rescue_from StandardError do |e|
      logger.error e.inspect
      logger.error e.backtrace.join("\n") unless e.is_a? ActionController::RoutingError
      raise e unless current_user && current_user.is_admin?

      render template: 'errors/admin_error', locals: { exception: e }
    end
  end

  def create_breadcrumbs
    @breadcrumb_paths = []
  end

  def add_to_breadcrumbs(name, path)
    @breadcrumb_paths << OpenStruct.new(name:, path:)
  end

  # before_action :debug_xforwarded_for
  # def debug_xforwarded_for
  #   Rails.logger.info "X-Forwarded-For is #{request.headers['X-Forwarded-For']}"
  #   request_headers = { 'HTTP_X_FORWARDED_FOR' => request.env['HTTP_X_FORWARDED_FOR'] }

  #   Rails.logger.info "Request headers are #{request_headers.inspect}"
  #   Rails.logger.info "Request headers are #{request.remote_ip}"
  #   Rails.logger.info "The HTTP_X_FORWARDED_HOST is #{request.env['HTTP_X_FORWARDED_HOST']}"
  # end

  # for some reason I can't get the headers x-forwarded-proto to go from AWS through traefik so turning
  # off for now
  # skip_forgery_protection
  def append_info_to_payload(payload)
    super
    payload[:environment] = Rails.env
    payload[:host] = request.host
    payload[:remote_ip] = request.remote_ip
    payload[:ip] = request.ip
  end

  private

  def log_requests?
    Rails.logger.warn "params are #{params.inspect}}"
    ENV['LOG_REQUESTS'] || Rails.env.development?
  end

  def log_everything
    log_headers
    yield
  ensure
    log_response
  end

  def log_headers
    http_envs = {}.tap do |envs|
      request.headers.each do |key, value|
        envs[key] = value if key.downcase.starts_with?('http')
      end
    end

    logger.info "Received #{request.method.inspect} to #{request.url.inspect} from #{request.remote_ip.inspect}. Processing with headers #{http_envs.inspect} and params #{params.inspect}"
  end

  def log_response
    logger.info "Responding with #{response.status.inspect} => #{response.body.inspect}"
  end

  def signup_user
    if current_user
      true
    else
      redirect_to new_user_registration_path and return
    end
  end

  def after_sign_in_path_for(resource_or_scope)
    res = stored_location_for(resource_or_scope) || super
    Rails.logger.info { "After sign in path is #{res}" }
    res
  end

  def set_csrf_cookie
    fat = form_authenticity_token
    Rails.logger.debug { "Setting CSRF cookie to #{fat}" }
    cookies['csrftoken'] = {
      value: fat,
      expires: 1.day.from_now,
      secure: false,
      httponly: false
    }
  end
end
