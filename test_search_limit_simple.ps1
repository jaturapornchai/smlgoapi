# Test script to verify search limit functionality
Write-Host "Testing SMLGOAPI Search Limit Functionality" -ForegroundColor Green
Write-Host "=============================================="

$apiUrl = "http://localhost:8008/v1/search-by-vector"

# Test 1: Search with limit 100 (should show the supplement logic)
Write-Host "`nTest 1: Searching with limit=100" -ForegroundColor Yellow
$body1 = @{
    query = "brake"
    limit = 100
    offset = 0
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body1 -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    Write-Host "Total Count: $($response1.data.TotalCount)" -ForegroundColor Cyan
    Write-Host "Returned Results: $($response1.data.Data.Count)" -ForegroundColor Cyan
    Write-Host "Query: $($response1.data.Query)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Search with limit 200 (should show more supplemental results)
Write-Host "`nTest 2: Searching with limit=200" -ForegroundColor Yellow
$body2 = @{
    query = "brake"
    limit = 200
    offset = 0
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body2 -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    Write-Host "Total Count: $($response2.data.TotalCount)" -ForegroundColor Cyan
    Write-Host "Returned Results: $($response2.data.Data.Count)" -ForegroundColor Cyan
    Write-Host "Query: $($response2.data.Query)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Search with default limit (should show vector DB results)
Write-Host "`nTest 3: Searching with default limit" -ForegroundColor Yellow
$body3 = @{
    query = "brake"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body3 -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    Write-Host "Total Count: $($response3.data.TotalCount)" -ForegroundColor Cyan
    Write-Host "Returned Results: $($response3.data.Data.Count)" -ForegroundColor Cyan
    Write-Host "Query: $($response3.data.Query)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nTesting completed!" -ForegroundColor Green
