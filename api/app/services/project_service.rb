# frozen_string_literal: true

class ProjectService
  # so we want to grab all of the workers for a project
  # we first take any that are assigned to the project with the correct worker image and the correct rebuild hash
  # if we don't have enough we take any that are assigned to the project with the correct worker image and no rebuild hash

  # if we need to assign workers we do that afterwards
  # then we try and assign the workers to the project - using a lock
  # if any additional workers are required we add them later

  # we also would like to have workers that are not on the same machine as each other
  # also we want to make sure we aren't going over the running memory allocation of the machine.

  # would like to make sure they have build commands run
  # but also would like to make sure they are not on a machine that is already running a test
  # so

  # sort them so we have a build commands run at as the first worker

  class OutOfWorkersError < StandardError; end

  def self.get_workers_for_project(jobrun)
    start_time = Time.now
    num_workers = jobrun.concurrency
    project = jobrun.project
    Rails.logger.info "get_workers_for_project: Reserving workers for project #{project.id} with #{num_workers} workers"
    Rails.logger.debug "get_workers_for_project: Worker image is #{jobrun.worker_image}"
    Rails.logger.debug "get_workers_for_project: Rebuild hash is #{jobrun.rebuild_hash}"

    # we oversubscribe workers and then balance at the end
    if project.workers.not_stale.in_use.busy.count + num_workers > project.max_workers
      Rails.logger.info "get_workers_for_project: Project #{project.id} has #{project.workers.not_stale.in_use.busy.count}
      workers and #{num_workers} requested. Max workers is #{project.max_workers} - jobrun info is
      #{project.workers.not_stale.in_use.busy.map(&:jobrun_id).group_by do |w|
          w
        end}"
      raise OutOfWorkersError,
            "Project #{project.project_token} has #{project.workers.not_stale.in_use.busy.count} workers currently in use and #{num_workers} requested. Max workers is #{project.max_workers} - (max supervisors is #{project.max_supervisors}) please try again later when your workers are free."
    end

    new_workers = []

    workers_with_builds = project.workers.not_stale.in_use.machine_not_draining.free_workers
                                 .hash_or_empty(jobrun.rebuild_hash)
                                 .with_worker_image(jobrun.worker_image).with_build_commands.includes({ machine: :workers }).uniq

    # free_cpu is a very rough estimate of how much cpu is free on the machine
    # we subtract the number of workers on the machine from the number of cores
    Rails.logger.debug("get_workers_for_project: Before we check memory space we have found #{workers_with_builds.to_a.size} workers")

    workers_with_builds_and_mem = workers_with_builds.reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end

    workers_with_builds_and_mem = workers_with_builds_and_mem.reject do |w|
                                    w.machine.free_cpu <= 0
                                  end.sort_by { |w| w.machine.free_cpu }.reverse

    Rails.logger.info "get_workers_for_project: timing Step 1 took #{Time.now - start_time} seconds"
    start_time = Time.now
    if workers_with_builds_and_mem.size > 0
      Rails.logger.info "get_workers_for_project: Found #{workers_with_builds_and_mem.size} free workers assigned to project #{project.id} with build commands"
      new_workers = Worker.safe_assign workers_with_builds_and_mem, project, num_workers, jobrun.supervisor_id,
                                       jobrun.id
      if new_workers.size >= num_workers
        Rails.logger.info "get_workers_for_project: Found #{new_workers.size} workers with memory space and build commands run and assigned them to the project"
        jobrun.assigned_concurrency = new_workers.size
        jobrun.save!
        return new_workers
      end
    end

    Rails.logger.info "get_workers_for_project: Not enough workers with memory space and build commands run - we have #{new_workers.size} workers and need #{num_workers} workers"

    workers_without_build_commands = project.workers.not_stale.in_use.machine_not_draining.free_workers
                                            .hash_or_empty(jobrun.rebuild_hash)
                                            .with_worker_image(jobrun.worker_image).without_build_commands.includes({ machine: :workers }).uniq

    workers_no_builds_and_mem = workers_without_build_commands.reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end

    workers_no_builds_and_mem = workers_no_builds_and_mem.reject do |w|
                                  w.machine.free_cpu <= 0
                                end.sort_by { |w| w.machine.free_cpu }.reverse
    Rails.logger.info "get_workers_for_project: timing Step 2 took #{Time.now - start_time} seconds"
    start_time = Time.now
    if workers_no_builds_and_mem.size > 0
      Rails.logger.info "get_workers_for_project: Found #{workers_no_builds_and_mem.size} free workers assigned to project #{project.id} with no build commands"
      new_workers += Worker.safe_assign workers_no_builds_and_mem, project, num_workers - new_workers.size,
                                        jobrun.supervisor_id, jobrun.id
      if new_workers.size >= num_workers
        Rails.logger.info "get_workers_for_project: Found #{new_workers.size} workers with memory space and assigned them to the project"
        jobrun.assigned_concurrency = new_workers.size
        val = jobrun.save!
        Rails.logger.debug "get_workers_for_project: Saved jobrun #{val}"
        Rails.logger.debug { "get_workers_for_project: Jobrun is now  #{jobrun.inspect}" }

        return new_workers
      end
    end

    # so no more workers already assigned to the project with the correct worker image and rebuild hash
    # so we need to assign some more workers
    # and when we assign we must also balance by removing older workers that aren't being used anymore.

    # so now workers with no project, that are free, not stale and have memory space on their machine
    # .reject { |w| project.machines.running.map(&:id).include?(w.machine_id) }

    workers = Worker.free_workers.not_stale.in_use.machine_not_draining.with_no_project.with_worker_image(jobrun.worker_image)
                    .where.not(machine_id: project.running_machines.map(&:id))

    Rails.logger.info do
      "get_workers_for_project: Found #{workers.size} free workers with no project and the correct worker image"
    end

    workers_with_memory_space = workers.reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end
    # we also do not want to put them on the same machine as each other
    # so we sort them by free cpu and then uniq them by machine_id

    workers_with_memory_space = workers_with_memory_space.reject do |w|
                                  w.machine.free_cpu <= 0
                                end.sort_by { |w| w.machine.free_cpu }.reverse

    workers_with_memory_space_uniq_machine = workers_with_memory_space.uniq(&:machine_id)
    Rails.logger.info "get_workers_for_project: Found #{workers_with_memory_space_uniq_machine.size} workers with memory space and not sharing a machine "
    Rails.logger.info "get_workers_for_project: timing Step 3 took #{Time.now - start_time} seconds"
    start_time = Time.now
    new_workers += Worker.safe_assign workers_with_memory_space_uniq_machine, project, num_workers - new_workers.size,
                                      jobrun.supervisor_id, jobrun.id

    if new_workers.size >= num_workers
      Rails.logger.info "get_workers_for_project: Found #{new_workers.size} workers with memory space and assigned them to the project"
      jobrun.assigned_concurrency = new_workers.size
      val = jobrun.save!
      Rails.logger.debug "get_workers_for_project: Saved jobrun #{val}"
      Rails.logger.debug { "get_workers_for_project: Jobrun is now  #{jobrun.inspect}" }
      return new_workers
    end

    # if we still need more workers we will try and assign new ones that are not sharing a machine with the existing project workers
    Rails.logger.info "get_workers_for_project: trying to assign new non-project workers to project #{project.id} with worker image #{jobrun.worker_image} "
    workers_with_memory_space_uniq_project = Worker.free_workers.not_stale.in_use.machine_not_draining.with_no_project.with_worker_image(jobrun.worker_image).where.not(id: new_workers.map(&:id)).where.not(machine_id: project.workers.not_stale.in_use.map(&:machine_id).uniq).uniq(&:machine_id).reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end.reject do |w|
                                               w.machine.free_cpu <= 0
                                             end.sort_by do |w|
      w.machine.free_cpu
    end.reverse
    Rails.logger.info "get_workers_for_project: timing Step 4 took #{Time.now - start_time} seconds"
    start_time = Time.now

    new_workers += Worker.safe_assign(workers_with_memory_space_uniq_project, project,
                                      (num_workers - new_workers.size), jobrun.supervisor_id, jobrun.id)

    if new_workers.size >= num_workers
      Rails.logger.info "get_workers_for_project: Found #{new_workers.size} workers with additional new ones and assigned them to project"
      jobrun.assigned_concurrency = new_workers.size
      val = jobrun.save!
      Rails.logger.debug "get_workers_for_project: Saved jobrun #{val}"
      Rails.logger.debug { "get_workers_for_project: Jobrun is now  #{jobrun.inspect}" }
      return new_workers
    end

    # if we still need more workers we will try and assign new ones
    Rails.logger.info "get_workers_for_project: trying to assign new non-project workers to project #{project.id} with worker image #{jobrun.worker_image} "
    workers_with_memory_space_uniq_machine = Worker.in_use.free_workers.machine_not_draining.with_no_project.with_worker_image(jobrun.worker_image).where.not(id: new_workers.map(&:id)).where.not(machine_id: new_workers.map(&:machine_id).uniq).uniq(&:machine_id).reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end.reject do |w|
                                               w.machine.free_cpu <= 0
                                             end.sort_by do |w|
      w.machine.free_cpu
    end.reverse
    Rails.logger.info "get_workers_for_project: timing Step 5 took #{Time.now - start_time} seconds"

    start_time = Time.now

    new_workers += Worker.safe_assign(workers_with_memory_space_uniq_machine, project,
                                      (num_workers - new_workers.size), jobrun.supervisor_id, jobrun.id)

    if new_workers.size >= num_workers
      Rails.logger.info "get_workers_for_project: Step 5 Found #{new_workers.size} workers and assigned them to the project"
      jobrun.assigned_concurrency = new_workers.size
      val = jobrun.save!
      Rails.logger.debug "get_workers_for_project: Saved jobrun #{val}"
      Rails.logger.debug { "get_workers_for_project: Jobrun is now  #{jobrun.inspect}" }
      return new_workers
    end

    Rails.logger.info "get_workers_for_project: Not enough uniq machineworkers with memory space - we have #{new_workers.size} workers and need #{num_workers} workers. Time to just assign what we can including non-uniq machines from the project"

    workers_with_memory_space_not_uniq = Worker.free_workers.not_stale.in_use.machine_not_draining.with_worker_image(jobrun.worker_image).where.not(id: new_workers.map(&:id)).reject do |w|
      w.machine.memory_used + project.memory_requirement > w.machine.memory
    end
    # This is where we make sure they have free CPU for us
    workers_with_memory_space_not_uniq = workers_with_memory_space_not_uniq.reject do |w|
                                           w.machine.free_cpu <= 0
                                         end.sort_by { |w| w.machine.free_cpu }.reverse
    Rails.logger.info "get_workers_for_project: timing Step 6 took #{Time.now - start_time} seconds"

    new_workers += Worker.safe_assign(workers_with_memory_space_not_uniq, project,
                                      (num_workers - new_workers.size), jobrun.supervisor_id, jobrun.id)
    jobrun.assigned_concurrency = new_workers.size
    jobrun.save!
    new_workers
  end
end
