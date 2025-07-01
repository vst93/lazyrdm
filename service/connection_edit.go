package service

import (
	"fmt"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

type LTRConnectionEditComponent struct {
	name                string
	title               string
	ConnectionConfig    types.Connection
	ConnectionConfigBak types.Connection
	viewList            []string
	viewBeginX          int
	viewBeginY          int
	viewEndX            int
	viewEndY            int
	viewNowLine         int
	viewNowCurrent      string
}

type LTRConnectionEditComponentFormViewConfig struct {
	name      string
	title     string
	value     EditorInput
	isNewLine bool
	viewType  string
	xBeing    int
	xEnd      int
}

func InitConnectionEditComponent(con types.Connection) *LTRConnectionEditComponent {
	connectionEditComponent := &LTRConnectionEditComponent{
		name:                "connection_edit",
		title:               " Connection Edit ",
		ConnectionConfig:    con,
		ConnectionConfigBak: con,
	}
	return connectionEditComponent
}

func (c *LTRConnectionEditComponent) Layout() *LTRConnectionEditComponent {
	GlobalApp.Gui.Cursor = false
	theX0 := 0
	theY0 := 0
	theX1 := GlobalApp.maxX - 1
	theY1 := GlobalApp.maxY - 2
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	c.viewBeginX, c.viewBeginY, c.viewEndX, c.viewEndY = theX0, theY0, theX1, theY1
	v, _ := GlobalApp.Gui.SetView(c.name, c.viewBeginX, c.viewBeginY, c.viewEndX, c.viewEndY)
	v.Title = c.title
	v.Editable = true
	v.Wrap = true
	c.viewNowLine = 0
	c.viewList = []string{}
	c.viewBeginY = 1

	// json, _ := json.Marshal(c.ConnectionConfig)
	// PrintLn(string(json))

	// name
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:  c.name + "_name",
		title: "Name",
		value: EditorInput{BindValString: &c.ConnectionConfig.Name},
	})
	// group
	if c.ConnectionConfig.Name == "" {
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_group",
			title:     "Group",
			value:     EditorInput{BindValString: &c.ConnectionConfig.Group},
			isNewLine: true,
		})
	}
	// network
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_network",
		title:     "Network",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Network},
		isNewLine: true,
	})
	if c.ConnectionConfig.Network == "unix" {
		// Sock
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_sock",
			title:     "Sock",
			value:     EditorInput{BindValString: &c.ConnectionConfig.Sock},
			isNewLine: true,
		})
	} else {
		// host
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_host",
			title:     "Host",
			value:     EditorInput{BindValString: &c.ConnectionConfig.Addr},
			isNewLine: true,
			xEnd:      c.viewEndX - 21,
		})
		// port
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_port",
			title:     "Port",
			value:     EditorInput{BindValInt: &c.ConnectionConfig.Port},
			isNewLine: false,
			xBeing:    c.viewEndX - 20,
		})
	}
	// username

	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_username",
		title:     "Username",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Username},
		isNewLine: true,
	})
	// password
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_password",
		title:     "Password",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Password},
		isNewLine: true,
	})

	lineWitdh := (c.viewEndX - c.viewBeginX) / 3
	// reset
	c.formBtn(c.name+"_enter", "Save", 0, lineWitdh, true)
	c.formBtn(c.name+"_cancel", "Cancel", c.viewBeginX+lineWitdh, lineWitdh, false)
	c.formBtn(c.name+"_reset", "Reset", c.viewBeginX+lineWitdh*2, lineWitdh, false)

	//表单选项选中
	if c.viewNowCurrent == "" {
		c.viewNowCurrent = c.viewList[0]
	}

	v.SetCursor(0, 0)
	v.Clear()
	c.KeyBind()

	GlobalTipComponent.Layout(c.KeyMapTip())
	return c
}

