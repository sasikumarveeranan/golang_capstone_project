# Register Order Service, Product Service and Gateway with Consul.
# Start all three services and Consul before running this script.

$services = @(
  @{ID="order-service-1"; Name="order-service"; Address="host.docker.internal"; Port=8080},
  @{ID="product-service-1"; Name="product-service"; Address="host.docker.internal"; Port=8081},
  @{ID="gateway-1"; Name="gateway"; Address="host.docker.internal"; Port=8088}
)

foreach ($s in $services) {
  $body = $s | ConvertTo-Json
  Invoke-RestMethod -Method Put -Uri "http://localhost:8500/v1/agent/service/register" -ContentType "application/json" -Body $body
}

Write-Host "Services registered. Check http://localhost:8500"
