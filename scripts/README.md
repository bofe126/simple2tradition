# Windows 安装包构建

本目录提供两个一键构建脚本，会自动：

1. 清理 `build\bin`
2. 安装/更新前端依赖并构建前端 (`npm install` + `npm run build`)
3. 调用 `wails build -nsis -platform windows/amd64 -clean` 编译 Go 程序
4. 调用 NSIS 把上一步的可执行文件封装成安装包

## 前置环境

| 工具           | 安装方式                                                                              |
| -------------- | ------------------------------------------------------------------------------------- |
| Go (≥ 1.23)    | <https://go.dev/dl/>                                                                  |
| Wails CLI      | `go install github.com/wailsapp/wails/v2/cmd/wails@latest`                            |
| Node.js (≥ 18) | <https://nodejs.org/>                                                                 |
| NSIS 3         | <https://nsis.sourceforge.io/Download> — 安装时勾选「Add NSIS to the system PATH」    |
| WebView2 Runtime | Windows 10+ 自带；若要支持更老系统，安装包会自动调用 bootstrapper 下载                |

## 使用方法

### PowerShell（推荐）

```powershell
# 在项目根目录
.\scripts\build-windows.ps1
```

### cmd / 双击运行

```bat
:: 在项目根目录
scripts\build-windows.bat
```

## 产物

执行成功后会在 `build\bin\` 下生成：

- `simple2tradition-amd64-installer.exe` — **分发的安装包**
- `simple2tradition.exe` — 裸的可执行文件（不带安装器）

## 安装包行为

- 默认安装到 `%ProgramFiles%\bofe126\Simple2Tradition`
- 创建开始菜单快捷方式和桌面快捷方式
- 安装结束页有一个 **"Launch Simple2Tradition after install"** 复选框，默认未勾选
- 如果用户机器上没装 WebView2 Runtime，会静默调用官方 bootstrapper 拉取
- 卸载时清理 `%AppData%` 下的 WebView2 用户数据
- 在「应用和功能」中有正常的卸载入口

## 修改安装器文本

安装器的 UI 文本在 `build\windows\installer\project.nsi` 顶部定义（`MUI_*` 系列宏）。
**重要：** 这台机器上的 NSIS 是 ANSI 版本，会把脚本当成 GBK/CP936 解析，因此脚本中
所有字符串字面量必须是 ASCII。若要中文 UI，请把 `MUI_LANGUAGE "SimpChinese"` 那行
取消注释，并新增一个 `SimpChinese.nsh` 用 `LangString` 覆盖默认文本。

## 手动 NSIS 调试

若只想重新打包而不重编 Go 程序：

```bat
cd build\windows\installer
makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\simple2tradition.exe project.nsi
```

## 同时打包 amd64 + arm64

修改 `wails build` 那一行为：

```bash
wails build -nsis -platform windows/amd64,windows/arm64 -clean
```

会同时输出 `simple2tradition-amd64-arm64-installer.exe`（一份安装器同时支持两种架构）。
