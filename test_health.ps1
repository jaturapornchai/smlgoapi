#!/usr/bin/env pwsh

Write-Host "Testing health endpoint..." -ForegroundColor Yellow

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8008/v1/health" -Method Get
    Write-Host "Health check successful!" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Green
    Write-Host "Version: $($response.version)" -ForegroundColor Green
    Write-Host "Database: $($response.database)" -ForegroundColor Green
} catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
}
