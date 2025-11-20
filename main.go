package main

import (
	"fmt"
	"lazyrdm/service"
	"log"
	"os"
	"strings"
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
	g.SelFrameColor = gocui.ColorGreen
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

	// 查看当前快捷键信息
	service.GuiSetKeysbindingConfirmWithVIEditor(g, "", []any{'?'}, "", func() string {
		infoText := service.GlobalTipComponent.GetLastTipString()
		infoText = strings.ReplaceAll(infoText, " | ", "\n")
		infoText = fmt.Sprintf("Shortcut Keys Reference (View Only)\n----------------------------------\n%s", infoText)
		return infoText
	}, nil, nil, true)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
