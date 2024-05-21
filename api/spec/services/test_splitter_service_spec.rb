require 'rails_helper'

describe TestSplitterService do
  describe 'split_files_with_test_files' do
    # this project doesn't have any previous test files
    let(:test_files) do
      [TestFile.new(filename: 'test1', runtime: 10), TestFile.new(filename: 'test2', runtime: 20),
       TestFile.new(filename: 'test3', runtime: 30)]
    end
    let(:project) { FactoryBot.create(:project, test_files:) }

    it 'splits files in 2' do
      test_splitter = TestSplitterService.new(project, 2, test_files.map(&:filename))

      expect(test_splitter.default_split.map { |t| t.map(&:filename) }).to eq [['test3'], %w[test2 test1]]
    end
  end

  describe 'split_files' do
    # let(:project) { FactoryBot.create(:project) }
    let(:project) do
      FactoryBot.create(:project,
                        test_files: [TestFile.new(filename: 'test1', runtime: 10), TestFile.new(filename: 'test2', runtime: 20),
                                     TestFile.new(filename: 'test3', runtime: 30)])
    end

    let(:files) do
      [TestFile.new(filename: 'test1', runtime: 10, project:), TestFile.new(filename: 'test2', runtime: 20, project:),
       TestFile.new(filename: 'test3', runtime: 30, project:)]
    end

    it 'splits files in 3' do
      test_splitter = TestSplitterService.new(project, 3, files.map { |f| f[:filename] })
      expect(test_splitter.default_split.map do |t|
               t.map(&:filename)
             end).to eq([[TestFile.new(filename: 'test3', runtime: 30)],
                         [TestFile.new(filename: 'test2', runtime: 20)], [TestFile.new(filename: 'test1', runtime: 10)]].map do |t|
                          t.map(&:filename)
                        end)
    end

    it 'splits files in 1' do
      test_splitter = TestSplitterService.new(project, 1, files.map { |f| f[:filename] })
      expect(test_splitter.default_split.map do |t|
               t.map(&:filename)
             end).to eq([[TestFile.new(filename: 'test3', runtime: 30),
                          TestFile.new(filename: 'test2', runtime: 20), TestFile.new(filename: 'test1', runtime: 10)]].map do |t|
                          t.map(&:filename)
                        end)
    end

    it 'splits files in 2' do
      test_splitter = TestSplitterService.new(project, 2, files.map { |f| f[:filename] })
      expect(test_splitter.default_split.map do |t|
               t.map(&:filename)
             end).to eq([[TestFile.new(filename: 'test3', runtime: 30)],
                         [TestFile.new(filename: 'test2', runtime: 20),
                          TestFile.new(filename: 'test1', runtime: 10)]].map do |t|
                          t.map(&:filename)
                        end)
    end
  end
end
