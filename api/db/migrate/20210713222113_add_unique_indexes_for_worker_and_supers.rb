# frozen_string_literal: true

class AddUniqueIndexesForWorkerAndSupers < ActiveRecord::Migration[6.1]
  def change
    add_index :workers, :uid, unique: true, name: 'unque_index_worker_uid'
    add_index :supervisors, :uid, unique: true, name: 'unque_index_supervisors_uid'
  end
end
