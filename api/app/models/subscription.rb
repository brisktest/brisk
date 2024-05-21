# a siubscription ties together a plan and an org
class Subscription < ApplicationRecord
  belongs_to :plan
  belongs_to :account
  has_one :org, through: :account
end
