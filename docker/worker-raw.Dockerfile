ARG ARCH=
FROM golang:1.22 as worker-builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o build/worker.bin server/main.go 



FROM cimg/node:18.10.0


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
RUN mkdir -p /home/brisk
RUN chown -R brisk /home/brisk
RUN usermod -d /home/brisk brisk
COPY --from=worker-builder /src/build/worker.bin ./

RUN  chmod +u worker.bin

## Our start command which kicks off
## our newly created binary executable
ADD ./certs /app/certs
RUN chown -R brisk /app/certs
RUN chmod -R 700 /app/certs
# RUN npm install npm@latest -g && \
#   npm install n -g && \
#   n latest
# RUN npm install --global yarn
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
# confirm installation
RUN curl -o  /usr/local/share/ca-certificates/moz-all.pem https://curl.se/ca/cacert.pem
RUN cat /usr/local/share/ca-certificates/squid-ca-cert.crt >> /usr/local/share/ca-certificates/moz-all.pem
RUN npm config set cafile /usr/local/share/ca-certificates/moz-all.pem

RUN yarn config set cafile /usr/local/share/ca-certificates/moz-all.pem
RUN node -v
RUN npm -v






#potentially run my server in a different user than the one we use to run the commands
# some weirdness with permissions there

CMD  ["/app/worker.bin"]






# FROM cimg/openjdk:19.0.1

# ENV TINI_VERSION v0.18.0
# ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
# RUN sudo chmod +x /tini
# USER root
# SHELL [ "/bin/bash", "-l", "-c" ]

# RUN apt-get update && apt-get install ca-certificates -y
# ADD ./certs/AmazonRootCA1.cer /usr/local/share/ca-certificates/AmazonRootCA1.cer
# ADD ./certs/squid-ca-cert.pem /usr/local/share/ca-certificates/squid-ca-cert.crt

# RUN /usr/sbin/update-ca-certificates



# RUN  mkdir /app
# WORKDIR /app
# ADD docker/docker-entrypoint-worker.sh ./
# RUN chmod +x docker-entrypoint-worker.sh
# ENTRYPOINT [ "/tini","-v", "--" ,"/app/docker-entrypoint-worker.sh"]
# RUN usermod -l brisk circleci
# RUN groupmod -n brisk circleci
# RUN mv /home/circleci /home/brisk
# WORKDIR /home/brisk
# RUN find . -type f -exec sed -i 's/circleci/brisk/g' {} +
# RUN find .* -type f -exec sed -i 's/circleci/brisk/g' {} +

# WORKDIR /app
# RUN chown -R brisk /home/brisk
# RUN usermod -d /home/brisk brisk
# ADD ./build/worker.bin ./

# RUN  chmod +u worker.bin
# ADD ./certs /app/certs
# RUN chown -R brisk /app/certs
# RUN chmod -R 700 /app/certs

# RUN chown -R brisk /usr/local/lib /usr/local/share/ /usr/local/bin/
# RUN adduser brisk sudo
# RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
# RUN chown -R brisk /usr/local/

# USER brisk
# SHELL [ "/bin/bash", "-l", "-c" ]




# CMD  ["/app/worker.bin"]


