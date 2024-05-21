# frozen_string_literal: true

class MigrateUserData < ActiveRecord::Migration[6.1]
  def change
    Credential.all.each do |c|
      u = User.find c.user_id
      c.credentialable = u
      c.save!
    end
  end
end
