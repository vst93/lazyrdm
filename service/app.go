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

type MainApp struct {
	gui          *gocui.Gui
	maxX, maxY   int
	CurrentView  string
	ViewNameList []string
}

func NewMainApp(g *gocui.Gui) {
	mainApp := MainApp{
		gui:  g,
		maxX: 0,
		maxY: 0,
	}

	GlobalApp = &mainApp
	mainApp.maxX, mainApp.maxY = GlobalApp.gui.Size()
	GlobalConnectionComponent = InitConnectionComponent()
	GlobalApp.gui.SelFgColor = gocui.ColorGreen
	GlobalApp.gui.Highlight = true
}
