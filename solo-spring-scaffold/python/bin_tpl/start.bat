@echo off
REM ─── start.bat ───
REM 用法: bin\start.bat [dev|prod]

set APP_NAME={{projectName}}
set PROFILE=%1
if "%PROFILE%"=="" set PROFILE=dev

set JAR_DIR=%~dp0..
set JAR_FILE=%JAR_DIR%\%APP_NAME%-web\target\%APP_NAME%-web-1.0.0.jar
set LOG_DIR=%JAR_DIR%\logs

if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

echo Starting %APP_NAME% (profile=%PROFILE%)...
start "%~n0" java -Xms256m -Xmx512m ^
    -jar "%JAR_FILE%" ^
    --spring.profiles.active="%PROFILE%"
