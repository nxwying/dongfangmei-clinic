#!/bin/bash
cd "$(dirname "$0")"
DIR="$(cd "$(dirname "$0")" && pwd)"
echo "================================================"
echo "  东芳美诊所 — 授权码生成工具"
echo "================================================"
echo ""
# Find the binary - try app bundle first, then current directory
BIN="$DIR/授权码生成工具_v2.0/授权码生成器.app/Contents/MacOS/授权码生成器"
if [ ! -f "$BIN" ]; then
  BIN="$DIR/license-maker"
fi
if [ ! -f "$BIN" ]; then
  BIN="$DIR/授权码生成工具_v2.0/license-maker"
fi
if [ ! -f "$BIN" ]; then
  echo "❌ 找不到程序文件"
  echo "   请解压 授权码生成工具_v2.0.zip 后重试"
  echo ""
  read -p "按回车键退出..."
  exit 1
fi

chmod +x "$BIN"
"$BIN"
read -p "按回车键退出..."
