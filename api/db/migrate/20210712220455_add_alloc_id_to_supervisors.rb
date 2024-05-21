# frozen_string_literal: true

class AddAllocIdToSupervisors < ActiveRecord::Migration[6.1]
  def change
    add_column :supervisors, :uid, :text
  end
end
