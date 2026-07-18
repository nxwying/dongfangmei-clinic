# 东芳美诊所管理系统 v1.3 — Windows 安装脚本
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  东芳美诊所管理系统 v1.3 — Windows 安装" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

$pgDir = "C:\Program Files\PostgreSQL\16"
$pgBin = "$pgDir\bin"
$psql = "$pgBin\psql.exe"
$installDir = "$env:LOCALAPPDATA\ClinicMgmt"
$serverExe = "$installDir\server.exe"
$dataDir = "$installDir\data"

# Step 1: Check PostgreSQL
Write-Host "[1/5] 检查 PostgreSQL..." -ForegroundColor Yellow
$pgFound = $false
if (Test-Path $psql) {
    $pgFound = $true
    Write-Host "  ✓ 已找到 PostgreSQL 16" -ForegroundColor Green
} else {
    # Try to find any PostgreSQL version
    $pgPaths = @(
        "C:\Program Files\PostgreSQL\*\bin\psql.exe",
        "C:\Program Files (x86)\PostgreSQL\*\bin\psql.exe"
    )
    foreach ($pattern in $pgPaths) {
        $files = Get-ChildItem $pattern -ErrorAction SilentlyContinue
        if ($files) {
            $psql = $files[0].FullName
            $pgBin = Split-Path $psql -Parent
            $pgDir = Split-Path $pgBin -Parent
            $pgFound = $true
            Write-Host "  ✓ 已找到 PostgreSQL ($pgDir)" -ForegroundColor Green
            break
        }
    }
}

if (-not $pgFound) {
    Write-Host "  ✗ 未找到 PostgreSQL" -ForegroundColor Red
    Write-Host ""
    Write-Host "请先安装 PostgreSQL 16，然后再运行此脚本。" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "下载地址：https://www.postgresql.org/download/windows/" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "安装时请注意：" -ForegroundColor White
    Write-Host "  1. 保持默认安装路径（C:\Program Files\PostgreSQL\16）"
    Write-Host "  2. 设置密码为：clinic123"
    Write-Host "  3. 端口保持默认：5432"
    Write-Host ""
    $choice = Read-Host "是否现在打开下载页面？(Y/N)"
    if ($choice -eq 'Y' -or $choice -eq 'y') {
        Start-Process "https://www.postgresql.org/download/windows/"
    }
    Read-Host "安装好 PostgreSQL 后，按 Enter 继续"
    
    # Check again
    if (Test-Path $psql) { $pgFound = $true }
}

# Step 2: Create installation directory
Write-Host "[2/5] 创建安装目录..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
New-Item -ItemType Directory -Force -Path $dataDir | Out-Null
Copy-Item "$PSScriptRoot\server.exe" $serverExe -Force
Write-Host "  ✓ 已复制到 $installDir" -ForegroundColor Green

# Step 3: Initialize database
Write-Host "[3/5] 初始化数据库..." -ForegroundColor Yellow
$env:Path = "$pgBin;$env:Path"

# Create user and database
& "$pgBin\psql" -p 5432 -d postgres -c "CREATE USER clinic WITH PASSWORD 'clinic123';" 2>$null
$databaseCreated = $?
& "$pgBin\createdb" -p 5432 clinic 2>$null
& "$pgBin\psql" -p 5432 -d postgres -c "ALTER DATABASE clinic OWNER TO clinic;" 2>$null

# Test connection with new user
$env:PGPASSWORD = 'clinic123'
& "$pgBin\psql" -p 5432 -d clinic -U clinic -c "SELECT 1;" 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ! 数据库连接测试失败，可能需要手动配置" -ForegroundColor Yellow
    Write-Host "  请确认 PostgreSQL 正在运行（服务名：postgresql-x64-16）" -ForegroundColor Yellow
    Write-Host "  可在服务管理器中启动或重启该服务" -ForegroundColor Yellow
} else {
    Write-Host "  ✓ 数据库就绪" -ForegroundColor Green
}
Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue

# Step 4: Create startup scripts
Write-Host "[4/5] 创建启动脚本..." -ForegroundColor Yellow

@"
@echo off
title 东芳美诊所管理系统
echo ================================================
echo   东芳美诊所管理系统 v1.3
echo ================================================
echo.
echo 正在启动服务，请稍候...
echo.
echo 启动后请打开浏览器访问：
echo   http://localhost:8080
echo.
echo 账号：admin / admin123
echo.
echo 按 Ctrl+C 停止服务
echo ================================================
echo.
cd /d "$installDir"
"$serverExe"
pause
"@ | Out-File -FilePath "$installDir\启动.bat" -Encoding ASCII

@"
@echo off
title 停止东芳美诊所管理系统
echo 正在停止服务...
taskkill /f /im server.exe 2>nul
echo 已停止
pause
"@ | Out-File -FilePath "$installDir\停止.bat" -Encoding ASCII

# Create desktop shortcut
$WshShell = New-Object -ComObject WScript.Shell
$shortcut = $WshShell.CreateShortcut("$env:USERPROFILE\Desktop\东芳美诊所管理系统.lnk")
$shortcut.TargetPath = "$installDir\启动.bat"
$shortcut.WorkingDirectory = "$installDir"
$shortcut.Description = "东芳美诊所管理系统"
$shortcut.Save()
Write-Host "  ✓ 已创建桌面快捷方式" -ForegroundColor Green

# Step 5: Configure firewall
Write-Host "[5/5] 配置防火墙..." -ForegroundColor Yellow
New-NetFirewallRule -DisplayName "东芳美诊所管理系统 (8080)" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow -ErrorAction SilentlyContinue 2>$null
Write-Host "  ✓ 防火墙规则已添加（TCP 8080）" -ForegroundColor Green

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  安装完成！" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "启动方式：" -ForegroundColor White
Write-Host "  1. 双击桌面「东芳美诊所管理系统」快捷方式"
Write-Host "  2. 或打开 $installDir 双击「启动.bat」"
Write-Host ""
Write-Host "浏览器打开：http://localhost:8080" -ForegroundColor Cyan
Write-Host "账号：admin / admin123" -ForegroundColor Cyan
Write-Host ""
pause
