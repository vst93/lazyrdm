package service

import (
	"fmt"
	"strconv"

	"github.com/jroimartin/gocui"
)

type LTRListDBComponent struct {
	name       string
	title      string
	LayoutMaxH int
	SelectedDB int // 当前打开数据库
	CurrenDB   int // 当前光标选中的据库
	view       *gocui.View
	minH       int
}

func InitDBComponent() {
	GlobalDBComponent = &LTRListDBComponent{
		name:       "db_list",
		title:      "DB",
		LayoutMaxH: 0,
		SelectedDB: -1,
		CurrenDB:   0,
		view:       nil,
		minH:       2,
	}

	GlobalDBComponent.Layout().KeyBind()
	InitKeyComponent()
	InitKeyInfoComponent()
	InitKeyInfoDetailComponent()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalDBComponent.name)
}

func (c *LTRListDBComponent) Layout() *LTRListDBComponent {
	if len(GlobalConnectionComponent.dbs) == 0 {
		return c
	}
	theY1 := GlobalApp.maxY * 2 / 10
	if GlobalApp.CurrentView != c.name {
		theY1 = c.minH
	}
	v, err := GlobalApp.gui.SetView(c.name, 0, 0, GlobalApp.maxX*2/10, theY1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = " " + c.title + " "
		v.Editable = false
		v.Frame = true
		_, c.LayoutMaxH = v.Size()
	}

	printString := ""
	currenLine := 0
	totalLine := 0
	for index, db := range GlobalConnectionComponent.dbs {
		if c.CurrenDB == index {
			// printString += fmt.Sprintf("\x1b[1;37;44m%s\x1b[0m\n", ""+db.Name+""+SPACE_STRING)
			printString += NewColorString(db.Name+" ("+strconv.Itoa(db.MaxKeys)+")"+SPACE_STRING+"\n", "white", "blue")
			totalLine++
			currenLine = totalLine
		} else {
			printString += fmt.Sprintf("%s\n", ""+db.Name+" ("+strconv.Itoa(db.MaxKeys)+")"+SPACE_STRING)
			totalLine++
		}
	}
	if GlobalApp.CurrentView != c.name {
		v.SetOrigin(0, c.SelectedDB)
	} else if currenLine > c.LayoutMaxH/2 {
		originLine := currenLine - c.LayoutMaxH/2
		if originLine < 0 {
			originLine = 0
		}
		if originLine > totalLine-c.LayoutMaxH {
			originLine = totalLine - c.LayoutMaxH
		}
		v.SetOrigin(0, originLine)
	} else {
		v.SetOrigin(0, 0)
	}
	v.Clear()
	v.Write([]byte(printString))
	c.view = v

	if GlobalApp.CurrentView == c.name {
		// c.KeyBind()
		GlobalApp.gui.SetCurrentView(c.name)

	} else {
		// GlobalApp.gui.DeleteKeybindings(c.name)
	}

	return c
}

func (c *LTRListDBComponent) KeyBind() *LTRListDBComponent {
	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CurrenDB++
		if c.CurrenDB > len(GlobalConnectionComponent.dbs)-1 {
			c.CurrenDB = 0
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CurrenDB--
		if c.CurrenDB < 0 {
			c.CurrenDB = len(GlobalConnectionComponent.dbs) - 1
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.SelectedDB = c.CurrenDB
		GlobalApp.CurrentView = "key_list"
		c.Layout()
		GlobalKeyComponent.LoadKeys().Layout()
		return nil
	})

	return c
}
