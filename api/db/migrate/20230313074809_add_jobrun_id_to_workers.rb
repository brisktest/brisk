class AddJobrunIdToWorkers < ActiveRecord::Migration[7.0]
  def change
    add_column :workers, :jobrun_id, :integer
  end
end
