#!/bin/bash
# 医美诊所管理系统 — 一键启动脚本（单端口部署）
set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

echo "================================================"
echo "  医美诊所管理系统 — 环境检查与启动"
echo "================================================"

# Check Go
command -v go >/dev/null 2>&1 || { echo "[✗] Go 未安装，请先安装: brew install go"; exit 1; }
echo "[✓] Go $(go version | grep -oP 'go\S+' || go version | awk '{print $3}')"

# Check PostgreSQL
command -v psql >/dev/null 2>&1 || { echo "[✗] PostgreSQL 未安装，请先安装: brew install postgresql@16"; exit 1; }
echo "[✓] PostgreSQL 已安装"
pg_isready -q && echo "[✓] PostgreSQL 运行中" || { echo "[✗] PostgreSQL 未运行，启动中..."; brew services start postgresql@16; sleep 2; }

# Build frontend
echo ""
echo "正在构建前端..."
cd web
npm install --silent
npx vite build --silent
cd ..

# Build server with embedded frontend
echo "正在编译后端..."
go mod tidy
go build -o server .

echo ""
echo "================================================"
echo "  启动成功！"
echo "  访问地址:  http://localhost:8080"
echo "  登录账号:  admin / admin123"
echo "  退出请按 Ctrl+C"
echo "================================================"
echo ""

# Start server (single binary, frontend embedded)
./server
