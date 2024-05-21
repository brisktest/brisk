class Admin::WorkersController < Admin::AdminController
  def workers_by_project
    @project_ids = Jobrun.order('created_at desc').distinct(:project_id)
    @projects = Project.where(id: @project_ids).include
  end

  def worker_details
    @image = Image.find(params[:image_id])
    @workers = @image.workers.not_stale
    @busy_workers = @image.workers.not_stale.busy
    @free_workers = @image.workers.not_stale.free
    @assigned_workers = @image.workers.assigned.not_stale.group_by(&:project)
  end
end
