#!/bin/bash

echo "Running tests for all services..."

services=("auth-service" "restaurant-service" "order-service" "delivery-service" "payment-service" "tracking-service")

for service in "${services[@]}"; do
    service_path="services/$service"
    if [ -d "$service_path" ]; then
        echo "Testing $service..."
        cd "$service_path"
        go test ./tests/... -v
        if [ $? -ne 0 ]; then
            echo "Tests failed for $service"
            exit 1
        fi
        cd ../..
    else
        echo "Warning: $service_path not found"
    fi
done

# Run integration tests
echo "Running integration tests..."
go test ./tests/integration/... -v
if [ $? -ne 0 ]; then
    echo "Integration tests failed"
    exit 1
fi

echo "All tests complete!"