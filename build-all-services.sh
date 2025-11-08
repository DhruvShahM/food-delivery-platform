#!/bin/bash

echo "Building all services..."

services=("auth-service" "restaurant-service" "order-service" "delivery-service" "payment-service" "tracking-service")

for service in "${services[@]}"; do
    service_path="services/$service"
    if [ -d "$service_path" ]; then
        echo "Building $service..."
        cd "$service_path"
        go mod tidy
        go build -o bin/server ./cmd/server
        cd ../..
    else
        echo "Warning: $service_path not found"
    fi
done

echo "All services built!"