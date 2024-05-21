p = Plan.create! name: 'Developer', amount_cents: 0, monthly_concurrency: 1_000_000_000

Image.create name: 'node-lts', version: '1'
Image.create name: 'rails', version: '1'
Image.create name: 'python', version: '1'
Image.create name: 'raw', version: '1'

user = User.create! password: 'my-password123', email: 'devuser@brisktest.com'
user.confirmed_at = Time.now
user.save!

# account = Account.create!

# Org.create! account: account
# user.plan.account = p

# plan = Plan.last

# p = user.projects.create! framework: "Jest", org: user.owned_orgs.first, name: "default-project-#{SecureRandom.hex(4)}", worker_concurrency: 5
# p.update(project_token: "cUh3FEZ9tu")
user.credential.update(api_key: 'dYho0h93lNfD/u/P', api_token: 'AfzWBMS8oy')
# Project.create(user: user, org: Org.first, name: "First").update(project_token: "cUh3FEZ9tu")
# Project.create(user: user, org: Org.first, name: "Second").update(project_token: "sdFiuB34n2")
# Plan.create(name: "Developer", trial_period: "14 days")

[{ 'name' => 'Super Register', 'grpc_method_name' => 'api.supers.register' },
 { 'name' => 'Super De Register', 'grpc_method_name' => 'api.supers.de_register' },
 { 'name' => 'Worker De Register', 'grpc_method_name' => 'api.workers.de_register' },
 { 'name' => 'Worker Register', 'grpc_method_name' => 'api.workers.register' },
 { 'name' => 'Recently Deregistered',
   'grpc_method_name' => 'api.workers.get_recently_deregistered' },

 { 'name' => 'Infra Machine Register', 'grpc_method_name' => 'api.infra.register_machine' },
 { 'name' => 'Infra Machine DeRegister', 'grpc_method_name' => 'api.infra.de_register_machine' },
 { 'name' => 'Infra Machine Draining', 'grpc_method_name' => 'api.infra.drain_machine' }].each do |api_action|
  ApiAction.create!(api_action)
end

# no run
# BRISK_CONFIG_WARNINGS=true BRISK_APITOKEN=AfzWBMS8oy BRISK_APIKEY=dYho0h93lNfD/u/P  DEV=true BRISK_DEV=true BRISK_APIENDPOINT=localhost:9001  brisk project init node

## really only for nomad
# ApiAction.all.map do |a|
#   [a.name,
#    { api_key: a.api_actions_credentials.first.credential.api_key,
#      api_token: a.api_actions_credentials.first.credential.api_token }]
# end

# export_strings = []
# { "SUPER_REG_ROUTE" => "api.supers.register",
#   "SUPER_DEREG_ROUTE" => "api.supers.de_register",
#   "WORKER_REG_ROUTE" => "api.workers.register",
#   "WORKER_DEREG_ROUTE" => "api.workers.de_register",
#   "WORKER_RECENT_DEREG_ROUTE" => "api.workers.get_recently_deregistered",
#   "INFRA_REG_ROUTE" => "api.infra.register_machine",
#   "INFRA_DE_REG_ROUTE" => "api.infra.de_register_machine",
#   "INFRA_DRAIN_ROUTE" => "api.infra.drain_machine" }.each do |key, value|
#   c = ApiAction.find_by(grpc_method_name: value).api_actions_credentials.first

#   export_strings << "export NOMAD_VAR_#{key}_KEY=\"#{c.credential.api_key}\""
#   export_strings << "export NOMAD_VAR_#{key}_TOKEN=\"#{c.credential.api_token}\""
# end

# puts export_strings.join("\n")
