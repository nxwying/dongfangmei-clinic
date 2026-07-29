#!/bin/bash
set -e
APP_TARGET="/Applications/东芳美诊所管理系统.app"
APP_HOME="$HOME/.clinic-mgmt"
PORT="8080"

RED='\033[0;31m'; GREEN='\033[0;32m'; NC='\033[0m'

echo "================================================"
echo "  东芳美诊所管理系统 — 卸载"
echo "================================================"
echo ""

# Stop server
lsof -ti :$PORT 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 1

echo -e "${YELLOW}将删除以下内容:${NC}"
echo "  $APP_TARGET"
echo "  $APP_HOME"
echo ""

read -p "确认卸载？(y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
  echo "已取消"
  exit 0
fi

rm -rf "$APP_TARGET" 2>/dev/null || true
echo -e "${GREEN}  [✓] 已移除: $APP_TARGET${NC}"
rm -rf "$APP_HOME" 2>/dev/null || true
echo -e "${GREEN}  [✓] 已移除: $APP_HOME${NC}"

echo ""
echo -e "${GREEN}卸载完成${NC}"
