# PowerShell script to fix all service go.mod files

Write-Host "Fixing go.mod files for all services..." -ForegroundColor Green

$services = @("auth-service", "restaurant-service", "order-service", "delivery-service", "payment-service", "tracking-service")

foreach ($service in $services) {
    $servicePath = Join-Path "services" $service
    if (Test-Path $servicePath) {
        Write-Host "Fixing $service..." -ForegroundColor Yellow
        Push-Location $servicePath
        
        # Remove problematic indirect dependency line if it exists
        $goModPath = Join-Path (Get-Location) "go.mod"
        $content = Get-Content $goModPath -Raw
        $content = $content -replace '^\s*google\.golang\.org/genproto/googleapis/rpc\s+v[^\s]+.*$', ''
        Set-Content -Path $goModPath -Value $content -NoNewline
        
        # Run go mod tidy
        go mod tidy
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✓ $service fixed successfully" -ForegroundColor Green
        } else {
            Write-Host "  ✗ Error fixing $service" -ForegroundColor Red
        }
        
        Pop-Location
    }
}

Write-Host "Done!" -ForegroundColor Green

