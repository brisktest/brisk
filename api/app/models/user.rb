# frozen_string_literal: true

require 'open-uri'

class User < ApplicationRecord
  # include SimpleDiscussion::ForumUser
  paginates_per 30
  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :trackable and :omniauthable
  devise :database_authenticatable, :registerable,
         :recoverable, :rememberable, :validatable, :confirmable, :omniauthable, omniauth_providers: %i[github google_oauth2]
  has_one :credential, as: :credentialable
  has_one :account
  has_many :memberships
  has_many :orgs, lambda {
                    where.not('memberships.accepted_at' => nil).where('memberships.cancelled_at' => nil)
                  }, through: 'memberships', class_name: 'Org', source: 'org'

  has_many :projects, through: :orgs

  has_many :jobruns, through: :projects

  has_many :owned_orgs, foreign_key: :owner_id, class_name: 'Org'

  after_create :generate_credentials

  # after_create :generate_default_org
  after_create_commit :queue_for_add_to_sendgrid

  has_many :subscriptions, through: :projects

  before_validation :generate_default_account, on: :create

  after_commit :send_welcome_email, on: :create

  has_one_attached :profile_image

  validate :acceptable_profile_image

  def self.from_omniauth(auth)
    # if the user has the same email as one of our users, just return the user
    # cause it's very difficult to spoof this and if you could you could just
    # spoof the rest of the auth info
    Rails.logger.debug "Auth: #{auth.inspect}"
    if auth.info.email.present?
      u = where(email: auth.info.email).first
      if u
        begin
          if !u.profile_image.attached? && auth.info.image.present?
            u.import_image_from_url(auth.info.image)
            u.save!
          end
        rescue StandardError => e
          Rails.logger.info "Error importing image from url: #{e.message}"
        end
        return u
      end
    end

    u = where(provider: auth.provider, uid: auth.uid).first_or_create! do |user|
      user.email = auth.info.email
      user.import_image_from_url(auth.info.image)
      user.password = Devise.friendly_token[0, 20]
      user.skip_confirmation!
      user.name = auth.info.name
    end

    begin
      if !u.profile_image.attached? && auth.info.image.present?
        u.import_image_from_url(auth.info.image)
        u.save!
      end
    rescue StandardError => e
      Rails.logger.info "Error importing image from url: #{e.message}"
    end
    u
  end

  def import_image_from_url(url)
    profile_image.attach(io: URI.parse(url).open, filename: "profile_image-#{id.to_s.hash}.jpg")
  end

  def acceptable_profile_image
    return unless profile_image.attached?

    errors.add(:profile_image, 'is too big') unless profile_image.blob.byte_size <= 1.megabyte
    acceptable_types = ['image/jpeg', 'image/png', 'image/jpg', 'image/gif', 'image/webp']
    return if acceptable_types.include?(profile_image.content_type)

    errors.add(:profile_image, 'must be a valid image type')
  end

  def authorized_projects
    projects
  end

  def fullname
    name
  end

  def to_user_for_js
    {
      id:,
      name:,
      email:,
      imageUrl: hasProfilePic? ? profile_image.url : gravatar_image_url,
      hasProfilePic: hasProfilePic?
    }
  end

  def hasProfilePic?
    profile_image.attached?
  end

  def gravatar_image_url
    "https://www.gravatar.com/avatar/#{Digest::MD5.hexdigest(email)}?s=200"
  end

  def send_welcome_email
    AdminMailer.new_user(id).deliver_later
  end

  def generate_default_account
    build_account if account.nil?
  end

  def generate_credentials
    self.credential = Credential.new
    credential.save!
  end

  def generate_default_org
    owned_orgs.create! name: "default#{SecureRandom.hex(9)}"
  end

  # def get_default_org
  #   generate_default_org || owned_orgs.first
  # end

  def is_admin?
    admin
  end

  def queue_for_add_to_sendgrid
    AddToSendgridJob.perform_later(id)
  rescue StandardError => e
    Rails.logger.error("Error in queue_for_add_to_sendgrid #{e} - proceeding")
    Sentry.capture_exception(e)
  end

  def add_to_sendgrid
    sg = SendGrid::API.new(api_key: ENV['SENDGRID_API_KEY'])

    data = { contacts: [{ email: }] }.to_json

    response = sg.client.marketing.contacts.put(request_body: data)
    puts response.body
    Rails.logger.error "Error adding #{email} to sendgrid: #{response.body}" if response.status_code.to_i > 299
    response.status_code.to_i < 300
  end

  def can_manage_org?(org)
    org.is_manager?(self)
  end
end
