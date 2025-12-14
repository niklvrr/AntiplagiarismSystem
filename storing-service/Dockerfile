FROM golang:1.25-alpine3.21 AS builder
RUN apk --no-cache add bash make gcc g++
WORKDIR /app
COPY storing-service/go.mod storing-service/go.sum ./storing-service/
COPY analysis-service/go.mod analysis-service/go.sum ./analysis-service/
WORKDIR /app/storing-service
RUN go mod download
WORKDIR /app
COPY storing-service/ ./storing-service/
COPY analysis-service/ ./analysis-service/
WORKDIR /app/storing-service
RUN go build -o /storing-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /storing-service ./
COPY --from=builder /app/storing-service/migrations ./migrations
EXPOSE 50051
CMD ["./storing-service"]
