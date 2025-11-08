# PowerShell script to stop all running services

Write-Host "Stopping all services..." -ForegroundColor Yellow

# Service ports
$ports = @(50051, 50052, 50053, 50054, 50055, 50056, 8080, 8081, 8082, 8083, 8084, 8085)

foreach ($port in $ports) {
    $connections = netstat -ano | findstr ":$port"
    if ($connections) {
        $processIds = ($connections | ForEach-Object {
            if ($_ -match '\s+(\d+)$') {
                $matches[1]
            }
        }) | Select-Object -Unique
        
        foreach ($pid in $processIds) {
            try {
                $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                if ($process -and $process.ProcessName -like "*go*" -or $process.ProcessName -like "*server*") {
                    Write-Host "Stopping process $pid on port $port..." -ForegroundColor Yellow
                    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
                }
            } catch {
                # Process already stopped or doesn't exist
            }
        }
    }
}

Write-Host "Done!" -ForegroundColor Green


