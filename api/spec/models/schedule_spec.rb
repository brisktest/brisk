require 'rails_helper'

RSpec.describe Schedule, type: :model do
  pending "add some examples to (or delete) #{__FILE__}"

  describe 'default schedule' do
    it 'has a default schedule' do
      expect(Schedule.default_schedule).not_to be_nil
    end
  end
end
