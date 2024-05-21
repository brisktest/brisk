class AddFreedAtToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :freed_at, :datetime
  end
end
