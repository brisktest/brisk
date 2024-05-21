class MigratePlanNames < ActiveRecord::Migration[7.0]
  def change
    p = Plan.where(name: 'Developer').first
    if p
      p.name = 'Team'
      p.save!
    end
    p2 = Plan.where(name: 'Free').first
    return unless p2

    p2.name = 'Developer'
    p2.save!
  end
end
