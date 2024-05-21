class AddDataToPlans < ActiveRecord::Migration[6.1]
  def change
    Plan.create! name: 'Developer', amount_cents: 0, trial_period: 'none',
                 description: 'Experience concurrency and unlimited 5X runs'

    Plan.create! name: 'Team', amount_cents: 12_900, trial_period: '14 days',
                 description: 'Perfect for a solo developer'
    Plan.create! name: 'Enterprise', amount_cents: 29_900, trial_period: '14 days',
                 description: 'Super charge your development team'
  end
end
