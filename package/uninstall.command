#!/bin/bash
MYDIR="$(cd "$(dirname "$0")" && pwd)"
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "================================================"
echo -e "  ${RED}卸载东芳美诊所管理系统${NC}"
echo "================================================"
echo ""
echo -e "${YELLOW}注意：此操作将删除所有数据和配置！${NC}"
echo ""
read -p "确定要卸载吗？(y/N): " CONFIRM
[ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ] && { echo "取消卸载"; exit 0; }

echo ""
echo -e "${YELLOW}正在停止服务...${NC}"
# Stop server
lsof -ti :8080 2>/dev/null | xargs kill -9 2>/dev/null || true

# Stop PostgreSQL
PG_HOME="$HOME/.clinic-mgmt"
if [ -f "$PG_HOME/pg-path.txt" ]; then
  PG_DIR=$(cat "$PG_HOME/pg-path.txt")
  PG_DATA="$PG_HOME/data"
  if [ -f "$PG_DIR/bin/pg_ctl" ] && [ -f "$PG_DATA/PG_VERSION" ]; then
    "$PG_DIR/bin/pg_ctl" -D "$PG_DATA" stop 2>/dev/null || true
  fi
fi
sleep 1

echo -e "${YELLOW}正在删除文件...${NC}"
# Remove app
rm -rf "/Applications/东芳美诊所管理系统.app" 2>/dev/null || true

# Remove data and config
rm -rf "$PG_HOME" 2>/dev/null || true

echo -e "${GREEN}[✓] 卸载完成${NC}"
echo ""
echo "如需重新安装，请重新打开安装包运行 install.command"
read -p "按回车退出..."
