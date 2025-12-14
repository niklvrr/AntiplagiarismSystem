FROM golang:1.25-alpine3.21 AS builder
RUN apk --no-cache add bash make gcc g++
WORKDIR /app
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
COPY storing-service/go.mod storing-service/go.sum ./storing-service/
COPY analysis-service/go.mod analysis-service/go.sum ./analysis-service/
WORKDIR /app/api-gateway
RUN go mod download
WORKDIR /app
COPY api-gateway/ ./api-gateway/
COPY storing-service/ ./storing-service/
COPY analysis-service/ ./analysis-service/
WORKDIR /app/api-gateway
RUN go build -o /api-gateway ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /api-gateway ./
COPY api-gateway/api ./api
EXPOSE 8080
CMD ["./api-gateway"]
