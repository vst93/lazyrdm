package service

import (
	"github.com/jroimartin/gocui"
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
