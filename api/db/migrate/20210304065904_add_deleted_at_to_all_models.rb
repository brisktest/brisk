# frozen_string_literal: true

class AddDeletedAtToAllModels < ActiveRecord::Migration[6.0]
  def change
    Rails.application.eager_load!
    ApplicationRecord.descendants.each do |table|
      next unless table_exists?(table)

      unless column_exists?(table.table_name, :deleted_at)
        add_column table.table_name, :deleted_at, :datetime
        puts 'added a column?'
      end
    end
  end
end
