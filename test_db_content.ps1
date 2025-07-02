# Test what products are available in PostgreSQL
$apiUrl = "http://localhost:8008/v1/pgselect"

# Test 1: Check for brake-related products  
Write-Host "Testing for brake-related products..." -ForegroundColor Yellow
$body1 = @{
    query = "SELECT name, code FROM ic_inventory WHERE name ILIKE '%brake%' OR name ILIKE '%เบรค%' OR code ILIKE '%brake%' LIMIT 10"
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body1 -ContentType "application/json"
    Write-Host "Found $($response1.row_count) brake products:" -ForegroundColor Green
    foreach($item in $response1.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Check for common car parts
Write-Host "`nTesting for common car parts..." -ForegroundColor Yellow
$body2 = @{
    query = "SELECT name, code FROM ic_inventory WHERE name ILIKE '%oil%' OR name ILIKE '%filter%' OR name ILIKE '%แปรง%' LIMIT 10"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body2 -ContentType "application/json"
    Write-Host "Found $($response2.row_count) oil/filter products:" -ForegroundColor Green
    foreach($item in $response2.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Get sample products to see what's in the database
Write-Host "`nGetting sample products..." -ForegroundColor Yellow
$body3 = @{
    query = "SELECT name, code FROM ic_inventory ORDER BY name LIMIT 10"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body3 -ContentType "application/json"
    Write-Host "Sample products ($($response3.row_count)):" -ForegroundColor Green
    foreach($item in $response3.data) {
        Write-Host "  - $($item.code): $($item.name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}
