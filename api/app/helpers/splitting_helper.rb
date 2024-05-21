module SplittingHelper
  # this is a 2d array so need to split it out to TestFiles objects
  def to_split_response(files, split_method)
    ::Api::SplitResponse.new({ split_method:, file_lists: files.map do |f|
                                                            ::Api::TestFiles.new({ filenames: f.map(&:filename) })
                                                          end })
  end
end
