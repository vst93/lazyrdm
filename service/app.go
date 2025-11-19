package service

import (
	"github.com/awesome-gocui/gocui"
)

var SPACE_STRING = "                                                                                                                             "
var GlobalApp *MainApp

//	var GlobalConnection struct {
//		dbs     []types.ConnectionDB
//		view    int
//		lastDB  int
//		version string
//	}
var GlobalConnectionComponent *LTRConnectionComponent
var GlobalDBComponent *LTRListDBComponent
var GlobalKeyComponent *LTRListKeyComponent
var GlobalKeyInfoComponent *LTRKeyInfoComponent
var GlobalKeyInfoDetailComponent *LTRKeyInfoDetailComponent
var GlobalTipComponent *LTRTipComponent

type MainApp struct {
	Gui        *gocui.Gui
	maxX, maxY int
	// CurrentView  string
	ViewNameList []string
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
	GlobalApp.Gui.SelFgColor = gocui.ColorGreen
	GlobalApp.Gui.Highlight = true
	// GlobalApp.Gui.Cursor = true
	InitTipComponent()
	InitConnectionComponent()
}

func (app *MainApp) ForceUpdate(setViewName string) {
	if setViewName != "" {
		GlobalApp.Gui.SetCurrentView(setViewName)
		switch setViewName {
		case "connection_list":
			GlobalConnectionComponent.Layout()
		default:
			GlobalDBComponent.Layout()
			GlobalKeyComponent.Layout()
			// GlobalKeyInfoComponent.Layout()
			// GlobalKeyInfoDetailComponent.Layout()
			GlobalKeyInfoComponent.LayoutTitle()
			GlobalKeyInfoDetailComponent.LayoutTitle()
		}
		GlobalTipComponent.LayComponentTips()
	}
	GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
}
