# frozen_string_literal: true

class AddUsernameToProjects < ActiveRecord::Migration[6.0]
  def change
    add_column :projects, :username, :string
  end
end
