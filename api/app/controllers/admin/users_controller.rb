class Admin::UsersController < Admin::AdminController
  def index
    @users = User.all.page(params[:page]).order('created_at desc')
  end

  def daily_new_users
    @users = User.where('created_at > ?', 1.day.ago).order('created_at desc').page(params[:page])
  end

  def weekly_new_users
    @users = User.where('created_at > ?', 1.week.ago).order('created_at desc').page(params[:page])
  end

  def daily_active_users
    @users = User.joins(:jobruns).where('jobruns.created_at > ?', 1.day.ago).distinct.page(params[:page])
  end

  def weekly_active_users
    @users = User.joins(:jobruns).where('jobruns.created_at > ?', 1.week.ago).distinct.page(params[:page])
  end
end
