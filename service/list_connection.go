package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/jroimartin/gocui"
	"github.com/vrischmann/userdir"
)

type LTRConnectionComponent struct {
	Name                                  string
	title                                 string
	ConnectionList                        types.Connections
	ConnectionListSelectedGroupIndex      int // 当前光标在的组的索引
	ConnectionListCurrentGroupIndex       int // 当前选择的组的索引
	ConnectionListSelectedConnectionIndex int
	LayoutMaxY                            int
	ConnectionListSelectedConnectionInfo  types.Connection
	dbs                                   []types.ConnectionDB
	view                                  int
	lastDB                                int
	version                               string
	isConnecting                          bool
	GUIView                               *gocui.View
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
		ConnectionList:                        types.Connections{},
		ConnectionListCurrentGroupIndex:       -1,
		ConnectionListSelectedConnectionIndex: -1,
	}
	// 兼容一级非目录的配置
	connectionList := GlobalConnectionComponent.ConnectionList
	noGriupConnection := []types.Connection{}
	for _, group := range connectionListJson.Data.(types.Connections) {
		if len(group.Connections) == 0 && group.Type != "group" {
			// GlobalConnectionComponent.ConnectionList[i].Connections = append(GlobalConnectionComponent.ConnectionList[i].Connections, group)
			noGriupConnection = append(noGriupConnection, group)
		} else if group.Type == "group" && group.Name == "" {
			// 空组的名称的移动到 "NO_GROUP"
			noGriupConnection = append(noGriupConnection, group.Connections...)
		} else {
			connectionList = append(connectionList, group)
		}
	}
	if len(noGriupConnection) > 0 {
		connectionList = append(connectionList, types.Connection{
			ConnectionConfig: types.ConnectionConfig{
				Name: "NO_GROUP",
			},
			Type:        "group",
			Connections: noGriupConnection,
		})
	}
	GlobalConnectionComponent.ConnectionList = connectionList
	GlobalApp.ViewNameList = []string{GlobalConnectionComponent.Name}
	GlobalConnectionComponent.Layout().KeyBind()
	GlobalApp.Gui.SetCurrentView(GlobalConnectionComponent.Name)
}

