# SMLGOAPI Docker Management Script for Windows PowerShell

param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "run", "stop", "logs", "clean", "status", "shell", "help")]
    [string]$Command = "help"
)

$ImageName = "smlgoapi:latest"

function Show-Help {
    Write-Host "SMLGOAPI Docker Management Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\docker.ps1 [COMMAND]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  build       Build Docker image" -ForegroundColor White
    Write-Host "  run         Run container with ClickHouse" -ForegroundColor White
    Write-Host "  stop        Stop all services" -ForegroundColor White
    Write-Host "  logs        Show application logs" -ForegroundColor White
    Write-Host "  clean       Remove containers and images" -ForegroundColor White
    Write-Host "  status      Show container status" -ForegroundColor White
    Write-Host "  shell       Open shell in running container" -ForegroundColor White
    Write-Host "  help        Show this help message" -ForegroundColor White
    Write-Host ""
}

function Invoke-Build {
    Write-Host "🏗️  Building Docker image..." -ForegroundColor Blue
    docker build -t $ImageName .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build complete!" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Start-Services {
    Write-Host "🚀 Starting services..." -ForegroundColor Blue
    docker-compose up -d
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Services started!" -ForegroundColor Green
        Write-Host "📊 API: http://localhost:8080" -ForegroundColor Cyan
        Write-Host "🗄️  ClickHouse: http://localhost:8123" -ForegroundColor Cyan
        Write-Host "💚 Health: http://localhost:8080/health" -ForegroundColor Cyan
    }
    else {
        Write-Host "❌ Failed to start services!" -ForegroundColor Red
        exit 1
    }
}

function Stop-Services {
    Write-Host "⏹️  Stopping services..." -ForegroundColor Blue
    docker-compose down
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Services stopped!" -ForegroundColor Green
    }
}

function Show-Logs {
    Write-Host "📋 Showing logs..." -ForegroundColor Blue
    docker-compose logs -f smlgoapi
}

function Remove-All {
    Write-Host "🧹 Cleaning up..." -ForegroundColor Blue
    docker-compose down --rmi all --volumes --remove-orphans
    docker image rm $ImageName -ErrorAction SilentlyContinue
    Write-Host "✅ Cleanup complete!" -ForegroundColor Green
}

function Show-Status {
    Write-Host "📊 Container status:" -ForegroundColor Blue
    docker-compose ps
}

function Open-Shell {
    Write-Host "🐚 Opening shell in container..." -ForegroundColor Blue
    docker-compose exec smlgoapi sh
}

switch ($Command) {
    "build" { Invoke-Build }
    "run" { Start-Services }
    "stop" { Stop-Services }
    "logs" { Show-Logs }
    "clean" { Remove-All }
    "status" { Show-Status }
    "shell" { Open-Shell }
    default { Show-Help }
}
