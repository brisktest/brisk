ARG ARCH=
FROM golang:1.22 as worker-builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o build/worker.bin server/main.go 


FROM cimg/ruby:3.0.1-browsers

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN sudo chmod +x /tini
USER root
SHELL [ "/bin/bash", "-l", "-c" ]
RUN apt-get update && apt-get install -y \
  software-properties-common \
  npm \
  curl \
  sudo \
  libsqlite3-dev \
  libxml2-dev \
  postgresql-client

RUN apt-get update -q && \
  apt-get install -qy procps curl ca-certificates gnupg2 build-essential --no-install-recommends && apt-get clean



RUN apt-get update && apt-get install ca-certificates -y
ADD ./certs/AmazonRootCA1.cer /usr/local/share/ca-certificates/AmazonRootCA1.cer
ADD ./certs/squid-ca-cert.pem /usr/local/share/ca-certificates/squid-ca-cert.crt

RUN /usr/sbin/update-ca-certificates

RUN  mkdir /app
WORKDIR /app
ADD ../docker/docker-entrypoint-worker.sh ./
RUN chmod +x docker-entrypoint-worker.sh
ENTRYPOINT [ "/tini","-v", "--" ,"/app/docker-entrypoint-worker.sh"]
RUN usermod -l brisk circleci 
RUN groupmod -n brisk circleci 
RUN mv /home/circleci /home/brisk
WORKDIR /home/brisk
RUN find . -type f -exec sed -i 's/circleci/brisk/g' {} +
RUN find .* -type f -exec sed -i 's/circleci/brisk/g' {} +

WORKDIR /app
RUN chown -R brisk /home/brisk
RUN usermod -d /home/brisk brisk
COPY --from=worker-builder /src/build/worker.bin ./

RUN  chmod +u worker.bin


ADD ./certs /app/certs
RUN chown -R brisk /app/certs
RUN chmod -R 700 /app/certs

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=$(dpkg --print-architecture)] https://dl-ssl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
RUN apt update && apt install -y google-chrome-stable --no-install-recommends


RUN chown -R brisk /usr/local/lib /usr/local/share/ /usr/local/bin/
RUN adduser brisk sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN chown -R brisk /usr/local/

USER brisk
SHELL [ "/bin/bash", "-l", "-c" ]

RUN curl --silent -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.0/install.sh | bash


ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 14.18.0
# install node and npm
RUN source $NVM_DIR/nvm.sh \
  && nvm install $NODE_VERSION \
  && nvm alias default $NODE_VERSION \
  && nvm use default
# add node and npm to path so the commands are available
ENV NODE_PATH $NVM_DIR/v$NODE_VERSION/lib/node_modules
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

RUN echo \. "$NVM_DIR/nvm.sh" >> /home/brisk/.bashrc

ENV GEM_HOME /home/brisk/.rubygems
RUN unset GEM_HOME
RUN gpg2 --keyserver hkp://keyserver.ubuntu.com:80  --recv-keys 0x409B6B1796C275462A1703113804BB82D39DC0E3 0x7D2BAF1CF37B13E2069D6956105BD0E739499BDB
RUN sudo curl -sSL https://get.rvm.io | bash -s
# RUN rvm get stable --auto-dotfiles
RUN rvm install 3.0.2





RUN curl -o  /usr/local/share/ca-certificates/moz-all.pem https://curl.se/ca/cacert.pem
RUN cat /usr/local/share/ca-certificates/squid-ca-cert.crt >> /usr/local/share/ca-certificates/moz-all.pem
RUN npm config set cafile /usr/local/share/ca-certificates/moz-all.pem
RUN yarn config set cafile /usr/local/share/ca-certificates/moz-all.pem

RUN node -v
RUN npm -v


CMD  ["/app/worker.bin"]


