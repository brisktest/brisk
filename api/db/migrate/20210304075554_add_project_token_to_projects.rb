# frozen_string_literal: true

class AddProjectTokenToProjects < ActiveRecord::Migration[6.0]
  def change
    add_column :projects, :project_token, :text
  end
end
