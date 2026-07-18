#!/bin/bash
# 构建安装包 — v1.2（含证件档案模块）
set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$DIR/.." && pwd)"

cd "$ROOT"

echo "================================================"
echo "  东芳美诊所管理系统 v1.2 — 构建安装包"
echo "================================================"
echo ""

# 1. Build frontend
echo "[1/4] 构建前端..."
cd web
npm install --silent 2>/dev/null
npx vite build
cd ..

# 2. Build Go binary for both architectures
echo "[2/4] 编译后端..."
rm -f package/server-arm64 package/server-amd64

# Apple Silicon (arm64)
GOCACHE=/private/tmp/go-cache GOARCH=arm64 GOOS=darwin go build -o package/server-arm64 .
echo "  ✓ arm64 编译完成"

# Intel (amd64)
GOCACHE=/private/tmp/go-cache GOARCH=amd64 GOOS=darwin go build -o package/server-amd64 .
echo "  ✓ amd64 编译完成"

# 3. Create universal binary
echo "[3/4] 创建通用二进制..."
lipo -create -output package/Clinic.app/Contents/MacOS/clinic-server \
  package/server-arm64 package/server-amd64
rm -f package/server-arm64 package/server-amd64
echo "  ✓ universal binary 已生成 ($(file package/Clinic.app/Contents/MacOS/clinic-server | grep -o 'Mach-O.*'))"

# 4. Create distribution zip
echo "[4/4] 打包分发..."
cd package
rm -f ../clinic-mgmt-dist.zip
zip -r ../clinic-mgmt-dist.zip Clinic.app install.command install.sh README.txt

echo ""
echo "================================================"
echo "  ✓ 安装包已生成"
echo "  📦 clinic-mgmt-dist.zip ($(du -h ../clinic-mgmt-dist.zip | cut -f1))"
echo "================================================"
echo ""
echo "分发方式："
echo "  把 clinic-mgmt-dist.zip 拷贝给客户"
echo "  解压后双击 install.command 即可安装"
echo ""
