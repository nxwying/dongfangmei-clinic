东芳美诊所 - 授权文件生成工具 v2.0

功能：为东芳美诊所管理系统生成授权文件（license.json）

使用方式：
  双击运行本程序，自动打开浏览器
  在页面中输入客户机器码 -> 客户名称 -> 过期日期 -> 生成授权文件
  下载授权文件后发给客户，在软件中导入激活

跨平台：
  - macOS: 双击 license-generator-darwin-arm64
  - Linux: 终端运行 ./license-generator-linux-amd64
  - Windows: 双击 license-generator-windows-amd64.exe

默认端口: 9090（可通过环境变量 PORT 修改）

注意：请妥善保管私钥（已内置于程序中），丢失后无法再生成授权
