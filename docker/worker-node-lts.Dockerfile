ARG ARCH=
FROM golang:1.22 as worker-builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o build/worker.bin server/main.go 


FROM cimg/node:18.10.0
#FROM cimg/ruby:3.0.1
#RUN rm /bin/sh && ln -s /bin/bash /bin/sh

#FROM nestybox/ubuntu-bionic-systemd

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN sudo chmod +x /tini
USER root

RUN apt-get update && apt-get install -y \
  software-properties-common \
  npm \
  curl \
  sudo
RUN apt-get update && apt-get install -y libgtk2.0-0 libgtk-3-0 libgbm-dev libnotify-dev libgconf-2-4 libnss3 libxss1 libasound2 libxtst6 xauth xvfb




RUN  mkdir /app
WORKDIR /app
RUN usermod -l brisk circleci
RUN groupmod -n brisk circleci
RUN mkdir -p /home/brisk
RUN chown -R brisk /home/brisk
RUN usermod -d /home/brisk brisk



RUN chown -R brisk /usr/local/lib /usr/local/share/ /usr/local/bin/
RUN adduser brisk sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN chown -R brisk /usr/local/
USER brisk
ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 14.18.0

RUN curl --silent -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.0/install.sh | bash
# install node and npm
RUN source $NVM_DIR/nvm.sh \
  && nvm install $NODE_VERSION \
  && nvm alias default $NODE_VERSION \
  && nvm use default

ENV NODE_VERSION 16.7
RUN source $NVM_DIR/nvm.sh \
  && nvm install $NODE_VERSION \
  && nvm alias default $NODE_VERSION \
  && nvm use default

ENV NODE_VERSION 18.9
RUN source $NVM_DIR/nvm.sh \
  && nvm install $NODE_VERSION \
  && nvm alias default $NODE_VERSION \
  && nvm use default
# add node and npm to path so the commands are available
ENV NODE_PATH $NVM_DIR/v$NODE_VERSION/lib/node_modules
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

RUN echo \. "$NVM_DIR/nvm.sh" >> /home/brisk/.bashrc
RUN echo \. "$NVM_DIR/nvm.sh" >> /home/brisk/.bash_profile
RUN echo \. "$NVM_DIR/nvm.sh" >> /home/brisk/.profile

RUN node -v
RUN npm -v

USER root
COPY --from=worker-builder /src/build/worker.bin /app/worker.bin

RUN  chmod +u /app/worker.bin
USER brisk



#potentially run my server in a different user than the one we use to run the commands
# some weirdness with permissions there

CMD  ["/app/worker.bin"]


