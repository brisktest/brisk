# frozen_string_literal: true

module Api
  class ProjectsController < Api::GrufController
    bind ::Api::Projects::Service
    include WorkersHelper
    include SupersHelper
    include ProjectsHelper
    include ApiHelper
    include JobrunsHelper
    include Rails.application.routes.url_helpers

    def init_project
      Rails.logger.debug "In init_project: request is #{request.inspect}"
      Rails.logger.info "In init_project: message is #{request.message}"
      fail!(:unauthenticated, :no_credentials_provided, 'Need to provide credentials') unless api_current_user
      mesg = request.message

      unless ::Project::FRAMEWORKS.include? mesg.framework
        fail!(:not_found, :framework_not_found,
              "#{mesg.framework} not recognised valid values are #{::Project::FRAMEWORKS.join(',')}")
      end
      plan = Plan.default_plan
      @org = if api_current_user.owned_orgs.any?
               api_current_user.owned_orgs.first
             else
               api_current_user.generate_default_org
             end
      # subscription = Subscription.new(plan: plan, user: api_current_user)
      # subscription.save!

      @project = @org.projects.build(framework: mesg.framework,
                                     org_id: api_current_user.owned_orgs.first.id,
                                     name: "default-project-#{SecureRandom.hex(4)}", worker_concurrency: 5, user: api_current_user)
      @project.save!
      @project.assign_default_image
      Api::InitProjectResp.new(project: to_project(@project))
      # rescue StandardError => e
      #   fail!(:internal, :could_not_create_project, e.message)
    end

    def get_project
      project = request.metadata[:project]

      Api::GetProjectResp.new(project: to_project(project))
      # rescue StandardError => e
      #   Rails.logger.error("exception is:#{e.inspect}")
      #   Rails.logger.error(e.backtrace)

      #   set_debug_info(e.message, e.backtrace)
      #   fail!(:not_found, :project_not_found, "Failed to find Project with token #{request.message.project_token}")
      #   set_debug_info(e.message, e.backtrace)
    end

    # For use after every run on the server
    # we log the worker run info
    # this prevents us from having situations where we miss logs
    # means we have logs faster and can spot gaps quicker
    # and workers can be freed as soon as they are finished
    def log_run
      Rails.logger.info "In log_run: request is #{request.inspect}"
      project = request.metadata[:project]

      wri = request.message.worker_run_info
      raise 'No command found in log_run' unless request.message.command

      no_test_files = request.message.command.noTestFiles
      Rails.logger.debug "In log_run no_test_files is #{no_test_files} for command #{request.message.command.inspect}"

      jobrun = Jobrun.find(wri.jobrun_id)
      Rails.logger.debug "In log_run: jobrun is #{jobrun.inspect}"
      raise 'wrong project' unless jobrun.project == project

      Rails.logger.debug "In log_run: message is #{request.message}"
      Rails.logger.debug "In log_run: worker_run_info is #{wri.inspect}"
      Rails.logger.debug "In log_run: worker_run_info is #{wri.to_json}"
      Rails.logger.debug "In log_run: this is one of #{jobrun.worker_run_infos.count}"
      Rails.logger.debug "In log_run: this is one of #{jobrun.supervisor.workers.count}"

      raise 'No jobrun found' unless jobrun

      worker = project.workers.find_by_id wri.worker_id
      raise 'No worker found' unless worker

      worker_run_info = worker.log_run(wri, jobrun.supervisor, no_test_files)
      unless worker_run_info.succeeded?
        Rails.logger.debug "In log_run: worker is #{worker.inspect} returning failure"
        jobrun.fail!
      end
      # what happens if this release comes in very late?
      # we can't assign it unless it's freed so it will still be locked
      # unless it comes in twice.
      # actually the unique constraint on worker run info should blow up and prevent doubles
      worker.free_from_super
      Rails.logger.debug { "In log_run: worker is #{worker.inspect} returning success" }
      if jobrun.worker_run_infos.count == jobrun.assigned_concurrency && jobrun.worker_run_infos.all?(&:succeeded?)
        Rails.logger.debug "In log_run: jobrun is #{jobrun.inspect} returning success"
        jobrun.finish!
      end
      Api::LogRunResp.new success: true
    end

    # we also need to ensure that everything gets freed
    def finish_run
      Rails.logger.info "In finish_run: request is #{request.inspect}"
      project = request.metadata[:project]

      unless request.message
        Rails.logger.error("request causing problems is #{request.inspect}")
        raise 'No FinishRunRequest found'
      end
      raise 'No project found' unless project
      raise 'No jobrun found' unless request.message.jobrun_id && request.message.jobrun_id > 0
      raise 'finish_run: No supervisor found' unless request.message.supervisor_uid.present?

      # worker_run_info = request.message.worker_run_info
      # Rails.logger.debug { "worker_run_info is #{worker_run_info.inspect}" }
      supervisor = project.supervisors.find_by_uid(request.message.supervisor_uid)

      Rails.logger.info("Finish run for project #{project.id} and jobrun #{request.message.jobrun_id}")
      if request.message.sync_failed_workers.size > 0
        Rails.logger.info("Failing workers are #{request.message.sync_failed_workers}")
        supervisor.workers.where(id: request.message.sync_failed_workers.map(&:id)).each do |worker|
          Rails.logger.info("Failing worker #{worker.id}")
          worker.de_register!
        end
      end

      jobrun = Jobrun.find(request.message.jobrun_id)
      if (request.message.final_worker_count || 0) < jobrun.assigned_concurrency
        Rails.logger.error("jobrun #{jobrun.id} has #{request.message.final_worker_count} workers but assigned_concurrency is #{jobrun.assigned_concurrency}")
        jobrun.update!(assigned_concurrency: request.message.final_worker_count)
        jobrun.add_note("assigned_concurrency was #{jobrun.assigned_concurrency} but only #{request.message.final_worker_count} workers were available after sync and rebuild check")
      end

      if jobrun.worker_run_infos.count != jobrun.assigned_concurrency
        Rails.logger.error("jobrun #{jobrun.id} has #{jobrun.worker_run_infos.count} worker_run_infos but assigned_concurrency is #{jobrun.assigned_concurrency}")
        jobrun.update!(state: 'failed', finished_at: Time.now, exit_code: 1,
                       error: 'all tests not passing', output: 'not all tests have finished')
      elsif jobrun.worker_run_infos.any? { |wri| !wri.succeeded? }
        jobrun.update!(state: 'failed', finished_at: Time.now, exit_code: request.message.exit_code,
                       error: 'not all workers finished with 0 exit code', output: '')
      else
        jobrun.update!(state: request.message.status, finished_at: Time.now, exit_code: request.message.exit_code,
                       error: request.message.error, output: request.message.output)
      end

      # supervisor.free_workers(worker_run_info, jobrun)
      unless supervisor.workers.empty?
        Rails.logger.error("supervisor #{supervisor.uid} still has workers #{supervisor.workers.map(&:id)}")
        supervisor.workers.each do |w|
          w.free_from_super
        end
      end
      supervisor.release

      Api::FinishRunResponse.new(state: jobrun.state)
      # rescue StandardError => e
      #   Rails.logger.error("metadata which caused the error was #{request.metadata.inspect}")
      #   Rails.logger.error("exception is:#{e.inspect}")
      #   Rails.logger.error(e.backtrace)
      #   # fail!(:internal, :error, "#{e.message} for jobrun #{request.message.jobrun_id}")
      #   raise e
    end

    # maybe need to include the supervisor here so we can release it later?
    # perhaps we can do all of this in the stage where we get the supervisor.....
    # get_super_for_project might assign a bunch of workers too.
    # get super for project, gets the relevant workers and returns em all

    # TODO: speed this up - takes nearly 800 ms , we can do better
    def get_workers_for_project
      project = request.metadata[:project]
      Rails.logger.info "get_workers_for_project: Get workers for project #{request.inspect}"

      if request.message.repo_info
        Rails.logger.info "get_workers_for_project: Repo info is #{request.message.repo_info.inspect}"
      end

      num_workers = request.message.num_workers
      worker_image = request.message.worker_image
      super_uid = request.message.supervisor_uid
      rebuild_hash = request.message.rebuild_hash
      log_uid = request.message.log_uid

      if log_uid.present?
        Rails.logger.info("Log uid is #{log_uid}")
      else
        # don't want blank strings
        log_uid = nil
      end

      if project.worker_concurrency < num_workers
        raise "Max concurrency allowed is #{project.allowed_concurrency} but you requested #{num_workers}"
      end

      raise 'get_workers_for_project: No project found' unless project
      raise "get_workers_for_project: No supervisor provided super_uid=#{super_uid}" unless super_uid.present?

      supervisor = project.supervisors.find_by_uid(super_uid)
      raise "get_workers_for_project: No supervisor found with uid #{super_uid}" unless supervisor

      # this is a little messy cause we do a check here but also in the
      # get workers for project
      if project.get_capacity > num_workers
        @jobrun = project.jobruns.create!(state: :starting, concurrency: num_workers, worker_image:,
                                          trace_key: request.metadata['trace-key'], api_version: request.metadata['brisk_api_version'], rebuild_hash:, supervisor_id: supervisor.id, log_uid:)
        if request.message.repo_info
          puts "Repo info is #{request.message.repo_info.inspect}"

          converted_hash = request.message.repo_info.to_h.map { |k, v| [k.to_s.underscore.to_sym, v] }.to_h
          puts "Converted hash is #{converted_hash}"
          repo_info = @jobrun.create_repo_info(commit_hash: converted_hash[:commit_hash],
                                               commit_author: converted_hash[:commit_author], commit_author_email: converted_hash[:commit_author_email], commit_message: converted_hash[:commit_message], branch: converted_hash[:branch])

          repo_info.save!
        end
      else
        Rails.logger.info 'This project has not got enough capacity'
        fail!(:resource_exhausted, :out_of_concurrency,
              "This project has not got enough capacity remaining. We are requiring #{num_workers}, however we only have #{project.get_capacity} remaining of #{project.plan.monthly_concurrency} with #{project.used_concurrency} units used ")
      end
      workers = []

      LockManager.redis_lock(@jobrun.worker_image) do
        # couple of ways to take care of contention
        # if we are struggling for workers - we can delay low priority workers
        # this will have the effect of making the high priority workers get more bites at the cherry than the low priority ones
        # without having to do any locking.
        # potentially can do the worker blocking in the super ?

        # Rails.logger.info "get_workers_for_project: Locking on #{worker_image}"
        start_get_workers = Time.now
        workers = ProjectService.get_workers_for_project(@jobrun)
        Rails.logger.info "get_workers_for_project: took #{Time.now - start_get_workers} to get workers for project"
      end
      project.balance_workers

      Rails.logger.debug { "get_workers_for_project: Jobrun after get_workers_for_project is #{@jobrun.inspect}" }
      Rails.logger.debug do
        "get_workers_for_project: Jobrun assigned concurrency is #{@jobrun.assigned_concurrency} after get_workers_for_project "
      end
      Rails.logger.debug do
        "get_workers_for_project: Jobrun assigned concurrency is #{@jobrun.reload.assigned_concurrency} after get_workers_for_project and reload "
      end
      Rails.logger.debug do
        "get_workers_for_project: Jobrun after get_workers_for_project and reload is #{@jobrun.reload.inspect}"
      end

      Rails.logger.debug { "get_workers_for_project: We are returning #{workers.inspect} from get_workers_for_project" }
      Rails.logger.debug { "get_workers_for_project: The workers in order are :- #{workers.map(&:id)}" }
      Rails.logger.debug do
        "get_workers_for_project: Workers without build commands run are :- #{workers.reject(&:build_commands_run_at).map(&:id)}"
      end

      # project.reload.workers.where(id: workers.map(&:id)).update_all(jobrun_id: @jobrun.id)

      if @jobrun.assigned_concurrency < @jobrun.concurrency
        Rails.logger.info("get_workers_for_project: We don't have as many workers as requested for this jobrun")
        # Rails.logger.info("We don't have as many workers as requested for this jobrun (#{@jobrun.id}) #{@jobrun.workers.count} < #{@jobrun.concurrency}  - details are : #{@jobrun.debug_worker_info}")
      end

      if @jobrun.reload.not_enough_workers?
        Rails.logger.info("get_workers_for_project: We don't have enough workers for this jobrun (#{@jobrun.id}) #{@jobrun.assigned_concurrency} < 90% of  #{@jobrun.concurrency}")
        @jobrun.unfulfill!
        Rails.logger.debug("get_workers_for_project: We are cleaning up unfulfilled jobrun #{@jobrun.id} but we shouldn't have to do it here")
        @jobrun.cleanup_unfulfilled

        fail!(:resource_exhausted, :temp_out_of_workers,
              "We don't have enough workers for this run #{@jobrun.assigned_concurrency} < 90% of  #{@jobrun.concurrency}")
      else
        workers.each do |w|
          w.update!(rebuild_hash:)
        end
      end

      Api::GetWorkersResp.new(jobrun_id: @jobrun.id, workers: workers.map do |w|
                                                                to_worker(w)
                                                              end, jobrun_link: get_jobrun_link(@jobrun))
    rescue ProjectService::OutOfWorkersError => e
      @jobrun.unfulfill! if @jobrun && @jobrun.starting?
      fail!(:resource_exhausted, :out_of_workers, e.message)
    end

    def clear_workers_for_project
      Rails.logger.info "In clear_workers_for_project: request is #{request.inspect}"
      project = request.metadata[:project]

      raise 'No project' if project.nil?

      if request.message.supervisor_uid.present?
        Rails.logger.info "SupervisorID passed to clear_workers_for_project - clearing workers for supervisor #{request.message.supervisor_uid}"
        supervisor = project.supervisors.find_by_uid(request.message.supervisor_uid)

        if supervisor.nil?
          raise "No supervisor found for project #{project.id} with uid #{request.message.supervisor_uid}"
        end

        supervisor.workers.each(&:de_register!)
      else
        Rails.logger.info "No supervisorID passed to clear_workers_for_project - clearing workers for project #{project.id}"
        project.workers.assigned.map(&:de_register!)
      end

      Api::ClearWorkersResp.new(status: :ok)
    end

    # When we get a supervisor for a project.
    # We first check if there is a supervisor assigned for the specfic affinity - if there is we return it.
    # Supervisors are shared among projects with the same afinity.
    # If there is no affinity then we return the first free supervisor that was already assigned to the project.
    # If there is no free supervisor then we create a new one and assign it to the project.
    # if we can't create a supervisor then we return a previously assigned supervisor (and let them queue behind it)

    def get_super_for_project
      Rails.logger.info "In get_super_for_project: request is #{request.inspect}"
      project = request.metadata[:project]
      # unique_instance_id = request.message.unique_instance_id
      affinity = request.message.affinity
      Rails.logger.debug { "Affinity is #{affinity}" }

      raise 'No project' if project.nil?

      # raise 'No unique_instance_id' if unique_instance_id.nil?

      Rails.logger.debug do
        "relevant state is affinity_present? #{affinity.present?} and project.assigned_supervisor(affinity) #{project.assigned_supervisor(affinity)} \n
      project.free_assigned_supervisor #{project.free_assigned_supervisor} \n
      project.can_assign_supervisor? #{project.can_assign_supervisor?} \n
      Supervisor.ready.size > 0 #{Supervisor.ready.size > 0} \n
      project.assigned_supervisors.count  #{project.assigned_supervisors.count} \n
      "
      end

      if project.free_assigned_supervisor
        Rails.logger.debug 'We have a free supervisor assigned to this project'
        sup = project.free_assigned_supervisor
        # elsif affinity.present? && project.assigned_supervisor(affinity)
        #   Rails.logger.debug "We have an affinity ='#{affinity}' and a supervisor assigned to it"
        #   # if we have multiple people using the same affinity then they may queue up here but that is ok.
        #   sup = project.assigned_supervisor(affinity)
      elsif project.can_assign_supervisor? && Supervisor.ready.size > 0
        Rails.logger.debug 'Our project can assign a supervisor and we have a free supervisor and can assign it to this project'
        Supervisor.transaction do
          s = Supervisor.ready.first.lock!

          Rails.logger.debug "Assigning supervisor the supervisor is #{s.inspect}"

          s.assign_to_project(project)

          s.save!
          sup = s
        end

        Rails.logger.info "This project #{project.id} has no free supervisors, assigning a new one"
      elsif project.assigned_supervisors.count > 0
        Rails.logger.info "This project #{project.id} has no free supervisors, assigning one of the existing ones - #{project.max_supervisors} supervisors assigned."

        # we grab the one that was updated last - should be freest
        sup = project.assigned_supervisors.order('updated_at asc').first
      else
        Rails.logger.error 'No supervisors available '
        fail!(:resource_exhausted, :out_of_supervisors, 'No supervisors available - contact support')
      end
      # maybe make this a little more robust - but the worst that happens is we reuse a super...
      sup.in_use = Time.now

      sup.affinity = affinity if affinity.present?
      sup.save!
      Api::GetSuperResp.new(super: to_super(sup.reload))
    end

    def get_all_projects
      Rails.logger.info "In get_all_projects: request is #{request.inspect}"
      fail!(:unauthenticated, :no_credentials_provided, 'Need to provide credentials') unless api_current_user

      Api::GetAllProjectsResp.new(projects: api_current_user.projects.map { |p| to_project(p) })
    end

    def de_register_workers
      Rails.logger.info "In rebuild_workers: request is #{request.message}"
      project = request.metadata[:project]
      Rails.logger.info "Rebuilding workers #{request.message.workers} for project #{project.id}"

      raise 'No project' if project.nil?
      raise 'No workers' if request.message.workers.empty?

      project.workers.where(id: request.message.workers.map(&:id)).each do |w|
        Rails.logger.debug("de registering worker #{w.id}")
        w.de_register!
      end
      Api::DeRegisterWorkersResp.new(status: :ok)
    end
  end
end
