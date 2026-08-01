@echo off
title Exam Arena - Starting Services
setlocal

echo ============================================
echo   Exam Arena - Starting All Services
echo ============================================
echo.

:: Check backend .env exists
if not exist "%~dp0backend\.env" (
    echo [ERROR] backend\.env not found!
    echo        Copy backend\.env.example to backend\.env and fill in your secrets.
    pause
    exit /b 1
)

:: Check frontend .env.local exists
if not exist "%~dp0frontend\.env.local" (
    echo [WARN]  frontend\.env.local not found.
    echo         Backend API will default to http://localhost:8080
    echo         Create frontend\.env.local with:
    echo           NEXT_PUBLIC_API_URL=http://localhost:8080
    echo           NEXT_PUBLIC_WS_URL=ws://localhost:8080
    echo.
)

:: Start Backend (Go)
echo [1/2] Starting Backend (Go) on port 8080...
start "Exam Arena - Backend" cmd /k "cd /d %~dp0backend && go run ./cmd/server"

:: Wait for backend to initialize before frontend tries to connect
echo       Waiting 4 seconds for backend to initialize...
timeout /t 4 /nobreak > nul

:: Start Frontend (Next.js)
echo [2/2] Starting Frontend (Next.js) on port 3000...
start "Exam Arena - Frontend" cmd /k "cd /d %~dp0frontend && npm run dev"

echo.
echo ============================================
echo   Services starting in separate windows:
echo   Backend REST  -> http://localhost:8080
echo   Backend WS    -> ws://localhost:8080/ws
echo   Backend Health-> http://localhost:8080/health
echo   Frontend App  -> http://localhost:3000
echo ============================================
echo.
echo   Auth:        Email+Password (Go JWT) or Google (Firebase)
echo   Matchmaking: Backend WS first, local simulation fallback
echo   Leaderboard: Backend /api/v1/leaderboard/{category}
echo.
echo Press any key to open the app in browser...
pause > nul

start http://localhost:3000
