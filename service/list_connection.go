package service

import (
	"context"
	"fmt"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

type LTRConnectionComponent struct {
	Name                                  string
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

func InitConnectionComponent() {
	connSvc := services.Connection()
	browserSvc := services.Browser()
	ctx := context.Background()
	connSvc.Start(ctx)
	browserSvc.Start(ctx)
	connectionListJson := connSvc.ListConnection()
	GlobalConnectionComponent = &LTRConnectionComponent{
		Name:                                  "connection_list",
		title:                                 "Connection List",
		ConnectionList:                        connectionListJson.Data.(types.Connections),
		ConnectionListCurrentGroupIndex:       -1,
		ConnectionListSelectedConnectionIndex: -1,
	}
	// 兼容一级非目录的配置
	for i, group := range GlobalConnectionComponent.ConnectionList {
		if len(group.Connections) == 0 && group.Type != "group" {
			GlobalConnectionComponent.ConnectionList[i].Connections = append(GlobalConnectionComponent.ConnectionList[i].Connections, group)
		}
	}
	GlobalApp.ViewNameList = []string{GlobalConnectionComponent.Name}
	GlobalConnectionComponent.Layout().KeyBind()
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
	v, err := GlobalApp.Gui.SetView(c.Name, theX0, theY0, theX1, theY1)
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

		GlobalApp.Gui.SetCurrentView(c.Name)
		// GlobalTipComponent.Layout()
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

	if GlobalApp.Gui.CurrentView().Name() == c.Name {
		GlobalTipComponent.Layout()
	}

	return c
}

func (c *LTRConnectionComponent) KeyBind() *LTRConnectionComponent {

	GlobalApp.Gui.SetKeybinding(c.Name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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

	GlobalApp.Gui.SetKeybinding(c.Name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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

	// 打开连接
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []interface{}{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.ConnectionListCurrentGroupIndex >= 0 {
			if GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name != "" {
				// 关闭之前的连接
				services.Browser().CloseConnection(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
			}
			// connection selected
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo = c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex]
			connectionInfo := services.Browser().OpenConnection(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
			// PrintLn(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
			if connectionInfo.Success {
				GlobalConnectionComponent.dbs = connectionInfo.Data.(map[string]any)["db"].([]types.ConnectionDB)
				GlobalConnectionComponent.view = connectionInfo.Data.(map[string]any)["view"].(int)
				GlobalConnectionComponent.lastDB = connectionInfo.Data.(map[string]any)["lastDB"].(int)
				GlobalConnectionComponent.version = connectionInfo.Data.(map[string]any)["version"].(string)
				GlobalApp.Gui.DeleteView(c.Name)
				GlobalApp.Gui.DeleteKeybindings(c.Name)
				GlobalApp.ViewNameList = []string{} // 清空视图列表
				InitDBComponent()
			}
			return nil
		} else {
			c.ConnectionListCurrentGroupIndex = c.ConnectionListSelectedGroupIndex
			c.ConnectionListSelectedConnectionIndex = 0
			v.Clear()
			c.Layout()
		}
		return nil
	})

	// 编辑连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []interface{}{gocui.KeyCtrlE}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if GlobalConnectionComponent.ConnectionListCurrentGroupIndex >= 0 {
			connectionComponent := InitConnectionEditComponent()
			connectionComponent.ConnectionConfig = GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex]
			connectionComponent.Layout()
			return nil
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []interface{}{gocui.KeyArrowLeft, gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
		v.Clear()
		c.Layout()
		return nil
	})

	GlobalApp.Gui.SetKeybinding(c.Name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return nil
	})
	return c
}

func (c *LTRConnectionComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Select", "↑↓"},
		{"Up", "←"},
		{"Enter", "<Enter>/→"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
		i++
	}
	// return "connection_list: " + ret
	return ret
}
