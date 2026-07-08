package main

import (
	"flag"
	"fmt"
	"lazyrdm/service"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"tinyrdm/backend/services"

	"github.com/awesome-gocui/gocui"
)

func main() {
	// 命令行参数解析
	showVersion := flag.Bool("version", false, "print version and exit")
	showVersionShort := flag.Bool("v", false, "print version and exit (shorthand)")
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Println(service.APP_VERSION)
		return
	}

	// 只在 Windows 平台检查
	if runtime.GOOS == "windows" {
		checkAndRelaunchInWT()
	}

	// 设置三方包的日志输出为 /dev/null
	if devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666); err == nil {
		defer devNull.Close()
		log.SetOutput(devNull)
	}

	g, err := gocui.NewGui(gocui.OutputNormal, false)
	g.Mouse = true
	g.SelFrameColor = gocui.ColorGreen
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		services.Browser().Stop()
		g.Close()
	}()

	service.NewMainApp(g)
	service.GlobalApp.StartResizeWatcher()

	// 退出程序
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// 切换视图（板块）
	service.GuiSetKeysbinding(g, "", []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// 切换视图（板块）
		if len(service.GlobalApp.ViewNameList) < 2 {
			return nil
		}
		currentView := service.GlobalApp.Gui.CurrentView()
		if currentView == nil || currentView.Name() == "page_confirm" {
			return nil
		}
		currentViewNameIndex := -1
		for i, name := range service.GlobalApp.ViewNameList {
			if name == currentView.Name() {
				currentViewNameIndex = i
				break
			}
		}
		currentViewNameIndex++
		if currentViewNameIndex >= len(service.GlobalApp.ViewNameList) {
			currentViewNameIndex = 0
		}
		nextViewName := service.GlobalApp.ViewNameList[currentViewNameIndex]
		service.GlobalApp.ForceUpdate(nextViewName)
		// service.GlobalTipComponent.Layout("")
		return nil
	})

	service.GuiSetKeysbinding(g, "", []any{gocui.KeyCtrlW}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		currentView := service.GlobalApp.Gui.CurrentView()
		if currentView == nil {
			return nil
		}
		if currentView.Name() == "connection_list" || currentView.Name() == "page_confirm" {
			return nil
		}
		service.ExitCurrentConnectionToList()
		return nil
	})

	service.GuiSetKeysbinding(g, "", []any{'?'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		service.OpenHelpPage()
		return nil
	})

	// 发送一个访问统计, 仅用于统计使用情况
	// go http.Get("https://finicounter.eu.org/counter?host=github.com/vst93/lazyrdm")
	go service.SendAppStats()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	if service.IsConfirmModalActive(g) {
		_, _ = g.SetCurrentView("page_confirm")
		return nil
	}
	return gocui.ErrQuit
}

func checkAndRelaunchInWT() {
	// 如果已经在 WT 中，直接返回
	if os.Getenv("WT_SESSION") != "" {
		return
	}

	// 简单检测：如果是双击启动，尝试在 WT 中重启
	if len(os.Args) == 1 { // 没有命令行参数，可能是双击启动
		relaunchInWindowsTerminal()
	}
}

func relaunchInWindowsTerminal() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// 检查 wt.exe 是否存在
	_, err = exec.LookPath("wt.exe")
	if err != nil {
		// Windows Terminal 未安装或不在 PATH 中
		fmt.Println("Windows Terminal 未安装或不在 PATH 中")
		return
	}

	// 使用绝对路径指向可执行文件
	absExePath := filepath.Join(exeDir, exeName)

	// 尝试在 Windows Terminal 中启动
	cmd := exec.Command("wt.exe", "-d", exeDir, absExePath)

	if err := cmd.Start(); err == nil {
		fmt.Println("在 Windows Terminal 中启动应用程序...")
		os.Exit(0)
	} else {
		fmt.Printf("无法启动 Windows Terminal，将在当前窗口运行: %v\n", err)
	}
}
