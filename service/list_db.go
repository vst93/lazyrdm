package service

import (
	"fmt"
	"strconv"
	"tinyrdm/backend/services"

	"github.com/awesome-gocui/gocui"
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
	if GlobalDBComponent != nil {
		GlobalApp.Gui.DeleteKeybindings(GlobalDBComponent.name)
	}
	GlobalDBComponent = &LTRListDBComponent{
		name:       "db_list",
		title:      "DB",
		LayoutMaxH: 0,
		SelectedDB: -1,
		CurrenDB:   0,
		view:       nil,
		minH:       2,
	}
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalDBComponent.name)
	GlobalDBComponent.Layout().KeyBind()
	GlobalTipComponent.AppendList(GlobalDBComponent.name, GlobalDBComponent.KeyMapTip())
	GlobalApp.Gui.SetCurrentView(GlobalDBComponent.name)
	GlobalDBComponent.Layout()
	InitKeyComponent()
	InitKeyInfoComponent()
	InitKeyInfoDetailComponent()
	// GlobalTipComponent.Layout()
	GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil }) // 刷新界面
}

func (c *LTRListDBComponent) Layout() *LTRListDBComponent {
	if len(GlobalConnectionComponent.dbs) == 0 {
		return c
	}
	theY1 := GlobalApp.maxY * 3 / 10
	if CurrentViewName() != c.name {
		theY1 = c.minH
	}
	var err error
	c.view, err = SetViewSafe(c.name, 0, 0, GlobalApp.maxX*2/10, theY1, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}

	if CurrentViewName() == c.name {
		c.view.Title = " [" + c.title + "] "
	} else {
		c.view.Title = " " + c.title + " "
	}

	c.view.Editable = false
	c.view.Frame = true
	_, c.LayoutMaxH = c.view.Size()

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

	c.view.Clear()
	c.view.Write([]byte(printString))

	if CurrentViewName() != c.name && c.SelectedDB >= 0 {
		c.view.SetOrigin(0, c.SelectedDB)
	} else if currenLine > c.LayoutMaxH/2 {
		originLine := currenLine - c.LayoutMaxH/2
		if originLine < 0 {
			originLine = 0
		}
		if originLine > totalLine-c.LayoutMaxH {
			originLine = totalLine - c.LayoutMaxH
		}
		c.view.SetOrigin(0, originLine)
	} else {
		c.view.SetOrigin(0, 0)
	}
	// c.view = v
	if CurrentViewName() == c.name && GlobalTipComponent != nil {
		GlobalTipComponent.Layout(c.KeyMapTip())
	}

	return c
}

func (c *LTRListDBComponent) KeyBind() *LTRListDBComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CurrenDB++
		if c.CurrenDB > len(GlobalConnectionComponent.dbs)-1 {
			c.CurrenDB = 0
		}

		v.Clear()
		c.Layout()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.CurrenDB--
		if c.CurrenDB < 0 {
			c.CurrenDB = len(GlobalConnectionComponent.dbs) - 1
		}

		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyEnter, gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		oldSelectedDB := c.SelectedDB
		c.SelectedDB = c.CurrenDB
		GlobalApp.Gui.SetCurrentView(GlobalKeyComponent.name)
		c.Layout()
		if c.SelectedDB != oldSelectedDB {
			services.Browser().OpenDatabase(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name, c.SelectedDB)
			GlobalKeyComponent.IsEnd = false
			GlobalKeyComponent.keys = []any{}
			GlobalKeyComponent.Current = 0
		}

		GlobalKeyComponent.LoadKeys().Layout()
		return nil
	})

	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'i'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		OpenServerInfoPage()
		return nil
	})

	// Redis Console
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{':'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		OpenConsolePage()
		return nil
	})

	// FlushDB
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'F'}, "Flush current database? ALL keys in this DB will be deleted permanently!", func() {
		if c.SelectedDB < 0 {
			GlobalTipComponent.LayoutTemporary("No database selected", 3, TipTypeWarning)
			return
		}
		res := services.Browser().FlushDB(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			c.SelectedDB,
			false,
		)
		if res.Success {
			GlobalTipComponent.LayoutTemporary("Database flushed", 2, TipTypeSuccess)
			GlobalKeyComponent.RefreshList()
		} else {
			GlobalTipComponent.LayoutTemporary("FlushDB failed: "+res.Msg, 5, TipTypeError)
		}
	}, func() {
		GlobalTipComponent.LayoutTemporary("FlushDB cancelled", 2, TipTypeWarning)
	})

	return c
}

func (c *LTRListDBComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Select", "↑/↓/j/k"},
		{"Open DB", "<Enter>/l/→"},
		{"Server Info", "<i>"},
		{"Console", "<:>"},
		{"FlushDB", "<F>"},
		{"Pane", "<Tab>"},
		{"Conn/Quit/Help", "<Ctrl+w>/<Ctrl+q>/<?>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	// return "db_list: " + ret
	return ret
}
