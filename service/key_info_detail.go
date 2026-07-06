package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/duke-git/lancet/v2/validator"

	"github.com/awesome-gocui/gocui"
)

type LTRKeyInfoDetailComponent struct {
	name           string
	title          string
	LayoutMaxY     int
	view           *gocui.View
	keyValueFormat string
	viewOriginY    int // view origin y
	keyValueMaxY   int // value real total height
	CopyString     string
	lineView       *gocui.View
	selectedRow    int
	structuredRows []keyDetailRow
	structuredMode bool
	currentKeyType string
	listFilter     string
	listFiltered   []keyDetailRow
	listFilterEdit string
}

const listFilterViewName = "key_detail_list_filter"

type keyDetailRow struct {
	Index int
	Field string
	Value string
	Score string
}

type keyOpDialogField struct {
	Label       string
	Placeholder string
	Value       string
}

type keyOpDialogSchema struct {
	Title       string
	Description string
	Fields      []keyOpDialogField
	BuildJSON   func(values map[string]string) (string, error)
}

var keyValueFormatList = []string{"Raw", "JSON", "Unicode JSON"}

func InitKeyInfoDetailComponent() {
	GlobalKeyInfoDetailComponent = &LTRKeyInfoDetailComponent{
		name:           "key_info_detail",
		title:          "Detail",
		LayoutMaxY:     0,
		keyValueFormat: "Raw",
	}
	GlobalKeyInfoDetailComponent.Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoDetailComponent.name)
	GlobalTipComponent.AppendList(GlobalKeyInfoDetailComponent.name, GlobalKeyInfoDetailComponent.KeyMapTip())
}

func (c *LTRKeyInfoDetailComponent) LayoutTitle() *LTRKeyInfoDetailComponent {
	if c.view != nil && CurrentViewName() == c.name {
		c.view.Title = " [" + c.title + "] "
		c.lineView.FrameColor = gocui.ColorGreen
	} else {
		c.view.Title = " " + c.title + " "
		c.lineView.FrameColor = gocui.ColorDefault
	}
	return c
}

func (c *LTRKeyInfoDetailComponent) Layout() *LTRKeyInfoDetailComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	theVal := ""
	maxLine := 0
	lineStr := ""
	lineStrNo := 1
	lineViewWidth := 0
	lineViewWidthStr := "1"
	// show key detail
	c.view, err = SetViewSafe(c.name, theX0+1, 3, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyValueMaxY = 0
		c.view.Wrap = true
		// c.view.Title = " " + c.title + " "
		if CurrentViewName() == c.name {
			c.view.Title = " [" + c.title + "] "
		} else {
			c.view.Title = " " + c.title + " "
		}
		keyDetail := services.Browser().GetKeyDetail(types.KeyDetailParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
		})
		if keyDetail.Success {
			keyDetailData := keyDetail.Data.(types.KeyDetail)
			theVal = c.buildDisplayValue(keyDetailData)
			theValSlice := strings.Split(theVal, "\n")
			maxLine = len(theValSlice) - 1
			if maxLine < 0 {
				maxLine = 0
			}
			lineViewWidth = len(strconv.Itoa(maxLine))
			lineViewWidthStr = strconv.Itoa(lineViewWidth)
			if c.structuredMode {
				lineViewWidth = 0
				lineViewWidthStr = "1"
				c.view.Wrap = false
				c.view, _ = SetViewSafe(c.name, theX0+1, 3, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
			} else {
				// reset view x0 , later affects the view width
				c.view, _ = SetViewSafe(c.name, theX0+1+lineViewWidth, 3, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
			}
			theViewX, _ := c.view.Size()
			for k, line := range theValSlice {
				if k == maxLine {
					// 跳过最后一行
					break
				}
				lineLen := DisplayWidth(line)

				if lineLen > theViewX {
					theRealHeight := 0
					theRealHeight = lineLen / theViewX
					if lineLen%theViewX > 0 {
						theRealHeight++
					}
					c.keyValueMaxY += theRealHeight
					for i := 0; i < theRealHeight; i++ {
						if i == 0 {
							lineStr += fmt.Sprintf("%"+lineViewWidthStr+"d", lineStrNo) + "\n"
						} else {
							lineStr += "\n"
						}
					}
				} else {
					c.keyValueMaxY++
					lineStr += fmt.Sprintf("%"+lineViewWidthStr+"d", lineStrNo) + "\n"
				}
				lineStrNo++
			}
			// PrintLn(c.keyValueMaxY)
		} else {
			theVal = fmt.Sprintln("")
		}
	}
	if maxLine > 0 {
		subtitle := " Lines: " + strconv.Itoa(maxLine) + " "
		if len(c.structuredRows) > 0 {
			rows := c.getActiveSelectionRows()
			if len(rows) == 0 {
				subtitle += " Row: 0/0 "
			} else {
				subtitle += " Row: " + strconv.Itoa(c.selectedRow+1) + "/" + strconv.Itoa(len(rows)) + " "
			}
			if c.currentKeyType == "list" && strings.TrimSpace(c.listFilter) != "" {
				subtitle += " Filtered " + strconv.Itoa(len(rows)) + "/" + strconv.Itoa(len(c.structuredRows)) + " "
			}
		}
		if c.currentKeyType != "" {
			subtitle = " Type: " + c.currentKeyType + " " + subtitle
		}
		c.view.Subtitle = subtitle
	} else {
		c.view.Subtitle = ""
	}
	c.view.Clear()
	// theValRune = theValRune[:GlobalApp.maxX-theX0-2]
	// theVal = string(theValRune)
	// theVal = text.TrimSpace(theVal)
	c.CopyString = theVal
	// c.view.Write(DisposeMultibyteString(theVal))
	c.view.Write([]byte(theVal))

	// show format select
	formatStr := " Format: " + c.keyValueFormat + " "
	formatSelectView, err := SetViewSafe("key_value_format", GlobalApp.maxX-len(formatStr)-2, GlobalApp.maxY-4, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		formatSelectView.Clear()
		formatSelectView.Write([]byte(formatStr))
	}
	formatSelectView.Frame = false
	formatSelectView.BgColor = gocui.ColorGreen
	c.layoutListFilterView(theX0, len(formatStr))

	if c.structuredMode {
		c.viewOriginY = 0
	}
	c.view.SetOrigin(0, c.viewOriginY)

	// line view
	c.lineView, err = SetViewSafe("key_detail_line", theX0, 3, theX0+6, GlobalApp.maxY-2, 1)
	if err == nil || err != gocui.ErrUnknownView {
		// c.lineView.FrameColor = gocui.NewRGBColor(149, 165, 166)
		c.lineView.FgColor = gocui.NewRGBColor(78, 142, 166)
		c.lineView.Clear()
		if !c.structuredMode {
			c.lineView.Write([]byte(lineStr))
		}
		c.lineView.SetOrigin(0, 0)
	}
	c.lineView.FrameRunes = []rune{'─', '│', '┌', '─', '└', '─'}
	if c.structuredMode {
		c.lineView.Frame = false
	} else {
		c.lineView.Frame = true
	}

	// reset view x0 and x1
	if c.structuredMode {
		c.lineView, _ = SetViewSafe("key_detail_line", theX0, 3, theX0+1, GlobalApp.maxY-2, 1)
	} else {
		c.lineView, _ = SetViewSafe("key_detail_line", theX0, 3, theX0+1+lineViewWidth, GlobalApp.maxY-2, 1)
	}
	if CurrentViewName() == c.name && GlobalTipComponent != nil {
		GlobalTipComponent.Layout(c.KeyMapTip())
	}

	return c
}

func (c *LTRKeyInfoDetailComponent) layoutListFilterView(theX0 int, formatWidth int) {
	if c.currentKeyType != "list" {
		GlobalApp.Gui.DeleteKeybindings(listFilterViewName)
		GlobalApp.Gui.DeleteView(listFilterViewName)
		return
	}

	if strings.TrimSpace(c.listFilterEdit) == "" {
		c.listFilterEdit = c.listFilter
	}

	x0 := theX0 + 2
	x1 := GlobalApp.maxX - formatWidth - 4
	if x1 <= x0+10 {
		x1 = GlobalApp.maxX - 3
	}
	v, err := SetViewSafe(listFilterViewName, x0, GlobalApp.maxY-4, x1, GlobalApp.maxY-2, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return
	}
	showText := c.listFilterEdit
	if strings.TrimSpace(showText) == "" {
		showText = "(type keyword then Enter)"
	}
	v.Clear()
	v.Write([]byte(" Filter: " + showText + " "))
	v.Frame = false
	v.Editable = CurrentViewName() == listFilterViewName
	if v.Editable {
		v.BgColor = gocui.ColorYellow
		v.Editor = &EditorInput{BindValString: &c.listFilterEdit}
		_ = v.SetCursor(len([]rune(c.listFilterEdit))+9, 0)
	} else {
		v.BgColor = gocui.ColorBlack
	}
}

