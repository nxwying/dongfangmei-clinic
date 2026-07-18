#!/bin/bash
# 医美诊所管理系统 — Terminal 安装脚本
# 在 Terminal 中运行：
#   cd /path/to/clinic-mgmt-dist
#   chmod +x install.sh && ./install.sh

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

echo "================================================"
echo "  医美诊所管理系统 v1.3 - 安装程序"
echo "================================================"
echo ""

# Remove quarantine from all files first
if command -v xattr &>/dev/null; then
    xattr -rd com.apple.quarantine "$DIR" 2>/dev/null || true
fi

exec "$DIR/install.command"
