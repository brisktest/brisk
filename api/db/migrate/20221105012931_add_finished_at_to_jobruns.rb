class AddFinishedAtToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :finished_at, :datetime
  end
end
