# frozen_string_literal: true

class AddExternalEndpointToSupervisors < ActiveRecord::Migration[6.1]
  def change
    add_column :supervisors, :external_endpoint, :text
  end
end
