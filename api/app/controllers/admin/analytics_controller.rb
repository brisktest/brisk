class Admin::AnalyticsController < Admin::AdminController
  def index
    @projects = Project.all
    @users = User.all
    @new_users = User.where('created_at > ?', 1.week.ago).load_async
    @daily_new_users = User.where('created_at > ?', 1.day.ago).load_async
    @weekly_active_users = Jobrun.where('jobruns.created_at > ?', 1.week.ago).joins(:user).distinct.load_async
    @daily_active_users = Jobrun.where('jobruns.created_at > ?', 1.day.ago).joins(:user).distinct.load_async
    @new_projects = Project.where('created_at > ?', 1.week.ago).load_async
    @daily_new_projects = Project.where('created_at > ?', 1.day.ago).load_async
    @jobruns = Jobrun.where('created_at > ?', 1.week.ago).order(created_at: :asc).load_async
    @daily_jobruns = Jobrun.where('created_at > ?', 1.day.ago).order(created_at: :asc).load_async
    @weekly_jobruns = Jobrun.where('created_at > ?', 1.week.ago).order(created_at: :asc).load_async
    @workers = Worker.active.not_stale.load_async
    rspec_image = Image.where name: 'rails'
    jest_image = Image.where name: 'node-lts'
    python_image = Image.where name: 'python'

    @rspec_workers = rspec_image.first.workers.not_stale.active.load_async
    @jest_workers = jest_image.first.workers.active.not_stale.load_async
    @python_workers = python_image&.first&.workers&.active&.not_stale&.load_async

    @assigned_workers = Worker.assigned.load_async
    @assigned_rspec_workers = rspec_image.first.workers.assigned.load_async
    @assigned_jest_workers = jest_image.first.workers.assigned.load_async
    @assigned_python_workers = python_image&.first&.workers&.assigned&.load_async

    @stale_workers = Worker.stale.in_use.load_async
    @stale_rspec_workers = rspec_image.first.workers.stale.in_use.load_async
    @stale_jest_workers = jest_image.first.workers.stale.in_use.load_async
    @stale_python_workers = python_image&.first&.workers&.stale&.in_use&.load_async

    @free_supers = Supervisor.ready.load_async
    @assigned_supers = Supervisor.assigned.load_async
    @all_supers = Supervisor.ready_and_assigned.load_async
  end
end
