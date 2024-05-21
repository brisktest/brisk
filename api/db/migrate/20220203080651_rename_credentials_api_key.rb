class RenameCredentialsApiKey < ActiveRecord::Migration[7.0]
  def change
    rename_column :credentials, :api_key_ciphertext, :api_key
  end
end