func (c *LTRConnectionEditComponent) formView(config LTRConnectionEditComponentFormViewConfig) {
	name := config.name
	title := config.title
	isNewLine := config.isNewLine
	// viewType := config.viewType
	valueEditor := config.value
	xBeing := config.xBeing
	xEnd := config.xEnd
	c.viewList = append(c.viewList, name)
	// (*line)++
	if isNewLine {
		c.viewNowLine = c.viewNowLine + 1
	}
	if xBeing == 0 {
		xBeing = c.viewBeginX + 1
	}
	if xEnd == 0 {
		xEnd = c.viewEndX - 1
	}
	view, _ := GlobalApp.Gui.SetView(name, xBeing, c.viewBeginY+c.viewNowLine*3+1, xEnd, c.viewBeginY+c.viewNowLine*3+3)
	view.Title = " " + title + " "
	view.Frame = true
	// view.Wrap = true
	view.FgColor = gocui.ColorWhite
	view.Clear()
	if c.viewNowCurrent == name || (c.viewNowCurrent == "" && len(c.viewList) == 1) {
		// view.FgColor = gocui.ColorWhite
		view.BgColor = gocui.ColorBlue
		view.Editable = true
		view.Editor = &valueEditor
		// view.SetCursor(0, 0)
		GlobalApp.Gui.SetCurrentView(name)
		GlobalApp.Gui.Cursor = true
	} else {
		view.BgColor = gocui.ColorBlack
	}
	theValue := ""
	if valueEditor.BindValString != nil {
		theValue = *valueEditor.BindValString
	} else if valueEditor.BindValInt != nil {
		theValue = fmt.Sprintf("%d", *valueEditor.BindValInt)
	}
	view.Write([]byte(theValue))
	GlobalApp.Gui.DeleteKeybindings(name)
}

func (c *LTRConnectionEditComponent) formBtn(name string, title string, xBeing int, width int, isNewLine bool) {
	c.viewList = append(c.viewList, name)
	if isNewLine {
		c.viewNowLine = c.viewNowLine + 1
	}
	if xBeing == 0 {
		xBeing = c.viewBeginX
	}
	view, _ := GlobalApp.Gui.SetView(name, xBeing, c.viewBeginY+c.viewNowLine*3+1, xBeing+width, c.viewBeginY+c.viewNowLine*3+3)
	view.Frame = false
	view.FgColor = gocui.ColorWhite
	view.Clear()
	if c.viewNowCurrent == name || (c.viewNowCurrent == "" && len(c.viewList) == 1) {
		// view.FgColor = gocui.ColorWhite
		view.BgColor = gocui.ColorBlue
		GlobalApp.Gui.SetCurrentView(name)
	} else {
		view.BgColor = gocui.Attribute(termbox.ColorDarkGray)
	}
	leftSpace := (width - len(title)) / 2
	// theTitle := " \n"
	theTitle := ""
	for i := 0; i < leftSpace; i++ {
		theTitle += " "
	}
	theTitle += title
	view.Write([]byte(theTitle))
	GlobalApp.Gui.DeleteKeybindings(name)
}

func (c *LTRConnectionEditComponent) KeyBind() *LTRConnectionEditComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.keyBindTab(1)
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		switch c.viewNowCurrent {
		case c.name + "_enter":
		case c.name + "_cancel":
		case c.name + "_reset":
			c.ConnectionConfig = c.ConnectionConfigBak
			c.Layout()
		default:
			c.keyBindTab(1)
		}
		return nil
	})
	return c
}

func (c *LTRConnectionEditComponent) keyBindTab(mod int) {
	if len(c.viewList) > 1 {
		if mod >= 0 {
			if c.viewNowCurrent == c.viewList[len(c.viewList)-1] {
				c.viewNowCurrent = c.viewList[0]
			} else {
				for i, viewName := range c.viewList {
					if viewName == c.viewNowCurrent {
						c.viewNowCurrent = c.viewList[i+1]
						break
					}
				}
			}
		} else {
			if c.viewNowCurrent == c.viewList[0] {
				c.viewNowCurrent = c.viewList[len(c.viewList)-1]
			} else {
				for i, viewName := range c.viewList {
					if viewName == c.viewNowCurrent {
						c.viewNowCurrent = c.viewList[i-1]
						break
					}
				}
			}
		}

		c.Layout()
	}
}

func (c *LTRConnectionEditComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>/<Enter>"},
		{"Submit", "<Enter>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
		i++
	}
	return ret
}