func (c *LTRKeyInfoDetailComponent) KeyBind() {
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalApp.Gui.DeleteKeybindings("key_value_format")
	GlobalApp.Gui.DeleteKeybindings(listFilterViewName)
	// format switch
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'f'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.switchKeyValueFormat()
		return nil
	})

	//copy
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'c'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		theVal := c.CopyString
		if theVal == "" {
			GlobalTipComponent.LayoutTemporary("No value to copy", 2, TipTypeWarning)
			return nil
		}
		clipboard.WriteAll(theVal)
		GlobalTipComponent.LayoutTemporary("Copied value to clipboard", 3, TipTypeSuccess)
		return nil
	})
	// scroll
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isCurrentListType() {
			c.moveDetailRowSelection(-1)
			return nil
		}
		c.scroll(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isCurrentListType() {
			c.moveDetailRowSelection(1)
			return nil
		}
		c.scroll(1)
		return nil
	})
	// scroll page
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowLeft, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isCurrentListType() {
			c.moveDetailRowSelection(-10)
			return nil
		}
		c.scroll(-GlobalApp.maxY + 9)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isCurrentListType() {
			c.moveDetailRowSelection(10)
			return nil
		}
		c.scroll(GlobalApp.maxY - 9)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'/'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.isCurrentListType() {
			GlobalTipComponent.LayoutTemporary("Filter is available in list detail mode", 2, TipTypeWarning)
			return nil
		}
		c.listFilterEdit = c.listFilter
		c.Layout()
		if _, err := GlobalApp.Gui.SetCurrentView(listFilterViewName); err != nil {
			GlobalTipComponent.LayoutTemporary("Open filter input failed", 3, TipTypeError)
			return nil
		}
		if fv, ferr := GlobalApp.Gui.View(listFilterViewName); ferr == nil {
			fv.Editable = true
			fv.Editor = &EditorInput{BindValString: &c.listFilterEdit}
			_ = fv.SetCursor(len([]rune(c.listFilterEdit))+9, 0)
		}
		GlobalApp.Gui.Cursor = true
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'x'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.isCurrentListType() {
			return nil
		}
		c.listFilter = ""
		c.applyListFilter()
		c.Layout()
		GlobalTipComponent.LayoutTemporary("List filter cleared", 2, TipTypeSuccess)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, listFilterViewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.listFilter = strings.TrimSpace(c.listFilterEdit)
		c.applyListFilter()
		c.Layout()
		_, _ = GlobalApp.Gui.SetCurrentView(c.name)
		GlobalApp.Gui.Cursor = false
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, listFilterViewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.listFilterEdit = c.listFilter
		c.Layout()
		_, _ = GlobalApp.Gui.SetCurrentView(c.name)
		GlobalApp.Gui.Cursor = false
		return nil
	})

	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

	// key_value_format
	GuiSetKeysbinding(GlobalApp.Gui, "key_value_format", []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.switchKeyValueFormat()
		return nil
	})

	// 刷新
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'r'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	// 粘贴-修改值
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'p'}, "Replace value using clipboard content?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or unavailable", 3, TipTypeError)
			return
		}
		if GlobalKeyInfoComponent.keyName == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeWarning)
			return
		}
		keyType, typeErr := c.getCurrentSetKeyType()
		if typeErr != nil {
			GlobalTipComponent.LayoutTemporary("Failed to read key type: "+typeErr.Error(), 3, TipTypeError)
			return
		}
		if isCollectionKeyType(keyType) {
			GlobalTipComponent.LayoutTemporary("Use <e>/<a>/<u>/<d> form dialog for this key type", 4, TipTypeWarning)
			return
		}
		if err := c.applyPrimaryEdit(theClipboardValue); err != nil {
			GlobalTipComponent.LayoutTemporary("Failed to apply edit: "+err.Error(), 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Edit applied", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Value update cancelled", 3, TipTypeWarning)
	})

	// 修改值
	GuiSetKeysbindingConfirmWithVIEditor(GlobalApp.Gui, c.name, []any{'e'}, "", func() string {
		return c.getPrimaryEditTemplate()
	}, func(editorResult string) {
		if GlobalKeyInfoComponent.keyName == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeWarning)
			return
		}

		if err := c.applyPrimaryEdit(editorResult); err != nil {
			GlobalTipComponent.LayoutTemporary("Failed to apply edit: "+err.Error(), 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Edit applied", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Value update cancelled", 3, TipTypeWarning)
	}, false, func() bool {
		if strings.TrimSpace(GlobalKeyInfoComponent.keyName) == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeWarning)
			return false
		}
		keyType, err := c.getCurrentSetKeyType()
		if err == nil && isCollectionKeyType(keyType) {
			if openErr := c.openTypeOperationDialog("update", ""); openErr != nil {
				GlobalTipComponent.LayoutTemporary("Failed to open dialog: "+openErr.Error(), 4, TipTypeError)
			}
			return false
		}
		return true
	})

	c.bindTypeOperationKeys()
}

