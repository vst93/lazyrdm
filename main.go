package main

import (
	"lazyrdm/service"
	"log"
	"os"
	"tinyrdm/backend/services"

	"github.com/awesome-gocui/gocui"
)

func main() {

	// 设置三方包的日志输出为 /dev/null
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	defer devNull.Close()
	log.SetOutput(devNull)

	g, err := gocui.NewGui(gocui.OutputNormal, false)
	g.Mouse = true
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		services.Browser().Stop()
		g.Close()
	}()

	service.NewMainApp(g)

	// 退出程序
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// service.GuiSetKeysbinding(g, "", []any{gocui.KeyCtrlQ}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if service.GlobalApp.Gui.CurrentView().Name() == "connection_list" {
	// 		return nil
	// 	}
	// 	service.GlobalApp.Gui.SetCurrentView(service.GlobalConnectionComponent.Name)
	// 	service.GlobalApp.ViewNameList = []string{} // 清空视图列表
	// 	service.InitConnectionComponent()
	// 	return nil
	// })

	// 切换视图（板块）
	service.GuiSetKeysbinding(g, "", []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// 切换视图（板块）
		if len(service.GlobalApp.ViewNameList) < 2 {
			return nil
		}
		currentViewNameIndex := -1
		for i, name := range service.GlobalApp.ViewNameList {
			if name == service.GlobalApp.Gui.CurrentView().Name() {
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

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
