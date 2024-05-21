class AddIndexOnWorkersLastCheckedAtAndState < ActiveRecord::Migration[7.0]
  def change
    add_index :workers, %i[last_checked_at state]
  end
end
