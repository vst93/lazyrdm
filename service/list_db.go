package service

import (
	"fmt"
	"strconv"

	"github.com/jroimartin/gocui"
)

type LTRListDBComponent struct {
	name       string
	title      string
	LayoutMaxY int
	SelectedDB int // 当前打开数据库
	CurrenDB   int // 当前光标选中的据库
	view       *gocui.View
}

func InitDBComponent() *LTRListDBComponent {
	c := &LTRListDBComponent{
		name:       "db_list",
		title:      "DB",
		LayoutMaxY: 0,
		SelectedDB: -1,
		CurrenDB:   0,
		view:       nil,
	}

	// GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, c.name)
	c.Layout().KeyBind()
	return c
}

func (c *LTRListDBComponent) Layout() *LTRListDBComponent {
	if len(GlobalConnectionComponent.dbs) == 0 {
		return c
	}
	// maxX, maxY := GlobalApp.gui.Size()
	v, err := GlobalApp.gui.SetView(c.name, 0, 0, GlobalApp.maxX*2/10, GlobalApp.maxY*2/10)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = " " + c.title + " "
		v.Editable = false
		v.Frame = true
		_, c.LayoutMaxY = v.Size()

		// GlobalApp.gui.SetCurrentView(c.name)

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
		GlobalKeyComponent = InitKeyComponent()
		return nil
	})

	return c
}
