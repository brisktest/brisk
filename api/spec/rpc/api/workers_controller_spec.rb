# frozen_string_literal: true

require 'rails_helper'
describe Api::WorkersController do
  describe 'register' do
    subject { run_rpc(:Register, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.workers.register') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end
    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json }
    end
    let(:machine) { FactoryBot.create(:machine) }
    let!(:image) { FactoryBot.create(:image, name: 'rails') }

    let(:request_proto) do
      ::Api::WorkerRegReq.new(host_ip: machine.ip_address, ip_address: '123.12.211.2', uid: rand.to_s,
                              worker_image: 'rails', host_uid: machine.uid, sync_port: '1234')
    end

    it 'registers' do
      expect(subject).to be_a(Api::WorkerResponse)
      puts(subject.inspect)

      worker = Worker.find(subject.worker.id)
      expect(worker.machine).to eq machine
      expect(worker.state).to eq 'active'
    end
  end

  describe 'deregister' do
    subject { run_rpc(:DeRegister, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.workers.de_register') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end
    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json }
    end
    let(:project) { FactoryBot.create(:project) }
    let(:worker) { FactoryBot.create(:worker, state: :assigned, project:) }
    let(:request_proto) { ::Api::WorkerReq.new(uid: worker.uid) }

    it 'deregisters' do
      expect(subject).to be_a(Api::WorkerResponse)
      puts(subject.inspect)

      worker = Worker.find(subject.worker.id)

      expect(worker.state).to eq 'finished'
    end
  end

  describe 'build_commands_run_at' do
    subject { run_rpc(:BuildCommandsRun, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let(:worker) { FactoryBot.create(:worker, state: :assigned, project:) }
    let!(:request_proto) { ::Api::CommandsRunReq.new(id: worker.id) }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'sets build commands run at for a worker' do
      expect(subject).to be_a(Api::WorkerResponse)
      puts(subject.inspect)

      worker = Worker.find(subject.worker.id)
      puts worker.inspect
      expect(worker.build_commands_run_at.nil?).to be false
    end
  end

  describe 'get_recently_deregistered_workers' do
    subject { run_rpc(:GetRecentlyDeregistered, request_proto, active_call_options: { metadata: }) }

    let(:user) { FactoryBot.create(:user) }
    let(:project) { FactoryBot.create(:project) }
    let(:worker) { FactoryBot.create(:worker, state: :finished, project:, updated_at: Time.now) }
    let!(:request_proto) { ::Api::WorkersReq.new }
    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.workers.get_recently_deregistered') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end
    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json }
    end

    it 'returns recently deregistered' do
      worker

      expect(subject).to be_a(Api::WorkersResp)
      puts(subject.inspect)

      expect(subject.workers.size).to eq(1)
      expect(subject.workers.first.uid).to eq(worker.uid)
    end
  end
end
