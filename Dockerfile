FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /spelling-server ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /spelling-server /app/spelling-server

ENV RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/"
ENV STATS_PATH="/app/data/stats.json"
ENV GRPC_PORT="50051"

EXPOSE 50051

ENTRYPOINT ["/app/spelling-server"]
