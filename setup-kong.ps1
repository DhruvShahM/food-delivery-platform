# setup-kong.ps1 - Kong API Gateway Setup Script

Write-Host "🚀 Setting up Kong API Gateway..." -ForegroundColor Green

# Check if kong.yml exists
if (!(Test-Path "kong.yml")) {
    Write-Host "❌ kong.yml not found. Please create kong.yml first." -ForegroundColor Red
    exit 1
}

# Start Kong
Write-Host "📦 Starting Kong API Gateway..." -ForegroundColor Yellow
docker-compose up -d kong

# Wait for Kong to be healthy
Write-Host "⏳ Waiting for Kong to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0

while ($attempt -lt $maxAttempts) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8001" -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ Kong is ready!" -ForegroundColor Green
            break
        }
    } catch {
        # Ignore errors and retry
    }
    
    $attempt++
    Start-Sleep -Seconds 2
}

if ($attempt -eq $maxAttempts) {
    Write-Host "❌ Kong failed to start properly" -ForegroundColor Red
    exit 1
}

# Test the API Gateway
Write-Host "`n🧪 Testing API Gateway..." -ForegroundColor Cyan

# Test Kong proxy
try {
    $proxyResponse = Invoke-WebRequest -Uri "http://localhost:8000/health" -TimeoutSec 10
    Write-Host "✅ Proxy endpoint working: $($proxyResponse.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Proxy endpoint not responding (services may not be running)" -ForegroundColor Yellow
}

# Test Kong admin API
try {
    $adminResponse = Invoke-WebRequest -Uri "http://localhost:8001" -TimeoutSec 5
    Write-Host "✅ Admin API working: $($adminResponse.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "❌ Admin API not working" -ForegroundColor Red
}

Write-Host "`n🎉 Kong API Gateway Setup Complete!" -ForegroundColor Green
Write-Host "`n📋 Access Points:" -ForegroundColor Cyan
Write-Host "🌐 API Gateway Proxy: http://localhost:8000" -ForegroundColor White
Write-Host "⚙️  Kong Admin API:   http://localhost:8001" -ForegroundColor White
Write-Host "🖥️  Kong Admin GUI:   http://localhost:8002" -ForegroundColor White

Write-Host "`n📖 Usage Examples:" -ForegroundColor Cyan
Write-Host "🔐 Login:     POST http://localhost:8000/auth/login" -ForegroundColor White
Write-Host "📋 Orders:    GET  http://localhost:8000/api/v1/orders" -ForegroundColor White
Write-Host "💳 Payments:  POST http://localhost:8000/api/v1/payments" -ForegroundColor White

Write-Host "`n💡 Next Steps:" -ForegroundColor Yellow
Write-Host "1. Start your microservices with: .\run-all-services.ps1" -ForegroundColor White
Write-Host "2. Test API endpoints through Kong proxy" -ForegroundColor White
Write-Host "3. Check Kong Admin GUI for monitoring" -ForegroundColor White
