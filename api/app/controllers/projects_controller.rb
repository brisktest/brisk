# frozen_string_literal: true

class ProjectsController < ApplicationController
  include JobrunsHelper
  # before_action :set_project, only: %i[show edit update destroy show_files]
  before_action :authenticate_user!
  before_action :set_user

  before_action :ensure_user_can_access, except: %i[new create index]
  before_action :set_breadcrumb, except: %i[new create index]

  before_action :project_nav_bar

  def index
    raise ActionController::RoutingError, 'Not Found'
    @projects = current_user.projects
  end

  def project_nav_bar
    @no_nav_bar = true
  end

  def ensure_user_can_access
    @project = Project.find_by_project_token params[:project_token]
    if @project.nil?
      flash[:error] = 'Project not found'
      redirect_to :root and return
    end
    if current_user.is_admin?
      Rails.logger.info("User is admin, allowing access for project token #{@project.project_token}}")
      return true
    end

    if @project.org.is_member?(current_user)
      Rails.logger.debug("User is member of org, allowing access for project token #{@project.project_token}")
      return true
    end

    # probably have to remove this check
    unless current_user.authorized_projects.include? @project
      Rails.logger.warn("User #{current_user.id} #{current_user.email} is not authorized to access project token #{@project.project_token}}")
      redirect_to org_request_membership_path(@project.org) and return
    end
    false
  end

  def set_breadcrumb
    add_to_breadcrumbs 'Project', project_path(@project) if @project
  end

  def project_dashboard
    add_to_breadcrumbs 'Dashboard', project_dashboard_path(@project.project_token) if @project

    @jobruns = @project.jobruns.order('created_at desc').page(params[:page])
  end

  def dashboard_jobruns
    add_to_breadcrumbs 'Dashboard', project_dashboard_path(@project.project_token) if @project

    if @project && params[:jobrun_uid]
      add_to_breadcrumbs 'Run',
                         project_jobrun_path(@project.project_token,
                                             params[:jobrun_uid])
    end
    @jobrun = @project.jobruns.includes(worker_run_infos: [:test_files]).find_by_uid(params[:jobrun_uid])

    return unless @jobrun.log_uid

    @run_log_url = S3Service.new.get_presigned_url(ENV['S3_LOG_BUCKET'] ||= 'brisk-output-logs',
                                                   @jobrun.log_uid).first
  end

  def dashboard_jobruns_with_logs
    @worker_run_info = WorkerRunInfo.find_by_uid! params[:worker_run_info_uid]
    add_to_breadcrumbs 'Dashboard', project_dashboard_path(@project.project_token) if @project

    if @project && @worker_run_info
      add_to_breadcrumbs 'Run',
                         project_jobrun_path(@project.project_token,
                                             @worker_run_info.jobrun.uid)
    end
    if @project && params[:worker_run_info_uid]
      add_to_breadcrumbs 'Workers',
                         project_jobrun_with_logs_path(@project.project_token,
                                                       params[:worker_run_info_uid])
    end

    # this is where we check that the user has access to the project
    @jobrun = @worker_run_info.jobrun
    # second is headers if we need them
    @log_url = S3Service.new.get_presigned_url(ENV['S3_LOG_BUCKET'] ||= 'brisk-output-logs',
                                               @worker_run_info.uid).first
  end

  # GET /projects/1 or /projects/1.json
  def show; end

  def set_user
    @user = current_user
  end

  # GET /projects/new
  def new
    @project = Project.new
    @user = current_user
  end

  # GET /projects/1/edit
  def edit; end

  # POST /projects or /projects.json
  def create
    @project = Project.new(project_params.merge(user_id: current_user.id))
    @project.assign_default_image

    @user = current_user
    respond_to do |format|
      if @project.save
        format.html { redirect_to @project, notice: 'Project was successfully created.' }
        format.json { render :show, status: :created, location: @project }
        if current_user.projects.size == 1
          redirect_to getting_started_path and return
        else
          redirect_to dashboard_path and return
        end
      else
        format.html { render :new, status: :unprocessable_entity }
        format.json { render json: @project.errors, status: :unprocessable_entity }
      end
    end
  end

  # PATCH/PUT /projects/1 or /projects/1.json
  def update
    respond_to do |format|
      if @project.update(project_params)
        format.html { redirect_to @project, notice: 'Project was successfully updated.' }
        format.json { render :show, status: :ok, location: @project }
      else
        format.html { render :edit, status: :unprocessable_entity }
        format.json { render json: @project.errors, status: :unprocessable_entity }
      end
    end
  end

  # DELETE /projects/1 or /projects/1.json
  def destroy
    @project.destroy
    respond_to do |format|
      format.html { redirect_to projects_url, notice: 'Project was successfully destroyed.' }
      format.json { head :no_content }
    end
  end

  def show_files
    @files = @project.test_files.order('runtime DESC')
  end

  def show_runs
    @jobruns = @project.jobruns.order('created_at desc').page(params[:page])
  end

  def show_worker_runs
    jobrun_id = params[:jobrun_id]
    @jobrun = Jobrun.find(jobrun_id)
    @worker_runs = @jobrun.worker_run_infos
  end

  private

  # Use callbacks to share common setup or constraints between actions.

  # Only allow a list of trusted parameters through.
  def project_params
    params.require(:project).permit(:name, :user_id, :framework, :git_hosting_provider, :git_org, :git_repo_name)
  end
end
