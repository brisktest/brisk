class SubscriptionsController < ApplicationController
  before_action :signup_user, only: :join
  before_action :authenticate_user!

  def join
    params.require(:plan_id)
    plan_id = params[:plan_id]
    plan = Plan.find plan_id

    Rails.logger.debug "switching user #{current_user.id}to plan #{plan.id}"
    @sub = current_user.account.subscription
    @sub.update(plan:)
    session[:new_sub] = @sub.id
    @sub.save!
    redirect_to subscriptions_welcome_path and return
  end

  def welcome; end
end
