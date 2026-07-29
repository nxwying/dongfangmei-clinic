#!/bin/bash
set -e
MYDIR="$(cd "$(dirname "$0")" && pwd)"

APP_NAME="东芳美诊所管理系统.app"
APP_SOURCE="$MYDIR/$APP_NAME"
APP_TARGET="/Applications/$APP_NAME"
APP_HOME="$HOME/.clinic-mgmt"
PORT="8080"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "================================================"
echo -e "  ${GREEN}东芳美诊所管理系统 v1.3 -- 安装${NC}"
echo "================================================"
echo ""

[ "$(uname)" != "Darwin" ] && { echo -e "${RED}[✗] 仅支持 macOS${NC}"; exit 1; }
echo -e "${GREEN}[✓] macOS $(sw_vers -productVersion 2>/dev/null || echo "?")$(echo " / $(uname -m)")${NC}"

# Stop old server
echo ""
echo -e "${YELLOW}[1/3] 停止旧版本...${NC}"
lsof -ti :$PORT 2>/dev/null | xargs kill -9 2>/dev/null || true
launchctl remove com.clinic.server 2>/dev/null || true
sleep 1
echo -e "${GREEN}  [✓] 已停止${NC}"

# Copy to Applications
echo -e "${YELLOW}[2/3] 安装到应用程序目录...${NC}"
rm -rf "$APP_TARGET" 2>/dev/null || true
mkdir -p "$APP_HOME/data"
cp -R "$APP_SOURCE" "$APP_TARGET"
echo -e "${GREEN}  [✓] 已安装到 $APP_TARGET${NC}"

# Start server via launchctl
echo -e "${YELLOW}[3/3] 启动服务...${NC}"
cat > /tmp/com.clinic.server.plist << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clinic.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>$APP_TARGET/Contents/MacOS/clinic-server</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$APP_HOME</string>
    <key>KeepAlive</key>
    <false/>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$APP_HOME/server.log</string>
    <key>StandardErrorPath</key>
    <string>$APP_HOME/server.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>DB_DRIVER</key>
        <string>sqlite</string>
        <key>DB_DSN</key>
        <string>$APP_HOME/data/clinic.db</string>
    </dict>
</dict>
</plist>
PLISTEOF

launchctl load /tmp/com.clinic.server.plist
sleep 3

# Verify
lsof -ti :$PORT >/dev/null 2>&1 && echo -e "${GREEN}  [✓] 服务已启动${NC}" || echo -e "${RED}  [✗] 启动失败，查看日志: $APP_HOME/server.log${NC}"

# Create start/stop scripts
cat > "$APP_HOME/start.command" << 'SCEOF'
#!/bin/bash
launchctl load /tmp/com.clinic.server.plist 2>/dev/null
sleep 2
open "http://localhost:8080/"
SCEOF
chmod +x "$APP_HOME/start.command"

cat > "$APP_HOME/stop.command" << 'SCEOF'
#!/bin/bash
launchctl remove com.clinic.server 2>/dev/null
lsof -ti :8080 2>/dev/null | xargs kill -9 2>/dev/null
echo "服务已停止"
SCEOF
chmod +x "$APP_HOME/stop.command"

echo ""
echo "================================================"
echo -e "  ${GREEN}[✓] 安装完成！${NC}"
echo ""
echo "  服务已启动: http://localhost:$PORT/"
echo "  默认账号: admin    密码: admin123"
echo "  数据目录: $APP_HOME/data"
echo ""
echo "  启动: $APP_HOME/start.command"
echo "  停止: $APP_HOME/stop.command"
echo "================================================"
echo ""
open "http://localhost:$PORT/"
