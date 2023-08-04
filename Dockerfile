FROM golang:1.20-alpine as build
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY main.go .
RUN go main -o ./backup .


FROM alpine

COPY --from=build /app/backup /db-backup

RUN apk add --no-cache postgresql-client ca-certificates && rm -rf /var/cache/apk/*

VOLUME ["/backup"]

CMD ["/db-backup"]