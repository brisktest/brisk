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

RUN chown -R brisk /tmp/
RUN chown -R brisk /var/lib
RUN  mkdir /app
WORKDIR /app



RUN  chown -R  brisk /home/brisk
# Need to copy this from the build stage
COPY --from=super-builder /src/build/super.bin ./

RUN  chmod +u super.bin
USER brisk
CMD  ["/app/super.bin"]