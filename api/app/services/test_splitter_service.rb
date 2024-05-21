require 'digest/sha1'

class TestSplitterService
  def clear_split
    REDIS_POOL.with do |conn|
      conn.del(key_for_project_and_files_and_buckets)
    end
  end

  def self.clear_all_splits
    REDIS_POOL.with do |conn|
      conn.keys('test_splitter_service:*').each do |key|
        conn.del(key)
      end
    end
  end

  def get_files
    @files
  end

  def get_project
    @project
  end

  def key_for_project_and_files_and_buckets
    "test_splitter_service:#{Digest::SHA1.hexdigest(@project.id.to_s + @num_buckets.to_s + @files.map(&:filename).sort.join)}"
  end

  def store_split
    REDIS_POOL.with do |conn|
      conn.set(key_for_project_and_files_and_buckets, Marshal.dump(split_files))
    end
  end

  def retrieve_split
    begin
      REDIS_POOL.with do |conn|
        data = conn.get(key_for_project_and_files_and_buckets)
        if data
          Rails.logger.info("Found data in redis for project_id: #{@project.id} with num_buckets: #{@num_buckets} #{key_for_project_and_files_and_buckets}")
          Rails.logger.debug("Data: #{data.inspect}")
          return Marshal.load(data)
        end
      end
    rescue StandardError => e
      Rails.logger.error("Error loading data from redis: #{e}")
      Sentry.capture_exception(e)
    end
    Rails.logger.info("No data found in redis for #{key_for_project_and_files_and_buckets}")
    nil
  end

  def initialize(project, num_buckets, files_to_split, lookback = 2)
    @lookback = ENV.fetch('TEST_SPLITTER_LOOKBACK_WINDOW', lookback).to_i
    project_test_files = project.test_files
    Rails.logger.debug("project_test_files: #{project_test_files.sort_by(&:filename).inspect}")
    Rails.logger.debug("num_buckets: #{num_buckets}")
    Rails.logger.debug("files_to_split: #{files_to_split.sort.inspect}")
    averageTimeValue = if project_test_files.empty?
                         1
                       else
                         project_test_files.map { |f| f[:runtime] || 10_000 }.sum / project_test_files.count
                       end
    Rails.logger.debug("averageTimeValue: #{averageTimeValue}")

    @files = files_to_split.map do |filename|
      found_file = project_test_files.find { |file| file.filename == filename }
      if found_file
        OpenStruct.new(filename: found_file.filename,
                       runtime: found_file.runtime)
      else
        OpenStruct.new(
          filename:, runtime: averageTimeValue
        )
      end
    end

    Rails.logger.debug("files: #{@files.inspect}")
    @num_buckets = num_buckets
    @project = project
  end

  # each test file has a runtime
  # we want to split the files into num_buckets
  # we want to optimize so that the runtime of the slowest bucket is lowest

  # we can sort the files by runtime
  # then we can start with the longest files
  # and add them to the buckets
  # we can keep track of the total time for each bucket
  # and the total time for all buckets
  # then we can add the next file to the bucket with the lowest total time
  # and we can keep track of the total time for all buckets
  # and we can repeat until we have added all files
  # then we can return the buckets

  def previous_split(lookback = 2)
    fetch_previous_split(@project, @num_buckets, @files, lookback)
  end

  def split_files
    previous_jr = previous_split @lookback

    if previous_jr
      Rails.logger.debug("using previous split #{previous_jr.id}")
      wris = previous_jr.worker_run_infos.map do |wri|
        OpenStruct.new(ms_time_taken: wri.ms_time_taken, test_files: wri.test_files.map do |tf|
                                                                       OpenStruct.new(filename: tf.filename, runtime: tf.runtime)
                                                                     end)
      end

      result = partition_test_files(@num_buckets, wris)
      Rails.logger.info("old_split: #{wris.sort_by(&:ms_time_taken).map(&:test_files).inspect}, new split_files: #{result.inspect}")
      return result
    else
      Rails.logger.debug("No previous jobrun for project id: #{@project.id} with #{@num_buckets} buckets and files #{@files} ")
    end
    nil
  end

  def default_split
    files = @files.sort_by { |f| f[:runtime] || 50_000 }.reverse
    buckets = Array.new(@num_buckets) { [] }
    bucket_times = Array.new(@num_buckets) { 0 }
    total_time = 0

    files.each do |file|
      bucket = bucket_times.index(bucket_times.min)
      buckets[bucket] << file
      bucket_times[bucket] += file[:runtime] || 50_000
      total_time += file[:runtime] || 50_000
    end
    buckets
  end

  # test all this stuff in rspec cause it's pretty fragile and we want to make sure it works and keeps working

  # if we have a previous split with the right number of buckets and the same files we can partition
  # otherwise use our learned info

  def fetch_previous_split(project, num_buckets, files, lookback = 2)
    Rails.logger.debug("fetching previous split for project id: #{project.id} with #{num_buckets} buckets and files #{files} (looking back #{lookback} jobruns) ")
    my_files = project.test_files.where(filename: files.map(&:filename))
    # if we don't have all the files we can't use the previous split
    return nil if my_files.size != files.size

    project.jobruns.where(assigned_concurrency: num_buckets, state: 'completed').last

    #   ps = project.jobruns
    #     .where(assigned_concurrency: num_buckets, state: "completed")
    #     .order(Arel.sql("finished_at - created_at ASC"))
    #     .includes(:test_files)
    #     .includes(:worker_run_infos => :execution_infos)
    #     .joins(:test_files).where(test_files: { filename: files.map(&:filename) })
    #     .limit(lookback)
    #     .sort_by { |jr|
    #     jr.worker_run_infos.max { |wri|
    #       wri.execution_infos.select { |ei| ei.command.stage == "Run" }
    #         .map(&:duration).sum
    #     }
    #   }

    #   return nil if ps.nil? || ps.worker_run_infos.size != num_buckets
    #   Rails.logger.info("Found previous split #{ps.id} with #{ps.uid} for project id: #{project.id} with #{num_buckets} buckets and files #{files} (looking back #{lookback} jobruns) ")
    #   return ps
    # rescue => e
    #   Rails.logger.error("Error fetching previous split - not splitting : #{e.message}")
    #   return nil
  end

  def partition_test_files(_num_buckets, previous_split)
    out = []
    i = previous_split.size
    previous_split = previous_split.sort_by(&:ms_time_taken)

    # remove the slow ones that only have one file because they'll pair with a fast one
    no_need_to_split = previous_split.select.with_index { |pr, i| i > previous_split.size && pr.test_files.size == 1 }
    out << no_need_to_split.map(&:test_files)

    needs_split = previous_split.reject.with_index { |pr, i| i > previous_split.size && pr.test_files.size == 1 }

    j = 0
    i = needs_split.size
    while j <= (i - 1) / 2
      if j == i - j - 1
        out << [needs_split[j].test_files]
      elsif needs_split[i - j - 1].test_files.size == 1
        # if our big only has one file we don't split it we just return the two
        out << [needs_split[j].test_files, needs_split[i - j - 1].test_files]
      else
        # we can swap the slowest file from the big bucket to the small bucket
        Rails.logger.debug("swapping files in the loop j:#{j} and big_side:#{i - j - 1}")
        out << swap(needs_split[j], needs_split[i - j - 1])
      end

      j += 1
    end

    out = out.flatten(1)
    Rails.logger.debug("paritioning test files we have #{out.size} buckets with #{out.map(&:size)} files in each")
    out
  end

  def swap(small, big)
    big_test_files = big.test_files.sort_by(&:runtime)

    small_test_files = small.test_files

    Rails.logger.debug("big test files has #{big_test_files.size} files")
    file = big_test_files.shift
    Rails.logger.debug("big test files now has #{big_test_files.size} files")
    Rails.logger.debug("swapping #{file.filename} between big and small")

    Rails.logger.debug("small now has #{small_test_files.size} files")
    small_test_files.push(file)
    Rails.logger.debug("small now has #{small_test_files.size} files")

    Rails.logger.debug("small now has #{small_test_files.size} files big now has #{big_test_files.size} files
    small: #{small_test_files.map(&:filename).inspect}
    big: #{big_test_files.map(&:filename).inspect}
      ")

    [small_test_files, big_test_files]
  end
end
