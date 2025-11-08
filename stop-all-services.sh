#!/bin/bash

echo "Stopping all services..."

# Stop services
echo "Stopping microservices..."
# Add your service stop commands here

# Stop infrastructure
echo "Stopping infrastructure..."
docker compose down

echo "All services stopped!"