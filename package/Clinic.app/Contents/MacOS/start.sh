#!/bin/bash
DIR="$(cd "$(dirname "$0")" && pwd)"
PG_HOME="$HOME/.clinic-mgmt"
PG_DATA="$PG_HOME/data"
PG_DIR=""

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo ""
echo "================================================"
echo "  东芳美诊所管理系统"
echo "================================================"
echo ""

# Find PostgreSQL
if [ -f "$PG_HOME/pg-path.txt" ]; then
  PG_DIR=$(cat "$PG_HOME/pg-path.txt")
elif [ -d "$PG_HOME/pg/bin" ]; then
  PG_DIR="$PG_HOME/pg"
fi

if [ -n "$PG_DIR" ]; then
  export DYLD_LIBRARY_PATH="$PG_DIR/lib:$DYLD_LIBRARY_PATH"
  export PATH="$PG_DIR/bin:$PATH"
fi

# Start PostgreSQL if not running
if command -v pg_isready &>/dev/null && pg_isready -q 2>/dev/null; then
  echo -e "${GREEN}[✓] PostgreSQL 运行中${NC}"
else
  echo -e "${YELLOW}正在启动 PostgreSQL...${NC}"
  if [ -f "$PG_DIR/bin/pg_ctl" ]; then
    "$PG_DIR/bin/pg_ctl" -D "$PG_DATA" -l "$PG_DATA/pg.log" start 2>/dev/null || true
  elif command -v brew &>/dev/null; then
    brew services start postgresql@16 2>/dev/null || true
  fi
  sleep 2
  if command -v pg_isready &>/dev/null && pg_isready -q 2>/dev/null; then
    echo -e "${GREEN}[✓] PostgreSQL 已启动${NC}"
  else
    echo -e "${RED}[✗] PostgreSQL 启动失败，请检查${NC}"
  fi
fi

# Ensure database exists
if [ -f "$PG_DIR/bin/psql" ]; then
  "$PG_DIR/bin/psql" -p 5432 -U clinic -d clinic -c "SELECT 1" &>/dev/null || {
    "$PG_DIR/bin/psql" -p 5432 -d postgres -c "CREATE USER clinic WITH PASSWORD 'clinic123';" &>/dev/null || true
    "$PG_DIR/bin/createdb" -p 5432 clinic &>/dev/null || true
    "$PG_DIR/bin/psql" -p 5432 -d postgres -c "ALTER DATABASE clinic OWNER TO clinic;" &>/dev/null || true
    echo -e "${GREEN}[✓] 数据库已初始化${NC}"
  }
fi

# Kill any existing server process on port 8080
lsof -ti :8080 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 1

# Get LAN IP
LAN_IP=""
command -v ipconfig &>/dev/null && LAN_IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null)
[ -z "$LAN_IP" ] && LAN_IP=$(ifconfig 2>/dev/null | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)

echo ""
echo -e "  本机访问：${GREEN}http://localhost:8080${NC}"
[ -n "$LAN_IP" ] && echo -e "  局域网访问：${GREEN}http://${LAN_IP}:8080${NC}"
echo -e "  登录账号：${YELLOW}admin / admin123${NC}"
echo "================================================"
echo ""

# Start server
cd "$DIR"
./clinic-server &
CLINIC_PID=$!

# Wait for server to be ready
for i in 1 2 3 4 5; do
  sleep 1
  if curl -s http://localhost:8080/api/v1/auth/login -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' &>/dev/null; then
    echo -e "${GREEN}[✓] 服务已就绪${NC}"
    break
  fi
done

# Open browser
open http://localhost:8080 2>/dev/null || true

# Wait for process
trap "kill $CLINIC_PID 2>/dev/null; exit 0" SIGINT SIGTERM
wait $CLINIC_PID
