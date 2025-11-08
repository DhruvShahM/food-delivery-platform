# PowerShell script to build all services

Write-Host "Building all services..." -ForegroundColor Green

$services = @("auth-service", "restaurant-service", "order-service", "delivery-service", "payment-service", "tracking-service")

foreach ($service in $services) {
    $servicePath = "services\$service"
    if (Test-Path $servicePath) {
        Write-Host "Building $service..." -ForegroundColor Yellow
        Push-Location $servicePath
        go mod tidy
        if ($LASTEXITCODE -eq 0) {
            go build -o bin/server ./cmd/server
            if ($LASTEXITCODE -ne 0) {
                Write-Host "Error building $service" -ForegroundColor Red
            }
        } else {
            Write-Host "Error running go mod tidy for $service" -ForegroundColor Red
        }
        Pop-Location
    } else {
        Write-Host "Warning: $servicePath not found" -ForegroundColor Red
    }
}

Write-Host "Build complete!" -ForegroundColor Green


