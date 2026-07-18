@echo off
title 东芳美诊所管理系统
chcp 65001 >nul
echo ================================================
echo   东芳美诊所管理系统 v1.3
echo ================================================
echo.
echo 正在启动服务，请稍候...
echo 启动后请打开浏览器访问 http://localhost:8080
echo.
echo 账号：admin / admin123
echo.
echo 按 Ctrl+C 停止服务
echo ================================================
echo.
start "" http://localhost:8080
server.exe
pause
