class AddTraceApiVersionToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :api_version, :text
  end
end
