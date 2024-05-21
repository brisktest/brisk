# frozen_string_literal: true

class AddValidThroughToCredentials < ActiveRecord::Migration[6.1]
  def change
    add_column :credentials, :valid_through, :datetime
  end
end
