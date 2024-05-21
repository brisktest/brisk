class AddUserAsMemberOfOrgs < ActiveRecord::Migration[7.0]
  def change
    Org.all.each do |org|
      next if org.memberships.where(user_id: org.owner_id).any?

      mem = org.memberships.create(user_id: org.owner_id, role: 'admin', accepted_at: Time.now)
      next if mem.save

      puts "Error: #{mem.errors.full_messages}"
      puts "User: #{org.owner.inspect}"
      puts "Membership: #{mem.inspect}"
      raise 'Error saving membership'
    end
  end
end
