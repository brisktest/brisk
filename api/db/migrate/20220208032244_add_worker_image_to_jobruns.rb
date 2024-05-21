class AddWorkerImageToJobruns < ActiveRecord::Migration[7.0]
  def change
    add_column :jobruns, :worker_image, :text
  end
end
