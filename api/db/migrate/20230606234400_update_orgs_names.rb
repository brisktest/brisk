class UpdateOrgsNames < ActiveRecord::Migration[7.0]
  def change
    Org.all.each do |org|
      next unless Org.where(name: org.name.downcase).size > 1

      org.name = org.name + org.id.to_s
      org.name = org.name.downcase
      org.save!
    end
  end
end