func (c *LTRKeyInfoDetailComponent) KeyMapTip() string {
	keyType, _ := c.getCurrentSetKeyType()
	editDesc := "Edit/Save"
	addDesc := "Add Row"
	updateDesc := "Edit Row"
	deleteDesc := "Delete Row"
	switch keyType {
	case "hash":
		addDesc = "Add Field"
		updateDesc = "Edit Field"
		deleteDesc = "Delete Field"
	case "set":
		addDesc = "Add Member"
		updateDesc = "Replace Member"
		deleteDesc = "Delete Member"
	case "zset":
		addDesc = "Add Member+Score"
		updateDesc = "Edit Member+Score"
		deleteDesc = "Delete Member"
	case "list":
		addDesc = "Add Item"
		updateDesc = "Edit By Index"
		deleteDesc = "Delete By Index"
	case "stream":
		addDesc = "Add Entry"
		updateDesc = "N/A"
		deleteDesc = "Delete Entry"
	}
	if keyType == "string" || keyType == "json" {
		addDesc = "Type Add(<a>)"
		updateDesc = "Primary Edit"
		deleteDesc = "Type Del(<d>)"
	}

	keyMap := []KeyMapStruct{
		{"Scroll/Select", "↑/↓/j/k"},
		{"Scroll Page/Jump", "←/→/h/l"},
		{"Select Row", "j/k"},
		{"Filter", "</>/<x>"},
		{"Switch Format", "<f>"},
		{editDesc, "<e>"},
		{addDesc + "/" + updateDesc + "/" + deleteDesc, "<a>/<u>/<d>"},
		{"Copy", "<c>"},
		{"Paste", "<p>"},
		{"Refresh", "<r>"},
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
	// return "key_detail: " + ret
	return ret
}

func (c *LTRKeyInfoDetailComponent) scroll(n int) {
	c.viewOriginY += n
	if c.viewOriginY < 0 {
		c.viewOriginY = 0
	}
	_, theViewY := c.view.Size()
	if c.keyValueMaxY-theViewY <= c.viewOriginY {
		c.viewOriginY = c.keyValueMaxY - theViewY
	}
	c.view.SetOrigin(0, c.viewOriginY)
	c.lineView.SetOrigin(0, c.viewOriginY)
}

func (c *LTRKeyInfoDetailComponent) switchKeyValueFormat() {
	nextIndex := 0
	for i, format := range keyValueFormatList {
		if format == c.keyValueFormat {
			nextIndex = i + 1
			break
		}
	}
	if nextIndex >= len(keyValueFormatList) {
		nextIndex = 0
	}
	c.keyValueFormat = keyValueFormatList[nextIndex]
	c.viewOriginY = 0
	c.Layout()
}

func (c *LTRKeyInfoDetailComponent) normalizeSelectedRow() {
	rows := c.getActiveSelectionRows()
	if len(rows) == 0 {
		c.selectedRow = 0
		return
	}
	if c.selectedRow < 0 {
		c.selectedRow = 0
	}
	if c.selectedRow >= len(rows) {
		c.selectedRow = len(rows) - 1
	}
}

func (c *LTRKeyInfoDetailComponent) getSelectedStructuredRow() *keyDetailRow {
	rows := c.getActiveSelectionRows()
	if len(rows) == 0 {
		return nil
	}
	c.normalizeSelectedRow()
	row := rows[c.selectedRow]
	return &row
}

func (c *LTRKeyInfoDetailComponent) moveDetailRowSelection(step int) {
	rows := c.getActiveSelectionRows()
	if len(rows) == 0 {
		GlobalTipComponent.LayoutTemporary("No row to select", 2, TipTypeWarning)
		return
	}
	c.selectedRow += step
	c.normalizeSelectedRow()
	// For structured (list) mode, re-render from cached rows without re-fetching
	// from Redis. This is critical for trackpad scroll smoothness — every scroll
	// event would otherwise trigger a GetKeyDetail API round-trip.
	if c.structuredMode && len(c.structuredRows) > 0 {
		c.renderFromCache()
		return
	}
	c.Layout()
}

func (c *LTRKeyInfoDetailComponent) renderFromCache() {
	if c.view == nil || len(c.structuredRows) == 0 {
		c.Layout()
		return
	}
	c.applyListFilter()
	c.normalizeSelectedRow()
	var theVal string
	switch c.currentKeyType {
	case "list":
		theVal = c.renderListRowsFromCache()
	case "hash":
		theVal = c.renderHashRowsFromCache()
	case "set":
		theVal = c.renderSetRowsFromCache()
	case "zset":
		theVal = c.renderZSetRowsFromCache()
	case "stream":
		theVal = c.renderStreamRowsFromCache()
	default:
		c.Layout()
		return
	}
	// update subtitle
	rows := c.getActiveSelectionRows()
	if len(c.structuredRows) > 0 {
		subtitle := " Lines: " + strconv.Itoa(len(strings.Split(theVal, "\n"))-1) + " "
		if len(rows) == 0 {
			subtitle += " Row: 0/0 "
		} else {
			subtitle += " Row: " + strconv.Itoa(c.selectedRow+1) + "/" + strconv.Itoa(len(rows)) + " "
		}
		if c.currentKeyType == "list" && strings.TrimSpace(c.listFilter) != "" {
			subtitle += " Filtered " + strconv.Itoa(len(rows)) + "/" + strconv.Itoa(len(c.structuredRows)) + " "
		}
		if c.currentKeyType != "" {
			subtitle = " Type: " + c.currentKeyType + " " + subtitle
		}
		c.view.Subtitle = subtitle
	} else {
		c.view.Subtitle = ""
	}
	c.CopyString = theVal
	c.view.Clear()
	c.view.Write([]byte(theVal))
	c.view.SetOrigin(0, 0)
}

// renderListRowsFromCache renders the list detail using already-populated structuredRows.
func (c *LTRKeyInfoDetailComponent) renderListRowsFromCache() string {
	rows := c.getActiveSelectionRows()
	var b strings.Builder
	b.WriteString("Type: list (Explorer Mode)\n")
	b.WriteString("Actions: ↑/↓ select  ←/→ jump  </> filter  <x> clear filter  <a>/<e>/<u>/<d> CRUD\n")
	if strings.TrimSpace(c.listFilter) != "" {
		b.WriteString("Filter: \"" + c.listFilter + "\" | matched " + strconv.Itoa(len(rows)) + "/" + strconv.Itoa(len(c.structuredRows)) + "\n")
	}
	b.WriteString("================================================================================\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String()
	}
	if len(rows) == 0 {
		b.WriteString("No rows match current filter. Press <x> to clear or </> to update filter.\n")
		return b.String()
	}
	start := c.selectedRow - 6
	if start < 0 {
		start = 0
	}
	end := start + 12
	if end > len(rows) {
		end = len(rows)
		start = end - 12
		if start < 0 {
			start = 0
		}
	}
	b.WriteString("List Rows\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-2s %-8s %-s\n", "", "INDEX", "VALUE(PREVIEW)"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	for i := start; i < end; i++ {
		row := rows[i]
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		preview := truncateByRuneCount(strings.ReplaceAll(row.Value, "\n", " <NL> "), 90)
		b.WriteString(fmt.Sprintf("%s %-8d %s\n", prefix, row.Index, preview))
	}
	selected := rows[c.selectedRow]
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("Selected Index: %d (row %d/%d)\n", selected.Index, c.selectedRow+1, len(rows)))
	b.WriteString("Selected Value\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	for _, line := range strings.Split(selected.Value, "\n") {
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) renderHashRowsFromCache() string {
	var b strings.Builder
	b.WriteString("Type: hash\n")
	b.WriteString("Actions: <a> Add Field  <u>/<e> Edit Field  <d> Delete Field\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-24s %-s\n", "FIELD", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String()
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-24s %s\n", prefix, row.Field, row.Value))
	}
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) renderSetRowsFromCache() string {
	var b strings.Builder
	b.WriteString("Type: set\n")
	b.WriteString("Actions: <a> Add Member  <u>/<e> Replace Member  <d> Delete Member\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-8s %-s\n", "ROW", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String()
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-8d %s\n", prefix, row.Index, row.Value))
	}
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) renderZSetRowsFromCache() string {
	var b strings.Builder
	b.WriteString("Type: zset\n")
	b.WriteString("Actions: <a> Add Member+Score  <u>/<e> Edit Member+Score  <d> Delete Member\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-10s %-s\n", "SCORE", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String()
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-10s %s\n", prefix, row.Score, row.Value))
	}
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) renderStreamRowsFromCache() string {
	var b strings.Builder
	b.WriteString("Type: stream\n")
	b.WriteString("Actions: <a> Add Entry  <d> Delete Entry(by ID)\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-26s %-s\n", "ENTRY ID", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String()
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-26s %s\n", prefix, row.Field, truncateByRuneCount(strings.ReplaceAll(row.Value, "\n", " <NL> "), 80)))
	}
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) isCurrentListType() bool {
	// Use the cached key type set by Layout() to avoid a Redis API call on every
	// scroll event. If the type hasn't been loaded yet, fall back to a one-shot fetch.
	if c.currentKeyType != "" {
		return c.currentKeyType == "list"
	}
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return false
	}
	return keyType == "list"
}

func (c *LTRKeyInfoDetailComponent) getActiveSelectionRows() []keyDetailRow {
	if c.currentKeyType == "list" && strings.TrimSpace(c.listFilter) != "" {
		return c.listFiltered
	}
	return c.structuredRows
}

func (c *LTRKeyInfoDetailComponent) applyListFilter() {
	if c.currentKeyType != "list" {
		c.listFiltered = nil
		return
	}
	keyword := strings.TrimSpace(c.listFilter)
	if keyword == "" {
		c.listFiltered = nil
		c.normalizeSelectedRow()
		return
	}
	needle := strings.ToLower(keyword)
	filtered := make([]keyDetailRow, 0, len(c.structuredRows))
	for _, row := range c.structuredRows {
		if strings.Contains(strings.ToLower(row.Value), needle) || strings.Contains(strconv.Itoa(row.Index), keyword) {
			filtered = append(filtered, row)
		}
	}
	c.listFiltered = filtered
	c.normalizeSelectedRow()
}

func (c *LTRKeyInfoDetailComponent) bindTypeOperationKeys() {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'a'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if err := c.openTypeOperationDialog("add", ""); err != nil {
			GlobalTipComponent.LayoutTemporary("Open add dialog failed: "+err.Error(), 4, TipTypeError)
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'u'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if err := c.openTypeOperationDialog("update", ""); err != nil {
			GlobalTipComponent.LayoutTemporary("Open update dialog failed: "+err.Error(), 4, TipTypeError)
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'d'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if err := c.openTypeOperationDialog("delete", ""); err != nil {
			GlobalTipComponent.LayoutTemporary("Open delete dialog failed: "+err.Error(), 4, TipTypeError)
		}
		return nil
	})
}

func (c *LTRKeyInfoDetailComponent) openTypeOperationDialog(operation, prefillValue string) error {
	if !c.canEditCurrentKey() {
		return fmt.Errorf("no key selected")
	}

	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return err
	}

	if !isCollectionKeyType(keyType) {
		return fmt.Errorf("this operation dialog is available for list/hash/set/zset")
	}

	schema, err := c.buildKeyOpDialogSchema(keyType, operation, prefillValue)
	if err != nil {
		return err
	}

	return c.showKeyOpDialog(schema, func(values map[string]string) error {
		payload, buildErr := schema.BuildJSON(values)
		if buildErr != nil {
			return buildErr
		}
		return c.applyTypeOperation(operation, payload)
	})
}

