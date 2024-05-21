class AdminMailer < ApplicationMailer
  default from: 'rackattack@brisktest.com', to: 'sean@brisktest.com'

  def rack_attack_notification(name, start, finish, request_id, remote_ip, path, request_headers)
    @name = name
    @start = start
    @finish = finish
    @request_id = request_id
    @remote_ip = remote_ip
    @path = path
    @request_headers = request_headers.inspect

    mail(subject: "[Rack::Attack][Blocked] remote_ip: #{remote_ip}")
  end

  def new_user(user_id)
    @user = User.find(user_id)
    mail(subject: "[New User] #{@user.email}", from: 'admin_mailer@brisktest.com')
  end

  def job_run_cleaned_up(job_run_id)
    @job_run = Jobrun.find(job_run_id)
    mail(subject: "[Job Run Cleaned Up] #{@job_run.id}", from: 'admin_mailer@brisktest.com', to: 'sean@brisktest.com')
  end
end
