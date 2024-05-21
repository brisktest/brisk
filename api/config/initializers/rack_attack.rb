require 'ipaddr'

class Rack::Attack
  class Request < ::Rack::Request
    def remote_ip
      @remote_ip ||= begin
        env['action_dispatch.remote_ip'].to_s.strip
      rescue StandardError
        nil
      end
      # Rails.logger.debug "remote_ip is #{@remote_ip}"
      @remote_ip
    end

    def allowed_ip?
      allowed_ips = ['127.0.0.1', '::1']
      allowed_ips.include?(remote_ip)
    end

    def in_vpc_cidr?
      vpc_cidr = IPAddr.new(ENV['VPC_CIDR'] || '172.31.0.0/16')
      vpc_cidr.include?(IPAddr.new(remote_ip))
    rescue StandardError
      false
    end
  end

  safelist('allow from vpc_cidr') do |req|
    Rails.logger.debug "remote_ip is in range for #{req.remote_ip}" if req.in_vpc_cidr?
    req.in_vpc_cidr?
  end

  safelist('allow from localhost') do |req|
    req.allowed_ip?
  end

  blocklist('fail2ban') do |req|
    Rack::Attack::Fail2Ban.filter("fail2ban-#{req.remote_ip}", maxretry: 1, findtime: 1.day, bantime: 1.day) do
      CGI.unescape(req.query_string) =~ %r{/etc/passwd} ||
        req.path.include?('/etc/passwd') ||
        req.path.include?('wp-admin') ||
        req.path.include?('wp-login') ||
        req.path.include?('wp-content') ||
        req.path.include?('wp-config') ||
        req.path.include?('.env') ||
        /\S+\.php/.match?(req.path)
    end
  end

  blocklist('fail2ban-user-agent') do |req|
    Rack::Attack::Fail2Ban.filter("fail2ban-user-agent-#{req.remote_ip}", maxretry: 1, findtime: 1.day,
                                                                          bantime: 7.days) do
      req.user_agent && req.user_agent.include?('python-requests')
    end
  end

  throttle('limit logins per email', limit: 5, period: 20.seconds) do |req|
    if req.path == '/users/sign_in' && req.post? && ((req.params['user'].to_s.size > 0) and (req.params['user']['email'].to_s.size > 0))
      req.params['user']['email']
    end
  end

  throttle('limit signups', limit: 2, period: 1.minute) do |req|
    req.remote_ip if req.path == '/users' && req.post?
  end

  throttle('limit signups_hourly', limit: 10, period: 1.hour) do |req|
    req.remote_ip if req.path == '/users' && req.post?
  end

  # python-requests

  # Exponential backoff for all requests to "/" path
  #
  # Allows 240 requests/IP in ~8 minutes
  #        480 requests/IP in ~1 hour
  #        960 requests/IP in ~8 hours (~2,880 requests/day)
  (3..5).each do |level|
    throttle("req/ip/#{level}",
             limit: (30 * (2**level)),
             period: (0.9 * (8**level)).to_i.seconds) do |req|
      req.remote_ip unless req.path.starts_with?('/assets') || req.path.starts_with?('/health')
    end
  end
end

ActiveSupport::Notifications.subscribe(/rack_attack/) do |name, start, finish, request_id, payload|
  req = payload[:request]

  unless req.allowed_ip? || req.in_vpc_cidr?
    request_headers = { 'HTTP_X_FORWARDED_FOR' => req.env['HTTP_X_FORWARDED_FOR'] }

    Rails.logger.info "[Rack::Attack][Blocked] remote_ip: #{req.remote_ip}, path: #{req.path}"

    rack_attack_sample_rate = begin
      (ENV['RACK_ATTACK_SAMPLE_RATE'] || 0.01).to_f
    rescue StandardError
      0.01
    end
    if rand < rack_attack_sample_rate.to_f
      AdminMailer.rack_attack_notification(name, start, finish, request_id, req.remote_ip, req.path,
                                           request_headers).deliver_later
    else
      Rails.logger.debug "Not sending email about rack attack sample rate is #{rack_attack_sample_rate}"
    end
  end
end
