# lazyrdm

## 介绍
**lazyrdm** 是一个终端形式下的 redis 管理工具，也可以理解为 tiny-rdm 的终端界面🙂。

项目地址： https://github.com/vst93/lazyrdm

![screenshot.png](https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/screenshot.png)

### 特性
- 基于 go 语言开发，跨平台支持，意味着支持 **Termux** 下使用
- 底层服务直接引用开源项目 tiny-rdm 项目( https://github.com/tiny-craft/tiny-rdm )，意味着如果你正在使用 tiny-rdm 管理 redis ，那么 lazyrdm 可以直接使用连接配置，同时两边的调整同步（因为读取和使用的同一个配置文件）
- 使用 gocui ( https://github.com/awesome-gocui/gocui ) 绘制界面


## 说明
- 已完成基本的功能使用
- 由于 https://github.com/awesome-gocui/gocui 和 https://github.com/jroimartin/gocui 都基本停止维护，复杂交互难以实现
- windows 系统下建议在有  Windows Terminal 的 windows11 下使用，CMD 下显示效果很差
- macos 中的 arm 版本未经测试，如果不能使用请反馈，然后尝试 amd 版本

### 安装与卸载
``` bash
# brew 
# 安装 
brew install vst93/tap/lazyrdm
# 卸载 
brew uninstall lazyrdm


# shell 
# 安装 
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"

```

--------------------------------------

## English Introduction

**lazyrdm** is a Redis management tool designed for the terminal, which can also be thought of as the command-line interface version of **tiny-rdm** 🙂.

Project Address: https://github.com/vst93/lazyrdm

### Features
- Developed in Go, with cross-platform support, meaning it is compatible with **Termux**.
- The underlying service directly utilizes the open-source project **tiny-rdm** (https://github.com/tiny-craft/tiny-rdm). If you are already using **tiny-rdm** to manage Redis, **lazyrdm** can directly use the same connection configurations, and changes will sync between both tools (as they read from and use the same configuration file).
- The UI is built with **gocui** (https://github.com/awesome-gocui/gocui).

## Notes
- Basic functionality has been implemented.
- Due to limited maintenance of both https://github.com/awesome-gocui/gocui and https://github.com/jroimartin/gocui, implementing complex interactions is challenging.
- On Windows, it is recommended to use **lazyrdm** with **Windows Terminal** on Windows 11 for the best experience, as the display performance in CMD is poor.
- The ARM version for macOS has not been thoroughly tested. If it does not work, please provide feedback and try the AMD version instead.

### Install & Uninstall
``` bash
# brew 
# install 
brew install vst93/tap/lazyrdm
# uninstall 
brew uninstall lazyrdm


# shell 
# install 
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"

```