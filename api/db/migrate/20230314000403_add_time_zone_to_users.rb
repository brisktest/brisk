class AddTimeZoneToUsers < ActiveRecord::Migration[7.0]
  def change
    add_column :users, :time_zone, :string, default: 'Pacific Time (US & Canada)'
  end
end
