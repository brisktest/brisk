class OrgsController < ApplicationController
  before_action :set_org
  before_action :authenticate_user!
  before_action :ensure_user_can_access, except: %i[index]
  before_action :ensure_user_can_manage, only: %i[edit update destroy]

  def show
    @projects = @org.projects
  end

  def index
    @orgs = current_user.orgs.uniq
  end

  def edit; end

  def update
    # can update name (by id)
    @org.update!(org_params)
    redirect_to org_path(@org)
  end

  private

  def org_params
    params.require(:org).permit(:name)
  end

  def set_org
    Rails.logger.debug(" params[:org] #{params[:org]}")
    if params[:name].present?
      @org = Org.find_by_name!(params[:name])
    elsif params[:org].present? && params[:org][:name].present?
      @org = Org.find_by_name!(params[:org][:name])
    end
  end

  def ensure_user_can_access
    return true if current_user.admin?
    raise 'not allowed' unless @org.is_member?(current_user)
  end

  def ensure_user_can_manage
    return true if current_user.admin?
    raise 'not allowed' unless @org.is_manager? current_user
  end
end
