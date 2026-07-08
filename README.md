# lazyrdm

[English](#english) | [中文](#中文)

![screenshot.png](https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/screenshot.png)

---

## English

**lazyrdm** is a terminal-based Redis management tool — think of it as the TUI version of [tiny-rdm](https://github.com/tiny-craft/tiny-rdm) 🙂

**Project:** https://github.com/vst93/lazyrdm

### Features

- **Cross-platform** — Built in Go, runs on Linux, macOS, Windows, and **Termux** (Android)
- **Shared config with tiny-rdm** — Directly uses tiny-rdm's connection profiles; changes sync between both tools
- **Full Redis type support** — String, List, Hash, Set, ZSet, Stream, and JSON (RedisJSON)
- **Structured detail view** — Tabular rendering for collection types with inline filtering, row selection, and a detail pane
- **Format switching** — Toggle between Raw / JSON / Unicode JSON display formats with `<f>`
- **Key management** — Create keys with selectable type, rename, edit TTL, delete with confirmation
- **Inline editing** — Add/edit/delete entries in collections via dialog forms (no external editor needed)
- **Redis console** — Execute commands with history navigation and formatted output
- **Cross-platform paste** — `Ctrl+V` paste support in all input fields
- Built with [gocui](https://github.com/awesome-gocui/gocui)

### Screenshots

The main interface shows:
- **DB list** — Select databases with key counts
- **Key list** — Browse, search, filter by type, create and delete keys
- **Info bar** — Key type, size, length, and TTL
- **Detail pane** — Value display with format switching, structured tables for collections

### Install

**Homebrew:**
```bash
brew install vst93/tap/lazyrdm
```

**Shell script:**
```bash
# Install latest release
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"

# Uninstall
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)" uninstall

# Check version
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)" version
```

**Build from source:**
```bash
git clone https://github.com/vst93/lazyrdm.git
cd lazyrdm
go build -o lazyrdm .
# Version will be "dev" for local builds
```

### Version Display

The version number is shown in the connection list title bar. Release binaries have the version injected at build time (e.g. `v1.2`). Local development builds show `vdev`.

Check version from command line:
```bash
lazyrdm --version
```

### Notes

- Windows: use **Windows Terminal** on Windows 11 for best rendering (CMD has poor display)
- macOS ARM: untested — if it doesn't work, try the AMD64 build
- Both [gocui](https://github.com/awesome-gocui/gocui) and [jroimartin/gocui](https://github.com/jroimartin/gocui) are essentially unmaintained, so complex interactions are limited

### Uninstall

```bash
# Homebrew
brew uninstall lazyrdm

# Manual: remove the binary from your PATH
```

---

## 中文

**lazyrdm** 是一个终端下的 Redis 管理工具，可以理解为 [tiny-rdm](https://github.com/tiny-craft/tiny-rdm) 的终端界面版 🙂

**项目地址：** https://github.com/vst93/lazyrdm

### 特性

- **跨平台** — Go 语言开发，支持 Linux、macOS、Windows，以及 **Termux**（Android）
- **与 tiny-rdm 共享配置** — 直接读取 tiny-rdm 的连接配置文件，两边修改自动同步
- **全类型支持** — String、List、Hash、Set、ZSet、Stream、JSON（RedisJSON）
- **结构化详情** — 集合类型以表格形式展示，支持行内过滤、行选择、详情面板
- **格式切换** — `<f>` 在 Raw / JSON / Unicode JSON 之间切换显示格式
- **Key 管理** — 新建 key 可选类型、重命名、修改 TTL、确认删除
- **内联编辑** — 通过弹窗表单增删改集合元素，无需外部编辑器
- **Redis 控制台** — 执行命令，支持历史导航和格式化输出
- **跨平台粘贴** — 所有输入框支持 `Ctrl+V` 粘贴
- 基于 [gocui](https://github.com/awesome-gocui/gocui) 绘制界面

### 安装

**Homebrew：**
```bash
brew install vst93/tap/lazyrdm
```

**脚本安装：**
```bash
# 安装最新版本
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"

# 卸载
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)" uninstall

# 查看版本信息
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)" version
```

**源码编译：**
```bash
git clone https://github.com/vst93/lazyrdm.git
cd lazyrdm
go build -o lazyrdm .
# 本地编译版本号显示为 dev
```

### 版本说明

版本号显示在连接列表标题栏。Release 版本在编译时注入版本号（如 `v1.2`），本地开发编译显示 `vdev`。

命令行查看版本：
```bash
lazyrdm --version
```

### 说明

- Windows 下建议使用 **Windows Terminal**（Windows 11），CMD 显示效果较差
- macOS ARM 版本未经测试，如不可用请尝试 AMD 版本
- [gocui](https://github.com/awesome-gocui/gocui) 和 [jroimartin/gocui](https://github.com/jroimartin/gocui) 均已停止维护，复杂交互实现受限

### 卸载

```bash
# Homebrew
brew uninstall lazyrdm

# 脚本卸载
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)" uninstall
```
