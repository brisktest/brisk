module JobrunsHelper
  def link_to_online_repo(jobrun)
    if jobrun.repo_info && jobrun.project.is_github_repo?
      link_to jobrun.repo_info.commit_hash, "https://github.com/#{jobrun.project.git_org}/#{jobrun.project.git_repo_name}/commit/#{jobrun.repo_info.commit_hash}"
    elsif jobrun.repo_info && jobrun.project.is_gitlab_repo?
      # https://gitlab.com/brisktest1/runner/-/commit/1a477ac77f679e9a5395feaeb7d48e29a5504407
      link_to jobrun.repo_info.commit_hash, "https://gitlab.com/#{jobrun.project.git_org}/#{jobrun.project.git_repo_name}/-/commit/#{jobrun.repo_info.commit_hash}"
    else
      ''
    end
  end

  def text_for_link_to_online_repo(jobrun)
    if jobrun.repo_info && jobrun.project.is_github_repo?
      "https://github.com/#{jobrun.project.git_org}/#{jobrun.project.git_repo_name}/commit/#{jobrun.repo_info.commit_hash}"
    elsif jobrun.repo_info && jobrun.project.is_gitlab_repo?
      # https://gitlab.com/brisktest1/runner/-/commit/1a477ac77f679e9a5395feaeb7d48e29a5504407
      "https://gitlab.com/#{jobrun.project.git_org}/#{jobrun.project.git_repo_name}/-/commit/#{jobrun.repo_info.commit_hash}"
    else
      ''
    end
  end

  def link_to_honeycomb_trace_for_jobrun(jobrun)
    trace_id = jobrun.parse_trace_key
    return '' unless trace_id

    link_to_honeycomb_trace(trace_id, jobrun.created_at.to_i, jobrun.updated_at.to_i)
  end

  def link_to_honeycomb_trace(trace_id, start_time, _end_time, dataset = 'brisk-cli', environment = 'production',
                              team = 'brisktest')
    trace_id = ERB::Util.url_encode(trace_id)

    # return "https://ui.honeycomb.io/#{team}/environments/#{environment}/datasets/#{dataset}/trace?trace_id=#{trace_id}&trace_start_ts=#{start_time}&trace_end_ts=#{end_time}"
    "https://ui.honeycomb.io/#{team}/environments/#{environment}/datasets/#{dataset}/trace?trace_id=#{trace_id}&trace_start_ts=#{start_time}"
  end

  def get_jobrun_link(jobrun)
    project_jobrun_url(jobrun.project.project_token, jobrun.uid)
  end
end
