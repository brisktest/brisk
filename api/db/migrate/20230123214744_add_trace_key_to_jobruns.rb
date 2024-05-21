class AddTraceKeyToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :trace_key, :text
  end
end
