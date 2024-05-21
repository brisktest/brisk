class AddDefaultValuesForOldJobruns < ActiveRecord::Migration[7.0]
  def change
    Jobrun.all.each do |jr|
      jr.set_uid
      jr.save(validate: false)
    end
  end
end
