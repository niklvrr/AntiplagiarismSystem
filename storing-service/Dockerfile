FROM golang:1.25-alpine3.21 AS builder
RUN apk --no-cache add bash make gcc g++
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /storing-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /storing-service ./
COPY --from=builder /app/migrations ./migrations
EXPOSE 50051
CMD ["./storing-service"]
