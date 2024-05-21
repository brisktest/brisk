# not sure where to put it... spec/support/zeitwerk_spec.rb perhaps?
require 'rails_helper'
describe 'Zeitwerk' do
  it 'eager loads all files' do
    expect { Zeitwerk::Loader.eager_load_all }.not_to raise_error
  end
end
