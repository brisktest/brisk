# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[7.0].define(version: 2024_04_05_213145) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "accounts", force: :cascade do |t|
    t.integer "user_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "plan_id"
  end

  create_table "active_storage_attachments", force: :cascade do |t|
    t.string "name", null: false
    t.string "record_type", null: false
    t.bigint "record_id", null: false
    t.bigint "blob_id", null: false
    t.datetime "created_at", null: false
    t.index ["blob_id"], name: "index_active_storage_attachments_on_blob_id"
    t.index ["record_type", "record_id", "name", "blob_id"], name: "index_active_storage_attachments_uniqueness", unique: true
  end

  create_table "active_storage_blobs", force: :cascade do |t|
    t.string "key", null: false
    t.string "filename", null: false
    t.string "content_type"
    t.text "metadata"
    t.string "service_name", null: false
    t.bigint "byte_size", null: false
    t.string "checksum"
    t.datetime "created_at", null: false
    t.index ["key"], name: "index_active_storage_blobs_on_key", unique: true
  end

  create_table "active_storage_variant_records", force: :cascade do |t|
    t.bigint "blob_id", null: false
    t.string "variation_digest", null: false
    t.index ["blob_id", "variation_digest"], name: "index_active_storage_variant_records_uniqueness", unique: true
  end

  create_table "api_actions", force: :cascade do |t|
    t.text "name"
    t.text "grpc_method_name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "api_actions_credentials", force: :cascade do |t|
    t.integer "credential_id"
    t.integer "api_action_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "cli_login_attempts", force: :cascade do |t|
    t.string "nonce"
    t.string "token"
    t.integer "user_id"
    t.datetime "valid_until"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "commands", force: :cascade do |t|
    t.text "commandline"
    t.text "args"
    t.text "work_directory"
    t.json "environment"
    t.boolean "is_test_run"
    t.integer "sequence_number"
    t.integer "worker_number"
    t.text "stage"
    t.boolean "is_list_test"
    t.text "test_framework"
    t.boolean "background"
    t.integer "total_worker_count"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "execution_info_id"
    t.boolean "no_test_files"
    t.string "command_id"
    t.integer "command_concurrency"
    t.index ["execution_info_id"], name: "index_commands_on_execution_info_id"
  end

  create_table "credentials", force: :cascade do |t|
    t.text "api_token"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.text "api_key"
    t.datetime "deleted_at", precision: nil
    t.integer "credentialable_id"
    t.string "credentialable_type"
    t.datetime "valid_through", precision: nil
    t.datetime "shown_to_user_at"
  end

  create_table "execution_infos", force: :cascade do |t|
    t.datetime "started"
    t.datetime "finished"
    t.integer "exit_code"
    t.text "rebuild_hash"
    t.string "output"
    t.string "text"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "worker_run_info_id"
    t.index ["finished"], name: "index_execution_infos_on_finished"
    t.index ["started"], name: "index_execution_infos_on_started"
    t.index ["worker_run_info_id"], name: "index_execution_infos_on_worker_run_info_id"
  end

  create_table "images", force: :cascade do |t|
    t.text "name"
    t.text "version"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.text "description"
  end

  create_table "jobruns", force: :cascade do |t|
    t.integer "project_id"
    t.integer "user_id"
    t.text "state"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "concurrency"
    t.text "worker_image"
    t.integer "assigned_concurrency"
    t.integer "exit_code"
    t.text "error"
    t.datetime "finished_at"
    t.text "output"
    t.text "trace_key"
    t.text "api_version"
    t.integer "supervisor_id"
    t.text "rebuild_hash"
    t.text "uid"
    t.text "notes"
    t.text "split_method"
    t.uuid "log_uid"
    t.index ["created_at"], name: "index_jobruns_on_created_at"
    t.index ["log_uid"], name: "index_jobruns_on_log_uid"
    t.index ["uid"], name: "index_jobruns_on_uid", unique: true
  end

  create_table "machines", force: :cascade do |t|
    t.integer "project_id"
    t.text "ip_address"
    t.text "state"
    t.text "instance_id"
    t.text "instance_type"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at", precision: nil
    t.text "host_ip"
    t.text "uid"
    t.text "os_info"
    t.text "host_uid"
    t.text "image"
    t.text "type"
    t.integer "cpus"
    t.bigint "memory"
    t.bigint "disk"
    t.jsonb "json_data"
    t.datetime "drain"
    t.datetime "finished_at"
    t.datetime "drained_at"
    t.datetime "last_ping_at"
    t.index ["uid"], name: "index_machines_on_uid", unique: true
  end

  create_table "memberships", force: :cascade do |t|
    t.text "invited_email"
    t.integer "user_id"
    t.datetime "accepted_at"
    t.integer "invited_by"
    t.string "role", default: "member"
    t.integer "org_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "cancelled_at"
    t.integer "inviter_id"
    t.text "token"
  end

  create_table "orgs", force: :cascade do |t|
    t.integer "owner_id"
    t.string "name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at", precision: nil
    t.integer "account_id"
  end

  create_table "plans", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.text "name"
    t.integer "amount_cents"
    t.text "currency"
    t.text "description"
    t.text "period"
    t.text "trial_period"
    t.text "status"
    t.integer "monthly_concurrency"
    t.integer "minimum_capacity", default: 5
  end

  create_table "projects", force: :cascade do |t|
    t.bigint "org_id", null: false
    t.string "name"
    t.bigint "user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at", precision: nil
    t.text "project_token"
    t.string "username"
    t.text "framework"
    t.integer "worker_concurrency", default: 1
    t.integer "image_id"
    t.integer "max_supervisors", default: 1
    t.integer "startup_time_in_ms"
    t.text "git_hosting_provider"
    t.text "git_org"
    t.text "git_repo_name"
    t.integer "memory_requirement"
    t.index ["org_id"], name: "index_projects_on_org_id"
    t.index ["user_id"], name: "index_projects_on_user_id"
  end

  create_table "repo_infos", force: :cascade do |t|
    t.text "commit_hash"
    t.text "repo"
    t.text "branch"
    t.text "tag"
    t.text "commit_message"
    t.text "commit_author"
    t.text "commit_author_email"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "jobrun_id"
  end

  create_table "schedules", force: :cascade do |t|
    t.string "org_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.decimal "min_worker_percent"
  end

  create_table "subscriptions", force: :cascade do |t|
    t.integer "plan_id"
    t.integer "account_id"
    t.integer "address_id"
    t.text "status"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "project_id"
  end

  create_table "supervisors", force: :cascade do |t|
    t.integer "project_id"
    t.string "ip_address"
    t.string "port"
    t.string "state"
    t.integer "machine_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "host_ip"
    t.string "sync_port"
    t.string "sync_endpoint"
    t.text "external_endpoint"
    t.text "uid"
    t.datetime "setup_run_at", precision: nil
    t.text "unique_instance_id"
    t.text "affinity"
    t.datetime "in_use"
    t.index ["affinity"], name: "index_supervisors_on_affinity"
    t.index ["created_at"], name: "index_supervisors_on_created_at"
    t.index ["machine_id"], name: "index_supervisors_on_machine_id"
    t.index ["project_id"], name: "index_supervisors_on_project_id"
    t.index ["state"], name: "index_supervisors_on_state"
    t.index ["uid"], name: "unque_index_supervisors_uid", unique: true
  end

  create_table "test_file_runs", force: :cascade do |t|
    t.integer "test_file_id"
    t.integer "ms_time_taken"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "worker_run_info_id"
    t.index ["test_file_id", "worker_run_info_id"], name: "index_test_file_runs_on_test_file_id_and_worker_run_info_id"
    t.index ["test_file_id"], name: "index_test_file_runs_on_test_file_id"
    t.index ["worker_run_info_id"], name: "index_test_file_runs_on_worker_run_info_id"
  end

  create_table "test_files", force: :cascade do |t|
    t.string "filename"
    t.string "version"
    t.string "language"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "project_id"
    t.float "timing_confidence"
    t.integer "runtime"
    t.index ["filename"], name: "index_test_files_on_filename"
  end

  create_table "users", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "email", default: "", null: false
    t.string "encrypted_password", default: "", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at", precision: nil
    t.datetime "remember_created_at", precision: nil
    t.integer "sign_in_count", default: 0, null: false
    t.datetime "current_sign_in_at", precision: nil
    t.datetime "last_sign_in_at", precision: nil
    t.inet "current_sign_in_ip"
    t.inet "last_sign_in_ip"
    t.datetime "deleted_at", precision: nil
    t.text "name"
    t.string "confirmation_token"
    t.datetime "confirmed_at", precision: nil
    t.datetime "confirmation_sent_at", precision: nil
    t.string "unconfirmed_email"
    t.boolean "moderator"
    t.boolean "admin"
    t.text "profile_image"
    t.string "time_zone", default: "Pacific Time (US & Canada)"
    t.string "provider"
    t.string "uid"
    t.index ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true
    t.index ["email"], name: "index_users_on_email", unique: true
    t.index ["reset_password_token"], name: "index_users_on_reset_password_token", unique: true
  end

  create_table "worker_run_infos", force: :cascade do |t|
    t.text "rebuild_hash"
    t.text "exit_code"
    t.text "output"
    t.datetime "finished_at"
    t.datetime "started_at"
    t.text "error"
    t.integer "worker_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "project_id"
    t.integer "supervisor_id"
    t.integer "jobrun_id"
    t.integer "ms_time_taken"
    t.text "uid"
    t.text "log_encryption_key"
    t.text "log_location"
    t.integer "test_count"
    t.datetime "timing_processed"
    t.index ["created_at"], name: "index_worker_run_infos_on_created_at"
    t.index ["exit_code"], name: "index_worker_run_infos_on_exit_code"
    t.index ["finished_at"], name: "index_worker_run_infos_on_finished_at"
    t.index ["jobrun_id"], name: "index_worker_run_infos_on_jobrun_id"
    t.index ["project_id"], name: "index_worker_run_infos_on_project_id"
    t.index ["started_at"], name: "index_worker_run_infos_on_started_at"
    t.index ["supervisor_id"], name: "index_worker_run_infos_on_supervisor_id"
    t.index ["test_count"], name: "index_worker_run_infos_on_test_count"
    t.index ["uid"], name: "index_worker_run_infos_on_uid", unique: true
    t.index ["worker_id"], name: "index_worker_run_infos_on_worker_id"
  end

  create_table "workers", force: :cascade do |t|
    t.integer "project_id"
    t.text "ip_address"
    t.text "port"
    t.text "state"
    t.text "image_id"
    t.text "endpoint"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "machine_id"
    t.datetime "deleted_at", precision: nil
    t.text "host_ip"
    t.datetime "build_commands_run_at", precision: nil
    t.text "uid"
    t.datetime "last_checked_at", precision: nil
    t.text "worker_image"
    t.datetime "freed_at"
    t.text "rebuild_hash"
    t.integer "supervisor_id"
    t.datetime "reserved_at"
    t.integer "jobrun_id"
    t.integer "assigned_ram"
    t.text "sync_port"
    t.index ["created_at"], name: "index_workers_on_created_at"
    t.index ["freed_at"], name: "index_workers_on_freed_at"
    t.index ["host_ip"], name: "index_workers_on_host_ip"
    t.index ["image_id"], name: "index_workers_on_image_id"
    t.index ["last_checked_at", "state"], name: "index_workers_on_last_checked_at_and_state"
    t.index ["last_checked_at"], name: "index_workers_on_last_checked_at"
    t.index ["machine_id"], name: "index_workers_on_machine_id"
    t.index ["project_id"], name: "index_workers_on_project_id"
    t.index ["rebuild_hash"], name: "index_workers_on_rebuild_hash"
    t.index ["state"], name: "index_workers_on_state"
    t.index ["supervisor_id"], name: "index_workers_on_supervisor_id"
    t.index ["uid"], name: "unque_index_worker_uid", unique: true
    t.index ["updated_at"], name: "index_workers_on_updated_at"
    t.index ["worker_image"], name: "index_workers_on_worker_image"
  end

  add_foreign_key "active_storage_attachments", "active_storage_blobs", column: "blob_id"
  add_foreign_key "active_storage_variant_records", "active_storage_blobs", column: "blob_id"
  add_foreign_key "projects", "orgs"
  add_foreign_key "projects", "users"
end
