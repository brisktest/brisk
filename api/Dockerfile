# Dockerfile.production
FROM ruby:3.1.0
# Add Tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

ENV INSTALL_PATH /opt/app
RUN mkdir -p $INSTALL_PATH

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg -o /root/yarn-pubkey.gpg && apt-key add /root/yarn-pubkey.gpg
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" > /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install -y --no-install-recommends  yarn
RUN curl -sL https://deb.nodesource.com/setup_lts.x -o nodesource_setup.sh
RUN bash nodesource_setup.sh
RUN apt install nodejs
RUN apt-get update && apt-get -y --no-install-recommends install libvips &&  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN groupadd rails && useradd -mg rails rails
RUN chown -R rails:rails $INSTALL_PATH
RUN chmod g+s $INSTALL_PATH
#USER rails

ENV RAILS_ENV production
ENV RACK_ENV production
ENV SECRET_KEY_BASE "nonesensenonesensenonesensenonesensenonesense"
WORKDIR $INSTALL_PATH

RUN gem install rails bundler
RUN mkdir -p tmp/pids
ADD package.json yarn.lock ./
RUN yarn install --check-files
ADD Gemfile.lock ./
ADD Gemfile ./
RUN bundle config set --local without 'development test'
RUN bundle install  
ADD . .
RUN  bin/rails tailwindcss:build
RUN  yarn build:prod
RUN  bundle exec rake assets:precompile

CMD  STATEMENT_TIMEOUT=280s bundle exec rake db:migrate:ignore_concurrent &&  STATEMENT_TIMEOUT=5s bundle exec puma -p 3000 -C config/puma.rb
CMD ["bundle", "exec", "gruf"]
