ARG TARGETOS=darwin
ARG TARGETARCH=arm64
ARG ARCH=
FROM golang:1.22 as super-builder
WORKDIR /src
COPY . .
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o build/super.bin super/main.go
#RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o build/super.bin super/main.go

RUN CGO_ENABLED=0 go build -o build/super.bin super/main.go
FROM bitnami/minideb:bookworm

RUN adduser -uid 3434 brisk
RUN apt-get update && apt-get -y install rsync

RUN apt-get update && apt-get -y install openssh-client

RUN apt-get update && apt-get install ca-certificates -y
ADD ./certs/AmazonRootCA1.cer /usr/local/share/ca-certificates/AmazonRootCA1.cer
RUN /usr/sbin/update-ca-certificates

RUN chown -R brisk /tmp/
RUN chown -R brisk /var/lib
RUN  mkdir /app
WORKDIR /app
ADD ./certs /app/certs


RUN  chown -R  brisk /home/brisk


RUN  chown -R  brisk certs
RUN chmod -R 700 certs
# Need to copy this from the build stage
COPY --from=super-builder /src/build/super.bin ./

RUN  chmod +u super.bin
USER brisk
CMD  ["/app/super.bin"]