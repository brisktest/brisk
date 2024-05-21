## We specify the base image we need for our
## go application
FROM golang:1.16-buster
RUN apt-get -y update  && apt-get -y install rsync openssh-server
RUN  mkdir /app
## We copy everything in the root directory
## into our /app directory
#ADD . /app
## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app
## we run go build to compile the binary
## executable of our Go program

## Our start command which kicks off
## our newly created binary executable
CMD  while true; do sleep 12 ; done 
#CMD ["go run brisk-cli/main.go"]