FROM golang:alpine AS builder

# Change the architecture if needed
ENV CGO_ENABLED=0 \
    GOOS=linux

COPY . /build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN go build -o gamerdeathbot api.go channel.go conn.go db.go main.go

WORKDIR /dist

RUN cp /build/gamerdeathbot .

# Build a small image
FROM scratch

COPY --from=builder /dist/gamerdeathbot /

CMD ["./gamerdeathbot"]
