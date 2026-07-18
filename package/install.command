#!/bin/bash
MYDIR="$(cd "$(dirname "$0")" && pwd)"
set -e

APP_NAME="东芳美诊所管理系统.app"
APP_SOURCE="$MYDIR/$APP_NAME"
APP_TARGET="/Applications/$APP_NAME"
PG_SOURCE="$MYDIR/pg"
PG_HOME="$HOME/.clinic-mgmt"
PG_DIR="$PG_HOME/pg"
PG_DATA="$PG_HOME/data"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "================================================"
echo -e "  ${GREEN}东芳美诊所管理系统 v1.3 — 离线安装${NC}"
echo "================================================"
echo ""

# Check macOS
[ "$(uname)" != "Darwin" ] && { echo -e "${RED}[✗] 仅支持 macOS${NC}"; exit 1; }
ARCH=$(uname -m)
echo -e "${GREEN}[✓] macOS $(sw_ver=$(sw_vers -productVersion 2>/dev/null || echo "?"); echo $sw_ver) / ${ARCH}${NC}"

# Check bundled PostgreSQL
[ ! -f "$PG_SOURCE/bin/pg_ctl" ] && { echo -e "${RED}[✗] 安装包损坏：缺少 PostgreSQL${NC}"; exit 1; }
echo -e "${GREEN}[✓] 内置 PostgreSQL 就绪${NC}"

# Check for previous installation
if [ -d "$APP_TARGET" ] || [ -d "$PG_DIR" ]; then
  echo -e "${YELLOW}检测到已有安装，将覆盖更新...${NC}"
  # Stop any running services first
  lsof -ti :8080 2>/dev/null | xargs kill -9 2>/dev/null || true
  if [ -f "$PG_DIR/bin/pg_ctl" ] && [ -f "$PG_DATA/PG_VERSION" ]; then
    "$PG_DIR/bin/pg_ctl" -D "$PG_DATA" stop 2>/dev/null || true
    sleep 1
  fi
fi

# Step 1: Copy files
echo ""
echo -e "${YELLOW}[1/4] 复制文件到本地...${NC}"
mkdir -p "$PG_HOME"
rm -rf "$PG_DIR" 2>/dev/null || true
rm -rf "$APP_TARGET" 2>/dev/null || true

if command -v ditto &>/dev/null; then
  ditto --noqtn "$PG_SOURCE" "$PG_DIR"
  ditto --noqtn "$APP_SOURCE" "$APP_TARGET"
else
  cp -R "$PG_SOURCE" "$PG_DIR"
  cp -R "$APP_SOURCE" "$APP_TARGET"
fi
# Remove quarantine
xattr -rd com.apple.quarantine "$PG_HOME" 2>/dev/null || true
xattr -rd com.apple.quarantine "$APP_TARGET" 2>/dev/null || true
echo -e "${GREEN}[✓] 文件已复制${NC} ($APP_TARGET)"

# Step 2: Initialize database
echo -e "${YELLOW}[2/4] 初始化数据库...${NC}"
export DYLD_LIBRARY_PATH="$PG_DIR/lib:$DYLD_LIBRARY_PATH"
export PATH="$PG_DIR/bin:$PATH"
mkdir -p "$PG_DATA"

if [ ! -f "$PG_DATA/PG_VERSION" ]; then
  "$PG_DIR/bin/initdb" -D "$PG_DATA" --encoding=UTF8 --locale=C 2>/dev/null
  echo -e "${GREEN}[✓] 数据库已初始化${NC}"
else
  echo -e "${GREEN}[✓] 数据库已存在${NC}"
fi

# Step 3: Start PostgreSQL and create database
echo -e "${YELLOW}[3/4] 配置数据库...${NC}"
"$PG_DIR/bin/pg_ctl" -D "$PG_DATA" -l "$PG_DATA/pg.log" stop 2>/dev/null || true
sleep 1
"$PG_DIR/bin/pg_ctl" -D "$PG_DATA" -l "$PG_DATA/pg.log" start 2>/dev/null || true
sleep 2

# Wait for PostgreSQL to be ready
for i in 1 2 3 4 5; do
  if "$PG_DIR/bin/pg_isready" -q 2>/dev/null; then
    echo -e "${GREEN}[✓] PostgreSQL 运行中${NC}"
    break
  fi
  sleep 1
done

"$PG_DIR/bin/psql" -p 5432 -d postgres -c "CREATE USER clinic WITH PASSWORD 'clinic123';" 2>/dev/null || true
"$PG_DIR/bin/createdb" -p 5432 clinic 2>/dev/null || true
"$PG_DIR/bin/psql" -p 5432 -d postgres -c "ALTER DATABASE clinic OWNER TO clinic;" 2>/dev/null || true
echo -e "${GREEN}[✓] 数据库就绪${NC}"

# Step 4: Save config
echo "$PG_DIR" > "$PG_HOME/pg-path.txt"

# Done!
echo -e "${YELLOW}[4/4] 安装完成${NC}"
echo ""
echo "================================================"
echo -e "  ${GREEN}安装完成！${NC}"
echo "================================================"
echo ""
echo "打开方式："
echo "  1. 打开 Finder → 应用程序 → 东芳美诊所管理系统"
echo "  2. 双击启动，浏览器自动打开 http://localhost:8080"
echo ""
echo "账号：admin / admin123"
echo ""

# Launch Finder
open /Applications 2>/dev/null || true
