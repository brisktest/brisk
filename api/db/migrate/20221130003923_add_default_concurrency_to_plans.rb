class AddDefaultConcurrencyToPlans < ActiveRecord::Migration[7.0]
  def change
    Plan.where(monthly_concurrency: 0).update_all monthly_concurrency: 1200
    Plan.where(name: 'Developer').update_all monthly_concurrency: 1200
    Plan.where(name: 'Team').update_all monthly_concurrency: 3100
    Plan.where(name: 'Enterprise').update_all monthly_concurrency: 29_900
  end
end
