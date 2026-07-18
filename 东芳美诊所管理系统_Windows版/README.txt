东芳美诊所管理系统 v1.3 — Windows 版
=====================================

安装前提
--------
本系统依赖 PostgreSQL 16 数据库。如果尚未安装：

  1. 下载安装：https://www.postgresql.org/download/windows/
  2. 安装时密码设为 clinic123，端口保留 5432

安装步骤
--------
方法 A（推荐）：双击 install.ps1（右键 → 使用 PowerShell 运行）
  脚本会自动：检测 PostgreSQL → 创建数据库 → 复制文件 → 创建桌面快捷方式 → 配置防火墙

方法 B：手动安装
  1. 安装 PostgreSQL 16（密码 clinic123）
  2. 以管理员身份运行命令提示符：
     "C:\Program Files\PostgreSQL\16\bin\psql" -U postgres -c "CREATE USER clinic WITH PASSWORD 'clinic123';"
     "C:\Program Files\PostgreSQL\16\bin\createdb" clinic
     "C:\Program Files\PostgreSQL\16\bin\psql" -U postgres -c "ALTER DATABASE clinic OWNER TO clinic;"
  3. 双击「启动.bat」

登录信息
--------
  地址：http://localhost:8080
  账号：admin / 密码：admin123

功能模块
--------
  工作台 · 客户管理 · 预约管理 · 收银台 · 退款管理 · 会员管理
  库存管理 · 证件档案 · 数据中心 · 绩效中心 · 回访管理
  病历管理 · 营销工具 · 支出管理 · 系统设置
  运营分析 · KPI目标 · 排行榜 · 报表中心

局域网使用
----------
  同一局域网内其他电脑可通过 http://电脑IP:8080 访问。
  安装脚本会自动配置 Windows 防火墙放行 8080 端口。

系统要求
--------
  Windows 10 / 11（64位）
  PostgreSQL 16
  4GB 内存 / 500MB 硬盘
