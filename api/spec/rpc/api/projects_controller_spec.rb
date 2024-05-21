# frozen_string_literal: true

require 'rails_helper'
describe Api::ProjectsController do
  describe 'get_project' do
    FactoryBot.create(:plan)
    subject { run_rpc(:GetProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let(:request_proto) { Api::GetProjectReq.new }

    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the project' do
      expect(subject).to be_a(Api::GetProjectResp)
      puts(subject.inspect)

      expect(Project.find(subject.project.id).id).to eq project.id
      expect(Project.find(subject.project.id).name).to eq project.name
    end
  end

  describe 'init_project' do
    subject { run_rpc(:InitProject, request_proto, active_call_options: { metadata: }) }

    let(:user) { FactoryBot.create(:user) }
    let(:request_proto) { Api::InitProjectReq.new(framework: 'Python') }
    let(:metadata) do
      { 'authorization' => { api_token: user.credential.api_token, api_key: user.credential.api_key }.to_json }
    end
    let(:org) { FactoryBot.create(:org, owner: user) }

    FactoryBot.create(:image)

    it 'returns the project' do
      expect(subject).to be_a(Api::InitProjectResp)
      puts(subject.inspect)
    end

    it 'creates a project' do
      expect(subject.project.id).not_to be_nil
      expect(subject.project.name).not_to be_nil
      expect(subject.project.framework).to eq 'Python'
    end
  end

  describe '.get_workers_for_project' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) { FactoryBot.create(:worker) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 1, worker_image: 'rails', supervisor_uid: supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the workers' do
      puts '--------------------------------------------'
      puts project.inspect
      puts project.org.inspect
      puts project.account.inspect
      puts '******no************'
      puts project.subscription
      puts project.plan

      puts '^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^'
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id
      expect(subject.jobrun_id).not_to be_nil
    end
  end

  describe '.get_workers_for_project3' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, worker_concurrency: 2) }
    let(:machine) { FactoryBot.create(:machine) }
    let(:machine2) { FactoryBot.create(:machine) }
    let!(:worker) { FactoryBot.create(:worker, machine:) }
    let!(:worker2) { FactoryBot.create(:worker, machine: machine2) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns two projects if we have them' do
      expect(subject.workers.size).to eq 2
      puts(subject.inspect)
    end
  end

  describe '.get_workers_for_projectmore' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, worker_concurrency: 1) }
    let!(:worker) { FactoryBot.create(:worker) }
    let!(:worker2) { FactoryBot.create(:worker, host_ip: '122.232.12.2') }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'fails we have concurrency at 1' do
      expect { subject }.to raise_error(StandardError)
    end
  end

  # describe ".get_workers_for_project" do
  #   let(:project) { FactoryBot.create(:project) }
  #   let!(:worker) { FactoryBot.create(:worker, host_ip: "127.0.0.1") }
  #   let!(:worker2) { FactoryBot.create(:worker, host_ip: "127.0.0.1") }
  #   let(:supervisor) { FactoryBot.create(:supervisor, project: project) }
  #   let(:request_proto) { Api::GetWorkersReq.new(num_workers: 2, worker_image: "rails", supervisor_uid: supervisor.uid) }
  #   let(:metadata) do
  #     { "authorization" => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
  #                           api_key: project.users.first.credential.api_key }.to_json }
  #   end

  #   subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: metadata }) }

  #   it "will not return a project with the same host ip" do
  #     expect(worker.host_ip).to_not eq worker2.host_ip
  #     expect(subject.workers.size).to eq 2

  #     puts subject.workers[0].inspect
  #     puts subject.workers[0].host_ip
  #     puts subject.workers[1].host_ip
  #     expect(subject.workers[0].host_ip).to_not eq subject.workers[1].host_ip
  #     puts(subject.inspect)
  #   end
  # end

  describe '.get_workers_for_project2' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild', machine: FactoryBot.create(:machine), project_id: project.id,
                                 state: :assigned, build_commands_run_at: Time.now, last_checked_at: Time.now, worker_image: project.image.name, freed_at: 1.minute.ago)
    end
    let!(:worker2) do
      FactoryBot.create(:worker, machine: FactoryBot.create(:machine), host_ip: '12.323.12.12', project_id: nil,
                                 worker_image: project.image.name)
    end
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 1, worker_image: project.image.name, supervisor_uid: supervisor.uid,
                             rebuild_hash: 'rebuild')
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'prioritizes workers with our project id already' do
      expect(subject.workers.size).to eq 1
      expect(subject.workers[0].id).to eq worker.id
      puts(subject.inspect)
    end
  end

  describe 'clear_workers_for_project' do
    subject { run_rpc(:ClearWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) do
      FactoryBot.create(:worker, project_id: project.id, state: :assigned, build_commands_run_at: Time.now)
    end
    let!(:worker2) { FactoryBot.create(:worker, host_ip: '12.323.12.12', project_id: nil) }

    let(:request_proto) { Api::ClearWorkersReq.new }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'clears all workers for a project' do
      expect(project.reload.workers.assigned.size).to eq(1)

      expect(subject.is_a?(Api::ClearWorkersResp)).to eq(true)

      expect(project.reload.workers.assigned.size).to eq(0)
    end
  end

  # describe '.get_workers_for_project' do
  #   let(:project) { FactoryBot.create(:project) }
  #   let(:request_proto) { Api::GetWorkersReq.new }
  #   let(:metadata) do
  #     { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
  #                            api_key: project.users.first.credential.api_key }.to_json }
  #   end

  #   subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: metadata }) }

  #   it 'will return the project' do
  #     expect { subject }.to raise_rpc_error(GRPC::NotFound)
  #   end
  # end

  # describe 'get workers for project' do
  #   let(:project) { FactoryBot.create(:project) }
  #   let(:request_proto) { Api::GetWorkersReq.new }
  #   let(:metadata) do
  #     { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
  #                            api_key: project.users.first.credential.api_key }.to_json }
  #   end

  #   subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: metadata }) }

  #   it 'will error' do
  #     expect { subject }.to raise_rpc_error(GRPC::NotFound)
  #   end
  # end

  describe 'get super for project' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:sup) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token) }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      expect(Supervisor.find(super_id).project_id).to eq project.id
    end
  end

  describe 'get super for project' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:sup) { FactoryBot.create(:supervisor) }
    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token) }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      expect(Supervisor.find(super_id).project_id).to eq project.id
    end
  end

  describe 'finish_run releases the workers' do
    subject do
      worker.freed_at = nil
      worker.save!
      run_rpc(:FinishRun, request_proto, active_call_options: { metadata: })
    end

    let(:project) { FactoryBot.create(:project) }

    let!(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let!(:worker) { FactoryBot.create(:worker) }
    let!(:jobrun) { FactoryBot.create(:jobrun, supervisor:) }
    let!(:wris) do
      Api::RunInfo.new(worker_id: worker.id, rebuild_hash: 'rebuild', exit_code: '0', output: '', finished_at: Time.now,
                       started_at: Time.now, error: 'no')
    end
    let(:request_proto) do
      Api::FinishRunRequest.new(exit_code: 0, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid, worker_run_info: [wris])
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    # it "will update the jobrun" do
    #   expect(subject.state).to eq "completed"
    # end

    it 'updates the set the finished time' do
      subject
      expect(jobrun.reload.finished_at).not_to be_nil
    end

    it 'adds the worker_info' do
      subject
      expect(worker.reload.worker_run_infos.size == 1)
    end

    it 'does not release the worker' do
      subject
      expect(worker.reload.freed_at).to be_nil
    end
  end

  describe 'finish_run records the files' do
    subject { run_rpc(:FinishRun, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }

    let!(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let!(:worker) { FactoryBot.create(:worker) }
    let!(:jobrun) { FactoryBot.create(:jobrun, supervisor:) }
    let!(:wris) do
      Api::RunInfo.new(files: ['one', 'two.ts'], worker_id: worker.id, rebuild_hash: 'rebuild', exit_code: '0', output: '',
                       finished_at: Time.now, started_at: Time.now, error: 'no')
    end
    let(:request_proto) do
      Api::FinishRunRequest.new(exit_code: 0, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid, worker_run_info: [wris])
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    # it "add the files" do
    #   expect(subject.state).to eq "completed"
    #   expect(worker.reload.worker_run_infos.size == 1)
    #   # expect(worker.reload.worker_run_infos.first.test_files.size == 2)
    #   # expect(worker.reload.worker_run_infos.first.test_files.second.filename).to eq "two.ts"
    # end
  end

  describe 'finish_run' do
    subject { run_rpc(:FinishRun, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) { FactoryBot.create(:worker) }

    let!(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let!(:jobrun) { FactoryBot.create(:jobrun, supervisor:) }
    let(:request_proto) do
      Api::FinishRunRequest.new(exit_code: 0, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    # it "will update the jobrun" do
    #   expect(subject.state).to eq "completed"
    # end

    it 'updates the set the finished time' do
      subject
      expect(jobrun.reload.finished_at).not_to be_nil
    end
  end

  describe 'finish_run2' do
    subject { run_rpc(:FinishRun, request_proto, active_call_options: { metadata: }) }

    let(:worker) { FactoryBot.create(:worker) }
    let(:project) { FactoryBot.create(:project) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:jobrun) { create(:jobrun, project:, supervisor:) }

    let(:request_proto) do
      Api::FinishRunRequest.new(status: 'completed', exit_code: 0, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: jobrun.project.project_token, api_token: jobrun.project.users.first.credential.api_token,
                             api_key: jobrun.project.users.first.credential.api_key }.to_json }
    end

    # it "will update the jobrun" do
    #   expect(subject.state).to eq "completed"
    # end

    it 'updates the set the finished time' do
      subject
      expect(jobrun.reload.finished_at).not_to be_nil
    end
  end

  describe 'finish_run releases workers' do
    subject { run_rpc(:FinishRun, request_proto, active_call_options: { metadata: }) }

    let(:worker) { FactoryBot.create(:worker) }
    let(:worker2) { FactoryBot.create(:worker) }
    let(:project) { FactoryBot.create(:project) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:jobrun) { create(:jobrun, project:, supervisor:) }

    let(:request_proto) do
      Api::FinishRunRequest.new(status: 'completed', exit_code: 0, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: jobrun.project.project_token, api_token: jobrun.project.users.first.credential.api_token,
                             api_key: jobrun.project.users.first.credential.api_key }.to_json }
    end

    # it "will update the jobrun" do
    #   expect(subject.state).to eq "completed"
    # end

    it 'updates the set the finished time' do
      subject
      expect(jobrun.reload.finished_at).not_to be_nil
    end
  end

  describe 'error_run' do
    subject { run_rpc(:FinishRun, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) { FactoryBot.create(:worker) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let!(:jobrun) { FactoryBot.create(:jobrun, project:, supervisor:) }

    let(:request_proto) do
      Api::FinishRunRequest.new(status: 'failed', exit_code: 1, output: 'Finished all good', jobrun_id: jobrun.id,
                                supervisor_uid: jobrun.supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'updates the jobrun' do
      expect(subject.state).to eq 'failed'
    end

    it 'updates the set the finished time' do
      subject
      expect(jobrun.reload.finished_at).not_to be_nil
    end

    it 'updates the set the exit code' do
      subject
      expect(jobrun.reload.exit_code).to eq 1
    end
  end

  describe 'list_projects' do
    subject { run_rpc(:GetAllProjects, request_proto, active_call_options: { metadata: }) }

    let(:user) { FactoryBot.create(:user) }
    let(:org) { FactoryBot.create(:org, owner: user) }

    let(:request_proto) { Api::GetAllProjectsReq.new }
    let(:metadata) do
      { 'authorization' => { api_token: user.credential.api_token,
                             api_key: user.credential.api_key }.to_json }
    end

    it 'lists no projects' do
      expect(subject.projects).to eq([])
    end

    it 'lists all projects' do
      FactoryBot.create(:project, name: 'test1', org:)
      FactoryBot.create(:project, name: 'test2', org:)

      expect(subject.projects.size).to eq(2)
    end
  end

  describe 'auth_fails_for_wrong_key' do
    subject { run_rpc(:GetAllProjects, request_proto, active_call_options: { metadata: }) }

    let(:user) { FactoryBot.create(:user) }
    let(:request_proto) { Api::GetAllProjectsReq.new }
    let(:metadata) do
      { 'authorization' => { api_token: user.credential.api_token,
                             api_key: user.credential.api_key + '2' }.to_json }
    end

    it 'fails auth' do
      expect { subject }.to raise_error(GRPC::Unauthenticated)
    end
  end

  describe 'get super for project with affinity' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:sup) { FactoryBot.create(:supervisor, state: :assigned, project:) }
    let!(:sup2) { FactoryBot.create(:supervisor, state: :assigned, project:, affinity: 'SOME_TAG') }
    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      expect(super_id).to eq sup2.id

      expect(Supervisor.find(super_id).project_id).to eq project.id
      expect(Supervisor.find(super_id).affinity).to eq 'SOME_TAG'
    end
  end

  describe "get super for project with affinity order doesn't matter" do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:sup) { FactoryBot.create(:supervisor, state: :assigned, project:, affinity: 'SOME_TAG') }
    let!(:sup2) { FactoryBot.create(:supervisor, state: :assigned, project:) }
    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      expect(super_id).to eq sup.id
      expect(Supervisor.find(super_id).project_id).to eq project.id
      expect(Supervisor.find(super_id).state).to eq 'assigned'
      expect(Supervisor.find(super_id).affinity).to eq 'SOME_TAG'
    end
  end

  describe 'no super with affinity' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:sup) { FactoryBot.create(:supervisor, project:) }
    let!(:sup2) { FactoryBot.create(:supervisor, project:) }
    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      expect(Supervisor.find(super_id).project_id).to eq project.id
    end
  end

  describe 'no super with affinity, no free super,at max' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, max_supervisors: 2) }
    let!(:sup) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }
    let!(:sup2) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }

    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      s = Supervisor.find(super_id)
      expect(s.project_id).to eq project.id
      expect([sup.id, sup2.id]).to include(s.id)
    end
  end

  describe 'no super with affinity, no free super,not at max,no supers available ' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, max_supervisors: 3) }
    let!(:sup) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }
    let!(:sup2) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }

    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      s = Supervisor.find(super_id)
      expect(s.project_id).to eq project.id
      expect([sup.id, sup2.id]).to include(s.id)
    end
  end

  describe 'no super with affinity, no free super,not at max,with supers available ' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, max_supervisors: 3) }
    let!(:sup) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }
    let!(:sup2) { FactoryBot.create(:supervisor, project:, state: :assigned, in_use: 1.minute.ago) }
    let!(:sup3) { FactoryBot.create(:supervisor) }

    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the super' do
      expect(subject).to be_a(Api::GetSuperResp)
      puts(subject.inspect)

      super_id = subject.super.id
      s = Supervisor.find(super_id)
      expect(s.project_id).to eq project.id
      expect([sup.id, sup2.id]).not_to include(s.id)
    end
  end

  describe 'no supers assigned, none available ' do
    subject { run_rpc(:GetSuperForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project, max_supervisors: 3) }
    let(:project2) { FactoryBot.create(:project, max_supervisors: 3) }

    let!(:sup) { FactoryBot.create(:supervisor, project: project2, state: :assigned, in_use: 1.minute.ago) }
    let!(:sup2) { FactoryBot.create(:supervisor, project: project2, state: :assigned, in_use: 1.minute.ago) }

    let(:request_proto) { Api::GetSuperReq.new(project_token: project.project_token, affinity: 'SOME_TAG') }
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'raise resource exhausted' do
      expect { subject }.to raise_error GRPC::ResourceExhausted
    end
  end

  describe 'get_workers_for_project with two supers' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }

    let!(:worker1) { FactoryBot.create(:worker) }
    let!(:worker2) { FactoryBot.create(:worker) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }

    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 1, worker_image: 'rails', supervisor_uid: supervisor.uid)
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the workers' do
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id
      expect(subject.jobrun_id).not_to be_nil
      subject.workers.map { |w| expect([worker1.id, worker2.id]).to include(w.id) }
    end
  end

  describe 'get_workers_for_project with two supers, three workers and 2 concurrency' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'),
                                  max_supervisors: 3)
    end

    let!(:worker1) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild2', project:, freed_at: nil, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, worker_image: project.image.name)
    end
    let!(:worker2) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild2', project:, freed_at: nil, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, worker_image: project.image.name)
    end
    let!(:worker3) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild', project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, worker_image: project.image.name)
    end
    let!(:worker4) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild', project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, worker_image: project.image.name)
    end
    let!(:worker5) do
      FactoryBot.create(:worker, rebuild_hash: 'rebuild', project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, worker_image: project.image.name)
    end
    let!(:worker6) { FactoryBot.create(:worker) }

    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker1, worker2]) }
    let(:supervisor2) { FactoryBot.create(:supervisor, project:) }

    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor2.uid,
                             rebuild_hash: 'rebuild')
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'sets max_workers correctly' do
      expect(project.max_workers).to eq((3 * project.worker_concurrency * 1.1 * 2).ceil)
    end

    it 'returns the workers' do
      puts '--------------------------------------------'
      puts supervisor1.workers.inspect
      puts '^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^'
      puts supervisor2.workers.inspect
      puts '^^^^^^^'
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id

      expect(subject.jobrun_id).not_to be_nil
      expect(subject.workers.length).to eq 2
      subject.workers.uniq.map { |w| expect([worker3.id, worker4.id, worker5.id]).to include(w.id) }
    end
  end

  describe 'get_workers_for_project with two supers, three workers and 2 concurrency and rebuild hash' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'),
                                  max_supervisors: 3)
    end

    let!(:worker1) do
      FactoryBot.create(:worker, project:, state: :assigned, freed_at: nil, build_commands_run_at: nil,
                                 last_checked_at: 1.second.ago)
    end
    let!(:worker2) do
      FactoryBot.create(:worker, project:, state: :assigned, freed_at: nil, build_commands_run_at: nil,
                                 last_checked_at: 1.second.ago)
    end
    let!(:worker3) do
      FactoryBot.create(:worker, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago)
    end
    let!(:worker4) do
      FactoryBot.create(:worker, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value')
    end
    let!(:worker5) do
      FactoryBot.create(:worker, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value')
    end
    let!(:worker6) { FactoryBot.create(:worker) }

    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker1, worker2]) }
    let(:supervisor2) { FactoryBot.create(:supervisor, project:) }

    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor2.uid,
                             rebuild_hash: 'my-hash-value')
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the workers' do
      puts '--------------------------------------------'
      puts supervisor1.workers.inspect
      puts '^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^'
      puts supervisor2.workers.inspect
      puts '^^^^^^^'
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id

      expect(subject.jobrun_id).not_to be_nil
      expect(subject.workers.length).to eq 2
      subject.workers.uniq.map { |w| expect([worker3.id, worker4.id, worker5.id]).to include(w.id) }
    end
  end

  describe 'get_workers_for_project with two supers, three workers and 2 concurrency and one rebuild hash that are not the same' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'),
                                  max_supervisors: 3)
    end

    let!(:worker1) do
      FactoryBot.create(:worker, project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value', machine: FactoryBot.create(:machine, cpus: 0))
    end
    let!(:worker2) do
      FactoryBot.create(:worker, project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value', machine: FactoryBot.create(:machine, cpus: 0))
    end
    let!(:worker3) do
      FactoryBot.create(:worker, project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value', machine: FactoryBot.create(:machine))
    end
    let!(:worker4) do
      FactoryBot.create(:worker, project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value', machine: FactoryBot.create(:machine))
    end
    let!(:worker5) do
      FactoryBot.create(:worker, project:, freed_at: 1.minute.ago, state: :assigned,
                                 build_commands_run_at: 1.minute.ago, last_checked_at: 1.second.ago, rebuild_hash: 'my-hash-value', machine: FactoryBot.create(:machine))
    end
    let!(:worker6) { FactoryBot.create(:worker) }

    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker1, worker2]) }
    let(:supervisor2) { FactoryBot.create(:supervisor, project:) }

    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor2.uid,
                             rebuild_hash: 'my-hash-value')
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the workers' do
      puts '--------------------------------------------'
      puts supervisor1.workers.inspect
      puts '^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^'
      puts supervisor2.workers.inspect
      puts '^^^^^^^'
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id

      expect(subject.jobrun_id).not_to be_nil
      expect(subject.workers.length).to eq 2
      subject.workers.uniq.map { |w| expect([worker3.id, worker4.id, worker5.id]).to include(w.id) }
    end
  end

  describe 'get_workers_for_project unassigned with two supers, three workers and 2 concurrency and one rebuild hash that are not the same' do
    subject do
      run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: })
    end

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'),
                                  max_supervisors: 3)
    end

    let!(:worker1) do
      FactoryBot.create(:busy_worker, project_id: project.id, state: :assigned, build_commands_run_at: 1.minute.ago,
                                      last_checked_at: 1.second.ago)
    end
    let!(:free_worker) do
      worker1.update_attribute(:freed_at, nil)
    end
    let!(:worker2) do
      FactoryBot.create(:busy_worker, project_id: project.id, state: :assigned, build_commands_run_at: 1.minute.ago,
                                      last_checked_at: 1.second.ago)
    end
    let!(:free_worker2) do
      worker2.update_attribute(:freed_at, nil)
    end
    let!(:worker3) { FactoryBot.create(:worker, last_checked_at: 1.second.ago) }
    let!(:worker4) { FactoryBot.create(:worker, last_checked_at: 1.second.ago) }
    let!(:worker5) { FactoryBot.create(:worker, last_checked_at: 1.second.ago) }
    let!(:worker6) { FactoryBot.create(:worker, last_checked_at: 1.second.ago) }

    let(:supervisor2) { FactoryBot.create(:supervisor, project:) }

    let(:request_proto) do
      Api::GetWorkersReq.new(num_workers: 2, worker_image: 'rails', supervisor_uid: supervisor2.uid,
                             rebuild_hash: 'my-hash-value')
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'returns the workers' do
      expect(project.workers.busy.size).to eq 2
      expect(project.workers.free_workers.count).to eq 0
      expect(project.workers.free_workers).not_to include(worker1)
      expect(project.workers.free_workers).not_to include(worker2)

      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id

      expect(subject.jobrun_id).not_to be_nil
      expect(subject.workers.length).to eq 2
      puts(subject.workers.map { |w| w.inspect })
      subject.workers.uniq.map { |w| expect([worker3.id, worker4.id, worker5.id, worker6.id]).to include(w.id) }
    end
  end

  # test log_and_release
  describe 'log_run' do
    subject do
      worker.freed_at = nil
      run_rpc(:LogRun, request_proto, active_call_options: { metadata: })
    end

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'))
    end

    let(:worker) do
      FactoryBot.create(:worker, freed_at: nil, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago)
    end
    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker]) }
    let(:jobrun) { FactoryBot.create(:jobrun, supervisor: supervisor1) }
    let(:command) { Api::Command.new({ commandline: 'echo', args: ['hello'], environment: { foo: 'Bar' } }) }
    let(:execution_infos) do
      [Api::ExecutionInfo.new(rebuild_hash: '', command:, exit_code: 0, started: Time.now, finished: Time.now)]
    end
    let(:request_proto) do
      Api::LogRunReq.new({ command: Api::Command.new(noTestFiles: false),
                           worker_run_info: Api::RunInfo.new(worker_id: worker.id, exit_code: '0', jobrun_id: jobrun.id, started_at: Time.now,
                                                             finished_at: Time.now, execution_infos:) })
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'logs the run' do
      expect(subject).to be_a(Api::LogRunResp)
      expect(subject.success).to eq true
    end

    it 'creates the worker_run_info' do
      subject
      expect(worker.reload.worker_run_infos.count).to eq 1
      expect(worker.reload.worker_run_infos.first.exit_code).to eq '0'
    end

    it 'creates the execution_infos' do
      subject
      expect(worker.reload.worker_run_infos.count).to eq 1
      expect(worker.reload.worker_run_infos.first.execution_infos.last).not_to be_nil
      expect(worker.reload.worker_run_infos.first.execution_infos.last.exit_code).to eq 0
    end

    it 'sets the worker to freed' do
      subject
      expect(worker.reload.freed_at).not_to be_nil
    end

    it 'does not set the jobrun to finished if not all finished' do
      subject
      expect(jobrun.reload.finished_at).to be_nil
    end

    it 'sets the reserved at to nil' do
      subject
      expect(worker.reload.reserved_at).to be_nil
    end
  end

  describe 'log_run_with_no_test_files' do
    subject do
      worker.freed_at = nil
      run_rpc(:LogRun, request_proto, active_call_options: { metadata: })
    end

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'))
    end

    let(:worker) do
      FactoryBot.create(:worker, freed_at: nil, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago)
    end
    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker]) }
    let(:jobrun) { FactoryBot.create(:jobrun, supervisor: supervisor1) }
    let(:command) do
      Api::Command.new({ noTestFiles: true, commandline: 'echo', args: ['hello'], environment: { foo: 'Bar' } })
    end
    let(:execution_infos) do
      [Api::ExecutionInfo.new(rebuild_hash: '', command:, exit_code: 0, started: Time.now, finished: Time.now)]
    end
    let(:request_proto) do
      Api::LogRunReq.new({ command:,
                           worker_run_info: Api::RunInfo.new(worker_id: worker.id, exit_code: '0', jobrun_id: jobrun.id, started_at: Time.now,
                                                             finished_at: Time.now, execution_infos:) })
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'logs the run' do
      expect(subject).to be_a(Api::LogRunResp)
      expect(subject.success).to eq true
    end
  end

  describe 'log_run with test files' do
    subject do
      worker.freed_at = nil
      run_rpc(:LogRun, request_proto, active_call_options: { metadata: })
    end

    let(:project) do
      FactoryBot.create(:project, worker_concurrency: 2, image: FactoryBot.create(:image, name: 'rails'))
    end

    let(:worker) do
      FactoryBot.create(:worker, freed_at: nil, project:, state: :assigned, build_commands_run_at: 1.minute.ago,
                                 last_checked_at: 1.second.ago)
    end
    let(:supervisor1) { FactoryBot.create(:supervisor, project:, workers: [worker]) }
    let(:jobrun) { FactoryBot.create(:jobrun, supervisor: supervisor1) }

    let(:request_proto) do
      Api::LogRunReq.new({ command: Api::Command.new({}), worker_run_info: Api::RunInfo.new(files: ['one', 'two.ts'],
                                                                                            worker_id: worker.id, exit_code: '0', jobrun_id: jobrun.id, started_at: Time.now, finished_at: Time.now,
                                                                                            execution_infos: [Api::ExecutionInfo.new({ started: Google::Protobuf::Timestamp.new, finished: Google::Protobuf::Timestamp.new, command: Api::Command.new({}) })]) })
    end
    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'logs the run' do
      expect(subject).to be_a(Api::LogRunResp)
      expect(subject.success).to eq true
    end

    it 'creates the worker_run_info' do
      subject
      expect(worker.reload.worker_run_infos.count).to eq 1
      expect(worker.reload.worker_run_infos.first.exit_code).to eq '0'
    end

    it 'creates the test files' do
      subject
      expect(worker.reload.worker_run_infos.first.test_files.count).to eq 2
      expect(worker.reload.worker_run_infos.first.test_files.first.filename).to eq 'one'
    end
  end

  describe 'get_workers_for_project with repo_info' do
    subject { run_rpc(:GetWorkersForProject, request_proto, active_call_options: { metadata: }) }

    let(:project) { FactoryBot.create(:project) }
    let!(:worker) { FactoryBot.create(:worker) }
    let(:supervisor) { FactoryBot.create(:supervisor, project:) }
    let(:repo_info) do
      Api::RepoInfo.new(CommitHash: '123', Branch: 'master', Repo: 'repo', CommitMessage: 'i did it', CommitAuthor: 'sean',
                        CommitAuthorEmail: 'sean@email.com', IsGitRepo: true)
    end
    let(:request_proto) do
      Api::GetWorkersReq.new(repo_info:, num_workers: 1, worker_image: 'rails', supervisor_uid: supervisor.uid)
    end

    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    it 'sets the repo_info' do
      expect(subject).to be_a(Api::GetWorkersResp)
      puts(subject.inspect)

      worker_id = subject.workers[0].id
      expect(Worker.find(worker_id).project_id).to eq project.id

      jobrun = Jobrun.find(subject.jobrun_id)
      expect(jobrun.repo_info.commit_hash).to eq '123'
    end
  end

  describe '#deregister_workers' do
    subject { run_rpc(:DeRegisterWorkers, request_proto, active_call_options: { metadata: }) }

    let(:project) { create(:project) }
    let(:w1) { create(:worker, project:) }
    let(:w2) { create(:worker, project:) }
    let(:worker1) { Api::Worker.new({ id: w1.id }) }
    let(:worker2) { Api::Worker.new({ id: w2.id }) }

    let(:request_proto) { Api::DeRegisterWorkersReq.new(workers: [worker1, worker2]) }

    let(:metadata) do
      { 'authorization' => { project_token: project.project_token, api_token: project.users.first.credential.api_token,
                             api_key: project.users.first.credential.api_key }.to_json }
    end

    context 'when workers list is empty' do
      let(:request_proto) { Api::DeRegisterWorkersReq.new(workers: []) }

      it 'raises an exception' do
        expect { subject }.to raise_error(GRPC::Internal)
      end
    end

    context 'when project and workers are present' do
      it 'de-registers each worker and returns a response with status :ok' do
        expect(subject).to be_an_instance_of(Api::DeRegisterWorkersResp)
        expect(subject.status).to eq('ok')
      end
    end
  end
end
