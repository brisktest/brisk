# frozen_string_literal: true

class AddLastCheckedAtToWorkers < ActiveRecord::Migration[6.1]
  def change
    add_column :workers, :last_checked_at, :datetime
  end
end
