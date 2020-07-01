FROM golang:alpine AS builder

RUN apk add git

# Change the architecture if needed
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm

WORKDIR /build

COPY . .

RUN go get github.com/kelr/go-twitch-api/twitchapi github.com/lib/pq

RUN go build -o gamerdeathbot api.go channel.go conn.go db.go main.go

WORKDIR /dist

RUN cp /build/gamerdeathbot .



FROM iron/go

COPY --from=builder /dist/gamerdeathbot /

CMD ["/gamerdeathbot"]
