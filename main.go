package main

import (
	"log"
	"os"
	"tinyrdm-tui/service"
	"tinyrdm/backend/services"

	"github.com/jroimartin/gocui"
)

func main() {
	// 设置三方包的日志输出为 /dev/null
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	defer devNull.Close()
	log.SetOutput(devNull)

	g, err := gocui.NewGui(gocui.OutputNormal)
	g.Mouse = true
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		services.Browser().Stop()
		g.Close()
	}()

	service.NewMainApp(g)

	// g.SetManagerFunc(service.Layout)

	// 退出程序
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// 切换视图（板块）
	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// 切换视图（板块）
		if len(service.GlobalApp.ViewNameList) < 2 {
			return nil
		}
		// service.PrintLn(service.GlobalApp.ViewNameList)
		// service.PrintLn(service.GlobalApp.CurrentView)

		currentViewNameIndex := -1
		for i, name := range service.GlobalApp.ViewNameList {
			if name == service.GlobalApp.CurrentView {
				currentViewNameIndex = i
				break
			}
		}
		currentViewNameIndex++
		if currentViewNameIndex >= len(service.GlobalApp.ViewNameList) {
			currentViewNameIndex = 0
		}
		nextViewName := service.GlobalApp.ViewNameList[currentViewNameIndex]
		if _, err := g.SetCurrentView(nextViewName); err == nil {
			service.GlobalApp.CurrentView = nextViewName
		}
		return nil
	})

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
