# frozen_string_literal: true

class AddSetupRunAtToSupervisors < ActiveRecord::Migration[6.1]
  def change
    add_column :supervisors, :setup_run_at, :datetime
  end
end
