# Test what products are available in PostgreSQL
$apiUrl = "http://localhost:8008/v1/pgselect"

# Test 1: Get sample products to see what's in the database
Write-Host "Getting sample products..." -ForegroundColor Yellow
$body1 = @{
    query = "SELECT name, code FROM ic_inventory ORDER BY name LIMIT 10"
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body1 -ContentType "application/json"
    Write-Host "Sample products ($($response1.row_count)):" -ForegroundColor Green
    foreach($item in $response1.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Check for English brake products
Write-Host "`nTesting for brake products..." -ForegroundColor Yellow
$body2 = @{
    query = "SELECT name, code FROM ic_inventory WHERE name ILIKE '%brake%' OR code ILIKE '%brake%' LIMIT 10"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body2 -ContentType "application/json"
    Write-Host "Found $($response2.row_count) brake products:" -ForegroundColor Green
    foreach($item in $response2.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Check for oil products
Write-Host "`nTesting for oil products..." -ForegroundColor Yellow
$body3 = @{
    query = "SELECT name, code FROM ic_inventory WHERE name ILIKE '%oil%' OR code ILIKE '%oil%' LIMIT 10"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body3 -ContentType "application/json"
    Write-Host "Found $($response3.row_count) oil products:" -ForegroundColor Green
    foreach($item in $response3.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}
