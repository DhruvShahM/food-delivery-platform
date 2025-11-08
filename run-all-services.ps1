# PowerShell script to run all services in the background

Write-Host "Starting all services..." -ForegroundColor Green

$services = @("auth-service", "restaurant-service", "order-service", "delivery-service", "payment-service", "tracking-service")

foreach ($service in $services) {
    $servicePath = "services\$service"
    if (Test-Path $servicePath) {
        Write-Host "Starting $service..." -ForegroundColor Yellow
        $fullPath = Join-Path $PWD $servicePath
        Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$fullPath'; go run cmd/server/main.go" -WindowStyle Minimized
        Start-Sleep -Milliseconds 500
    } else {
        Write-Host "Warning: $servicePath not found" -ForegroundColor Red
    }
}

Write-Host "All services started! Check individual PowerShell windows." -ForegroundColor Green
Write-Host "Press Ctrl+C to stop all services (close the individual windows)." -ForegroundColor Yellow

