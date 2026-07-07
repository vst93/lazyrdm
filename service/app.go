package service

import (
	"fmt"
	"time"
	"tinyrdm/backend/services"

	"github.com/awesome-gocui/gocui"
)

var SPACE_STRING = "                                                                                                                             "
var GlobalApp *MainApp
var GlobalConnectionComponent *LTRConnectionComponent
var GlobalDBComponent *LTRListDBComponent
var GlobalKeyComponent *LTRListKeyComponent
var GlobalKeyInfoComponent *LTRKeyInfoComponent
var GlobalKeyInfoDetailComponent *LTRKeyInfoDetailComponent
var GlobalTipComponent *LTRTipComponent

type MainApp struct {
	Gui            *gocui.Gui
	maxX, maxY     int
	ViewNameList   []string
	watchingResize bool
}

func CurrentViewName() string {
	if GlobalApp == nil || GlobalApp.Gui == nil {
		return ""
	}
	currentView := GlobalApp.Gui.CurrentView()
	if currentView == nil {
		return ""
	}
	return currentView.Name()
}

func SetViewSafe(name string, x0 int, y0 int, x1 int, y1 int, overwrite byte) (*gocui.View, error) {
	if GlobalApp == nil || GlobalApp.Gui == nil {
		return nil, fmt.Errorf("gui not initialized")
	}
	maxX := GlobalApp.maxX
	maxY := GlobalApp.maxY
	if maxX < 2 || maxY < 2 {
		return nil, fmt.Errorf("terminal too small")
	}
	maxContentY := maxY - 1
	if name != "key_map_tip" && maxY >= 6 {
		maxContentY = maxY - 2
	}
	if maxContentY < 1 {
		maxContentY = maxY - 1
	}
	y0Max := maxContentY - 1
	y1Max := maxContentY
	if name == "key_map_tip" {
		y0Max = maxY - 1
		y1Max = maxY
	}

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x0 > maxX-2 {
		x0 = maxX - 2
	}
	if y0 > y0Max {
		y0 = y0Max
	}

	if x1 < x0+1 {
		x1 = x0 + 1
	}
	if y1 < y0+1 {
		y1 = y0 + 1
	}
	if x1 > maxX-1 {
		x1 = maxX - 1
	}
	if y1 > y1Max {
		y1 = y1Max
	}
	if x1 <= x0 || y1 <= y0 {
		return nil, fmt.Errorf("invalid point")
	}

	return GlobalApp.Gui.SetView(name, x0, y0, x1, y1, overwrite)
}

func NewMainApp(g *gocui.Gui) {
	g.InputEsc = true
	GlobalApp = &MainApp{
		Gui:          g,
		maxX:         0,
		maxY:         0,
		ViewNameList: []string{},
	}

	// GlobalApp = &mainApp
	GlobalApp.maxX, GlobalApp.maxY = GlobalApp.Gui.Size()
	GlobalApp.Gui.FrameColor = gocui.ColorCyan
	GlobalApp.Gui.SelFrameColor = themeFrameActive
	GlobalApp.Gui.SelFgColor = gocui.ColorGreen
	GlobalApp.Gui.Highlight = true
	GlobalApp.Gui.SetManagerFunc(GlobalApp.LayoutManager)
	// GlobalApp.Gui.Cursor = true
	InitTipComponent()
	InitConnectionComponent()
}

func (app *MainApp) ForceUpdate(setViewName string) {
	app.applyViewLayout(setViewName, true)
	GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
}

func (app *MainApp) applyViewLayout(setViewName string, setCurrent bool) {
	if setViewName == "" {
		return
	}
	if setCurrent {
		if _, err := GlobalApp.Gui.SetCurrentView(setViewName); err != nil {
			return
		}
	}
	switch setViewName {
	case "connection_list":
		if GlobalConnectionComponent != nil {
			GlobalConnectionComponent.Layout()
		}
	default:
		if GlobalDBComponent != nil {
			GlobalDBComponent.Layout()
		}
		if GlobalKeyComponent != nil {
			GlobalKeyComponent.Layout()
		}
		if GlobalKeyInfoComponent != nil {
			GlobalKeyInfoComponent.LayoutTitle()
		}
		if GlobalKeyInfoDetailComponent != nil {
			GlobalKeyInfoDetailComponent.LayoutTitle()
		}
	}
	if GlobalTipComponent != nil {
		GlobalTipComponent.LayComponentTips()
	}
}

