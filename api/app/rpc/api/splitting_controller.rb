module Api
  class SplittingController < Api::GrufController
    bind ::Api::Splitting::Service
    include SplittingHelper

    def split_for_project # api.splitting.split_for_project
      mesg = request.message
      Rails.logger.debug("Got split_for_project with mesg : #{mesg.inspect}")
      raise GRPC::InvalidArgument.new('num_buckets must be > 0') if mesg.num_buckets < 1
      raise GRPC::InvalidArgument.new('no files provided') if mesg.filenames.empty?
      if mesg.filenames.size < mesg.num_buckets.to_i
        raise GRPC::InvalidArgument.new("num_buckets must be <= number of files but #{mesg.filenames.size} ! <= #{mesg.num_buckets}")
      end

      num_buckets = mesg.num_buckets
      project = request.metadata[:project]

      files, split_method = project.split_files num_buckets, mesg.filenames
      # we don't know the jr yet
      # jr.split_method = split_method
      # jr.save!
      unless files.size == num_buckets
        Rails.logger.error("We don't have enough buckets for the files we have #{files.size} != #{num_buckets}")
        raise "failure during split not enough buckets created #{files.size} != #{num_buckets}"
      end
      Rails.logger.debug('Got split_for_project returning files')
      files.each_with_index do |f, i|
        names = f.map(&:filename)
        Rails.logger.debug("Got split_for_project for bucket #{i} returning files #{names}")
      end

      to_split_response(files, split_method)
    end
  end
end
