# frozen_string_literal: true

# Be sure to restart your server when you modify this file.

# Define an application-wide content security policy
# For further information see the following documentation
# https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy

# Rails.application.config.content_security_policy do |policy|
#   policy.default_src :self, :https
#   policy.font_src    :self, :https, :data
#   policy.img_src     :self, :https, :data
#   policy.object_src  :none
#   policy.script_src  :self, :https
#   policy.style_src   :self, :https
#   # If you are using webpack-dev-server then specify webpack-dev-server host
#   policy.connect_src :self, :https, "http://localhost:3035", "ws://localhost:3035" if Rails.env.development?

#   # Specify URI for violation reports
#   # policy.report_uri "/csp-violation-report-endpoint"
# end

# If you are using UJS then enable automatic nonce generation
# Rails.application.config.content_security_policy_nonce_generator = -> request { SecureRandom.base64(16) }

# Set the nonce only to specific directives
# Rails.application.config.content_security_policy_nonce_directives = %w(script-src)

# Report CSP violations to a specified URI
# For further information see the following documentation:
# https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy-Report-Only
# Rails.application.config.content_security_policy_report_only = true
# Rails.application.config.content_security_policy_nonce_generator = ->(_request) { SecureRandom.base64(16) }
# Rails.application.config.content_security_policy_nonce_generator = ->(request) do
#   # use the same csp nonce for turbo requests
#   if request.env["HTTP_TURBO_REFERRER"].present?
#     request.env["HTTP_X_TURBO_NONCE"]
#   else
#     SecureRandom.base64(16)
#   end
# end
# Rails.application.config.content_security_policy_nonce_directives = %w[script-src style-src]
# if Rails.env.development?
#   Rails.application.config.content_security_policy_report_only = true
# else
Rails.application.config.content_security_policy do |policy|
  # shit doesn't work with turbo
  policy.default_src :self, :https, :unsafe_inline
  policy.script_src :self, :unsafe_inline, 'https://www.googletagmanager.com', 'https://www.google-analytics.com',
                    'https://asciinema.org/', 'https://player.vimeo.com/', 'https://bam.nr-data.net', 'https://js-agent.newrelic.com'

  policy.connect_src :self, 'https://bam.nr-data.net', 'https://www.googletagmanager.com',
                     'https://www.google-analytics.com', 'https://brisk-output-logs.s3.amazonaws.com/'
  policy.style_src :self, :https, :unsafe_inline
  policy.img_src :self, :https, :data
  policy.font_src :self, :https

  # end
end