func (app *MainApp) LayoutManager(g *gocui.Gui) error {
	if app == nil || g == nil {
		return nil
	}
	maxX, maxY := g.Size()
	if maxX == app.maxX && maxY == app.maxY {
		if GlobalTipComponent != nil {
			GlobalTipComponent.LayComponentTips()
		}
		return nil
	}
	app.maxX = maxX
	app.maxY = maxY

	if maxX < 20 || maxY < 8 {
		return nil
	}

	currentViewName := CurrentViewName()
	if currentViewName == "" {
		return nil
	}

	switch currentViewName {
	case "page_help":
		if GlobalHelpPageComponent != nil {
			GlobalHelpPageComponent.Layout()
		}
	case "page_server_info":
		if GlobalServerInfoPageComponent != nil {
			GlobalServerInfoPageComponent.Layout()
		}
	case "page_console":
		if GlobalConsoleComponent != nil {
			GlobalConsoleComponent.Layout()
		}
	case "page_confirm":
		return nil
	case "connection_list":
		if GlobalConnectionComponent != nil {
			GlobalConnectionComponent.Layout()
		}
	default:
		if GlobalDBComponent != nil {
			GlobalDBComponent.Layout()
		}
		if GlobalKeyComponent != nil {
			GlobalKeyComponent.Layout()
		}
		if GlobalKeyInfoComponent != nil {
			GlobalKeyInfoComponent.Layout()
		}
		if GlobalKeyInfoDetailComponent != nil {
			GlobalKeyInfoDetailComponent.Layout()
		}
		if GlobalTipComponent != nil {
			GlobalTipComponent.LayComponentTips()
		}
	}
	return nil
}

func (app *MainApp) StartResizeWatcher() {
	if app == nil || app.Gui == nil || app.watchingResize {
		return
	}
	app.watchingResize = true
	go func() {
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if app == nil || app.Gui == nil {
				return
			}
			maxX, maxY := app.Gui.Size()
			if maxX == app.maxX && maxY == app.maxY {
				if GlobalTipComponent != nil {
					app.Gui.Update(func(g *gocui.Gui) error {
						GlobalTipComponent.LayComponentTips()
						return nil
					})
				}
				continue
			}
			app.maxX = maxX
			app.maxY = maxY
			if maxX < 20 || maxY < 8 {
				continue
			}
			app.Gui.Update(func(g *gocui.Gui) error {
				app.relayoutCurrentViewOnResize()
				return nil
			})
		}
	}()
}

func (app *MainApp) relayoutCurrentViewOnResize() {
	currentViewName := CurrentViewName()
	if currentViewName == "" {
		return
	}
	switch currentViewName {
	case "page_help":
		if GlobalHelpPageComponent != nil {
			GlobalHelpPageComponent.Layout()
		}
	case "page_server_info":
		if GlobalServerInfoPageComponent != nil {
			GlobalServerInfoPageComponent.Layout()
		}
	case "page_console":
		if GlobalConsoleComponent != nil {
			GlobalConsoleComponent.Layout()
		}
	case "page_confirm":
		return
	case "connection_list":
		if GlobalConnectionComponent != nil {
			GlobalConnectionComponent.Layout()
		}
	default:
		if GlobalDBComponent != nil {
			GlobalDBComponent.Layout()
		}
		if GlobalKeyComponent != nil {
			GlobalKeyComponent.Layout()
		}
		if GlobalKeyInfoComponent != nil {
			GlobalKeyInfoComponent.Layout()
		}
		if GlobalKeyInfoDetailComponent != nil {
			GlobalKeyInfoDetailComponent.Layout()
		}
		if GlobalTipComponent != nil {
			GlobalTipComponent.LayComponentTips()
		}
	}
}

func ExitCurrentConnectionToList() {
	if GlobalApp == nil || GlobalApp.Gui == nil || GlobalConnectionComponent == nil {
		return
	}

	connectionName := GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name
	if connectionName == "" {
		return
	}

	services.Browser().CloseConnection(connectionName)

	for _, viewName := range GlobalApp.ViewNameList {
		GlobalApp.Gui.DeleteView(viewName)
		GlobalApp.Gui.DeleteKeybindings(viewName)
	}

	auxiliaryViews := []string{"key_list_line", "key_info_ttl", "key_detail_line", "key_value_format", "search_key"}
	for _, viewName := range auxiliaryViews {
		GlobalApp.Gui.DeleteView(viewName)
		GlobalApp.Gui.DeleteKeybindings(viewName)
	}

	GlobalApp.ViewNameList = []string{}
	InitConnectionComponent()
	GlobalTipComponent.LayoutTemporary("Disconnected. Switched to connection list.", 2, TipTypeSuccess)
}
