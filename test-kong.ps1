# test-kong.ps1 - Test Kong API Gateway

Write-Host "🧪 Testing Kong API Gateway..." -ForegroundColor Cyan

# Test 1: Kong Health
Write-Host "`n1. Testing Kong Health..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8001/status" -TimeoutSec 5
    $status = $response.Content | ConvertFrom-Json
    Write-Host "✅ Kong Status: $($status.server.connections_accepted) connections accepted" -ForegroundColor Green
} catch {
    Write-Host "❌ Kong not responding" -ForegroundColor Red
}

# Test 2: Routes
Write-Host "`n2. Testing Kong Routes..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8001/routes" -TimeoutSec 5
    $routes = $response.Content | ConvertFrom-Json
    Write-Host "✅ Routes configured: $($routes.data.Count) routes" -ForegroundColor Green
    
    foreach ($route in $routes.data) {
        Write-Host "   - $($route.paths[0]) → $($route.service.name)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Could not fetch routes" -ForegroundColor Red
}

# Test 3: Services
Write-Host "`n3. Testing Kong Services..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8001/services" -TimeoutSec 5
    $services = $response.Content | ConvertFrom-Json
    Write-Host "✅ Services configured: $($services.data.Count) services" -ForegroundColor Green
    
    foreach ($service in $services.data) {
        Write-Host "   - $($service.name): $($service.url)" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Could not fetch services" -ForegroundColor Red
}

# Test 4: Plugins
Write-Host "`n4. Testing Kong Plugins..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8001/plugins" -TimeoutSec 5
    $plugins = $response.Content | ConvertFrom-Json
    Write-Host "✅ Plugins configured: $($plugins.data.Count) plugins" -ForegroundColor Green
    
    $pluginTypes = $plugins.data | Group-Object -Property name
    foreach ($group in $pluginTypes) {
        Write-Host "   - $($group.Name): $($group.Count) instances" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Could not fetch plugins" -ForegroundColor Red
}

# Test 5: API Endpoints (if services are running)
Write-Host "`n5. Testing API Endpoints..." -ForegroundColor Yellow

$endpoints = @(
    @{Name="Health Check"; Url="http://localhost:8000/health"; Method="GET"},
    @{Name="Circuit Breakers"; Url="http://localhost:8000/circuit-breakers"; Method="GET"}
)

foreach ($endpoint in $endpoints) {
    try {
        $response = Invoke-WebRequest -Uri $endpoint.Url -Method $endpoint.Method -TimeoutSec 5
        Write-Host "✅ $($endpoint.Name): $($response.StatusCode)" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  $($endpoint.Name): Service not responding" -ForegroundColor Yellow
    }
}

Write-Host "`n🎯 Kong API Gateway Test Complete!" -ForegroundColor Green
Write-Host "`n💡 If services are not responding:" -ForegroundColor Yellow
Write-Host "1. Start services: .\run-all-services.ps1" -ForegroundColor White
Write-Host "2. Check service logs: docker-compose logs <service-name>" -ForegroundColor White
Write-Host "3. Verify network: docker network ls" -ForegroundColor White
