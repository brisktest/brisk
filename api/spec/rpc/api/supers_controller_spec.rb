# frozen_string_literal: true

require 'rails_helper'
describe Api::SupersController do
  describe 'register' do
    subject { run_rpc(:Register, request_proto, active_call_options: { metadata: }) }

    let(:machine) { FactoryBot.create(:machine) }
    let(:request_proto) do
      ::Api::SuperRegReq.new(host_ip: machine.ip_address, ip_address: '123.12.211.2', sync_port: '2222',
                             host_uid: machine.uid)
    end
    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.supers.register') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end
    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json }
    end

    it 'registers' do
      expect(subject).to be_a(Api::SuperResponse)
      puts(subject.inspect)

      sup = Supervisor.find(subject.super.id)
      expect(sup.machine).to eq machine
      expect(sup.state).to eq 'ready'
    end

    it 'registers the sync port' do
      expect(subject).to be_a(Api::SuperResponse)
      puts(subject.inspect)

      sup = Supervisor.find(subject.super.id)
      expect(sup.sync_port).to eq '2222'
    end
  end

  describe 'deregister' do
    subject { run_rpc(:DeRegister, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.supers.de_register') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end
    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json }
    end

    let(:project) { FactoryBot.create(:project) }
    let(:sup) { FactoryBot.create(:supervisor, state: :assigned, project:) }
    let(:request_proto) { ::Api::SuperReq.new(id: sup.id) }

    it 'deregisters' do
      expect(subject).to be_a(Api::SuperResponse)
      puts(subject.inspect)

      sup = Supervisor.find(subject.super.id)

      expect(sup.state).to eq 'finished'
    end
  end

  describe 'mark_super_as_unreachable' do
    subject { run_rpc(:MarkSuperAsUnreachable, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let(:request_proto) { Api::GetProjectReq.new(project_token: project.project_token) }

    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    let(:sup) { FactoryBot.create(:supervisor, state: :assigned, project:) }
    let(:request_proto) { ::Api::UnreacheableReq.new({ super: { id: sup.id } }) }

    it 'marks unreachable' do
      expect(subject).to be_a(Api::UnreachableResp)
      puts(subject.inspect)
      sup = Supervisor.find(subject.super.id)
      expect(sup.state).to eq 'finished'
    end
  end
end
