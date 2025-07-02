#!/usr/bin/env pwsh

Write-Host "Testing health endpoint..." -ForegroundColor Yellow

try {
    $headers = @{"Content-Type" = "application/json"}
    $response = Invoke-RestMethod -Uri "http://localhost:8008/v1/health" -Method Get -Headers $headers
    Write-Host "Health check successful!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 3 | Write-Host
} catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Full error details:" -ForegroundColor Yellow
    $_.Exception | Write-Host
}
