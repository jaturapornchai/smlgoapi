# PostgreSQL Endpoints Test Script (PowerShell)

$API_BASE = "http://localhost:8008"

Write-Host "🐘 Testing PostgreSQL Endpoints" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Test pgcommand endpoint
Write-Host ""
Write-Host "1. Testing /pgcommand endpoint..." -ForegroundColor Yellow
$body1 = @{
    query = "SELECT version()"
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri "$API_BASE/pgcommand" -Method POST -Body $body1 -ContentType "application/json"
    $response1 | ConvertTo-Json -Depth 10
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "2. Testing /v1/pgcommand endpoint..." -ForegroundColor Yellow
$body2 = @{
    query = "CREATE TABLE IF NOT EXISTS test_users (id SERIAL PRIMARY KEY, name VARCHAR(100), created_at TIMESTAMP DEFAULT NOW())"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri "$API_BASE/v1/pgcommand" -Method POST -Body $body2 -ContentType "application/json"
    $response2 | ConvertTo-Json -Depth 10
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Test pgselect endpoint
Write-Host ""
Write-Host "3. Testing /pgselect endpoint..." -ForegroundColor Yellow
$body3 = @{
    query = "SELECT current_database(), current_user, now() as current_time"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri "$API_BASE/pgselect" -Method POST -Body $body3 -ContentType "application/json"
    $response3 | ConvertTo-Json -Depth 10
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "4. Testing /v1/pgselect endpoint..." -ForegroundColor Yellow
$body4 = @{
    query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' LIMIT 5"
} | ConvertTo-Json

try {
    $response4 = Invoke-RestMethod -Uri "$API_BASE/v1/pgselect" -Method POST -Body $body4 -ContentType "application/json"
    $response4 | ConvertTo-Json -Depth 10
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "✅ PostgreSQL endpoints test completed!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Available endpoints:" -ForegroundColor Cyan
Write-Host "  - POST /pgcommand       (Legacy)" -ForegroundColor White
Write-Host "  - POST /v1/pgcommand    (Recommended)" -ForegroundColor White
Write-Host "  - POST /pgselect        (Legacy)" -ForegroundColor White
Write-Host "  - POST /v1/pgselect     (Recommended)" -ForegroundColor White
