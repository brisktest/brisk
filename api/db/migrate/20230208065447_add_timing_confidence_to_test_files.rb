class AddTimingConfidenceToTestFiles < ActiveRecord::Migration[7.0]
  def change
    add_column :test_files, :timing_confidence, :float
  end
end
