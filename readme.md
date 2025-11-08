# Food Delivery Platform

Microservices-based food delivery app with gRPC, Kafka, OTEL, Prometheus, Grafana.

## Setup
1. `make up-infra` - Start Postgres, Redis, Kafka, Jaeger, Prometheus, Grafana.
2. `make proto-all` - Generate protos.
3. Per service: `cd services/auth-service && go mod tidy && go build -o bin/server ./cmd/server`
4. `make run-all` - Run all services.
5. Test: grpcurl -plaintext -d '{"email":"test@example.com","password":"password"}' localhost:50051 auth.AuthService/Login

## Services
- Auth: User login, JWT.
- Restaurant: Menu management.
- Order: Place order, Kafka event.
- Delivery: Assign delivery, Kafka consume.
- Payment: Process payment.
- Tracking: Location streams.

## Monitoring
- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Tech
Go 1.21, gRPC, Kafka, Postgres, Redis, Gin, OTEL, Zap.