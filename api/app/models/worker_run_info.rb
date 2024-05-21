class WorkerRunInfo < ApplicationRecord
  belongs_to :worker
  belongs_to :project
  belongs_to :supervisor
  belongs_to :jobrun
  has_many :test_file_runs, dependent: :destroy
  has_many :test_files, through: :test_file_runs
  has_many :execution_infos, dependent: :destroy
  encrypts :log_encryption_key

  before_validation :set_uid, on: :create
  # before_create :set_encryption_key

  validates :uid, uniqueness: true, allow_nil: true

  scope :succeeded, -> { where(exit_code: '0') }
  scope :with_number_of_test_files_db, lambda { |count|
                                         joins(:test_file_runs).group(:id).having('COUNT(test_file_runs.id) = ?', count)
                                       }

  scope :with_number_of_test_files, lambda { |count|
                                      where('test_count = ?', count)
                                    }

  def set_test_file_count
    self.test_count = test_files.count
    save!
  end

  def get_log_bucket_key
    uid
  end

  def get_s3_encryption_key
    log_encryption_key
  end

  # this should be set from the other side but if not we need a uid to access it

  def set_uid
    self.uid = SecureRandom.uuid unless uid
  end

  # def set_encryption_key
  #   self.log_encryption_key = SecureRandom.alphanumeric(32) unless self.log_encryption_key
  # end

  def duration
    ms_time_taken
  end

  def test_run_duration
    execution_infos&.last&.duration
  end

  def succeeded?
    exit_code == '0'
  end

  def record_test_file_run(wi)
    MyAppTracer.in_span('record_test_file_run') do |_span|
      wi.files.each do |filename|
        # pytest doesn't use filenames
        next unless filename.present?

        MyAppTracer.in_span('record_test_file_run-each-file') do |_file_span|
          raise "we can't save an empty filename" if filename.blank?

          test_file = project.test_files.find_or_create_by!(filename:)
          # we have an adjustment to make here because we don't want to include the startup time
          # this will vary per project so we allow it to be configurable
          adjusted_time_taken = ((ms_time_taken - (project.startup_time_in_ms || 1750))) / number_of_contending_runs
          adjusted_time_taken = 100 if adjusted_time_taken < 100
          adjusted_time_taken /= wi.files.size

          test_file_runs.create! test_file:, ms_time_taken: adjusted_time_taken

          Rails.logger.debug("WRI Timing #{id} starting with #{test_file.reload.inspect}")
          Rails.logger.debug("WRI Timing #{id} #{test_file.filename} #{test_file.timing_confidence} #{test_file.runtime} #{adjusted_time_taken} #{wi.files.size} #{wi.files} #{number_of_contending_runs} #{ms_time_taken} #{project.startup_time_in_ms}")

          if succeeded?
            if wi.files.size == 1
              # we are somewhat confident if we just have a single file
              test_file.runtime = if test_file.runtime || 0 > 0
                                    # we have a previous value so we should average it
                                    (test_file.runtime + adjusted_time_taken) / 2
                                  else
                                    adjusted_time_taken
                                  end
              test_file.timing_confidence
              test_file.save!
            elsif (test_file.timing_confidence || 0) < 1
              # we are less confident if we have multiple files
              # we should average the previous value with the new value
              # we should also reduce the confidence
              test_file.runtime = if test_file.runtime || 0 > 0
                                    (test_file.runtime * 4 + adjusted_time_taken) / 6
                                  else
                                    adjusted_time_taken
                                  end
              test_file.timing_confidence = 1 / wi.files.size.to_f
              test_file.save!
            end
            Rails.logger.debug("WRI Timing #{id} finished with #{test_file.reload.inspect}")
            # now we can check previous runs to see if we have a difference of one file which will give us high confidence
          end
        end
      end
      save!
    end
  end

  # this gets a previous run that has one more test file than we do
  # we use this to compare the timings and see if we can put an accurate time on a test
  def get_previous_one_off_runs
    MyAppTracer.in_span('get_previous_one_off_runs') do |_span|
      filenames = test_files.map(&:filename)
      test_file_count = test_files.size + 1

      same_size_with_files = project.worker_run_infos.succeeded.with_number_of_test_files(test_file_count).joins(:jobrun).where("jobruns.state = 'completed'").joins(:test_files).where(test_files: { filename: filenames }).order('worker_run_infos.created_at desc').limit(10)

      # when we delete the larger array from the smaller we should wipe out the smaller showing that it is a super set
      # we are smaler than them
      same_size_with_files.select { |wri| (test_files - wri.test_files).empty? }
    end
  end

  # if we succeeded and have one more file than the previous run we can be pretty confident that we have the correct timing
  def compare_to_previous_run
    Rails.logger.debug "about to compare #{id} %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% \n"
    return 'we never succeeded' unless succeeded?
    # we account for single runs when recording the test file run
    return 'we have less than 1 files' unless test_files.size > 0

    #  I can search for the exact test files ...
    Rails.logger.debug "about to get previous run #{id} %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% \n"
    wri = get_previous_one_off_runs.first
    return 'no previous run' unless wri

    Rails.logger.debug "about to compare #{wri.id} and #{id} %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% \n"
    if wri.succeeded? && test_files.count + 1 == wri.test_files.count
      diff = wri.test_files - test_files
      Rails.logger.debug "diff is #{diff}  %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% \n"
      if diff.size == 1
        test_file = diff.first
        # we have a high confidence value
        our_contention = number_of_contending_runs
        their_contention = wri.number_of_contending_runs
        timing_estimate = (wri.test_run_duration / their_contention - test_run_duration / our_contention)
        if timing_estimate < 0
          Rails.logger.debug 'timing estimate is negative so we are going to return nil  '
          return nil
        end
        combined_contention = our_contention + their_contention
        if test_file.runtime > 0
          # we have a previous value so we should average it
          test_file.runtime = (test_file.runtime / combined_contention + (timing_estimate * (combined_contention - 1))) / combined_contention
        else
          test_file.runtime = timing_estimate
        end
        # high confidence with no contention
        test_file.timing_confidence = 2 / combined_contention.to_f
        test_file.save!
        Rails.logger.debug "Found a high confidence value for #{test_file.filename} with timing #{test_file.runtime} when comparing wris #{wri.id} and #{id}"
      else
        Rails.logger.debug "MOre than one difference between #{wri.id} and #{id} - should only be one %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% but we got #{diff.inspect} \n"
        Rails.logger.error "Found more than one difference between #{wri.id} and #{id} - should only be one"
        Rails.logger.error "other has #{wri.test_files} we have #{test_files} difference between them is #{(wri.test_files.to_a - test_files.to_a).inspect}}"
      end
    else
      Rails.logger.debug "We didn't succeed or we didn't have one more file"
    end
  end

  # an estimate of how many runs are contending for the same resources on the cpu
  def number_of_contending_runs
    @contending_runs ||= begin
      contending = (worker.machine.overlapping_worker_run_infos(self).count + 1) / (worker.machine.cpus.to_f / 2)
      Rails.logger.error("contending is #{contending} for #{id} on #{worker.machine.overlapping_worker_run_infos(self).count} overlapping runs")
      contending = 1 if contending < 1
      contending
    end
  end
end
