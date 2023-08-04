FROM golang:1.20 as build
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY main.go .
RUN go build -o ./backup .


FROM debian:12-slim

RUN groupadd -g 1001 backupp && useradd -mG backupp -u 1001 backupper

RUN apt update -y && apt install --no-install-recommends -y postgresql-client ca-certificates tzdata rclone && rm -rf /var/lib/{apt,dpkg,cache,log}/

COPY --from=build /app/backup /db-backup

VOLUME ["/backup"]

CMD ["/db-backup"]