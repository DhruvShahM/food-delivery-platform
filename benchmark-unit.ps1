Write-Host "Running unit benchmarks for all services..." -ForegroundColor Green

Get-ChildItem -Directory services | Where-Object { Test-Path (Join-Path $_.FullName "*.go") } | ForEach-Object {
    $serviceName = $_.Name
    Write-Host "Running benchmarks in $serviceName..." -ForegroundColor Yellow
    
    Push-Location $_.FullName
    try {
        go test -bench=. -benchmem -benchtime=5s ./...
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Benchmarks failed for $serviceName" -ForegroundColor Red
            exit 1
        }
    } finally {
        Pop-Location
    }
}

Write-Host "Unit benchmarks complete!" -ForegroundColor Green