package service

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type LTRListDBComponent struct {
	LayoutMaxY int
	SelectedDB int
	view       *gocui.View
}

func InitDBComponent() *LTRListDBComponent {
	c := &LTRListDBComponent{
		LayoutMaxY: 0,
		SelectedDB: 0,
		view:       nil,
	}
	c.Layout()
	return c
}

func (c *LTRListDBComponent) Layout() *LTRListDBComponent {
	if len(GlobalConnection.dbs) == 0 {
		return c
	}
	// maxX, maxY := GlobalApp.gui.Size()
	v, err := GlobalApp.gui.SetView("db_list", 0, 0, GlobalApp.maxX*2/10, GlobalApp.maxY*2/10)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = "DB"
		v.Editable = false
		v.Frame = true
		_, c.LayoutMaxY = v.Size()
	}
	printString := ""
	currenLine := 0
	totalLine := 0
	for index, db := range GlobalConnection.dbs {
		if c.SelectedDB == index {
			printString += fmt.Sprintf("\x1b[1;37;44m%s\x1b[0m\n", ""+db.Name+""+SPACE_STRING)
			totalLine++
			currenLine = totalLine
		} else {
			printString += fmt.Sprintf("%s\n", ""+db.Name+""+SPACE_STRING)
			totalLine++
		}
	}
	if currenLine > c.LayoutMaxY/2 {
		originLine := currenLine - c.LayoutMaxY/2
		if originLine < 0 {
			originLine = 0
		}
		if originLine > totalLine-c.LayoutMaxY {
			originLine = totalLine - c.LayoutMaxY
		}
		v.SetOrigin(0, originLine)
	} else {
		v.SetOrigin(0, 0)
	}
	v.Clear()
	v.Write([]byte(printString))
	c.view = v
	return c
}
