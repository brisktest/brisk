# frozen_string_literal: true

class AddPolymorphicToCredentials < ActiveRecord::Migration[6.1]
  def change
    add_column :credentials, :credentialable_id, :int
    add_column :credentials, :credentialable_type, :string
  end
end
