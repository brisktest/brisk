# frozen_string_literal: true

class AddWorkerImageToWorkers < ActiveRecord::Migration[6.1]
  def change
    add_column :workers, :worker_image, :text
  end
end
