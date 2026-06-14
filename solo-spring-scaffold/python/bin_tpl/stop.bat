@echo off
REM ─── stop.bat ───

set JAR_DIR=%~dp0..
set PID_FILE=%JAR_DIR%\logs\app.pid

if exist "%PID_FILE%" (
    set /p PID=<"%PID_FILE%"
    echo Stopping PID=%PID%...
    taskkill /F /PID %PID% 2>nul
    del "%PID_FILE%"
    echo Stopped.
) else (
    echo app.pid not found. Try: taskkill /F /IM java.exe
)
