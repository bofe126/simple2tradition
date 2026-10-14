# 简繁

剪贴板简繁体中文自动转换小工具。复制即转换，支持 **简 → 繁** 与 **繁 → 简** 双向切换。

<p align="center">
  <img src="build/appicon.png" width="128" alt="简繁 图标"/>
</p>

## 功能

- 后台监听剪贴板，复制简体（或繁体）中文后自动转换并写回
- 一键切换转换方向，界面配色、Dock 图标、菜单栏/托盘标题随方向联动
- 主窗口展示最近一次转换的原文与结果，以及历史记录（可一键复制）
- macOS 菜单栏图标 / Windows 系统托盘：显示主窗口、启用/暂停、退出
- 可选「仅预览」模式：只显示转换结果，不修改剪贴板

## 安装

前往 [Releases](https://github.com/bofe126/simple2tradition/releases) 下载：

| 平台 | 文件 |
| --- | --- |
| macOS（Apple Silicon / Intel） | `jianfan-macos-universal.zip` — 解压后将 `简繁.app` 拖入「应用程序」 |
| Windows 10+（amd64） | `jianfan-amd64-installer.exe` |

> **macOS 首次打开**：应用未经 Apple 公证，若提示"无法打开"，请右键 → 打开，或执行
> `xattr -d com.apple.quarantine /Applications/简繁.app`

## 使用

1. 启动后应用常驻菜单栏 / 托盘，默认方向为「简 → 繁」
2. 复制一段中文，剪贴板内容会在半秒内自动替换为转换结果
3. 点击主窗口右上角的方向胶囊（如「简 → 繁」）即可反转方向
4. 关闭窗口不会退出，macOS 用 `Cmd+Q` 或菜单栏「退出」，Windows 用托盘「退出」

## 从源码构建

依赖：Go ≥ 1.23、Node.js ≥ 18、[Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

```bash
# 开发模式
wails dev

# macOS 通用包
wails build -platform darwin/universal

# Windows 安装包（在 Windows 上，需 NSIS），详见 scripts/README.md
.\scripts\build-windows.ps1
```

产物位于 `build/bin/`。

## 技术栈

[Wails v2](https://wails.io) · Go · Svelte · [gojianfan](https://github.com/siongui/gojianfan)（转换引擎）· [fyne.io/systray](https://github.com/fyne-io/systray)（macOS 菜单栏）· [getlantern/systray](https://github.com/getlantern/systray)（Windows 托盘）
