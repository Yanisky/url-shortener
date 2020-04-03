FROM golang:1.14.1-alpine as builder

ENV GO111MODULE=on

RUN mkdir /build
WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build ./cmd/urlshortener/redis-postgres

WORKDIR /dist
RUN cp /build/redis-postgres ./server
RUN ls /dist

ENTRYPOINT ["/dist/server"]

