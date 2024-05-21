# frozen_string_literal: true

require 'sidekiq/web'
require 'sidekiq-scheduler/web'

Rails.application.routes.draw do
  # resources :machines
  # resources :workers
  case Rails.env
  when 'development'
    default_url_options({ host: 'brisktest.test' })
  when 'staging'
    default_url_options({ host: ENV['WEB_HOSTNAME'] || 'development-nomad.brisktesting.com' })
  when 'production'
    default_url_options({ host: 'brisktest.com' })
  end

  devise_for :users, controllers: { registrations: 'registrations', omniauth_callbacks: 'users/omniauth_callbacks' }

  # resources :credentials
  # resources :users
  resources :projects, param: :project_token
  get 'projects/:project_token/test_files', to: 'projects#show_files', as: :project_show_files

  get 'projects/:project_token/dashboard', to: 'projects#project_dashboard', as: :project_dashboard
  # resources :orgs
  root 'home#index'
  get 'demos', to: 'home#demos', as: :home_demos
  get 'pricing', to: 'home#pricing', as: :home_pricing
  get '/health', to: 'home#health'
  get '/dashboard', to: 'dashboard#show', as: :dashboard
  get '/getting_started', to: 'dashboard#getting_started', as: :getting_started
  # get '/dashboard/config_file/:project_token', to: 'dashboard#config_file', as: :config_file
  post '/join/:plan_id', to: 'subscriptions#join', as: :new_subscription
  get 'welcome', to: 'subscriptions#welcome', as: :subscriptions_welcome

  get '/api/cli/auth/:token', to: 'cli#auth', as: :cli_auth

  post '/api/cli/confirm_auth/:token', to: 'cli#do_confirm_auth', as: :cli_confirm_auth
  get '/api/cli/confirm_auth/:token', to: 'cli#confirm_auth', as: :confirm_auth
  # mount SimpleDiscussion::Engine => '/community'

  get 'project/:project_token/run/:jobrun_uid/', to: 'projects#dashboard_jobruns', as: :project_jobrun

  get 'project/:project_token/worker/:worker_run_info_uid', to: 'projects#dashboard_jobruns_with_logs',
                                                            as: :project_jobrun_with_logs

  get '/admin/analytics', to: 'admin/analytics#index', as: :admin_analytics

  get '/admin/projects/:id/jobruns', to: 'admin/jobruns#for_project', as: :admin_project_jobruns
  get '/admin/jobruns/for_week', to: 'admin/jobruns#for_week', as: :admin_jobruns_for_week
  get '/admin/jobruns/for_day', to: 'admin/jobruns#for_day', as: :admin_jobruns_for_day
  get '/admin/jobruns/for_running', to: 'admin/jobruns#for_running', as: :admin_jobruns_for_running
  get '/admin/jobruns/for_failed', to: 'admin/jobruns#for_failed', as: :admin_jobruns_for_failed
  get '/admin/jobruns/for_completed', to: 'admin/jobruns#for_completed', as: :admin_jobruns_for_completed
  get 'admin/jobruns/:id', to: 'admin/jobruns#show', as: :admin_jobrun

  get '/admin/projects', to: 'admin/projects#index', as: :admin_projects
  get '/admin/projects/daily_active_projects', to: 'admin/projects#daily_active_projects',
                                               as: :admin_daily_active_projects
  get '/admin/projects/weekly_active_projects', to: 'admin/projects#weekly_active_projects',
                                                as: :admin_weekly_active_projects
  get '/admin/projects/new_daily_projects', to: 'admin/projects#new_daily_projects', as: :admin_new_daily_projects
  get '/admin/projects/new_weekly_projects', to: 'admin/projects#new_weekly_projects', as: :admin_new_weekly_projects

  get '/admin/users', to: 'admin/users#index', as: :admin_users
  get '/admin/users/daily_active_users', to: 'admin/users#daily_active_users', as: :admin_daily_active_users
  get '/admin/users/weekly_active_users', to: 'admin/users#weekly_active_users', as: :admin_weekly_active_users
  get '/admin/users/daily_new_users', to: 'admin/users#daily_new_users', as: :admin_new_daily_users
  get '/admin/users/weekly_new_users', to: 'admin/users#weekly_new_users', as: :admin_new_weekly_users

  get '/download', to: 'home#download', as: :download
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html

  get '/admin/worker_details/:image_id', to: 'admin/workers#worker_details', as: :admin_worker_details

  get 'beacons_case_study', to: 'home#beacons_case_study', as: :beacons_case_study
  resources :orgs, param: :name do
    resources :memberships, only: %i[create new]
    put '/memberships/:token/change_role/:new_role', to: 'memberships#change_role', as: :membership_change_role
    delete '/memberships/:token', to: 'memberships#cancel', as: :membership
    get '/invite/:token/claim', to: 'memberships#claim', as: :claim_membership
    get 'add/:user_id', to: 'memberships#add_user', as: :add_membership
    post 'memberships/create_request', to: 'memberships#create_request', as: :create_request_membership
    get 'memberships/request', to: 'memberships#request_access', as: :request_membership
  end

  get '/profile/edit', to: 'profile#edit', as: :edit_profile
  put '/profile', to: 'profile#update', as: :update_profile

  get '/account', to: 'account_settings#settings', as: :account_settings
  post '/credentials/one_time_show', to: 'credentials#one_time_show', as: :one_time_show_credential
  post '/credentials/refresh', to: 'credentials#destroy', as: :delete_credential
  get '/credentials/one_time_show', to: 'account_settings#settings', as: :one_time_show_credential_reload

  authenticate :user, ->(u) { u.admin? } do
    mount Sidekiq::Web => '/admin/sidekiq'
  end
end
