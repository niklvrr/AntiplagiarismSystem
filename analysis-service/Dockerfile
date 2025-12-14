FROM golang:1.25-alpine3.21 AS builder
RUN apk --no-cache add bash make gcc g++
WORKDIR /app
COPY analysis-service/go.mod analysis-service/go.sum ./analysis-service/
WORKDIR /app/analysis-service
RUN go mod download
WORKDIR /app
COPY analysis-service/ ./analysis-service/
WORKDIR /app/analysis-service
RUN go build -o /analysis-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /analysis-service ./
COPY --from=builder /app/analysis-service/migrations ./migrations
EXPOSE 50052
CMD ["./analysis-service"]
