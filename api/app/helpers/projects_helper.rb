# frozen_string_literal: true

module ProjectsHelper
  def to_project(project)
    { id: project.id, name: project.name, project_token: project.project_token, username: project.username,
      org_id: project.org.id, user_id: project.user_id, framework: project.framework, image: project.image.name,
      concurrency: project.worker_concurrency }
  end
end
