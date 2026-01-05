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

## Detailed Help Commands

### Method 1: Using go run (Recommended)
Navigate to each service directory and run:

**1. Auth Service**
`cd services/auth-service && go run cmd/server/main.go`
- gRPC Port: :50051
- HTTP Port: :8080

**2. Restaurant Service**
`cd services/restaurant-service && go run cmd/server/main.go`
- gRPC Port: :50052
- HTTP Port: :8081

**3. Order Service**
`cd services/order-service && go run cmd/server/main.go`
- gRPC Port: :50053
- HTTP Port: :8082

**4. Delivery Service**
`cd services/delivery-service && go run cmd/server/main.go`
- gRPC Port: :50054
- HTTP Port: :8083

**5. Payment Service**
`cd services/payment-service && go run cmd/server/main.go`
- gRPC Port: :50055
- HTTP Port: :8084

**6. Tracking Service**
`cd services/tracking-service && go run cmd/server/main.go`
- gRPC Port: :50056
- HTTP Port: :8085

### Method 2: Build and Run
If you've run `make build-all`, you can run the binaries directly:
`cd services/auth-service && .\bin\server.exe`

### Quick Test Commands
`netstat -ano | findstr :50052` (Check port)
`taskkill /F /PID <PID>` (Kill process)

### Unit Testing
**Run all service tests:**
```powershell
.\test-all-services.ps1
# Or using the Makefile
make test-all
```

**Individual Service Testing:**
```bash
cd services/auth-service
go test ./tests/... -v
```

**Coverage:**
```bash
go test ./tests/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```