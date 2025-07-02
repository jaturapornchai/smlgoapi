# Test search functionality with actual products
$apiUrl = "http://localhost:8008/v1/search-by-vector"

# Test with a term that exists: "coil"
Write-Host "Testing search with 'coil' (should find results)..." -ForegroundColor Green
$body1 = @{
    query = "coil"
    limit = 100
    offset = 0
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body1 -ContentType "application/json"
    Write-Host "Success!" -ForegroundColor Green
    Write-Host "Total Count: $($response1.data.TotalCount)" -ForegroundColor Cyan
    Write-Host "Returned Results: $($response1.data.Data.Count)" -ForegroundColor Cyan
    Write-Host "Query: $($response1.data.Query)" -ForegroundColor Cyan
    if($response1.data.Data.Count -gt 0) {
        Write-Host "Sample results:" -ForegroundColor Yellow
        for($i = 0; $i -lt [Math]::Min(3, $response1.data.Data.Count); $i++) {
            $item = $response1.data.Data[$i]
            Write-Host "  $($i+1). [$($item.Code)] $($item.Name)" -ForegroundColor White
        }
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test with a term that exists and high limit
Write-Host "`nTesting search with 'coil' and limit=200..." -ForegroundColor Green
$body2 = @{
    query = "coil"
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

Write-Host "`nTesting completed!" -ForegroundColor Green