func (c *LTRKeyInfoDetailComponent) buildKeyOpDialogSchema(keyType, operation, prefillValue string) (keyOpDialogSchema, error) {
	base := keyOpDialogSchema{}
	selected := c.getSelectedStructuredRow()
	valueDefault := strings.TrimSpace(prefillValue)
	if strings.TrimSpace(valueDefault) == "" {
		if selected != nil && strings.TrimSpace(selected.Value) != "" {
			valueDefault = selected.Value
		} else {
			valueDefault = "value"
		}
	}
	fieldDefault := "field"
	if selected != nil && strings.TrimSpace(selected.Field) != "" {
		fieldDefault = selected.Field
	}
	indexDefault := "0"
	if selected != nil && selected.Index >= 0 {
		indexDefault = strconv.Itoa(selected.Index)
	}
	scoreDefault := "1"
	if selected != nil && strings.TrimSpace(selected.Score) != "" {
		scoreDefault = selected.Score
	}

	switch keyType {
	case "list":
		switch operation {
		case "add":
			base.Title = "List Add"
			base.Description = "Add new item to list"
			base.Fields = []keyOpDialogField{{Label: "Position(head/tail)", Placeholder: "tail", Value: "tail"}, {Label: "Value", Placeholder: "item", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				position := strings.TrimSpace(values["Position(head/tail)"])
				if position == "" {
					position = "tail"
				}
				obj := map[string]any{"position": position, "items": []string{values["Value"]}}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		case "update":
			base.Title = "List Update"
			base.Description = "Update list item by index"
			base.Fields = []keyOpDialogField{{Label: "Index", Placeholder: "0", Value: indexDefault}, {Label: "Value", Placeholder: "new value", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				idx, err := parseRequiredInt(values, "Index")
				if err != nil {
					return "", err
				}
				val, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"index": idx, "value": val}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		default:
			base.Title = "List Delete"
			base.Description = "Delete list item by index"
			base.Fields = []keyOpDialogField{{Label: "Index", Placeholder: "0", Value: indexDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				idx, err := parseRequiredInt(values, "Index")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"index": idx}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		}
	case "hash":
		switch operation {
		case "add":
			base.Title = "Hash Add"
			base.Description = "Add new field"
			base.Fields = []keyOpDialogField{{Label: "Field", Placeholder: "field", Value: fieldDefault}, {Label: "Value", Placeholder: "value", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				field, err := requireNonEmpty(values, "Field")
				if err != nil {
					return "", err
				}
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				obj := []map[string]any{{"field": field, "value": value}}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		case "update":
			base.Title = "Hash Update"
			base.Description = "Edit field/value"
			base.Fields = []keyOpDialogField{{Label: "Field", Placeholder: "old field", Value: fieldDefault}, {Label: "NewField", Placeholder: "same as field", Value: ""}, {Label: "Value", Placeholder: "new value", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				field, err := requireNonEmpty(values, "Field")
				if err != nil {
					return "", err
				}
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				newField := strings.TrimSpace(values["NewField"])
				if newField == "" {
					newField = field
				}
				obj := map[string]any{"field": field, "newField": newField, "value": value}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		default:
			base.Title = "Hash Delete"
			base.Description = "Delete field"
			base.Fields = []keyOpDialogField{{Label: "Field", Placeholder: "field", Value: fieldDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				field, err := requireNonEmpty(values, "Field")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"field": field}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		}
	case "set":
		switch operation {
		case "add":
			base.Title = "Set Add"
			base.Description = "Add member"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "member", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				buf, err := json.Marshal([]string{value})
				return string(buf), err
			}
		case "update":
			base.Title = "Set Update"
			base.Description = "Replace member"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "old member", Value: valueDefault}, {Label: "NewValue", Placeholder: "new member", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				newValue, err := requireNonEmpty(values, "NewValue")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"value": value, "newValue": newValue}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		default:
			base.Title = "Set Delete"
			base.Description = "Delete member"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "member", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				buf, err := json.Marshal([]string{value})
				return string(buf), err
			}
		}
	case "zset":
		switch operation {
		case "add":
			base.Title = "ZSet Add"
			base.Description = "Add member with score"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "member", Value: valueDefault}, {Label: "Score", Placeholder: "1", Value: scoreDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				score, err := parseRequiredFloat(values, "Score")
				if err != nil {
					return "", err
				}
				obj := []map[string]any{{"value": value, "score": score}}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		case "update":
			base.Title = "ZSet Update"
			base.Description = "Edit member/score"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "old member", Value: valueDefault}, {Label: "NewValue", Placeholder: "new member", Value: valueDefault}, {Label: "Score", Placeholder: "1", Value: scoreDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				newValue, err := requireNonEmpty(values, "NewValue")
				if err != nil {
					return "", err
				}
				score, err := parseRequiredFloat(values, "Score")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"value": value, "newValue": newValue, "score": score}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		default:
			base.Title = "ZSet Delete"
			base.Description = "Delete member"
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "member", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"value": value}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		}
	default:
		return keyOpDialogSchema{}, fmt.Errorf("unsupported key type: %s", keyType)
	}

	// stream operations handled separately below
	if keyType == "stream" {
		switch operation {
		case "add":
			base.Title = "Stream Add"
			base.Description = "Add entry to stream"
			base.Fields = []keyOpDialogField{{Label: "ID", Placeholder: "* (auto)", Value: "*"}, {Label: "Field", Placeholder: "field", Value: fieldDefault}, {Label: "Value", Placeholder: "value", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				id := strings.TrimSpace(values["ID"])
				if id == "" {
					id = "*"
				}
				field, err := requireNonEmpty(values, "Field")
				if err != nil {
					return "", err
				}
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"id": id, "field": field, "value": value}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		case "delete":
			base.Title = "Stream Delete"
			base.Description = "Delete stream entry by ID"
			idDefault := "*"
			if selected != nil && strings.TrimSpace(selected.Field) != "" {
				idDefault = selected.Field
			}
			base.Fields = []keyOpDialogField{{Label: "EntryID", Placeholder: "1234567890-0", Value: idDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				id, err := requireNonEmpty(values, "EntryID")
				if err != nil {
					return "", err
				}
				obj := map[string]any{"id": id}
				buf, err := json.Marshal(obj)
				return string(buf), err
			}
		default:
			return keyOpDialogSchema{}, fmt.Errorf("stream does not support %s operation", operation)
		}
	}

	return base, nil
}

func (c *LTRKeyInfoDetailComponent) showKeyOpDialog(schema keyOpDialogSchema, onSubmit func(values map[string]string) error) error {
	maskName := "key_op_dialog_mask"
	dialogName := "key_op_dialog"
	fieldPrefix := "key_op_dialog_field_"

	if _, err := SetViewSafe(maskName, 0, 0, GlobalApp.maxX-1, GlobalApp.maxY-1, 0); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	maskView, _ := GlobalApp.Gui.View(maskName)
	maskView.Frame = false
	maskView.Clear()

	width := 70
	if GlobalApp.maxX-4 < width {
		width = GlobalApp.maxX - 4
	}
	if width < 36 {
		width = 36
	}
	height := len(schema.Fields)*3 + 8
	if height > GlobalApp.maxY-2 {
		height = GlobalApp.maxY - 2
	}
	x0 := (GlobalApp.maxX - width) / 2
	y0 := (GlobalApp.maxY - height) / 2
	x1 := x0 + width - 1
	y1 := y0 + height - 1

	dlg, err := SetViewSafe(dialogName, x0, y0, x1, y1, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	dlg.Title = " " + schema.Title + " "
	dlg.Clear()
	dlg.Wrap = true
	dlg.Write([]byte(schema.Description + "\n"))
	dlg.Write([]byte("Tab/Up/Down switch fields | Enter submit | Esc cancel\n"))

	fieldNames := make([]string, 0, len(schema.Fields))
	fieldValues := make(map[string]*string, len(schema.Fields))
	for i, field := range schema.Fields {
		fieldViewName := fieldPrefix + strconv.Itoa(i)
		fy0 := y0 + 3 + i*3
		fy1 := fy0 + 2
		fv, ferr := SetViewSafe(fieldViewName, x0+2, fy0, x1-2, fy1, 0)
		if ferr != nil && ferr != gocui.ErrUnknownView {
			return ferr
		}
		fv.Title = " " + field.Label + " "
		fv.Clear()
		val := strings.TrimSpace(field.Value)
		if val == "" {
			val = field.Placeholder
		}
		fv.Write([]byte(val))
		bound := val
		fv.Editable = true
		fv.Editor = &EditorInput{BindValString: &bound}
		fieldNames = append(fieldNames, fieldViewName)
		fieldValues[field.Label] = &bound
	}

	currentIdx := 0
	focusField := func(idx int) {
		if idx < 0 {
			idx = len(fieldNames) - 1
		}
		if idx >= len(fieldNames) {
			idx = 0
		}
		currentIdx = idx
		for i, name := range fieldNames {
			view, viewErr := GlobalApp.Gui.View(name)
			if viewErr != nil {
				continue
			}
			if i == currentIdx {
				view.BgColor = gocui.ColorBlue
				GlobalApp.Gui.SetCurrentView(name)
				GlobalApp.Gui.Cursor = true
				view.SetCursor(0, 0)
			} else {
				view.BgColor = gocui.ColorBlack
			}
		}
	}

	closeDialog := func() {
		GlobalApp.Gui.DeleteView(dialogName)
		GlobalApp.Gui.DeleteKeybindings(dialogName)
		GlobalApp.Gui.DeleteView(maskName)
		GlobalApp.Gui.DeleteKeybindings(maskName)
		for _, name := range fieldNames {
			GlobalApp.Gui.DeleteView(name)
			GlobalApp.Gui.DeleteKeybindings(name)
		}
		GlobalApp.Gui.Cursor = false
		GlobalApp.Gui.SetCurrentView(c.name)
	}

	submit := func() {
		values := make(map[string]string, len(fieldValues))
		for label, ptr := range fieldValues {
			values[label] = strings.TrimSpace(*ptr)
		}
		if err := onSubmit(values); err != nil {
			GlobalTipComponent.LayoutTemporary("Apply failed: "+err.Error(), 4, TipTypeError)
			return
		}
		closeDialog()
		GlobalTipComponent.LayoutTemporary("Operation applied", 3, TipTypeSuccess)
		c.Layout()
	}

	bindNavKeys := func(viewName string) {
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyTab, gocui.KeyArrowDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			focusField(currentIdx + 1)
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			focusField(currentIdx - 1)
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			submit()
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			closeDialog()
			GlobalTipComponent.LayoutTemporary("Operation cancelled", 3, TipTypeWarning)
			return nil
		})
	}

	bindNavKeys(dialogName)
	bindNavKeys(maskName)
	for _, fieldName := range fieldNames {
		bindNavKeys(fieldName)
	}

	focusField(0)
	GlobalApp.Gui.SetViewOnTop(maskName)
	GlobalApp.Gui.SetViewOnTop(dialogName)
	for _, name := range fieldNames {
		GlobalApp.Gui.SetViewOnTop(name)
	}
	return nil
}

func requireNonEmpty(values map[string]string, key string) (string, error) {
	val := strings.TrimSpace(values[key])
	if val == "" {
		return "", fmt.Errorf("%s is required", strings.ToLower(key))
	}
	return val, nil
}

func parseRequiredInt(values map[string]string, key string) (int, error) {
	val, err := requireNonEmpty(values, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s must be integer", strings.ToLower(key))
	}
	return n, nil
}

func parseRequiredFloat(values map[string]string, key string) (float64, error) {
	val, err := requireNonEmpty(values, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be number", strings.ToLower(key))
	}
	return n, nil
}

func (c *LTRKeyInfoDetailComponent) canEditCurrentKey() bool {
	if strings.TrimSpace(GlobalKeyInfoComponent.keyName) == "" {
		GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeWarning)
		return false
	}
	return true
}

func (c *LTRKeyInfoDetailComponent) getPrimaryEditTemplate() string {
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return c.CopyString
	}
	if isCollectionKeyType(keyType) {
		return c.getTypeOperationTemplate("update")
	}
	return c.CopyString
}

func (c *LTRKeyInfoDetailComponent) applyPrimaryEdit(editorInput string) error {
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return err
	}
	if isCollectionKeyType(keyType) {
		return c.applyTypeOperation("update", editorInput)
	}
	return c.updateValueByEditorInput(editorInput)
}

func (c *LTRKeyInfoDetailComponent) getTypeOperationTemplate(operation string) string {
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return "{}"
	}

	switch keyType {
	case "string", "json":
		if operation != "update" {
			return "{}"
		}
		if keyType == "string" {
			return c.CopyString
		}
		if validator.IsJSON(c.CopyString) {
			pretty, _ := PrettyString(c.CopyString)
			return pretty
		}
		return c.CopyString
	case "list":
		switch operation {
		case "add":
			return "{\n  \"position\": \"tail\",\n  \"items\": [\n    \"new-item\"\n  ]\n}\n"
		case "update":
			return "{\n  \"index\": 0,\n  \"value\": \"updated-item\"\n}\n"
		default:
			return "{\n  \"index\": 0\n}\n"
		}
	case "hash":
		switch operation {
		case "add":
			return "[\n  {\n    \"field\": \"field1\",\n    \"value\": \"value1\"\n  }\n]\n"
		case "update":
			return "{\n  \"field\": \"field1\",\n  \"newField\": \"field1\",\n  \"value\": \"updated-value\"\n}\n"
		default:
			return "{\n  \"field\": \"field1\"\n}\n"
		}
	case "set":
		switch operation {
		case "add":
			return "[\n  \"member1\",\n  \"member2\"\n]\n"
		case "update":
			return "{\n  \"value\": \"member1\",\n  \"newValue\": \"member1-updated\"\n}\n"
		default:
			return "[\n  \"member1\"\n]\n"
		}
	case "zset":
		switch operation {
		case "add":
			return "[\n  {\n    \"value\": \"member1\",\n    \"score\": 1\n  }\n]\n"
		case "update":
			return "{\n  \"value\": \"member1\",\n  \"newValue\": \"member1\",\n  \"score\": 2\n}\n"
		default:
			return "{\n  \"value\": \"member1\"\n}\n"
		}
	case "stream":
		switch operation {
		case "add":
			return "{\n  \"id\": \"*\",\n  \"field\": \"field1\",\n  \"value\": \"value1\"\n}\n"
		default:
			return "{\n  \"id\": \"1234567890-0\"\n}\n"
		}
	default:
		return "{}"
	}
}

func (c *LTRKeyInfoDetailComponent) getCurrentSetKeyType() (string, error) {
	keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
		Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		DB:     GlobalDBComponent.SelectedDB,
		Key:    GlobalKeyInfoComponent.keyName,
	})
	if !keySummary.Success {
		return "", fmt.Errorf("failed to load key summary: %s", keySummary.Msg)
	}
	keySummaryData := keySummary.Data.(types.KeySummary)
	keyType := strings.ToLower(strings.TrimSpace(keySummaryData.Type))
	return normalizeSetKeyType(keyType), nil
}

