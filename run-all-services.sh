#!/bin/bash

echo "Running all services..."

# Start infrastructure first
echo "Starting infrastructure..."
docker compose up -d

echo "Waiting for services to be ready..."
sleep 10

# Run services (you'll need to implement the actual service startup logic)
echo "Starting microservices..."
# Add your service startup commands here

echo "All services running!"