require 'rails_helper'
describe Api::SplittingController do
  describe 'split_for_project_with no credentials' do
    subject { run_rpc(:SplitForProject, request_proto) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 30),
        create(:test_file, project:, filename: 'test4', runtime: 40)
      ]
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: test_files.map(&:filename)) }

    it 'is unauthenticated' do
      expect { subject }.to raise_error(GRPC::Unauthenticated)
    end
  end

  describe 'split_for_project' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 30)
      ]
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: test_files.map(&:filename)) }

    it('splits for 2') do
      expect(subject).to be_a(Api::SplitResponse)

      expect(subject.file_lists.map(&:filenames)).to eq [['test3'], %w[test2 test1]]
    end
  end

  describe 'split_a_long_complicated_one' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 30),
        create(:test_file, project:, filename: 'test04', runtime: 4)
      ]
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 4 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: test_files.map(&:filename)) }

    it('splits for 4 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [['test3'], ['test2'], ['test1'], ['test04']]
    end
  end

  describe 'only return files we pass' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 30),
        create(:test_file, project:, filename: 'test04', runtime: 4)
      ]
    end

    let(:passed_files) { %w[test1 test2] }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: passed_files) }

    it('splits for 2 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [['test2'], ['test1']]
    end
  end

  describe 'only return files we pass test' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 20),
        create(:test_file, project:, filename: 'test4', runtime: 20)

      ]
    end

    let(:passed_files) { %w[test1 test2] }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: passed_files) }

    it('splits for 2 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [['test2'], ['test1']]
    end
  end

  describe 'returns all the files we pass' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 10),
        create(:test_file, project:, filename: 'test2', runtime: 20),
        create(:test_file, project:, filename: 'test3', runtime: 20),
        create(:test_file, project:, filename: 'test4', runtime: 20)

      ]
    end

    let(:passed_files) { %w[test1 test2 test5] }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: passed_files) }

    it('splits for 2 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [['test2'], %w[test5 test1]]
    end
  end

  describe 'returns all the files we pass using an average value' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 50),
        create(:test_file, project:, filename: 'test2', runtime: 10),
        create(:test_file, project:, filename: 'test3', runtime: 10),
        create(:test_file, project:, filename: 'test4', runtime: 10)

      ]
    end

    let(:passed_files) { %w[test1 test2 test5] }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: passed_files) }

    it('splits for 2 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [['test1'], %w[test5 test2]]
    end
  end

  describe 'multiple unknowns all use average value' do
    subject { run_rpc(:SplitForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let!(:test_files) do
      [
        create(:test_file, project:, filename: 'test1', runtime: 50),
        create(:test_file, project:, filename: 'test2', runtime: 20)

      ]
    end

    let(:passed_files) { %w[test1 test2 test5 test6] }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:num_buckets) { 2 }
    let(:request_proto) { Api::SplitRequest.new(num_buckets:, filenames: passed_files) }

    it('splits for 2 correctly') do
      expect(subject).to be_a(Api::SplitResponse)
      expect(subject.file_lists.map(&:filenames).map(&:to_a)).to eq [%w[test1 test2], %w[test6 test5]]
    end
  end
end
