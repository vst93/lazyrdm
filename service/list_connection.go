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
	"github.com/awesome-gocui/gocui"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/pkg/browser"
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
	if GlobalApp.maxX < 2 || GlobalApp.maxY < 3 {
		return c
	}
	theX0 := 0
	theY0 := 0
	theX1 := GlobalApp.maxX - 1
	theY1 := GlobalApp.maxY - 2
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	if theX1 <= theX0 {
		theX1 = theX0 + 1
	}
	if theY1 <= theY0 {
		theY1 = theY0 + 1
	}
	v, err := SetViewSafe(c.Name, theX0, theY0, theX1, theY1, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = " " + c.title + " "
		v.Editable = false
		v.Frame = true
		_, c.LayoutMaxY = v.Size()
	}
	GlobalApp.Gui.SetCurrentView(c.Name)
	if c.ConnectionListCurrentGroupIndex >= 0 && c.ConnectionListCurrentGroupIndex < len(c.ConnectionList) {
		v.Subtitle = " Connection mode | Group: " + c.ConnectionList[c.ConnectionListCurrentGroupIndex].Name + " | Enter: connect | h: back to groups "
	} else {
		v.Subtitle = " Group mode | Enter: open group | j/k: move | e/n/d: edit/new/delete group "
	}

	builder := strings.Builder{}
	currenLine := 0
	totalLine := 0
	for index, group := range c.ConnectionList {
		if group.Name == "" {
			continue
		}

		groupPrefix := ">"
		if c.ConnectionListCurrentGroupIndex == index {
			groupPrefix = "v"
		}
		groupLine := fmt.Sprintf("%s [%s] (%d)\n", groupPrefix, group.Name, len(group.Connections))

		if c.ConnectionListSelectedGroupIndex == index {
			if c.ConnectionListSelectedConnectionIndex == -1 {
				builder.WriteString(NewColorString(groupLine, "white", "blue", "bold"))
				totalLine++
				currenLine = totalLine
			} else {
				builder.WriteString(NewColorString(groupLine, "white", "purple", "bold"))
				totalLine++
			}
		} else {
			builder.WriteString(groupLine)
			totalLine++
		}

		if c.ConnectionListCurrentGroupIndex == index {
			for key, item := range group.Connections {
				connectionLine := fmt.Sprintf("  - %s\n", item.Name)
				if key == c.ConnectionListSelectedConnectionIndex {
					builder.WriteString(NewColorString(connectionLine, "white", "blue", "bold"))
					totalLine++
					currenLine = totalLine
				} else {
					builder.WriteString(connectionLine)
					totalLine++
				}
			}
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
	v.Write([]byte(builder.String()))

	if CurrentViewName() == c.Name {
		GlobalTipComponent.Layout(c.KeyMapTip())
	}
	return c
}

func (c *LTRConnectionComponent) KeyBind() *LTRConnectionComponent {

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if c.ConnectionListCurrentGroupIndex >= 0 {
			c.moveConnectionSelection(1)
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

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if c.ConnectionListCurrentGroupIndex >= 0 {
			c.moveConnectionSelection(-1)
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyEnter, gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if c.ConnectionListCurrentGroupIndex >= 0 {
			if c.ConnectionListCurrentGroupIndex >= len(c.ConnectionList) || c.ConnectionListSelectedConnectionIndex < 0 || c.ConnectionListSelectedConnectionIndex >= len(c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections) {
				GlobalTipComponent.LayoutTemporary("No connection selected", 2, TipTypeWarning)
				return nil
			}
			GlobalTipComponent.LayoutTemporary("Connecting to Redis...", 10, TipTypeWarning)
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
					GlobalTipComponent.LayoutTemporary("Connected successfully", 2, TipTypeSuccess)
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
					GlobalTipComponent.LayoutTemporary("Failed to connect", 5, TipTypeError)
					GlobalApp.Gui.SetCurrentView(c.Name)
				}
				c.isConnecting = false
			}()
			return nil
		} else {
			if c.ConnectionListSelectedGroupIndex < 0 || c.ConnectionListSelectedGroupIndex >= len(c.ConnectionList) {
				return nil
			}
			if len(c.ConnectionList[c.ConnectionListSelectedGroupIndex].Connections) == 0 {
				GlobalTipComponent.LayoutTemporary("This group has no connections. Press n to create one.", 3, TipTypeWarning)
				return nil
			}
			c.ConnectionListCurrentGroupIndex = c.ConnectionListSelectedGroupIndex
			c.ConnectionListSelectedGroupIndex = c.ConnectionListCurrentGroupIndex
			c.ConnectionListSelectedConnectionIndex = 0
			v.Clear()
			c.Layout()
		}
		return nil
	})

	// 编辑连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'e'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'d'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if GlobalConnectionComponent.ConnectionListSelectedConnectionIndex >= 0 {
			NewPageComponentConfirm("Delete connection", "Are you sure to delete this connection?", func() {
				//删除光标选中连接
				apiResult := services.Connection().DeleteConnection(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListCurrentGroupIndex].Connections[c.ConnectionListSelectedConnectionIndex].Name)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Connection deleted", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Failed to delete connection", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				//取消删除
				GlobalTipComponent.LayoutTemporary("Delete cancelled", 2, TipTypeWarning)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})

		} else if GlobalConnectionComponent.ConnectionListSelectedGroupIndex >= 0 {
			if len(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListSelectedGroupIndex].Connections) > 0 {
				GlobalTipComponent.LayoutTemporary("Cannot delete non-empty group", 3, TipTypeError)
				return nil
			}
			NewPageComponentConfirm("Delete group", "Are you sure to delete this group?", func() {
				//删除光标选中的组
				apiResult := services.Connection().DeleteGroup(GlobalConnectionComponent.ConnectionList[GlobalConnectionComponent.ConnectionListSelectedGroupIndex].Name, false)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Group deleted", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Failed to delete group", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				//取消删除
				GlobalTipComponent.LayoutTemporary("Delete cancelled", 2, TipTypeWarning)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowLeft, gocui.KeyEsc, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'E'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		apiResult := c.ExportConnections()
		PrintLn(apiResult)
		if apiResult.Success {
			GlobalTipComponent.LayoutTemporary("Connections exported", 2, TipTypeSuccess)
			OpenFileManager(apiResult.Data.(struct {
				Path string `json:"path"`
			}).Path)
		} else {
			GlobalTipComponent.LayoutTemporary("Failed to export connections", 3, TipTypeError)
		}
		return nil
	})

	// 导入连接信息
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'I'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		//读取剪切板内容
		clipboardContent, _ := clipboard.ReadAll()
		clipboardContent = strings.TrimSpace(clipboardContent)
		noticeString := "Copy a Tiny RDM export ZIP path to your clipboard, then press y to import.\n\nClipboard: \"" + clipboardContent + "\""
		NewPageComponentConfirm("Import connections", noticeString, func() {
			// 导入连接信息
			if clipboardContent == "" {
				GlobalTipComponent.LayoutTemporary("Clipboard is empty", 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			if !strings.HasSuffix(clipboardContent, ".zip") {
				GlobalTipComponent.LayoutTemporary("Clipboard does not contain a .zip path", 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			apiResult := c.ImportConnections(clipboardContent)
			if !apiResult.Success {
				GlobalTipComponent.LayoutTemporary(apiResult.Msg, 5, TipTypeError)
				GlobalApp.Gui.SetCurrentView(c.Name)
				return
			}
			GlobalTipComponent.LayoutTemporary("Connections imported", 2, TipTypeSuccess)
			c.closeView()
			InitConnectionComponent()
		}, func() {
			// 取消导入
			GlobalTipComponent.LayoutTemporary("Import cancelled", 2, TipTypeWarning)
			GlobalApp.Gui.SetCurrentView(c.Name)
		})
		return nil
	})

	// 检查更新
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'u'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		haveNewVersion, msg := CheckOutNewVersion()
		if haveNewVersion {
			NewPageComponentConfirm("New version available", "A new version is available. Open the download page in your browser?", func() {
				// 浏览器打开
				browser.OpenURL(msg)
			}, func() {
				GlobalTipComponent.LayoutTemporary("Download cancelled", 2, TipTypeWarning)
			})
		} else {
			GlobalTipComponent.LayoutTemporary("No new version available: "+msg, 2, TipTypeSuccess)
		}
		return nil
	})

	return c
}

