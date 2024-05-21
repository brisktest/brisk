# frozen_string_literal: true

class AddUidToWorkers < ActiveRecord::Migration[6.1]
  def change
    add_column :workers, :uid, :text
  end
end
