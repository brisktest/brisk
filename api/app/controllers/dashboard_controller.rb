# frozen_string_literal: true

class DashboardController < ApplicationController
  before_action :authenticate_user!, except: %i[getting_started]
  before_action :set_user

  def show; end

  def getting_started; end

  def config_file
    # @project = current_user.projects.find_by_project_token params[:project_token]

    # @json = { listTestCommand: '', commands: { commandline: '', args: '' }, buildCommands: { commandLine: '', args: '' },
    #           environment: '', excludedFromSync: '', excludedFromWatch: '', projectToken: @project.project_token,
    #           api_key: current_user.credential.api_key, apiToken: current_user.credential.api_token, apiEndpoint: 'api.brisktest.com:50052', framework: @project.framework, image: @project.image.name }

    # respond_to do |format|
    #   format.html {}
    #   format.json do
    #     send_data JSON.pretty_generate(@json), type: :json, disposition: 'attachment', filename: 'brisk.json'
    #   end
    # end
  end

  private

  def set_user
    @user = current_user
  end
end