func (c *LTRConnectionComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{}
	if c.ConnectionListCurrentGroupIndex >= 0 {
		keyMap = []KeyMapStruct{
			{"Move", "↑/↓/j/k (cross-group)"},
			{"Connect", "<Enter>/l/→"},
			{"Back", "<Esc>/h/←"},
			{"Edit/New/Delete", "<e>/<n>/<d>"},
			{"Export/Import/Update", "<E>/<I>/<u>"},
			{"Quit/Help", "<Ctrl+q>/<?>"},
		}
	} else {
		keyMap = []KeyMapStruct{
			{"Move", "↑/↓/j/k"},
			{"Open Group", "<Enter>/l/→"},
			{"Edit/New/Delete", "<e>/<n>/<d>"},
			{"Export/Import/Update", "<E>/<I>/<u>"},
			{"Quit/Help", "<Ctrl+q>/<?>"},
		}
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	// return "connection_list: " + ret
	return ret
}

func (c *LTRConnectionComponent) moveConnectionSelection(step int) {
	if c.ConnectionListCurrentGroupIndex < 0 || c.ConnectionListCurrentGroupIndex >= len(c.ConnectionList) {
		return
	}
	currentConnections := c.ConnectionList[c.ConnectionListCurrentGroupIndex].Connections
	if len(currentConnections) == 0 {
		nextGroup := c.findNextGroupWithConnections(c.ConnectionListCurrentGroupIndex, step)
		if nextGroup < 0 {
			c.ConnectionListCurrentGroupIndex = -1
			c.ConnectionListSelectedGroupIndex = -1
			c.ConnectionListSelectedConnectionIndex = -1
			return
		}
		c.ConnectionListCurrentGroupIndex = nextGroup
		c.ConnectionListSelectedGroupIndex = nextGroup
		if step > 0 {
			c.ConnectionListSelectedConnectionIndex = 0
		} else {
			c.ConnectionListSelectedConnectionIndex = len(c.ConnectionList[nextGroup].Connections) - 1
		}
		return
	}

	if c.ConnectionListSelectedConnectionIndex < 0 || c.ConnectionListSelectedConnectionIndex >= len(currentConnections) {
		if step > 0 {
			c.ConnectionListSelectedConnectionIndex = 0
		} else {
			c.ConnectionListSelectedConnectionIndex = len(currentConnections) - 1
		}
		return
	}

	nextIndex := c.ConnectionListSelectedConnectionIndex + step
	if nextIndex >= 0 && nextIndex < len(currentConnections) {
		c.ConnectionListSelectedConnectionIndex = nextIndex
		return
	}

	nextGroup := c.findNextGroupWithConnections(c.ConnectionListCurrentGroupIndex, step)
	if nextGroup < 0 {
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
		return
	}

	c.ConnectionListCurrentGroupIndex = nextGroup
	c.ConnectionListSelectedGroupIndex = nextGroup
	if step > 0 {
		c.ConnectionListSelectedConnectionIndex = 0
	} else {
		c.ConnectionListSelectedConnectionIndex = len(c.ConnectionList[nextGroup].Connections) - 1
	}
}

func (c *LTRConnectionComponent) findNextGroupWithConnections(from int, step int) int {
	if len(c.ConnectionList) == 0 {
		return -1
	}
	listLen := len(c.ConnectionList)
	for i := 1; i <= listLen; i++ {
		next := from + i*step
		for next < 0 {
			next += listLen
		}
		next = next % listLen
		if len(c.ConnectionList[next].Connections) > 0 {
			return next
		}
	}
	return -1
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

		// 检查和创建 TinyRDM 目录
		if !fileutil.IsDir(path.Join(userdir.GetConfigHome(), "TinyRDM")) {
			if err = os.MkdirAll(path.Join(userdir.GetConfigHome(), "TinyRDM"), 0755); err != nil {
				resp.Msg = err.Error()
				return
			}
		}

		outputFile, err := os.Create(path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename))

		// PrintLn(path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename))
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