func (c *LTRKeyInfoDetailComponent) applyTypeOperation(operation, editorInput string) error {
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return err
	}

	trimmed := strings.TrimSpace(editorInput)
	switch keyType {
	case "string", "json":
		if operation != "update" {
			return fmt.Errorf("%s operation is only for list/hash/set/zset keys", operation)
		}
		return c.updateValueByEditorInput(editorInput)
	case "list":
		return c.applyListOperation(operation, trimmed)
	case "hash":
		return c.applyHashOperation(operation, trimmed)
	case "set":
		return c.applySetOperation(operation, trimmed)
	case "zset":
		return c.applyZSetOperation(operation, trimmed)
	case "stream":
		return c.applyStreamOperation(operation, trimmed)
	default:
		return fmt.Errorf("key type %q does not support type operations", keyType)
	}
}

func (c *LTRKeyInfoDetailComponent) applyListOperation(operation, raw string) error {
	switch operation {
	case "add":
		action := 1
		items := []any{}

		var addPayload struct {
			Position string `json:"position"`
			Items    []any  `json:"items"`
		}
		if err := json.Unmarshal([]byte(raw), &addPayload); err == nil && len(addPayload.Items) > 0 {
			items = addPayload.Items
			if strings.EqualFold(strings.TrimSpace(addPayload.Position), "head") {
				action = 0
			}
		} else {
			if err := json.Unmarshal([]byte(raw), &items); err != nil {
				return fmt.Errorf("list add requires JSON array or {position,items}")
			}
		}

		res := services.Browser().AddListItem(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			action,
			items,
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "update":
		var payload struct {
			Index int `json:"index"`
			Value any `json:"value"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("list update requires {index,value}")
		}

		res := services.Browser().SetListItem(types.SetListParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
			Index:  payload.Index,
			Value:  fmt.Sprintf("%v", payload.Value),
			Format: types.FORMAT_RAW,
			Decode: types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "delete":
		var payload struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("list delete requires {index}")
		}

		res := services.Browser().SetListItem(types.SetListParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
			Index:  payload.Index,
			Value:  "",
			Format: types.FORMAT_RAW,
			Decode: types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	default:
		return fmt.Errorf("unsupported list operation: %s", operation)
	}
}

func (c *LTRKeyInfoDetailComponent) applyHashOperation(operation, raw string) error {
	switch operation {
	case "add":
		fieldItems, err := parseHashFieldItems(raw)
		if err != nil {
			return err
		}
		res := services.Browser().AddHashField(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			0,
			fieldItems,
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "update":
		var payload struct {
			Field    string `json:"field"`
			NewField string `json:"newField"`
			Value    any    `json:"value"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("hash update requires {field,newField,value}")
		}
		if strings.TrimSpace(payload.NewField) == "" {
			payload.NewField = payload.Field
		}
		res := services.Browser().SetHashValue(types.SetHashParam{
			Server:   GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:       GlobalDBComponent.SelectedDB,
			Key:      GlobalKeyInfoComponent.keyName,
			Field:    payload.Field,
			NewField: payload.NewField,
			Value:    fmt.Sprintf("%v", payload.Value),
			Format:   types.FORMAT_RAW,
			Decode:   types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "delete":
		fields, err := parseDeleteFields(raw)
		if err != nil {
			return err
		}
		for _, field := range fields {
			res := services.Browser().SetHashValue(types.SetHashParam{
				Server:   GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				DB:       GlobalDBComponent.SelectedDB,
				Key:      GlobalKeyInfoComponent.keyName,
				Field:    field,
				NewField: "",
				Value:    "",
				Format:   types.FORMAT_RAW,
				Decode:   types.DECODE_NONE,
			})
			if !res.Success {
				return fmt.Errorf("%s", res.Msg)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported hash operation: %s", operation)
	}
}

func (c *LTRKeyInfoDetailComponent) applySetOperation(operation, raw string) error {
	switch operation {
	case "add", "delete":
		members, err := parseSetMembers(raw)
		if err != nil {
			return err
		}
		res := services.Browser().SetSetItem(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			operation == "delete",
			members,
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "update":
		var payload struct {
			Value    any `json:"value"`
			NewValue any `json:"newValue"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("set update requires {value,newValue}")
		}
		res := services.Browser().UpdateSetItem(types.SetSetParam{
			Server:   GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:       GlobalDBComponent.SelectedDB,
			Key:      GlobalKeyInfoComponent.keyName,
			Value:    fmt.Sprintf("%v", payload.Value),
			NewValue: fmt.Sprintf("%v", payload.NewValue),
			Format:   types.FORMAT_RAW,
			Decode:   types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	default:
		return fmt.Errorf("unsupported set operation: %s", operation)
	}
}

func (c *LTRKeyInfoDetailComponent) applyZSetOperation(operation, raw string) error {
	switch operation {
	case "add":
		valueScore, err := parseZSetValueScoreMap(raw)
		if err != nil {
			return err
		}
		res := services.Browser().AddZSetValue(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			0,
			valueScore,
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "update":
		var payload struct {
			Value    any     `json:"value"`
			NewValue any     `json:"newValue"`
			Score    float64 `json:"score"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("zset update requires {value,newValue,score}")
		}
		res := services.Browser().UpdateZSetValue(types.SetZSetParam{
			Server:   GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:       GlobalDBComponent.SelectedDB,
			Key:      GlobalKeyInfoComponent.keyName,
			Value:    fmt.Sprintf("%v", payload.Value),
			NewValue: fmt.Sprintf("%v", payload.NewValue),
			Score:    payload.Score,
			Format:   types.FORMAT_RAW,
			Decode:   types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "delete":
		member, err := parseZSetDeleteMember(raw)
		if err != nil {
			return err
		}
		res := services.Browser().UpdateZSetValue(types.SetZSetParam{
			Server:   GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:       GlobalDBComponent.SelectedDB,
			Key:      GlobalKeyInfoComponent.keyName,
			Value:    member,
			NewValue: "",
			Score:    0,
			Format:   types.FORMAT_RAW,
			Decode:   types.DECODE_NONE,
		})
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	default:
		return fmt.Errorf("unsupported zset operation: %s", operation)
	}
}

func (c *LTRKeyInfoDetailComponent) applyStreamOperation(operation, raw string) error {
	switch operation {
	case "add":
		var payload struct {
			ID    string `json:"id"`
			Field string `json:"field"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("stream add requires {id,field,value}")
		}
		id := strings.TrimSpace(payload.ID)
		if id == "" {
			id = "*"
		}
		field := strings.TrimSpace(payload.Field)
		if field == "" {
			return fmt.Errorf("field is required")
		}
		value := strings.TrimSpace(payload.Value)
		if value == "" {
			return fmt.Errorf("value is required")
		}
		fieldItems := []any{field, value}
		res := services.Browser().AddStreamValue(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			id,
			fieldItems,
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	case "delete":
		var payload struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("stream delete requires {id}")
		}
		entryID := strings.TrimSpace(payload.ID)
		if entryID == "" {
			return fmt.Errorf("entry id is required")
		}
		res := services.Browser().RemoveStreamValues(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			[]string{entryID},
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		return nil
	default:
		return fmt.Errorf("stream does not support %s operation", operation)
	}
}

func parseHashFieldItems(raw string) ([]any, error) {
	var listPayload []map[string]any
	if err := json.Unmarshal([]byte(raw), &listPayload); err == nil {
		ret := make([]any, 0, len(listPayload)*2)
		for _, item := range listPayload {
			field, ok := item["field"]
			if !ok {
				field, ok = item["key"]
			}
			if !ok {
				return nil, fmt.Errorf("hash field item requires field/key")
			}
			value, ok := item["value"]
			if !ok {
				return nil, fmt.Errorf("hash field item requires value")
			}
			ret = append(ret, fmt.Sprintf("%v", field), fmt.Sprintf("%v", value))
		}
		if len(ret) == 0 {
			return nil, fmt.Errorf("hash add payload is empty")
		}
		return ret, nil
	}

	var objPayload map[string]any
	if err := json.Unmarshal([]byte(raw), &objPayload); err == nil {
		ret := make([]any, 0, len(objPayload)*2)
		for field, value := range objPayload {
			ret = append(ret, field, fmt.Sprintf("%v", value))
		}
		if len(ret) == 0 {
			return nil, fmt.Errorf("hash add payload is empty")
		}
		return ret, nil
	}

	return nil, fmt.Errorf("hash add requires object or array of {field,value}")
}

func parseDeleteFields(raw string) ([]string, error) {
	var payload struct {
		Field string `json:"field"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && strings.TrimSpace(payload.Field) != "" {
		return []string{payload.Field}, nil
	}

	var fields []string
	if err := json.Unmarshal([]byte(raw), &fields); err == nil && len(fields) > 0 {
		return fields, nil
	}

	return nil, fmt.Errorf("hash delete requires {field} or [field1,field2]")
}

func parseSetMembers(raw string) ([]any, error) {
	var arr []any
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		ret := make([]any, 0, len(arr))
		for _, v := range arr {
			switch t := v.(type) {
			case map[string]any:
				member, ok := t["value"]
				if !ok {
					return nil, fmt.Errorf("set member object requires value")
				}
				ret = append(ret, fmt.Sprintf("%v", member))
			default:
				ret = append(ret, fmt.Sprintf("%v", t))
			}
		}
		if len(ret) == 0 {
			return nil, fmt.Errorf("set member list is empty")
		}
		return ret, nil
	}
	return nil, fmt.Errorf("set operation requires JSON array")
}

func parseZSetValueScoreMap(raw string) (map[string]float64, error) {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, fmt.Errorf("zset add requires array of {value,score}")
	}
	ret := make(map[string]float64, len(arr))
	for _, item := range arr {
		member, ok := item["value"]
		if !ok {
			member, ok = item["member"]
		}
		if !ok {
			return nil, fmt.Errorf("zset add item requires value/member")
		}
		score, ok := item["score"]
		if !ok {
			return nil, fmt.Errorf("zset add item requires score")
		}
		scoreStr := scoreAnyToString(score)
		if scoreStr == "" {
			return nil, fmt.Errorf("zset score is invalid")
		}
		scoreFloat, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return nil, fmt.Errorf("zset score parse failed: %w", err)
		}
		ret[fmt.Sprintf("%v", member)] = scoreFloat
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("zset add payload is empty")
	}
	return ret, nil
}

func parseZSetDeleteMember(raw string) (string, error) {
	var payload struct {
		Value  any `json:"value"`
		Member any `json:"member"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", fmt.Errorf("zset delete requires {value} or {member}")
	}
	if payload.Value != nil {
		return fmt.Sprintf("%v", payload.Value), nil
	}
	if payload.Member != nil {
		return fmt.Sprintf("%v", payload.Member), nil
	}
	return "", fmt.Errorf("zset delete requires value/member")
}

func (c *LTRKeyInfoDetailComponent) buildDisplayValue(detail types.KeyDetail) string {
	keyType := strings.ToLower(detail.KeyType)
	c.currentKeyType = keyType
	c.structuredMode = false
	c.structuredRows = nil
	if keyType == "string" {
		theVal := fmt.Sprintln(detail.Value)
		if c.keyValueFormat == "JSON" && validator.IsJSON(theVal) {
			theVal, _ = PrettyString(theVal)
		} else if c.keyValueFormat == "Unicode JSON" && validator.IsJSON(theVal) {
			theVal, _ = UnicodeSequenceToString(theVal)
			theVal, _ = PrettyString(theVal)
		}
		return theVal
	}

	if keyType == "json" {
		jsonStr, ok := detail.Value.(string)
		if ok {
			if validator.IsJSON(jsonStr) {
				pretty, _ := PrettyString(jsonStr)
				return pretty + "\n"
			}
			return jsonStr + "\n"
		}
	}

	if keyType == "list" {
		c.structuredMode = true
		if rendered, ok := c.renderListDetail(detail.Value); ok {
			return rendered
		}
	}

	if keyType == "hash" {
		c.structuredMode = true
		if rendered, ok := c.renderHashDetail(detail.Value); ok {
			return rendered
		}
	}

	if keyType == "set" {
		c.structuredMode = true
		if rendered, ok := c.renderSetDetail(detail.Value); ok {
			return rendered
		}
	}

	if keyType == "zset" {
		c.structuredMode = true
		if rendered, ok := c.renderZSetDetail(detail.Value); ok {
			return rendered
		}
	}

	if keyType == "stream" {
		c.structuredMode = true
		if rendered, ok := c.renderStreamDetail(detail.Value); ok {
			return rendered
		}
	}

	jsonBytes, err := json.MarshalIndent(detail.Value, "", "  ")
	if err != nil {
		return fmt.Sprintln(detail.Value)
	}
	return string(jsonBytes) + "\n"
}

func (c *LTRKeyInfoDetailComponent) renderListDetail(value any) (string, bool) {
	items := []types.ListEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for _, item := range items {
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Index: item.Index,
			Value: displayEntryValue(item.Value, item.DisplayValue),
		})
	}
	c.applyListFilter()
	c.normalizeSelectedRow()
	rows := c.getActiveSelectionRows()

	var b strings.Builder
	b.WriteString("Type: list (Explorer Mode)\n")
	b.WriteString("Actions: ↑/↓ select  ←/→ jump  </> filter  <x> clear filter  <a>/<e>/<u>/<d> CRUD\n")
	if strings.TrimSpace(c.listFilter) != "" {
		b.WriteString("Filter: \"" + c.listFilter + "\" | matched " + strconv.Itoa(len(rows)) + "/" + strconv.Itoa(len(c.structuredRows)) + "\n")
	}
	b.WriteString("================================================================================\n")
	if len(c.structuredRows) == 0 {
		b.WriteString("(empty)\n")
		return b.String(), true
	}
	if len(rows) == 0 {
		b.WriteString("No rows match current filter. Press <x> to clear or </> to update filter.\n")
		return b.String(), true
	}

	start := c.selectedRow - 6
	if start < 0 {
		start = 0
	}
	end := start + 12
	if end > len(rows) {
		end = len(rows)
		start = end - 12
		if start < 0 {
			start = 0
		}
	}

	b.WriteString("List Rows\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-2s %-8s %-s\n", "", "INDEX", "VALUE(PREVIEW)"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	for i := start; i < end; i++ {
		row := rows[i]
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		preview := truncateByRuneCount(strings.ReplaceAll(row.Value, "\n", " <NL> "), 90)
		b.WriteString(fmt.Sprintf("%s %-8d %s\n", prefix, row.Index, preview))
	}

	selected := rows[c.selectedRow]
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("Selected Index: %d (row %d/%d)\n", selected.Index, c.selectedRow+1, len(rows)))
	b.WriteString("Selected Value\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	for _, line := range strings.Split(selected.Value, "\n") {
		b.WriteString(line + "\n")
	}
	return b.String(), true
}

func (c *LTRKeyInfoDetailComponent) renderHashDetail(value any) (string, bool) {
	items := []types.HashEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for _, item := range items {
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Field: item.Key,
			Value: displayEntryValue(item.Value, item.DisplayValue),
		})
	}
	c.normalizeSelectedRow()

	var b strings.Builder
	b.WriteString("Type: hash\n")
	b.WriteString("Actions: <a> Add Field  <u>/<e> Edit Field  <d> Delete Field\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-24s %-s\n", "FIELD", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(items) == 0 {
		b.WriteString("(empty)\n")
		return b.String(), true
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-24s %s\n", prefix, row.Field, row.Value))
	}
	return b.String(), true
}

func (c *LTRKeyInfoDetailComponent) renderSetDetail(value any) (string, bool) {
	items := []types.SetEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for i, item := range items {
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Index: i,
			Value: displayEntryValue(item.Value, item.DisplayValue),
		})
	}
	c.normalizeSelectedRow()

	var b strings.Builder
	b.WriteString("Type: set\n")
	b.WriteString("Actions: <a> Add Member  <u>/<e> Replace Member  <d> Delete Member\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-8s %-s\n", "ROW", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(items) == 0 {
		b.WriteString("(empty)\n")
		return b.String(), true
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-8d %s\n", prefix, row.Index, row.Value))
	}
	return b.String(), true
}

func (c *LTRKeyInfoDetailComponent) renderZSetDetail(value any) (string, bool) {
	items := []types.ZSetEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for _, item := range items {
		score := item.ScoreStr
		if strings.TrimSpace(score) == "" {
			score = strconv.FormatFloat(item.Score, 'f', -1, 64)
		}
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Score: score,
			Value: displayEntryValue(item.Value, item.DisplayValue),
		})
	}
	c.normalizeSelectedRow()

	var b strings.Builder
	b.WriteString("Type: zset\n")
	b.WriteString("Actions: <a> Add Member+Score  <u>/<e> Edit Member+Score  <d> Delete Member\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-10s %-s\n", "SCORE", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(items) == 0 {
		b.WriteString("(empty)\n")
		return b.String(), true
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-10s %s\n", prefix, row.Score, row.Value))
	}
	return b.String(), true
}

func (c *LTRKeyInfoDetailComponent) renderStreamDetail(value any) (string, bool) {
	items := []types.StreamEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for i, item := range items {
		displayVal := item.DisplayValue
		if strings.TrimSpace(displayVal) == "" {
			if vb, err := json.Marshal(item.Value); err == nil {
				displayVal = string(vb)
			} else {
				displayVal = fmt.Sprintf("%v", item.Value)
			}
		}
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Index: i,
			Field: item.ID,
			Value: displayVal,
		})
	}
	c.normalizeSelectedRow()

	var b strings.Builder
	b.WriteString("Type: stream\n")
	b.WriteString("Actions: <a> Add Entry  <d> Delete Entry(by ID)\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-26s %-s\n", "ENTRY ID", "VALUE"))
	b.WriteString("--------------------------------------------------------------------------------\n")
	if len(items) == 0 {
		b.WriteString("(empty)\n")
		return b.String(), true
	}
	for i, row := range c.structuredRows {
		prefix := " "
		if i == c.selectedRow {
			prefix = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-26s %s\n", prefix, row.Field, truncateByRuneCount(strings.ReplaceAll(row.Value, "\n", " <NL> "), 80)))
	}
	return b.String(), true
}

func decodeTypedEntries(value any, out any) bool {
	buf, err := json.Marshal(value)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(buf, out); err != nil {
		return false
	}
	return true
}

func displayEntryValue(raw any, display string) string {
	if strings.TrimSpace(display) != "" {
		return display
	}
	return fmt.Sprintf("%v", raw)
}

func (c *LTRKeyInfoDetailComponent) updateValueByEditorInput(editorInput string) error {
	keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
		Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		DB:     GlobalDBComponent.SelectedDB,
		Key:    GlobalKeyInfoComponent.keyName,
	})
	if !keySummary.Success {
		return fmt.Errorf("failed to load key summary: %s", keySummary.Msg)
	}

	keySummaryData := keySummary.Data.(types.KeySummary)
	keyType := strings.ToLower(strings.TrimSpace(keySummaryData.Type))
	setKeyType := normalizeSetKeyType(keyType)
	value, err := parseEditorValueByKeyType(setKeyType, editorInput)
	if err != nil {
		return err
	}

	if isCollectionKeyType(setKeyType) && collectionValueIsEmpty(value) {
		return fmt.Errorf("redis %s cannot be set to empty; delete the key if needed", setKeyType)
	}

	if isCollectionKeyType(setKeyType) || setKeyType == "json" {
		delRes := services.Browser().DeleteKey(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			GlobalKeyInfoComponent.keyName,
			false,
		)
		if !delRes.Success {
			return fmt.Errorf("failed to replace current value: %s", delRes.Msg)
		}
	}

	ttl := keySummaryData.TTL
	if ttl < 0 {
		ttl = -1
	}

	res := services.Browser().SetKeyValue(
		types.SetKeyParam{
			Server:  GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:      GlobalDBComponent.SelectedDB,
			Key:     GlobalKeyInfoComponent.keyName,
			KeyType: setKeyType,
			Value:   value,
			TTL:     ttl,
			Format:  types.FORMAT_RAW,
			Decode:  types.DECODE_NONE,
		},
	)
	if !res.Success {
		return fmt.Errorf("%s", res.Msg)
	}
	return nil
}

func normalizeSetKeyType(keyType string) string {
	if keyType == "rejson-rl" {
		return "json"
	}
	return keyType
}

func isCollectionKeyType(keyType string) bool {
	switch keyType {
	case "list", "hash", "set", "zset", "stream":
		return true
	default:
		return false
	}
}

func collectionValueIsEmpty(value any) bool {
	slice, ok := value.([]any)
	if !ok {
		return false
	}
	return len(slice) == 0
}

func parseEditorValueByKeyType(keyType, editorInput string) (any, error) {
	trimmedInput := strings.TrimSpace(editorInput)
	switch keyType {
	case "string":
		return editorInput, nil
	case "list":
		return parseListEditorInput(trimmedInput)
	case "hash":
		return parseHashEditorInput(trimmedInput)
	case "set":
		return parseSetEditorInput(trimmedInput)
	case "zset":
		return parseZSetEditorInput(trimmedInput)
	case "json":
		var obj any
		if err := json.Unmarshal([]byte(trimmedInput), &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("editing for key type %q is not supported", keyType)
	}
}

func parseListEditorInput(raw string) ([]any, error) {
	var listItems []types.ListEntryItem
	if err := json.Unmarshal([]byte(raw), &listItems); err == nil {
		res := make([]any, 0, len(listItems))
		for i := len(listItems) - 1; i >= 0; i-- {
			res = append(res, listItems[i].Value)
		}
		return res, nil
	}

	var vals []any
	if err := json.Unmarshal([]byte(raw), &vals); err != nil {
		return nil, fmt.Errorf("list value must be a JSON array")
	}

	res := make([]any, 0, len(vals))
	for i := len(vals) - 1; i >= 0; i-- {
		switch item := vals[i].(type) {
		case map[string]any:
			value, ok := item["value"]
			if !ok {
				return nil, fmt.Errorf("list item object must contain 'value'")
			}
			res = append(res, fmt.Sprintf("%v", value))
		default:
			res = append(res, fmt.Sprintf("%v", item))
		}
	}
	return res, nil
}

func parseHashEditorInput(raw string) ([]any, error) {
	var hashItems []types.HashEntryItem
	if err := json.Unmarshal([]byte(raw), &hashItems); err == nil {
		res := make([]any, 0, len(hashItems)*2)
		for _, item := range hashItems {
			res = append(res, item.Key, item.Value)
		}
		return res, nil
	}

	var kvMap map[string]any
	if err := json.Unmarshal([]byte(raw), &kvMap); err == nil {
		res := make([]any, 0, len(kvMap)*2)
		for field, value := range kvMap {
			res = append(res, field, fmt.Sprintf("%v", value))
		}
		return res, nil
	}

	var vals []map[string]any
	if err := json.Unmarshal([]byte(raw), &vals); err != nil {
		return nil, fmt.Errorf("hash value must be a JSON object or field/value array")
	}

	res := make([]any, 0, len(vals)*2)
	for _, item := range vals {
		fieldRaw, hasField := item["field"]
		if !hasField {
			fieldRaw, hasField = item["key"]
		}
		if !hasField {
			return nil, fmt.Errorf("hash item must contain 'field' or 'key'")
		}
		valueRaw, hasValue := item["value"]
		if !hasValue {
			return nil, fmt.Errorf("hash item must contain 'value'")
		}
		res = append(res, fmt.Sprintf("%v", fieldRaw), fmt.Sprintf("%v", valueRaw))
	}
	return res, nil
}

func parseSetEditorInput(raw string) ([]any, error) {
	var setItems []types.SetEntryItem
	if err := json.Unmarshal([]byte(raw), &setItems); err == nil {
		res := make([]any, 0, len(setItems))
		for _, item := range setItems {
			res = append(res, item.Value)
		}
		return res, nil
	}

	var vals []any
	if err := json.Unmarshal([]byte(raw), &vals); err != nil {
		return nil, fmt.Errorf("set value must be a JSON array")
	}

	res := make([]any, 0, len(vals))
	for _, item := range vals {
		switch v := item.(type) {
		case map[string]any:
			value, ok := v["value"]
			if !ok {
				return nil, fmt.Errorf("set item object must contain 'value'")
			}
			res = append(res, fmt.Sprintf("%v", value))
		default:
			res = append(res, fmt.Sprintf("%v", v))
		}
	}
	return res, nil
}

func parseZSetEditorInput(raw string) ([]any, error) {
	var zsetItems []types.ZSetEntryItem
	if err := json.Unmarshal([]byte(raw), &zsetItems); err == nil {
		res := make([]any, 0, len(zsetItems)*2)
		for _, item := range zsetItems {
			res = append(res, item.Value, formatZSetScore(item.Score, item.ScoreStr))
		}
		return res, nil
	}

	var vals []map[string]any
	if err := json.Unmarshal([]byte(raw), &vals); err != nil {
		return nil, fmt.Errorf("zset value must be a JSON array")
	}

	res := make([]any, 0, len(vals)*2)
	for _, item := range vals {
		memberRaw, hasMember := item["member"]
		if !hasMember {
			memberRaw, hasMember = item["value"]
		}
		if !hasMember {
			return nil, fmt.Errorf("zset item must contain 'member' or 'value'")
		}
		scoreRaw, hasScore := item["score"]
		if !hasScore {
			scoreRaw, hasScore = item["scoreStr"]
		}
		if !hasScore {
			return nil, fmt.Errorf("zset item must contain 'score' or 'scoreStr'")
		}

		scoreStr := scoreAnyToString(scoreRaw)
		if scoreStr == "" {
			return nil, fmt.Errorf("zset score is invalid")
		}
		res = append(res, fmt.Sprintf("%v", memberRaw), scoreStr)
	}
	return res, nil
}

func formatZSetScore(score float64, scoreStr string) string {
	if strings.TrimSpace(scoreStr) != "" {
		return scoreStr
	}
	if math.IsInf(score, 1) {
		return "+inf"
	}
	if math.IsInf(score, -1) {
		return "-inf"
	}
	return strconv.FormatFloat(score, 'f', -1, 64)
}

func scoreAnyToString(score any) string {
	switch val := score.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		if math.IsInf(val, 1) {
			return "+inf"
		}
		if math.IsInf(val, -1) {
			return "-inf"
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}
