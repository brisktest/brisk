# frozen_string_literal: true

Rails.autoloaders.main.ignore(Rails.root.join('app/rpc/api/brisk-api_pb.rb'))
Rails.autoloaders.main.ignore(Rails.root.join('app/rpc/api/brisk-api_services_pb.rb'))
Rails.autoloaders.main.ignore(Rails.root.join('app/rpc/api/shared_types_pb.rb'))

require Rails.root.join('app/rpc/api/brisk-api_services_pb').to_s
# require Rails.root.join('app/rpc/interceptors/token_auth.rb').to_s

# Rails.autoloaders.each do |autoloader|
#   autoloader.inflector.inflect(
#     'brisk-api_services_pb.rb' => 'ApiProjectsService'
#   )
# end
