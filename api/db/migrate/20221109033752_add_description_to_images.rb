class AddDescriptionToImages < ActiveRecord::Migration[7.0]
  def change
    add_column :images, :description, :text
  end
end
