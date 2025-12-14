module api-gateway

go 1.25.1

require (
	analysis-service v0.0.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.77.0
	storing-service v0.0.0
)

require (
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	analysis-service => ../analysis-service
	storing-service => ../storing-service
)
