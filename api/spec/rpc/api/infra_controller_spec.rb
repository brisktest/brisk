# frozen_string_literal: true

require 'rails_helper'
describe Api::InfraController do
  describe 'register' do
    subject { run_rpc(:RegisterMachine, mesg, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.infra.register_machine') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end

    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json,
        authenticated_action: 'api.infra.register' }
    end
    let(:uid) { '1234' }
    let(:request_proto) do
      # Uid: client.ID, IpAddress: client.ID, HostIp: client.HTTPAddr, HostUid: client.ID, Image: client.Attributes["image"], Region: client.Attributes["region"], Cpus: uint32(*client.Resources.CPU), Memory: uint32(*client.Resources.MemoryMaxMB), Disk: uint32(*client.Resources.DiskMB), OsInfo: client.Attributes["os"]

      ::Api::Machine.new(uid:, ip_address: '134.212.122.32', host_ip: '122.32.122.33', host_uid: '1234',
                         image: 'ami-something', region: 'us-east-1', cpus: 2, memory: '204812212321', disk: '404812231232', os_info: 'ubuntu')
    end

    let(:mesg) { Api::MachineReq.new(machine: request_proto) }

    it 'registers' do
      expect(subject).to be_a(Api::MachineResponse)
      puts(subject.inspect)

      machine = Machine.find(subject.machine.id)
      expect(machine.uid).to eq uid
      expect(machine.memory).to eq 195_324
    end

    context 'when authenticated_action does not match' do
      subject { run_rpc(:RegisterMachine, request_proto, active_call_options: { metadata: }) }

      let(:metadata) { { authenticated_action: 'api.infra.different' } }

      it 'returns an error' do
        expect { subject }.to raise_error(GRPC::Unauthenticated)
      end
    end
  end

  describe 'de_register_machine' do
    subject { run_rpc(:DeRegisterMachine, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.infra.de_register_machine') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end

    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json,
        authenticated_action: 'api.infra.deregister' }
    end
    let(:uid) { '1234' }
    let(:m) { FactoryBot.create(:machine, uid:) }
    let(:machine) { ::Api::Machine.new(uid:) }

    let(:request_proto) { ::Api::MachineReq.new(machine:) }

    it 'fails to deregister non-existent machine' do
      expect { subject }.to raise_error(GRPC::Internal)
    end

    context 'when authenticated_action does not match' do
      subject { run_rpc(:DeRegisterMachine, request_proto, active_call_options: { metadata: }) }

      let(:metadata) { { authenticated_action: 'api.infra.different' } }

      it 'returns an error' do
        expect { subject }.to raise_error(GRPC::Unauthenticated)
      end
    end
  end

  describe 'de_register_machine' do
    subject { run_rpc(:DeRegisterMachine, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.infra.de_register_machine') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end

    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json,
        authenticated_action: 'api.infra.deregister' }
    end
    let(:uid) { '1234' }
    let!(:m) { FactoryBot.create(:machine, uid:) }
    let(:machine) { ::Api::Machine.new(uid:) }

    let(:request_proto) { ::Api::MachineReq.new(machine:) }

    it 'deregisters' do
      expect(subject).to be_a(Api::MachineResponse)
      puts(subject.inspect)

      machine = ::Machine.find_by_uid(m.uid)

      expect(machine.finished_at).not_to be_nil
    end

    context 'when authenticated_action does not match' do
      subject { run_rpc(:DeRegisterMachine, request_proto, active_call_options: { metadata: }) }

      let(:metadata) { { authenticated_action: 'api.infra.different' } }

      it 'returns an error' do
        expect { subject }.to raise_error(GRPC::Unauthenticated)
      end
    end
  end

  describe 'drain_machine' do
    subject { run_rpc(:DrainMachine, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.infra.drain_machine') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end

    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json,
        authenticated_action: 'api.infra.drain' }
    end
    let(:uid) { '1234' }
    let(:m) { FactoryBot.create(:machine, uid:) }
    let(:machine) { ::Api::Machine.new(uid:) }

    let(:request_proto) { ::Api::MachineReq.new(machine:) }

    it 'fails to drain non-existent machine' do
      expect { subject }.to raise_error(GRPC::Internal)
    end

    context 'when authenticated_action does not match' do
      subject { run_rpc(:DrainMachine, request_proto, active_call_options: { metadata: }) }

      let(:metadata) { { authenticated_action: 'api.infra.different' } }

      it 'returns an error' do
        expect { subject }.to raise_error(GRPC::Unauthenticated)
      end
    end
  end

  describe 'drain_machine_that_exists' do
    subject { run_rpc(:DrainMachine, request_proto, active_call_options: { metadata: }) }

    let(:api_action) { FactoryBot.create(:api_action, grpc_method_name: 'api.infra.drain_machine') }
    let(:credentials) do
      api_action.generate_new_credential
      api_action.credentials.first
    end

    let(:metadata) do
      { 'authorization' => { api_token: credentials.api_token,
                             api_key: credentials.api_key }.to_json,
        authenticated_action: 'api.infra.drain' }
    end
    let(:uid) { '1234' }
    let!(:m) { FactoryBot.create(:machine, uid:) }
    let(:machine) { ::Api::Machine.new(uid:) }

    let(:request_proto) { ::Api::MachineReq.new(machine:) }

    it 'drains machine' do
      expect(subject).to be_a(Api::MachineResponse)
      puts(subject.inspect)

      machine = ::Machine.find_by_uid(m.uid)

      expect(machine.drained_at).to be_truthy
    end

    context 'when authenticated_action does not match' do
      subject { run_rpc(:DrainMachine, request_proto, active_call_options: { metadata: }) }

      let(:metadata) { { authenticated_action: 'api.infra.different' } }

      it 'returns an error' do
        expect { subject }.to raise_error(GRPC::Unauthenticated)
      end
    end
  end
end
