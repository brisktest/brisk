# frozen_string_literal: true

class AddCiphertextForCredentials < ActiveRecord::Migration[6.0]
  def change
    add_column :credentials, :api_key_ciphertext, :text
  end
end
