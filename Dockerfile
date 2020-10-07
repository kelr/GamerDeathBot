# Build container
FROM golang:alpine AS builder

LABEL maintainer="Kyle Ruan <kyle.ruan@protonmail.com>"

# Change the architecture if needed
ENV CGO_ENABLED=0 \
    GOOS=linux

COPY . /build
WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN go build -o gamerdeathbot api.go channel.go irc.go db.go manager.go main.go

WORKDIR /dist

RUN cp /build/gamerdeathbot .

# Get certs
FROM alpine:latest as certs

RUN apk --update add ca-certificates

# Build a small image
FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /dist/gamerdeathbot /

CMD ["./gamerdeathbot"]
