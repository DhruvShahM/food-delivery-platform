# PowerShell script to test all services


Write-Host "Running tests for all services..." -ForegroundColor Green

$services = @("auth-service", "restaurant-service", "order-service", "delivery-service", "payment-service", "tracking-service")

foreach ($service in $services) {
    $servicePath = "services\$service"
    if (Test-Path $servicePath) {
        Write-Host "Testing $service..." -ForegroundColor Yellow
        Push-Location $servicePath
        go test ./tests/... -v
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Tests failed for $service" -ForegroundColor Red
        }
        Pop-Location
    } else {
        Write-Host "Warning: $servicePath not found" -ForegroundColor Red
    }
}

# Run integration tests
Write-Host "Running integration tests..." -ForegroundColor Yellow
go test ./tests/integration/... -v
if ($LASTEXITCODE -ne 0) {
    Write-Host "Integration tests failed" -ForegroundColor Red
}

Write-Host "All tests complete!" -ForegroundColor Green