func (c *LTRConnectionComponent) Layout() *LTRConnectionComponent {
	GlobalApp.Gui.Cursor = false
	theX0 := 0
	theY0 := 0
	theX1 := GlobalApp.maxX - 1
	theY1 := GlobalApp.maxY - 2
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	var err error
	if c.GUIView == nil {
		c.GUIView, err = GlobalApp.Gui.SetView(c.Name, theX0, theY0, theX1, theY1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return c
			}
			c.GUIView.Title = " " + c.title + " "
			c.GUIView.Editable = false
			c.GUIView.Frame = true
			_, c.LayoutMaxY = c.GUIView.Size()
		}
	}
	GlobalApp.Gui.SetCurrentView(c.Name)

	printString := ""
	currenLine := 0
	totalLine := 0
	for index, conn := range c.ConnectionList {
		if conn.Name == "" {
			continue
		}
		theConnectionsLen := len(conn.Connections)
		if c.ConnectionListSelectedGroupIndex == index {
			if c.ConnectionListSelectedConnectionIndex == -1 {
				printString += NewColorString("["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING+"\n", "white", "blue", "bold")
				totalLine++
				currenLine = totalLine
			} else {
				// printString += fmt.Sprintf("%s\n", "["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING)
				printString += NewColorString("["+conn.Name+"] ("+fmt.Sprintf("%d", theConnectionsLen)+")"+SPACE_STRING+"\n", "white", "purple", "bold")
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
		c.GUIView.SetOrigin(0, originLine)
	} else {
		c.GUIView.SetOrigin(0, 0)
	}
	c.GUIView.Clear()
	c.GUIView.Write([]byte(printString))

	if GlobalApp.Gui.CurrentView().Name() == c.Name {
		GlobalTipComponent.Layout(c.KeyMapTip())
	}

	return c
}

func (c *LTRConnectionComponent) KeyBind() *LTRConnectionComponent {

	GlobalApp.Gui.SetKeybinding(c.Name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
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
		if c.isConnecting {
			return nil
		}
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if c.ConnectionListCurrentGroupIndex >= 0 {
			GlobalTipComponent.LayoutTemporary("Connecting...", 10, TipTypeWarning)
			c.isConnecting = true
			if GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name != "" {
				// 关闭之前的连接
				services.Browser().CloseConnection(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
			}
			// connection selected
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo = c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex]
			go func() {
				connectionInfo := services.Browser().OpenConnection(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
				if connectionInfo.Success {
					GlobalTipComponent.LayoutTemporary("Open connection success", 2, TipTypeSuccess)
					GlobalConnectionComponent.dbs = connectionInfo.Data.(map[string]any)["db"].([]types.ConnectionDB)
					GlobalConnectionComponent.view = connectionInfo.Data.(map[string]any)["view"].(int)
					GlobalConnectionComponent.lastDB = connectionInfo.Data.(map[string]any)["lastDB"].(int)
					GlobalConnectionComponent.version = connectionInfo.Data.(map[string]any)["version"].(string)
					GlobalApp.Gui.DeleteView(c.Name)
					GlobalApp.Gui.DeleteKeybindings(c.Name)
					GlobalApp.ViewNameList = []string{} // 清空视图列表
					c.closeView()
					InitDBComponent()
				} else {
					GlobalTipComponent.LayoutTemporary("Open connection failed", 5, TipTypeError)
					GlobalApp.Gui.SetCurrentView(c.Name)
				}
				c.isConnecting = false
			}()
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'e', 'E'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if GlobalConnectionComponent.ConnectionListCurrentGroupIndex >= 0 {
			// 编辑光标选中的连接
			c.closeView()
			connectionComponent := InitConnectionEditComponent(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex])
			connectionComponent.Layout()
			return nil
		} else if GlobalConnectionComponent.ConnectionListSelectedGroupIndex >= 0 {
			// 编辑光标选中的组
			c.closeView()
			connectionComponent := InitConnectionEditComponent(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListSelectedGroupIndex])
			connectionComponent.Layout()
			return nil
		}
		return nil
	})

	// 新建连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'n', 'N'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if GlobalConnectionComponent.ConnectionListCurrentGroupIndex >= 0 {
			// 新建连接
			c.closeView()
			connectionComponent := InitConnectionEditComponent(types.Connection{
				ConnectionConfig: types.ConnectionConfig{
					Group: GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListCurrentGroupIndex].Name,
					Port:  6379,
					SSH: types.ConnectionSSH{
						Enable:    false,
						LoginType: "pwd",
						Port:      22,
					},
				},
			})
			connectionComponent.Layout()
			return nil
		} else if GlobalConnectionComponent.ConnectionListSelectedGroupIndex >= 0 {
			// 新建组
			c.closeView()
			connectionComponent := InitConnectionEditComponent(types.Connection{
				Type: "group",
			})
			connectionComponent.Layout()
			return nil
		}

		return nil
	})

	// 删除连接信息或分组
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'d', 'D'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if GlobalConnectionComponent.ConnectionListSelectedConnectionIndex >= 0 {
			NewPageComponentConfirm("Delete connection", "Are you sure to delete this connection?", func() {
				//删除光标选中连接
				apiResult := services.Connection().DeleteConnection(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex].Name)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Delete connection success", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Delete connection failed", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				//取消删除
				GlobalTipComponent.LayoutTemporary("Delete canceled", 2, TipTypeSuccess)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})

		} else if GlobalConnectionComponent.ConnectionListSelectedGroupIndex >= 0 {
			if len(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListSelectedGroupIndex].Connections) > 0 {
				GlobalTipComponent.LayoutTemporary("Group not empty, can not delete", 3, TipTypeError)
				return nil
			}
			NewPageComponentConfirm("Delete group", "Are you sure to delete this group?", func() {
				//删除光标选中的组
				apiResult := services.Connection().DeleteGroup(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListSelectedGroupIndex].Name, false)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Delete group success", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Delete group failed", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				//取消删除
				GlobalTipComponent.LayoutTemporary("Delete canceled", 2, TipTypeSuccess)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowLeft, gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
		v.Clear()
		c.Layout()
		return nil
	})

	// 导出连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyCtrlE}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		apiResult := c.ExportConnections()
		PrintLn(apiResult)
		if apiResult.Success {
			GlobalTipComponent.LayoutTemporary("Export connections success", 2, TipTypeSuccess)
			OpenFileManager(apiResult.Data.(struct {
				Path string `json:"path"`
			}).Path)
		} else {
			GlobalTipComponent.LayoutTemporary("Export connections failed", 3, TipTypeError)
		}
		return nil
	})

	// 导入连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyCtrlI}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		//读取剪切板内容
		clipboardContent, _ := clipboard.ReadAll()
		noticeString := "You need to first copy the path of the link file exported from Tiny RDM (in zip format) to the clipboard and select Confirm to import.\n\nCurrent clipboard: " + clipboardContent
		NewPageComponentConfirm("Import connections", noticeString, func() {
			// 导入连接信息
			if clipboardContent == "" {
				GlobalTipComponent.LayoutTemporary("Failed to get clipboard content", 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			if !strings.HasSuffix(clipboardContent, ".zip") {
				GlobalTipComponent.LayoutTemporary("The clipboard content is not a zip file path", 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			apiResult := c.ImportConnections(clipboardContent)
			if !apiResult.Success {
				GlobalTipComponent.LayoutTemporary(apiResult.Msg, 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			GlobalTipComponent.LayoutTemporary("Import connections success", 2, TipTypeSuccess)
			c.closeView()
			InitConnectionComponent()
		}, func() {
			// 取消导入
			GlobalTipComponent.LayoutTemporary("Import canceled", 2, TipTypeSuccess)
			GlobalApp.Gui.SetCurrentView(c.Name)
		})
		return nil
	})
	// GlobalApp.Gui.SetKeybinding(c.Name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if c.isConnecting {
	// 		return nil
	// 	}
	// 	return nil
	// })
	return c
}

func (c *LTRConnectionComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Select", "↑/↓"},
		{"Up", "←"},
		{"Enter", "<Enter>/→"},
		{"Edit", "<E>"},
		{"New", "<N>"},
		{"Delete", "<D>"},
		{"Export", "<Ctrl + E>"},
		{"[Global] Quit", "<Ctrl + Q>"},
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

func (c *LTRConnectionComponent) closeView() {
	GlobalApp.Gui.DeleteView(c.Name)
	GlobalApp.Gui.DeleteKeybindings(c.Name)
	GlobalApp.ViewNameList = []string{} // 清空视图列表
}

// ExportConnections 导出连接信息
func (c *LTRConnectionComponent) ExportConnections() (resp types.JSResp) {
	defaultFileName := "connections_" + time.Now().Format("20060102150405") + ".zip"

	// 获取用户下载目录
	userDownloadDir, err := GetDownloadPath()
	if err != nil {
		userDownloadDir = "~"
	}
	filepath := path.Join(userDownloadDir, defaultFileName)

	// compress the connections profile with zip
	const connectionFilename = "connections.yaml"
	connectionFilePath := path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename)
	inputFile, err := os.Open(connectionFilePath)
	if err != nil {
		resp.Msg = err.Error()
		return
	}
	defer inputFile.Close()

	err = fileutil.Zip(connectionFilePath, filepath)
	if err != nil {
		resp.Msg = err.Error()
		return
	}

	resp.Success = true
	resp.Data = struct {
		Path string `json:"path"`
	}{
		Path: filepath,
	}
	return
}

// ImportConnections import connections from local zip file
func (c *LTRConnectionComponent) ImportConnections(filepath string) (resp types.JSResp) {
	if !fileutil.IsZipFile(filepath) {
		resp.Msg = "The file is not a zip file"
		return
	}

	const connectionFilename = "connections.yaml"
	zipFile, err := zip.OpenReader(filepath)
	if err != nil {
		resp.Msg = err.Error()
		return
	}

	var file *zip.File
	for _, file = range zipFile.File {
		if file.Name == connectionFilename {
			break
		}
	}
	if file != nil {
		zippedFile, err := file.Open()
		if err != nil {
			resp.Msg = err.Error()
			return
		}
		defer zippedFile.Close()

		outputFile, err := os.Create(path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename))
		PrintLn(path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename))
		if err != nil {
			resp.Msg = err.Error()
			return
		}
		defer outputFile.Close()

		if _, err = io.Copy(outputFile, zippedFile); err != nil {
			resp.Msg = err.Error()
			return
		}
	}

	resp.Success = true
	return
}
