.PHONY: proto-all build-all run-all test-all benchmark-all benchmark-unit benchmark-integration benchmark-load up-infra down clean stop-all benchmark-auth benchmark-order benchmark-payment benchmark-restaurant benchmark-delivery benchmark-tracking

# Replace the bash version with PowerShell
proto-all:
	powershell -ExecutionPolicy Bypass -File generate-proto.ps1

build-all:
	powershell -ExecutionPolicy Bypass -File build-all-services.ps1

run-all:
	powershell -ExecutionPolicy Bypass -File run-all-services.ps1

test-all:
	powershell -ExecutionPolicy Bypass -File test-all-services.ps1

up-infra:
	docker compose up -d

down:
	docker compose down

stop-all:
	powershell -ExecutionPolicy Bypass -File stop-all-services.ps1

clean:
	docker compose down -v
	for /d %d in (services\*) do (if exist "%d\bin" rmdir /s /q "%d\bin")

benchmark-unit:
	powershell -Command "Get-ChildItem -Directory services | Where-Object { Test-Path (Join-Path $$_.FullName '*.go') } | ForEach-Object { Push-Location $$_.FullName; Write-Host \"Running benchmarks in $$($$_.Name)...\"; go test -bench=. -benchmem -benchtime=5s ./...; Pop-Location }"

benchmark-integration:
	cd tests/integration && go test -bench=. -benchmem -benchtime=10s .

benchmark-all: benchmark-unit benchmark-integration

benchmark-load:
	@echo "Starting load test with artillery..."
	artillery run benchmark/artillery-config.yml --output benchmark/results.json
	artillery report benchmark/results.json --output benchmark/report.html

# Individual service benchmarks
benchmark-auth:
	cd services/auth-service && go test -bench=. -benchmem -benchtime=5s .

benchmark-order:
	cd services/order-service && go test -bench=. -benchmem -benchtime=5s .

benchmark-payment:
	cd services/payment-service && go test -bench=. -benchmem -benchtime=5s .

benchmark-restaurant:
	cd services/restaurant-service && go test -bench=. -benchmem -benchtime=5s .

benchmark-delivery:
	cd services/delivery-service && go test -bench=. -benchmem -benchtime=5s .

benchmark-tracking:
	cd services/tracking-service && go test -bench=. -benchmem -benchtime=5s .