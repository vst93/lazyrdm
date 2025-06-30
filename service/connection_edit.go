package service

import (
	"strconv"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

type LTRConnectionEditComponent struct {
	name             string
	title            string
	ConnectionConfig types.Connection
	viewList         []string
	viewBeginX       int
	viewBeginY       int
	viewEndX         int
	viewEndY         int
	viewNowLine      int
	viewNowCurrent   string
}

type LTRConnectionEditComponentFormViewConfig struct {
	name      string
	title     string
	value     string
	isNewLine bool
	viewType  string
	xBeing    int
	xEnd      int
}

func InitConnectionEditComponent() *LTRConnectionEditComponent {
	connectionEditComponent := &LTRConnectionEditComponent{
		name:  "connection_edit",
		title: " Connection Edit ",
	}
	return connectionEditComponent
}

func (c *LTRConnectionEditComponent) Layout() *LTRConnectionEditComponent {
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

	// json, _ := json.Marshal(c.ConnectionConfig)
	// PrintLn(string(json))

	// name
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:  c.name + "_name",
		title: "Name",
		value: c.ConnectionConfig.Name,
	})
	// group
	if c.ConnectionConfig.Name == "" {
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_group",
			title:     "Group",
			value:     c.ConnectionConfig.Group,
			isNewLine: true,
		})
	}
	// network
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_network",
		title:     "Network",
		value:     c.ConnectionConfig.Network,
		isNewLine: true,
	})
	if c.ConnectionConfig.Network == "unix" {
		// Sock
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_sock",
			title:     "Sock",
			value:     c.ConnectionConfig.Sock,
			isNewLine: true,
		})
	} else {
		// host
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_host",
			title:     "Host",
			value:     c.ConnectionConfig.Addr,
			isNewLine: true,
			xEnd:      c.viewEndX - 21,
		})
		// port
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_port",
			title:     "Port",
			value:     strconv.Itoa(c.ConnectionConfig.Port),
			isNewLine: false,
			xBeing:    c.viewEndX - 20,
		})
	}
	// username

	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_username",
		title:     "Username",
		value:     c.ConnectionConfig.Username,
		isNewLine: true,
	})
	// password
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_password",
		title:     "Password",
		value:     c.ConnectionConfig.Password,
		isNewLine: true,
	})

	//表单选项选中
	if c.viewNowCurrent == "" {
		c.viewNowCurrent = c.viewList[0]
	}

	v.SetCursor(0, 0)
	v.Clear()
	// GlobalApp.Gui.SetCurrentView(c.name)
	// v.Write(c.ConnectionConfig.ToJSON())
	// GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if len(c.viewList) > 1 {
	// 		if c.viewNowCurrent == c.viewList[len(c.viewList)-1] {
	// 			c.viewNowCurrent = c.viewList[0]
	// 		} else {
	// 			for i, viewName := range c.viewList {
	// 				if viewName == c.viewNowCurrent {
	// 					c.viewNowCurrent = c.viewList[i+1]
	// 					break
	// 				}
	// 			}
	// 		}
	// 		c.Layout()
	// 	}
	// 	return nil
	// })
	PrintLn(c.viewList)
	c.KeyBind()

	return c
}

func (c *LTRConnectionEditComponent) formView(config LTRConnectionEditComponentFormViewConfig) {
	name := config.name
	title := config.title
	isNewLine := config.isNewLine
	// viewType := config.viewType
	value := config.value
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
		view.Editor = &EditorInput{}
		// view.SetCursor(0, 0)
		GlobalApp.Gui.SetCurrentView(name)
	} else {
		view.BgColor = gocui.ColorBlack
	}
	view.Write([]byte(value))
	GlobalApp.Gui.DeleteKeybindings(name)
}

func (c *LTRConnectionEditComponent) KeyBind() *LTRConnectionEditComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(c.viewList) > 1 {
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
			c.Layout()
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(c.viewList) == 0 {
			return nil
		}

		c.Layout()
		return nil
	})

	return c
}

func (c *LTRConnectionEditComponent) enterFormView(viewName string, value string) {
	

}