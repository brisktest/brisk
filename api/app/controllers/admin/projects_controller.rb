class Admin::ProjectsController < Admin::AdminController
  def index
    @projects = Project.all.page(params[:page]).order('created_at desc')
  end

  def daily_active_projects
    @projects = Project.joins(:jobruns).where('jobruns.created_at > ?', 1.day.ago).distinct.page(params[:page])
  end

  def weekly_active_projects
    @projects = Project.joins(:jobruns).where('jobruns.created_at > ?', 1.week.ago).distinct.page(params[:page])
  end

  def new_daily_projects
    @projects = Project.where('created_at > ?', 1.day.ago).order('created_at desc').page(params[:page])
  end

  def new_weekly_projects
    @projects = Project.where('created_at > ?', 1.week.ago).order('created_at desc').page(params[:page])
  end
end
