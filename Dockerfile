FROM golang:1.20-alpine as build
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY main.go .
RUN go build -o ./backup .


FROM alpine

RUN addgroup --system --gid 1001 backup && adduser --system --uid 1001 backupper

COPY --from=build /app/backup /db-backup

RUN apk add --no-cache postgresql-client ca-certificates && rm -rf /var/cache/apk/*

VOLUME ["/backup"]

CMD ["/db-backup"]