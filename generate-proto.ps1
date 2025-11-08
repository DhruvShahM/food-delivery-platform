# PowerShell script to generate protobuf files for all services

Write-Host "Generating protobuf files..." -ForegroundColor Green

# Generate common proto
Write-Host "Generating common proto..." -ForegroundColor Yellow
cd common/proto
protoc -I. --go_out=. --go-grpc_out=. common.proto
cd ../..

# Generate service protos
$services = @(
    @{Dir="auth-service"; Proto="auth.proto"},
    @{Dir="restaurant-service"; Proto="restaurant.proto"},
    @{Dir="order-service"; Proto="order.proto"},
    @{Dir="delivery-service"; Proto="delivery.proto"},
    @{Dir="payment-service"; Proto="payment.proto"},
    @{Dir="tracking-service"; Proto="tracking.proto"}
)

foreach ($service in $services) {
    $servicePath = "services\$($service.Dir)"
    if (Test-Path $servicePath) {
        Write-Host "Generating proto for $($service.Dir)..." -ForegroundColor Yellow
        $protoFile = "$servicePath\proto\$($service.Proto)"
        if (Test-Path $protoFile) {
            cd "$servicePath\proto"
            protoc -I . -I ..\..\..\common\proto --go_out=..\internal\proto --go-grpc_out=..\internal\proto $($service.Proto)
            cd ..\..\..
        } else {
            Write-Host "Warning: $protoFile not found" -ForegroundColor Red
        }
    }
}

Write-Host "Proto generation complete!" -ForegroundColor Green

