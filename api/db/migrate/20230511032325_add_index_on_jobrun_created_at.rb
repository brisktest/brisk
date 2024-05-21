class AddIndexOnJobrunCreatedAt < ActiveRecord::Migration[7.0]
  def change
    add_index :jobruns, :created_at
  end
end
