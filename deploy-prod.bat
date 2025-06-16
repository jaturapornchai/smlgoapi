@echo off
REM SMLGOAPI Production Deployment Script for Windows
REM This script deploys SMLGOAPI to production server

echo ========================================
echo SMLGOAPI Production Deployment
echo ========================================
echo.

REM Check if Git is available
git --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Git is not installed or not in PATH!
    pause
    exit /b 1
)

echo [INFO] Git is available...
echo.

REM Menu selection
echo Choose deployment option:
echo 1. Git only (push to repository)
echo 2. Production only (deploy to server)
echo 3. Full deployment (Git + Production)
echo 4. Check production status
echo.

set /p choice="Enter your choice (1-4): "

if "%choice%"=="1" goto git_deploy
if "%choice%"=="2" goto prod_deploy
if "%choice%"=="3" goto full_deploy
if "%choice%"=="4" goto check_status
echo [ERROR] Invalid choice!
pause
exit /b 1

:git_deploy
echo.
echo [INFO] Deploying with Git...
echo.

git add .
git commit -m "Deploying the latest changes"
git push

if %errorlevel% equ 0 (
    echo [SUCCESS] Git deployment completed!
) else (
    echo [ERROR] Git deployment failed!
)
goto end

:prod_deploy
echo.
echo [INFO] Deploying to production server...
echo [INFO] Connecting to root@143.198.192.64...
echo.

ssh root@143.198.192.64 "cd /data/vectorapi-dev/ && docker pull ghcr.io/smlsoft/vectordbapi:main && docker compose up -d"

if %errorlevel% equ 0 (
    echo [SUCCESS] Production deployment completed!
) else (
    echo [ERROR] Production deployment failed!
)
goto end

:full_deploy
echo.
echo [INFO] Starting full deployment...
echo.

echo [STEP 1/3] Git deployment...
git add .
git commit -m "Deploying the latest changes"
git push

if %errorlevel% neq 0 (
    echo [ERROR] Git deployment failed!
    goto end
)

echo [STEP 2/3] Waiting for CI/CD to build image...
timeout /t 10 /nobreak

echo [STEP 3/3] Production deployment...
ssh root@143.198.192.64 "cd /data/vectorapi-dev/ && docker pull ghcr.io/smlsoft/vectordbapi:main && docker compose up -d"

if %errorlevel% equ 0 (
    echo [SUCCESS] Full deployment completed!
) else (
    echo [ERROR] Production deployment failed!
)
goto end

:check_status
echo.
echo [INFO] Checking production server status...
echo.

echo Checking Docker containers:
ssh root@143.198.192.64 "cd /data/vectorapi-dev/ && docker compose ps"

echo.
echo Checking API health:
ssh root@143.198.192.64 "curl -s http://localhost:8080/health || echo 'API not responding'"

goto end

:end
echo.
pause
