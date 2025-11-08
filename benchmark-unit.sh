#!/bin/bash

echo "Running unit benchmarks for all services..."

for dir in services/*/; do
    if [ -d "$dir" ] && find "$dir" -name "*.go" -type f | read; then
        service_name=$(basename "$dir")
        echo "Running benchmarks in $service_name..."
        cd "$dir" && go test -bench=. -benchmem -benchtime=5s ./... && cd ../..
    fi
done

echo "Unit benchmarks complete!"
