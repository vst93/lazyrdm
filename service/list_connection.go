package service

import (
	"context"
	"fmt"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

type LTRConnectionComponent struct {
	name                                  string
	title                                 string
	ConnectionList                        types.Connections
	ConnectionListSelectedGroupIndex      int
	ConnectionListCurrentGroupIndex       int
	ConnectionListSelectedConnectionIndex int
	LayoutMaxY                            int
	ConnectionListSelectedConnectionInfo  types.Connection
	dbs                                   []types.ConnectionDB
	view                                  int
	lastDB                                int
	version                               string
}

func InitConnectionComponent() *LTRConnectionComponent {
	connSvc := services.Connection()
	browserSvc := services.Browser()
	ctx := context.Background()
	connSvc.Start(ctx)
	browserSvc.Start(ctx)
	connectionListJson := connSvc.ListConnection()
	c := &LTRConnectionComponent{
		name:                                  "connection_list",
		title:                                 "Connection List",
		ConnectionList:                        connectionListJson.Data.(types.Connections),
		ConnectionListCurrentGroupIndex:       -1,
		ConnectionListSelectedConnectionIndex: -1,
	}
	// 兼容一级非目录的配置
	for i, group := range c.ConnectionList {
		if len(group.Connections) == 0 && group.Type != "group" {
			c.ConnectionList[i].Connections = append(c.ConnectionList[i].Connections, group)
		}
	}
	GlobalApp.ViewNameList = []string{c.name}
	c.Layout().KeyBind()
	return c
}

func (c *LTRConnectionComponent) KeyBind() *LTRConnectionComponent {

	GlobalApp.gui.SetKeybinding(c.name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.ConnectionListCurrentGroupIndex >= 0 {
			c.ConnectionListSelectedConnectionIndex++
			if c.ConnectionListSelectedConnectionIndex > len(c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections)-1 {
				c.ConnectionListSelectedConnectionIndex = 0
			}
		} else {
			c.ConnectionListSelectedGroupIndex++
			if c.ConnectionListSelectedGroupIndex > len(c.ConnectionList)-1 {
				c.ConnectionListSelectedGroupIndex = 0
			}
			c.ConnectionListSelectedConnectionIndex = -1
		}

		v.Clear()
		c.Layout()
		return nil
	})

	GlobalApp.gui.SetKeybinding(c.name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.ConnectionListCurrentGroupIndex >= 0 {
			c.ConnectionListSelectedConnectionIndex--
			if c.ConnectionListSelectedConnectionIndex < 0 {
				c.ConnectionListSelectedConnectionIndex = len(c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections) - 1
			}
		} else {
			c.ConnectionListSelectedGroupIndex--
			if c.ConnectionListSelectedGroupIndex < 0 {
				c.ConnectionListSelectedGroupIndex = len(c.ConnectionList) - 1
			}
			c.ConnectionListSelectedConnectionIndex = -1
		}

		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []interface{}{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.ConnectionListCurrentGroupIndex >= 0 {
			// connection selected
			c.ConnectionListSelectedConnectionInfo = c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex]
			connectionInfo := services.Browser().OpenConnection(c.ConnectionListSelectedConnectionInfo.Name)
			GlobalConnectionComponent.dbs = connectionInfo.Data.(map[string]any)["db"].([]types.ConnectionDB)
			GlobalConnectionComponent.view = connectionInfo.Data.(map[string]any)["view"].(int)
			GlobalConnectionComponent.lastDB = connectionInfo.Data.(map[string]any)["lastDB"].(int)
			GlobalConnectionComponent.version = connectionInfo.Data.(map[string]any)["version"].(string)
			GlobalApp.gui.DeleteView(c.name)
			GlobalApp.gui.DeleteKeybindings(c.name)
			GlobalApp.CurrentView = "db_list"
			GlobalApp.ViewNameList = []string{} // 清空视图列表
			GlobalDBComponent = InitDBComponent()
			GlobalKeyComponent = InitKeyComponent()
			GlobalApp.ViewNameList = []string{GlobalDBComponent.name, GlobalKeyComponent.name}
			return nil
		} else {
			c.ConnectionListCurrentGroupIndex = c.ConnectionListSelectedGroupIndex
			c.ConnectionListSelectedConnectionIndex = 0
			v.Clear()
			c.Layout()
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []interface{}{gocui.KeyArrowLeft, gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
		v.Clear()
		c.Layout()
		return nil
	})

	GlobalApp.gui.SetKeybinding(c.name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return nil
	})
	return c
}

func (c *LTRConnectionComponent) Layout() *LTRConnectionComponent {
	theX0 := 0
	theY0 := 0
	theX1 := GlobalApp.maxX - 1
	theY1 := GlobalApp.maxY - 2
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	v, err := GlobalApp.gui.SetView(c.name, theX0, theY0, theX1, theY1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = " " + c.title + " "
		v.Editable = false
		// v.Wrap = true
		// v.Autoscroll = true
		v.Frame = true
		// v.FgColor = gocui.ColorGreen
		_, c.LayoutMaxY = v.Size()

		GlobalApp.gui.SetCurrentView(c.name)
	}

	printString := ""
	currenLine := 0
	totalLine := 0
	for index, conn := range c.ConnectionList {
		theConnectionsLen := len(conn.Connections)
		if c.ConnectionListSelectedGroupIndex == index {
			if c.ConnectionListSelectedConnectionIndex == -1 {
				printString += NewColorString("["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING+"\n", "white", "blue", "bold") // 白底黑字
				totalLine++
				currenLine = totalLine
			} else {
				printString += fmt.Sprintf("%s\n", "["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING)
				totalLine++
			}
			for key, item := range conn.Connections {
				if key == c.ConnectionListSelectedConnectionIndex {
					// printString += fmt.Sprintf(" - \x1b[1;37;44m%s\x1b[0m\n", item.Name+SPACE_STRING) // 白底黑字
					printString += NewColorString(" - "+item.Name+SPACE_STRING+"\n", "white", "blue", "bold")
					totalLine++
					currenLine = totalLine
				} else {
					printString += fmt.Sprintf(" - %s%s\n", item.Name, SPACE_STRING)
					totalLine++
				}
			}
		} else {
			printString += fmt.Sprintf("%s\n", "["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING)
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

	return c
}
