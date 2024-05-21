# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :authenticate_user!, except: %i[index health demos pricing download]
  newrelic_ignore only: [:health]
  # caches_action :index, :demos, :pricing, unless: -> { current_user }, expires_in: 1.hour.from_now

  def index; end

  def download; end

  def demos; end

  def pricing
    @free_plan = Plan.find_by_name 'Developer'
    @developer_plan = Plan.find_by_name 'Team'
  end

  def health
    Project.count
    # head :ok
  end

  def beacons_case_study
    send_file(
      "#{Rails.root.join('public/beacons-x-brisk-case-study.pdf')}",
      filename: 'beacons-x-brisk-case-study.pdf',
      type: 'application/pdf',
      disposition: 'inline'
    )
  end
end
