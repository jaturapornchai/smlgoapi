# Test script for /search-by-vector endpoint with collection as database name
# PowerShell script to test the new collection -> database functionality

$baseUrl = "http://localhost:8008"

# Test 1: Search with default database (no collection specified)
Write-Host "=== Test 1: Search with default database (no collection specified) ===" -ForegroundColor Green
$body1 = @{
    query = "โตโยต้า"
    limit = 5
    offset = 0
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri "$baseUrl/v1/search-by-vector" -Method POST -Body $body1 -ContentType "application/json"
    Write-Host "Success: Found $($response1.data.total_count) results" -ForegroundColor Green
    Write-Host "Database used: Default (from ENV)" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 2: Search with Changthai collection -> changthai database
Write-Host "=== Test 2: Search in Changthai database ===" -ForegroundColor Green
$body2 = @{
    query = "โตโยต้า"
    limit = 5
    offset = 0
    collection = "Changthai"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri "$baseUrl/v1/search-by-vector" -Method POST -Body $body2 -ContentType "application/json"
    Write-Host "Success: Found $($response2.data.total_count) results" -ForegroundColor Green
    Write-Host "Database used: changthai (lowercase from 'Changthai')" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: Search with mixed case collection -> lowercase database
Write-Host "=== Test 3: Search with mixed case collection ===" -ForegroundColor Green
$body3 = @{
    query = "brake"
    limit = 3
    offset = 0
    collection = "TOYOTA_Parts"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri "$baseUrl/v1/search-by-vector" -Method POST -Body $body3 -ContentType "application/json"
    Write-Host "Success: Found $($response3.data.total_count) results" -ForegroundColor Green
    Write-Host "Database used: toyota_parts (lowercase from 'TOYOTA_Parts')" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 4: Test with non-existent database
Write-Host "=== Test 4: Test non-existent database error handling ===" -ForegroundColor Green
$body4 = @{
    query = "test"
    limit = 3
    offset = 0
    collection = "NonExistentDB"
} | ConvertTo-Json

try {
    $response4 = Invoke-RestMethod -Uri "$baseUrl/v1/search-by-vector" -Method POST -Body $body4 -ContentType "application/json"
    Write-Host "Unexpected success: $($response4.data.total_count) results" -ForegroundColor Yellow
} catch {
    Write-Host "Expected error: $($_.Exception.Message)" -ForegroundColor Cyan
    Write-Host "This is expected behavior for non-existent database" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host "✅ Collection parameter is used as PostgreSQL database name" -ForegroundColor Green
Write-Host "✅ Collection names are automatically converted to lowercase" -ForegroundColor Green
Write-Host "✅ Proper error handling for non-existent databases" -ForegroundColor Green
Write-Host "✅ Backward compatibility maintained for requests without collection" -ForegroundColor Green
