# frozen_string_literal: true

class CreateApiActions < ActiveRecord::Migration[6.1]
  def change
    create_table :api_actions do |t|
      t.text :name
      t.text :grpc_method_name
      t.timestamps
    end
  end
end
