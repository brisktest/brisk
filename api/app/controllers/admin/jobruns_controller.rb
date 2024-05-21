class Admin::JobrunsController < Admin::AdminController
  def for_project
    @project = Project.find(params[:id])
    @jobruns = @project.jobruns.order('created_at desc').page(params[:page])
  end

  def for_day
    @jobruns = Jobrun.where('created_at > ?', Time.now - 1.day).order('created_at desc').page(params[:page])
  end

  def for_week
    @jobruns = Jobrun.where('created_at > ?', Time.now - 1.week).order('created_at desc').page(params[:page])
  end

  def for_running
    @jobruns = Jobrun.started.order('created_at desc').page(params[:page])
  end

  def for_failed
    @jobruns = Jobrun.failed.order('created_at desc').page(params[:page])
  end

  def for_completed
    @jobruns = Jobrun.completed.order('created_at desc').page(params[:page])
  end

  def show
    @jobrun = Jobrun.find(params[:id])
  end
end
