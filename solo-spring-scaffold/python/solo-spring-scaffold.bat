@echo off
REM solo-spring-scaffold.bat — 快捷启动脚本
REM 用法: solo-spring-scaffold init my-app --package com.example.myapp
REM
REM 安装：把这个 .bat 文件所在目录加到 PATH 中
REM       或在命令行中直接使用完整路径

set "SCRIPT_DIR=%~dp0"
set "PYTHON=%USERPROFILE%\AppData\Local\hermes\hermes-agent\.venv\Scripts\python.exe"

if not exist "%PYTHON%" (
    echo Error: Python not found at %PYTHON%
    echo 请修改本 .bat 文件中的 PYTHON 路径为你本机的 python.exe
    pause
    exit /b 1
)

"%PYTHON%" "%SCRIPT_DIR%solo-spring-scaffold" %*
