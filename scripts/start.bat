@echo off
setlocal
cd /d "%~dp0\.."

echo =====================================================================
echo   BIFROST SMARTROUTE - ONE-CLICK LAUNCHER
echo   Intelligent AI Model Routing ^& Cost Optimization System
echo =====================================================================

if not exist "bifrost-smartroute.exe" (
    echo [INFO] Building bifrost-smartroute binary...
    cd backend
    go build -o ..\bifrost-smartroute.exe ./cmd/server
    cd ..
)

echo [INFO] Starting Bifrost SmartRoute Server on :8080...
start "" http://localhost:8080
.\bifrost-smartroute.exe -port 8080

pause
