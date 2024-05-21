class Image < ApplicationRecord
  has_many :workers
  has_many :projects, through: :workers
  has_many :users, through: :projects

  validates :name, presence: true
  validates :version, presence: true
end
