package service

import (
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

var SPACE_STRING = "                                                                                                                             "
var GlobalApp *MainApp
var GlobalConnection struct {
	dbs     []types.ConnectionDB
	view    int
	lastDB  int
	version string
}
var GlobalConnectionComponent *LTRConnectionComponent
var GlobalDBComponent *LTRListDBComponent

type MainApp struct {
	gui        *gocui.Gui
	maxX, maxY int
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
	// GlobalDBComponent = InitDBComponent().Layout()
}
