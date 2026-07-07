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

// flatItem is a single visible row in the connection list.
type flatItem struct {
	groupIdx  int // index into ConnectionList, -1 for none
	connIdx   int // index into group.Connections, -1 for group header
	isGroup   bool
	groupName string
	connName  string
	connCount int
	expanded  bool
}

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

	// flat list model
	flatItems      []flatItem
	flatCursor     int // 当前光标在 flatItems 中的位置
	expandedGroups map[int]bool
	flatOrigin     int
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
		flatCursor:                            0,
		flatOrigin:                            0,
		expandedGroups:                        map[int]bool{},
	}
	// 兼容一级非目录的配置
	connectionList := GlobalConnectionComponent.ConnectionList
	noGriupConnection := []types.Connection{}
	for _, group := range connectionListJson.Data.(types.Connections) {
		if len(group.Connections) == 0 && group.Type != "group" {
			noGriupConnection = append(noGriupConnection, group)
		} else if group.Type == "group" && group.Name == "" {
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
	// 默认展开所有有连接的组
	for i, g := range connectionList {
		if g.Name != "" && len(g.Connections) > 0 {
			GlobalConnectionComponent.expandedGroups[i] = true
		}
	}
	GlobalConnectionComponent.rebuildFlatItems()
	GlobalApp.ViewNameList = []string{GlobalConnectionComponent.Name}
	GlobalConnectionComponent.Layout().KeyBind()
	GlobalApp.Gui.SetCurrentView(GlobalConnectionComponent.Name)
}

// rebuildFlatItems 根据当前分组展开状态重建扁平列表
func (c *LTRConnectionComponent) rebuildFlatItems() {
	c.flatItems = c.flatItems[:0]
	for i, group := range c.ConnectionList {
		if group.Name == "" {
			continue
		}
		expanded := c.expandedGroups[i]
		c.flatItems = append(c.flatItems, flatItem{
			groupIdx:  i,
			connIdx:   -1,
			isGroup:   true,
			groupName: group.Name,
			connCount: len(group.Connections),
			expanded:  expanded,
		})
		if expanded {
			for j, conn := range group.Connections {
				c.flatItems = append(c.flatItems, flatItem{
					groupIdx: i,
					connIdx:  j,
					isGroup:  false,
					connName: conn.Name,
				})
			}
		}
	}
	// 修正光标
	if c.flatCursor >= len(c.flatItems) {
		c.flatCursor = len(c.flatItems) - 1
	}
	if c.flatCursor < 0 {
		c.flatCursor = 0
	}
	if len(c.flatItems) == 0 {
		c.flatCursor = 0
	}
}

// currentFlatItem 返回当前光标所在项
func (c *LTRConnectionComponent) currentFlatItem() *flatItem {
	if c.flatCursor < 0 || c.flatCursor >= len(c.flatItems) {
		return nil
	}
	return &c.flatItems[c.flatCursor]
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
		v.TitleColor = gocui.ColorCyan
		_, c.LayoutMaxY = v.Size()
	}
	GlobalApp.Gui.SetCurrentView(c.Name)

	cur := c.currentFlatItem()
	if cur != nil {
		if cur.isGroup {
			v.Subtitle = " Group: " + cur.groupName + " (" + fmt.Sprintf("%d", cur.connCount) + ") | Enter: expand/collapse | e/n/d: edit/new/delete "
		} else {
			v.Subtitle = " Connection: " + cur.connName + " | Enter: connect | e/n/d: edit/new/delete "
		}
	} else {
		v.Subtitle = " No connections | n: new group "
	}

	builder := strings.Builder{}
	totalLine := 0
	cursorLine := 0
	for i, item := range c.flatItems {
		var line string
		if item.isGroup {
			arrow := "▸"
			if item.expanded {
				arrow = "▾"
			}
			line = fmt.Sprintf("%s [%s] (%d)\n", arrow, item.groupName, item.connCount)
		} else {
			line = fmt.Sprintf("    %s\n", item.connName)
		}

		if i == c.flatCursor {
			builder.WriteString(NewColorString(line, "black", "cyan", "bold"))
			cursorLine = totalLine
		} else {
			if item.isGroup {
				builder.WriteString(NewColorString(line, "cyan", "", "bold"))
			} else {
				builder.WriteString(line)
			}
		}
		totalLine++
	}

	// 自动滚动
	viewH := c.LayoutMaxY
	if viewH <= 0 {
		viewH = 1
	}
	if totalLine > viewH {
		originLine := c.flatOrigin
		if cursorLine < originLine {
			originLine = cursorLine
		}
		if cursorLine >= originLine+viewH {
			originLine = cursorLine - viewH + 1
		}
		if originLine > totalLine-viewH {
			originLine = totalLine - viewH
		}
		if originLine < 0 {
			originLine = 0
		}
		c.flatOrigin = originLine
		v.SetOrigin(0, originLine)
	} else {
		c.flatOrigin = 0
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
		if len(c.flatItems) == 0 {
			return nil
		}
		c.flatCursor++
		if c.flatCursor >= len(c.flatItems) {
			c.flatCursor = 0
		}
		c.syncLegacyIndices()
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		if len(c.flatItems) == 0 {
			return nil
		}
		c.flatCursor--
		if c.flatCursor < 0 {
			c.flatCursor = len(c.flatItems) - 1
		}
		c.syncLegacyIndices()
		v.Clear()
		c.Layout()
		return nil
	})

	// Enter / l / →: 展开/折叠组 或 连接
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyEnter, gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		cur := c.currentFlatItem()
		if cur == nil {
			return nil
		}
		if cur.isGroup {
			// 切换展开/折叠
			c.expandedGroups[cur.groupIdx] = !c.expandedGroups[cur.groupIdx]
			c.rebuildFlatItems()
			// 保持光标在同一组头
			for i, item := range c.flatItems {
				if item.isGroup && item.groupIdx == cur.groupIdx {
					c.flatCursor = i
					break
				}
			}
			c.syncLegacyIndices()
			v.Clear()
			c.Layout()
		} else {
			// 连接
			c.connectCurrent()
		}
		return nil
	})

	// 编辑
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'e'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		cur := c.currentFlatItem()
		if cur == nil {
			return nil
		}
		if cur.isGroup {
			// 编辑组
			c.closeView()
			connectionComponent := InitConnectionEditComponent(c.ConnectionList[cur.groupIdx])
			connectionComponent.Layout()
		} else {
			// 编辑连接
			c.closeView()
			connectionComponent := InitConnectionEditComponent(c.ConnectionList[cur.groupIdx].Connections[cur.connIdx])
			connectionComponent.Layout()
		}
		return nil
	})

	// 新建
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		cur := c.currentFlatItem()
		if cur != nil && !cur.isGroup {
			// 在当前组内新建连接
			c.closeView()
			connectionComponent := InitConnectionEditComponent(types.Connection{
				ConnectionConfig: types.ConnectionConfig{
					Group: c.ConnectionList[cur.groupIdx].Name,
					Port:  6379,
					SSH: types.ConnectionSSH{
						Enable:    false,
						LoginType: "pwd",
						Port:      22,
					},
				},
			})
			connectionComponent.Layout()
		} else if cur != nil && cur.isGroup {
			// 在当前组内新建连接
			c.closeView()
			connectionComponent := InitConnectionEditComponent(types.Connection{
				ConnectionConfig: types.ConnectionConfig{
					Group: c.ConnectionList[cur.groupIdx].Name,
					Port:  6379,
					SSH: types.ConnectionSSH{
						Enable:    false,
						LoginType: "pwd",
						Port:      22,
					},
				},
			})
			connectionComponent.Layout()
		} else {
			// 没有选中任何东西，新建组
			c.closeView()
			connectionComponent := InitConnectionEditComponent(types.Connection{
				Type: "group",
			})
			connectionComponent.Layout()
		}
		return nil
	})

	// 删除
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{'d'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		cur := c.currentFlatItem()
		if cur == nil {
			return nil
		}
		if cur.isGroup {
			// 删除组（必须空）
			if len(c.ConnectionList[cur.groupIdx].Connections) > 0 {
				GlobalTipComponent.LayoutTemporary("Cannot delete non-empty group", 3, TipTypeError)
				return nil
			}
			NewPageComponentConfirm("Delete group", "Are you sure to delete this group?", func() {
				apiResult := services.Connection().DeleteGroup(c.ConnectionList[cur.groupIdx].Name, false)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Group deleted", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Failed to delete group", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				GlobalTipComponent.LayoutTemporary("Delete cancelled", 2, TipTypeWarning)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})
		} else {
			// 删除连接
			NewPageComponentConfirm("Delete connection", "Are you sure to delete this connection?", func() {
				connName := c.ConnectionList[cur.groupIdx].Connections[cur.connIdx].Name
				apiResult := services.Connection().DeleteConnection(connName)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Connection deleted", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Failed to delete connection", 3, TipTypeError)
				}
				c.closeView()
				InitConnectionComponent()
			}, func() {
				GlobalTipComponent.LayoutTemporary("Delete cancelled", 2, TipTypeWarning)
				GlobalApp.Gui.SetCurrentView(c.Name)
			})
		}
		return nil
	})

	// h / ← / Esc: 折叠当前组（如果在组内）或什么都不做
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyArrowLeft, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		cur := c.currentFlatItem()
		if cur == nil {
			return nil
		}
		if !cur.isGroup {
			// 在连接行上，先跳到所属组头
			for i, item := range c.flatItems {
				if item.isGroup && item.groupIdx == cur.groupIdx {
					c.flatCursor = i
					break
				}
			}
			c.syncLegacyIndices()
			v.Clear()
			c.Layout()
		} else {
			// 在组头上，折叠
			c.expandedGroups[cur.groupIdx] = false
			c.rebuildFlatItems()
			for i, item := range c.flatItems {
				if item.isGroup && item.groupIdx == cur.groupIdx {
					c.flatCursor = i
					break
				}
			}
			c.syncLegacyIndices()
			v.Clear()
			c.Layout()
		}
		return nil
	})

	// Esc: 折叠所有组
	GuiSetKeysbinding(GlobalApp.Gui, c.Name, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isConnecting {
			return nil
		}
		c.expandedGroups = map[int]bool{}
		c.rebuildFlatItems()
		c.syncLegacyIndices()
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
		clipboardContent, _ := clipboard.ReadAll()
		clipboardContent = strings.TrimSpace(clipboardContent)
		noticeString := "Copy a Tiny RDM export ZIP path to your clipboard, then press y to import.\n\nClipboard: \"" + clipboardContent + "\""
		NewPageComponentConfirm("Import connections", noticeString, func() {
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

// connectCurrent 连接当前光标选中的连接
func (c *LTRConnectionComponent) connectCurrent() {
	cur := c.currentFlatItem()
	if cur == nil || cur.isGroup {
		return
	}
	if cur.groupIdx >= len(c.ConnectionList) || cur.connIdx < 0 || cur.connIdx >= len(c.ConnectionList[cur.groupIdx].Connections) {
		GlobalTipComponent.LayoutTemporary("No connection selected", 2, TipTypeWarning)
		return
	}
	GlobalTipComponent.LayoutTemporary("Connecting to Redis...", 10, TipTypeWarning)
	c.isConnecting = true
	conn := c.ConnectionList[cur.groupIdx].Connections[cur.connIdx]
	if c.ConnectionListSelectedConnectionInfo.Name != "" {
		services.Browser().CloseConnection(c.ConnectionListSelectedConnectionInfo.Name)
	}
	c.ConnectionListSelectedConnectionInfo = conn
	go func() {
		connectionInfo := services.Browser().OpenConnection(c.ConnectionListSelectedConnectionInfo.Name)
		if connectionInfo.Success {
			GlobalTipComponent.LayoutTemporary("Connected successfully", 2, TipTypeSuccess)
			c.dbs = connectionInfo.Data.(map[string]any)["db"].([]types.ConnectionDB)
			c.view = connectionInfo.Data.(map[string]any)["view"].(int)
			c.lastDB = connectionInfo.Data.(map[string]any)["lastDB"].(int)
			c.version = connectionInfo.Data.(map[string]any)["version"].(string)
			GlobalApp.Gui.DeleteView(c.Name)
			GlobalApp.Gui.DeleteKeybindings(c.Name)
			GlobalApp.ViewNameList = []string{}
			c.closeView()
			InitDBComponent()
		} else {
			GlobalTipComponent.LayoutTemporary("Failed to connect", 5, TipTypeError)
			GlobalApp.Gui.SetCurrentView(c.Name)
		}
		c.isConnecting = false
	}()
}

// syncLegacyIndices 同步旧索引字段，兼容外部引用
func (c *LTRConnectionComponent) syncLegacyIndices() {
	cur := c.currentFlatItem()
	if cur == nil {
		c.ConnectionListSelectedGroupIndex = -1
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
		return
	}
	if cur.isGroup {
		c.ConnectionListSelectedGroupIndex = cur.groupIdx
		c.ConnectionListCurrentGroupIndex = -1
		c.ConnectionListSelectedConnectionIndex = -1
	} else {
		c.ConnectionListSelectedGroupIndex = cur.groupIdx
		c.ConnectionListCurrentGroupIndex = cur.groupIdx
		c.ConnectionListSelectedConnectionIndex = cur.connIdx
	}
}

func (c *LTRConnectionComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Move", "↑/↓/j/k"},
		{"Expand/Collapse", "<Enter>/l/→"},
		{"Collapse Group", "<h>/←"},
		{"Collapse All", "<Esc>"},
		{"Connect", "<Enter> (on connection)"},
		{"Edit/New/Delete", "<e>/<n>/<d>"},
		{"Export/Import/Update", "<E>/<I>/<u>"},
		{"Quit/Help", "<Ctrl+q>/<?>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	return ret
}

func (c *LTRConnectionComponent) closeView() {
	GlobalApp.Gui.DeleteView(c.Name)
	GlobalApp.Gui.DeleteKeybindings(c.Name)
	GlobalApp.ViewNameList = []string{}
}

// ExportConnections 导出连接信息
func (c *LTRConnectionComponent) ExportConnections() (resp types.JSResp) {
	defaultFileName := "connections_" + time.Now().Format("20060102150405") + ".zip"

	userDownloadDir, err := GetDownloadPath()
	if err != nil {
		userDownloadDir = "~"
	}
	filepath := path.Join(userDownloadDir, defaultFileName)

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

		if !fileutil.IsDir(path.Join(userdir.GetConfigHome(), "TinyRDM")) {
			if err = os.MkdirAll(path.Join(userdir.GetConfigHome(), "TinyRDM"), 0755); err != nil {
				resp.Msg = err.Error()
				return
			}
		}

		outputFile, err := os.Create(path.Join(userdir.GetConfigHome(), "TinyRDM", connectionFilename))
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
