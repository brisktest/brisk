require 'rails_helper'

RSpec.describe TimingAnalysisJob, type: :job do
  let(:worker_run_info) { create(:worker_run_info, exit_code: '0') }
  let(:jobrun) { worker_run_info.jobrun }
  let(:machine) { worker_run_info.worker.machine }

  describe '#perform' do
    context 'when worker run has succeeded and no more still running workers' do
      before do
        jobrun.state = 'completed'
        jobrun.save!
        allow_any_instance_of(Machine).to receive(:no_still_running_workers_from).and_return(true)
        allow_any_instance_of(WorkerRunInfo).to receive(:compare_to_previous_run)
        allow_any_instance_of(Jobrun).to receive_message_chain(:worker_run_infos, :all?).and_return(true)
      end

      it 'compares timing to previous run' do
        expect_any_instance_of(WorkerRunInfo).to receive(:compare_to_previous_run)
        described_class.perform_now(worker_run_info.id)
      end

      it 'marks timing as processed' do
        described_class.perform_now(worker_run_info.id)
        expect(!!worker_run_info.reload.timing_processed).to be(true)
      end

      it 'queues PreSplitJob if all worker run infos have processed timing' do
        expect(PreSplitJob).to receive(:perform_later).with(jobrun.id)
        described_class.perform_now(worker_run_info.id)
      end
    end
  end
end
