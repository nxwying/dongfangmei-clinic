#!/bin/bash
MYDIR="$(cd "$(dirname "$0")" && pwd)"
APP_BIN="$MYDIR/东芳美诊所管理系统.app/Contents/MacOS/clinic-server"
APP_HOME="$HOME/.clinic-mgmt"
PORT="8080"

[ ! -f "$APP_BIN" ] && { echo "❌ 未找到程序文件"; exit 1; }

# Stop old
lsof -ti :$PORT 2>/dev/null | xargs kill -9 2>/dev/null
launchctl remove com.clinic.server 2>/dev/null
sleep 1

# Create plist and load via launchctl
mkdir -p "$APP_HOME/data"
cat > /tmp/com.clinic.server.plist << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clinic.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>$APP_BIN</string>
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
echo "系统正在启动..."
sleep 3

if lsof -ti :$PORT >/dev/null 2>&1; then
  echo "✓ 系统已启动！"
  open "http://localhost:$PORT/"
else
  echo "✗ 启动失败，查看日志: $APP_HOME/server.log"
fi
