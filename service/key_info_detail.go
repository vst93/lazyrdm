package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/duke-git/lancet/v2/validator"

	"github.com/awesome-gocui/gocui"
)

type LTRKeyInfoDetailComponent struct {
	name            string
	title           string
	view            *gocui.View
	keyValueFormat  string
	viewOriginY     int // view origin y
	keyValueMaxY    int // value real total height
	CopyString      string
	lineView        *gocui.View
	selectedRow     int
	structuredRows  []keyDetailRow
	structuredMode  bool
	currentKeyType  string
	listFilter      string
	listFiltered    []keyDetailRow
	listFilterEdit  string
	scrollOffset    int // scroll offset for structured row window
	detailScrollY   int // scroll offset for detail pane content
	detailExpanded  bool // whether the detail pane is expanded (full value)
	filterDirty      bool // true when filter changed and needs re-applying
	cachedHintLine   string // cached hint line to avoid NewColorString on every scroll
	cachedCols       []columnDef
	cachedKeyType    string
	cachedViewW      int
	cachedColHeader  string // cached column header line
	cachedSep        string // cached separator line
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
	SelectOpts  []string // if non-nil, render as a select (←/→ toggle) instead of text input
}

type keyOpDialogSchema struct {
	Title       string
	Description string
	Fields      []keyOpDialogField
	BuildJSON   func(values map[string]string) (string, error)
	ReturnView  string // optional: view to focus after dialog closes (default: c.name)
}

var keyValueFormatList = []string{"Raw", "JSON", "Unicode JSON"}

func InitKeyInfoDetailComponent() {
	GlobalKeyInfoDetailComponent = &LTRKeyInfoDetailComponent{
		name:           "key_info_detail",
		title:          "Detail",
		keyValueFormat: "Raw",
	}
	GlobalKeyInfoDetailComponent.Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoDetailComponent.name)
	GlobalTipComponent.AppendList(GlobalKeyInfoDetailComponent.name, GlobalKeyInfoDetailComponent.KeyMapTip())
}

func (c *LTRKeyInfoDetailComponent) LayoutTitle() *LTRKeyInfoDetailComponent {
	if c.view == nil {
		return c
	}
	if CurrentViewName() == c.name {
		c.view.Title = " [" + c.title + "] "
		if c.lineView != nil {
			c.lineView.FrameColor = gocui.ColorGreen
		}
	} else {
		c.view.Title = " " + c.title + " "
		if c.lineView != nil {
			c.lineView.FrameColor = gocui.ColorDefault
		}
	}
	return c
}

func (c *LTRKeyInfoDetailComponent) Layout() *LTRKeyInfoDetailComponent {
	if GlobalDBComponent == nil || GlobalDBComponent.view == nil {
		return c
	}
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
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	c.view.TitleColor = gocui.ColorCyan
	c.view.FrameRunes = frameSolid
	c.keyValueMaxY = 0
	c.view.Wrap = true
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
		if c.structuredMode {
			c.view.Wrap = false
			c.keyValueMaxY = len(strings.Split(theVal, "\n"))
		} else {
			theValSlice := strings.Split(theVal, "\n")
			maxLine = len(theValSlice) - 1
			if maxLine < 0 {
				maxLine = 0
			}
			lineViewWidth = len(strconv.Itoa(maxLine))
			lineViewWidthStr = strconv.Itoa(lineViewWidth)
			c.view, _ = SetViewSafe(c.name, theX0+1+lineViewWidth, 3, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
			c.view.TitleColor = gocui.ColorCyan
			c.view.Wrap = true
			if CurrentViewName() == c.name {
				c.view.Title = " [" + c.title + "] "
			} else {
				c.view.Title = " " + c.title + " "
			}
			theViewX, _ := c.view.Size()
			for k, line := range theValSlice {
				if k == maxLine {
					break
				}
				lineLen := DisplayWidth(line)
				if lineLen > theViewX {
					theRealHeight := lineLen / theViewX
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
		}
	} else {
		theVal = fmt.Sprintln("")
	}
	if maxLine > 0 {
		subtitle := ""
		if len(c.structuredRows) > 0 {
			subtitle = c.buildStructuredSubtitle()
		} else {
			subtitle = " Lines: " + strconv.Itoa(maxLine) + " "
		}
		if c.currentKeyType != "" {
			subtitle = " " + c.currentKeyType + " " + subtitle
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
	c.view.Write([]byte(theVal))

	// show format select
	formatStr := " Format: " + c.keyValueFormat + " "
	formatSelectView, err := SetViewSafe("key_value_format", GlobalApp.maxX-len(formatStr)-2, GlobalApp.maxY-4, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		formatSelectView.Clear()
		formatSelectView.Write([]byte(formatStr))
	}
	formatSelectView.Frame = false
	formatSelectView.BgColor = themeIndicatorBg
	c.layoutListFilterView(theX0, len(formatStr))

	if c.structuredMode {
		c.viewOriginY = 0
	}
	c.view.SetOrigin(0, c.viewOriginY)

	// line view
	c.lineView, err = SetViewSafe("key_detail_line", theX0, 3, theX0+6, GlobalApp.maxY-2, 1)
	if err == nil || err != gocui.ErrUnknownView {
		// c.lineView.FrameColor = gocui.NewRGBColor(149, 165, 166)
		c.lineView.FgColor = themeLineNum
		c.lineView.Clear()
		if !c.structuredMode {
			c.lineView.Write([]byte(lineStr))
		}
		c.lineView.SetOrigin(0, 0)
	}
	c.lineView.FrameRunes = frameHalfTL
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
	if !c.isStructuredType() {
		GlobalApp.Gui.DeleteKeybindings(listFilterViewName)
		GlobalApp.Gui.DeleteView(listFilterViewName)
		return
	}

	// Don't recreate/reposition the view if it's currently focused (being edited)
	if CurrentViewName() == listFilterViewName {
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
	v.Frame = false
	v.BgColor = gocui.ColorBlack
	v.Editable = false
	v.Clear()
	v.Write([]byte(" Filter: " + c.listFilter + " "))
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
		if c.isStructuredType() {
			// Structured type: copy selected row's value
			row := c.getSelectedStructuredRow()
			if row == nil {
				GlobalTipComponent.LayoutTemporary("No row selected", 2, TipTypeWarning)
				return nil
			}
			clipboard.WriteAll(row.Value)
			GlobalTipComponent.LayoutTemporary("Copied value to clipboard", 3, TipTypeSuccess)
			return nil
		}
		// String type: copy full value
		theVal := c.CopyString
		if theVal == "" {
			GlobalTipComponent.LayoutTemporary("No value to copy", 2, TipTypeWarning)
			return nil
		}
		clipboard.WriteAll(theVal)
		GlobalTipComponent.LayoutTemporary("Copied value to clipboard", 3, TipTypeSuccess)
		return nil
	})
	// Copy field/key name (for hash: field name, for zset: member, etc.)
	// Note: gocui strips ModShift from character keys, converting Shift+C to 'C' with ModNone.
	// So we bind 'C' (uppercase) with ModNone instead of 'c' with ModShift.
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'C'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			row := c.getSelectedStructuredRow()
			if row == nil {
				GlobalTipComponent.LayoutTemporary("No row selected", 2, TipTypeWarning)
				return nil
			}
			// Copy field/key depending on type
			copyText := ""
			keyType, _ := c.getCurrentSetKeyType()
			switch keyType {
			case "hash":
				copyText = row.Field
			case "zset":
				copyText = row.Value // member value
			case "list":
				copyText = strconv.Itoa(row.Index)
			case "set":
				copyText = row.Value
			case "stream":
				copyText = row.Field
			}
			if copyText == "" {
				GlobalTipComponent.LayoutTemporary("Nothing to copy", 2, TipTypeWarning)
				return nil
			}
			clipboard.WriteAll(copyText)
			GlobalTipComponent.LayoutTemporary("Copied key to clipboard", 3, TipTypeSuccess)
			return nil
		}
		return nil
	})
	// scroll
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.moveDetailRowSelection(-1)
			return nil
		}
		c.scroll(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.moveDetailRowSelection(1)
			return nil
		}
		c.scroll(1)
		return nil
	})
	// scroll page
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowLeft, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.moveDetailRowSelection(-10)
			return nil
		}
		c.scroll(-GlobalApp.maxY + 9)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.moveDetailRowSelection(10)
			return nil
		}
		c.scroll(GlobalApp.maxY - 9)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'/'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.isStructuredType() {
			GlobalTipComponent.LayoutTemporary("Filter is available in structured detail mode", 2, TipTypeWarning)
			return nil
		}
		c.listFilterEdit = c.listFilter
		// Ensure the filter view exists without triggering a Redis API round-trip.
		// Previously this called c.Layout() which re-fetches GetKeyDetail from Redis.
		c.renderFromCache()
		// Create/position the filter view
		if GlobalDBComponent != nil && GlobalDBComponent.view != nil {
			dbW, _ := GlobalDBComponent.view.Size()
			formatStr := " Format: " + c.keyValueFormat + " "
			c.layoutListFilterView(dbW+2, len(formatStr))
		} else {
			c.layoutListFilterView(0, 0)
		}
		fv, ferr := GlobalApp.Gui.View(listFilterViewName)
		if ferr != nil {
			GlobalTipComponent.LayoutTemporary("Open filter input failed", 3, TipTypeError)
			return nil
		}
		// Switch to edit mode: buffer contains ONLY the raw input text (no prefix)
		fv.Editable = true
		fv.Editor = &EditorInput{BindValString: &c.listFilterEdit}
		fv.BgColor = themeIndicatorBg
		fv.Frame = true
		fv.FrameRunes = frameDashed
		fv.Title = " Filter (Enter=apply Esc=cancel) "
		fv.TitleColor = gocui.ColorCyan
		fv.Clear()
		fv.Write([]byte(c.listFilterEdit))
		_ = fv.SetCursor(len([]rune(c.listFilterEdit)), 0)
		if _, err := GlobalApp.Gui.SetCurrentView(listFilterViewName); err != nil {
			GlobalTipComponent.LayoutTemporary("Open filter input failed", 3, TipTypeError)
			return nil
		}
		GlobalApp.Gui.Cursor = true
		return nil
	})
	// 'x' clears the filter
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'x'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.isStructuredType() {
			return nil
		}
		c.listFilter = ""
		c.listFilterEdit = ""
		c.filterDirty = true
		c.applyListFilter()
		c.filterDirty = false
		c.selectedRow = 0
		c.scrollOffset = 0
		if fv, ferr := GlobalApp.Gui.View(listFilterViewName); ferr == nil {
			fv.Editable = false
			fv.Frame = false
			fv.BgColor = gocui.ColorBlack
			fv.Clear()
			fv.Write([]byte(" Filter: (none) "))
		}
		c.renderFromCache()
		GlobalTipComponent.LayoutTemporary("Filter cleared", 2, TipTypeSuccess)
		return nil
	})
	// toggle detail pane expansion
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.isStructuredType() {
			return nil
		}
		c.detailExpanded = !c.detailExpanded
		c.detailScrollY = 0
		c.renderFromCache()
		return nil
	})
	// Scroll detail pane content with Shift+↑/↓, Shift+wheel, or PgUp/PgDn
	// Shift+arrow/wheel works on most terminals (gocui preserves ModShift for
	// non-character keys like arrows and mouse events).
	// PgUp/PgDn as fallback for terminals where Shift+arrow sends escape sequences.
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp}, gocui.ModShift, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(-1)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown}, gocui.ModShift, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(1)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelUp}, gocui.ModShift, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(-1)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelDown}, gocui.ModShift, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(1)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyPgup}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(-3)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyPgdn}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.isStructuredType() {
			c.scrollDetailPane(3)
			return nil
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, listFilterViewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.listFilter = strings.TrimSpace(c.listFilterEdit)
		c.filterDirty = true
		c.applyListFilter()
		c.filterDirty = false
		c.selectedRow = 0
		c.scrollOffset = 0
		// Restore filter view to non-editable display state
		v.Editable = false
		v.Frame = false
		v.BgColor = gocui.ColorBlack
		v.Clear()
		v.Write([]byte(" Filter: " + c.listFilter + " "))
		GlobalApp.Gui.Cursor = false
		_, _ = GlobalApp.Gui.SetCurrentView(c.name)
		c.renderFromCache()
		if c.listFilter != "" {
			GlobalTipComponent.LayoutTemporary("Filter applied: "+c.listFilter, 2, TipTypeSuccess)
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, listFilterViewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.listFilterEdit = c.listFilter
		// Restore filter view to non-editable display state
		v.Editable = false
		v.Frame = false
		v.BgColor = gocui.ColorBlack
		v.Clear()
		v.Write([]byte(" Filter: " + c.listFilter + " "))
		GlobalApp.Gui.Cursor = false
		_, _ = GlobalApp.Gui.SetCurrentView(c.name)
		c.renderFromCache()
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
			GlobalTipComponent.LayoutTemporary("Use <e>/<a>/<d> for this key type", 4, TipTypeWarning)
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
	deleteDesc := "Delete Row"
	switch keyType {
	case "hash":
		addDesc = "Add Field"
		editDesc = "Edit Field"
		deleteDesc = "Delete Field"
	case "set":
		addDesc = "Add Member"
		editDesc = "Replace Member"
		deleteDesc = "Delete Member"
	case "zset":
		addDesc = "Add Member+Score"
		editDesc = "Edit Member+Score"
		deleteDesc = "Delete Member"
	case "list":
		addDesc = "Add Item"
		editDesc = "Edit By Index"
		deleteDesc = "Delete By Index"
	case "stream":
		addDesc = "Add Entry"
		editDesc = "N/A"
		deleteDesc = "Delete Entry"
	}
	if keyType == "string" || keyType == "json" {
		addDesc = "Add(<a>)"
		editDesc = "Primary Edit"
		deleteDesc = "Del(<d>)"
	}

	keyMap := []KeyMapStruct{
		{"Scroll/Select", "↑/↓/j/k"},
		{"Scroll Page/Jump", "←/->/h/l"},
		{"Scroll Detail", "Shift+↑/↓/wheel or PgUp/PgDn"},
		{"Expand Detail", "<Enter>"},
		{"Filter", "</>/<x>"},
		{"Switch Format", "<f>"},
		{editDesc, "<e>"},
		{addDesc + "/" + deleteDesc, "<a>/<d>"},
		{"Copy Value", "<c>"},
		{"Copy Key", "<C>"},
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
	c.detailScrollY = 0
	// Update format display view
	formatStr := " Format: " + c.keyValueFormat + " "
	if formatSelectView, err := GlobalApp.Gui.View("key_value_format"); err == nil {
		formatSelectView.Clear()
		formatSelectView.Write([]byte(formatStr))
	}
	if c.structuredMode && len(c.structuredRows) > 0 {
		c.renderFromCache()
		GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
		return
	}
	c.Layout()
	GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
}

func (c *LTRKeyInfoDetailComponent) normalizeSelectedRow() {
	rows := c.getActiveSelectionRows()
	if len(rows) == 0 {
		c.selectedRow = 0
		c.scrollOffset = 0
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
		return
	}
	c.selectedRow += step
	c.normalizeSelectedRow()
	c.detailScrollY = 0
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
	// Only re-apply filter if it changed since last render
	if c.filterDirty {
		c.applyListFilter()
		c.filterDirty = false
	}
	c.normalizeSelectedRow()
	theVal := c.renderStructuredRows()
	c.view.Subtitle = c.buildStructuredSubtitle()
	c.view.Clear()
	c.CopyString = theVal
	c.view.Write([]byte(theVal))
	c.view.SetOrigin(0, c.viewOriginY)
	// Update format indicator
	if formatSelectView, err := GlobalApp.Gui.View("key_value_format"); err == nil {
		formatSelectView.Clear()
		formatSelectView.Write([]byte(" Format: " + c.keyValueFormat + " "))
		GlobalApp.Gui.SetViewOnTop("key_value_format")
	}
}

// buildStructuredSubtitle generates the subtitle string for structured detail mode.
// Shared between Layout() and renderFromCache() to avoid duplication.
func (c *LTRKeyInfoDetailComponent) buildStructuredSubtitle() string {
	rows := c.getActiveSelectionRows()
	subtitle := " Row: "
	if len(rows) == 0 {
		subtitle += "0/0 "
	} else {
		subtitle += strconv.Itoa(c.selectedRow+1) + "/" + strconv.Itoa(len(rows)) + " "
	}
	if strings.TrimSpace(c.listFilter) != "" {
		subtitle += "Filtered:" + strconv.Itoa(len(rows)) + "/" + strconv.Itoa(len(c.structuredRows)) + " "
	}
	if c.detailExpanded {
		subtitle += "[Expanded] "
	}
	if c.currentKeyType != "" {
		subtitle = " " + c.currentKeyType + " " + subtitle
	}
	return subtitle
}

// renderStructuredRows is the unified renderer for all collection types
// (list/hash/set/zset/stream). It produces a two-section layout:
//  1. A columnar table with windowed scrolling showing all rows
//  2. A detail pane showing the full value of the selected row
//
// Column widths are dynamically computed from the available terminal width.
func (c *LTRKeyInfoDetailComponent) renderStructuredRows() string {
	rows := c.getActiveSelectionRows()
	viewW, viewH := 0, 0
	if c.view != nil {
		viewW, viewH = c.view.Size()
	}
	if viewW < 20 {
		viewW = 60
	}
	if viewH < 6 {
		viewH = 20
	}

	var b strings.Builder
	kt := c.currentKeyType

	// Single header line: key hints (type/count info is already in the view subtitle)
	// Cached string — this is called on every scroll event
	if c.cachedHintLine == "" {
		c.cachedHintLine = " " + NewColorString("</>filter  <Enter>expand  <a>add  <e>edit  <d>del  <x>clear  <f>fmt  <c>copy  <r>refresh", "cyan", "", "")
	}
	b.WriteString(truncateByDisplayWidth(c.cachedHintLine, viewW) + "\n")

	// separator
	if c.cachedSep == "" {
		c.cachedSep = strings.Repeat("─", viewW)
	}
	sep := c.cachedSep
	b.WriteString(sep + "\n")

	if len(c.structuredRows) == 0 {
		b.WriteString(NewColorString(" (empty — press <a> to add)", "yellow", "", "bold") + "\n")
		return b.String()
	}
	if len(rows) == 0 {
		b.WriteString(NewColorString(" No rows match filter.", "yellow", "", "bold") + "\n")
		b.WriteString(" Press <x> to clear or </> to update.\n")
		return b.String()
	}

	// Determine column layout based on key type (cached — only changes when keyType or viewW changes)
	if c.cachedCols == nil || c.cachedKeyType != kt || c.cachedViewW != viewW {
		c.cachedCols = c.getColumnLayout(kt, viewW)
		c.cachedKeyType = kt
		c.cachedViewW = viewW
		c.cachedColHeader = c.renderColumnHeader(c.cachedCols)
		c.cachedSep = strings.Repeat("─", viewW)
	}
	cols := c.cachedCols

	// column header — dim cyan (cached)
	b.WriteString(c.cachedColHeader)
	b.WriteString(c.cachedSep + "\n")

	// Layout calculation:
	// Fixed overhead = 1 (hint) + 1 (sep) + 1 (col header) + 1 (sep) + 1 (detail sep) = 5
	// Plus 1 line for scroll indicator when rows exceed visible window.
	// Table gets at most 55% of view, detail gets at least 30% of view.
	detailLines := c.getDetailPaneHeight(viewH, rows)
	// Ensure detail pane has a reasonable minimum (30% of view)
	minDetail := viewH * 3 / 10
	if minDetail < 4 {
		minDetail = 4
	}
	if detailLines < minDetail {
		detailLines = minDetail
	}
	// Also cap detail at 50% so table still has room
	maxDetail := viewH / 2
	if detailLines > maxDetail {
		detailLines = maxDetail
	}
	overhead := 5
	tableHeight := viewH - overhead - detailLines
	// Account for scroll indicator line
	if len(rows) > tableHeight {
		tableHeight--
	}
	if tableHeight < 3 {
		tableHeight = 3
	}

	c.adjustScrollOffset(len(rows), tableHeight)
	start := c.scrollOffset
	end := start + tableHeight
	if end > len(rows) {
		end = len(rows)
	}

	for i := start; i < end; i++ {
		row := rows[i]
		isSelected := i == c.selectedRow
		b.WriteString(c.renderRowLine(isSelected, i, row, cols))
	}

	// Scroll position indicator (only show if there are more rows than visible)
	if len(rows) > tableHeight {
		scrollInfo := ""
		if start == 0 {
			scrollInfo = fmt.Sprintf(" [%d-%d/%d ↓]", start+1, end, len(rows))
		} else if end == len(rows) {
			scrollInfo = fmt.Sprintf(" [%d-%d/%d ↑]", start+1, end, len(rows))
		} else {
			scrollInfo = fmt.Sprintf(" [%d-%d/%d ↕]", start+1, end, len(rows))
		}
		scrollInfo = NewColorString(scrollInfo, "yellow", "", "")
		padLen := viewW - DisplayWidth(scrollInfo)
		if padLen > 0 {
			scrollInfo += strings.Repeat(" ", padLen)
		}
		b.WriteString(truncateByDisplayWidth(scrollInfo, viewW) + "\n")
	}

	// detail pane separator
	b.WriteString(sep + "\n")
	if c.selectedRow < 0 || c.selectedRow >= len(rows) {
		c.selectedRow = 0
	}
	if len(rows) == 0 {
		return b.String()
	}
	selected := rows[c.selectedRow]
	b.WriteString(c.renderDetailPane(selected, kt, viewW, detailLines))
	return b.String()
}

// columnDef defines a column in the structured table
type columnDef struct {
	header   string
	width    int
	getValue func(row keyDetailRow, idx int) string
}

func (c *LTRKeyInfoDetailComponent) getColumnLayout(keyType string, viewW int) []columnDef {
	// prefix ">" takes 1 char + 1 space = 2 chars reserved
	availW := viewW - 2
	switch keyType {
	case "list":
		idxW := 8
		if availW < 20 {
			idxW = 4
		}
		valW := availW - idxW
		if valW < 10 {
			valW = 10
		}
		return []columnDef{
			{"INDEX", idxW, func(r keyDetailRow, i int) string { return strconv.Itoa(r.Index) }},
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	case "hash":
		fieldW := 24
		if availW < 40 {
			fieldW = availW / 3
			if fieldW < 8 {
				fieldW = 8
			}
		}
		valW := availW - fieldW
		if valW < 10 {
			valW = 10
		}
		return []columnDef{
			{"FIELD", fieldW, func(r keyDetailRow, i int) string { return r.Field }},
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	case "set":
		rowW := 6
		valW := availW - rowW
		if valW < 10 {
			valW = 10
		}
		return []columnDef{
			{"#", rowW, func(r keyDetailRow, i int) string { return strconv.Itoa(i) }},
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	case "zset":
		scoreW := 12
		valW := availW - scoreW
		if valW < 10 {
			valW = 10
		}
		return []columnDef{
			{"SCORE", scoreW, func(r keyDetailRow, i int) string { return r.Score }},
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	case "stream":
		idW := 26
		if availW < 40 {
			idW = availW / 3
			if idW < 10 {
				idW = 10
			}
		}
		valW := availW - idW
		if valW < 10 {
			valW = 10
		}
		return []columnDef{
			{"ENTRY ID", idW, func(r keyDetailRow, i int) string { return r.Field }},
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	default:
		valW := availW
		return []columnDef{
			{"VALUE", valW, func(r keyDetailRow, i int) string { return r.Value }},
		}
	}
}

func (c *LTRKeyInfoDetailComponent) renderColumnHeader(cols []columnDef) string {
	var b strings.Builder
	b.WriteString("  ")
	for i, col := range cols {
		if i > 0 {
			b.WriteString(" ")
		}
		header := padRightDisplayWidth(col.header, col.width)
		b.WriteString(NewColorString(header, "cyan", "", ""))
	}
	b.WriteString("\n")
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) renderRowLine(selected bool, idx int, row keyDetailRow, cols []columnDef) string {
	var b strings.Builder
	if selected {
		b.WriteString(NewColorString("▶", "green", "", "bold"))
	} else {
		b.WriteString(" ")
	}
	b.WriteString(" ")
	for i, col := range cols {
		if i > 0 {
			b.WriteString(" ")
		}
		val := col.getValue(row, idx)
		val = strings.ReplaceAll(val, "\n", " ⏎ ")
		val = strings.ReplaceAll(val, "\r", "")
		if selected {
			// Selected row: truncate + pad + color
			val = truncateByDisplayWidth(val, col.width)
			padded := padRightDisplayWidth(val, col.width)
			b.WriteString(NewColorString(padded, "black", "cyan", "bold"))
		} else {
			// Unselected row: just truncate, no padding needed (saves DisplayWidth calls)
			b.WriteString(truncateByDisplayWidth(val, col.width))
		}
	}
	b.WriteString("\n")
	return b.String()
}

func (c *LTRKeyInfoDetailComponent) getDetailPaneHeight(viewH int, rows []keyDetailRow) int {
	if len(rows) == 0 {
		return 0
	}
	if c.selectedRow < 0 || c.selectedRow >= len(rows) {
		c.selectedRow = 0
	}
	selected := rows[c.selectedRow]

	// Apply format to count lines (same logic as renderDetailPane)
	val := selected.Value
	if c.keyValueFormat == "JSON" && validator.IsJSON(val) {
		if pretty, err := PrettyString(val); err == nil {
			val = pretty
		}
	} else if c.keyValueFormat == "Unicode JSON" && validator.IsJSON(val) {
		if unicode, err := UnicodeSequenceToString(val); err == nil {
			val = unicode
		}
		if pretty, err := PrettyString(val); err == nil {
			val = pretty
		}
	}

	// Count display lines after wrapping
	viewW := 60
	if c.view != nil {
		viewW, _ = c.view.Size()
	}
	wrapW := viewW - 2
	if wrapW < 10 {
		wrapW = 10
	}
	lineCount := 0
	for _, line := range strings.Split(val, "\n") {
		line = strings.ReplaceAll(line, "\r", "")
		if line == "" {
			lineCount++
		} else {
			lineCount += len(wrapText(line, wrapW))
		}
	}

	if c.detailExpanded {
		if lineCount > viewH/2 {
			return viewH / 2
		}
		if lineCount < 3 {
			return 4
		}
		return lineCount + 1
	}

	// collapsed: show up to 4 lines + 1 label
	if lineCount > 4 {
		return 5
	}
	return lineCount + 1
}

func (c *LTRKeyInfoDetailComponent) scrollDetailPane(n int) {
	c.detailScrollY += n
	if c.detailScrollY < 0 {
		c.detailScrollY = 0
	}
	c.renderFromCache()
}

func (c *LTRKeyInfoDetailComponent) renderDetailPane(row keyDetailRow, keyType string, viewW int, maxLines int) string {
	var b strings.Builder
	// label line with colored type-specific prefix
	label := ""
	switch keyType {
	case "list":
		label = NewColorString("Index ", "cyan", "", "") + strconv.Itoa(row.Index)
	case "hash":
		label = NewColorString("Field: ", "cyan", "", "") + row.Field
	case "zset":
		label = NewColorString("Score: ", "cyan", "", "") + row.Score
	case "stream":
		label = NewColorString("ID: ", "cyan", "", "") + row.Field
	case "set":
		label = NewColorString("#", "cyan", "", "") + strconv.Itoa(row.Index)
	}
	posInfo := NewColorString("  ["+strconv.Itoa(c.selectedRow+1)+"/"+strconv.Itoa(len(c.getActiveSelectionRows()))+"]", "yellow", "", "")
	label += posInfo
	if c.detailExpanded {
		label += NewColorString("  (Enter to collapse)", "green", "", "")
	} else {
		label += NewColorString("  (Enter to expand)", "green", "", "")
	}
	b.WriteString(truncateByDisplayWidth(label, viewW) + "\n")

	// Apply format to value (JSON/Unicode JSON) if applicable
	val := row.Value
	if c.keyValueFormat == "JSON" && validator.IsJSON(val) {
		if pretty, err := PrettyString(val); err == nil {
			val = pretty
		}
	} else if c.keyValueFormat == "Unicode JSON" && validator.IsJSON(val) {
		if unicode, err := UnicodeSequenceToString(val); err == nil {
			val = unicode
		}
		if pretty, err := PrettyString(val); err == nil {
			val = pretty
		}
	}

	// Wrap long lines to viewW-2 columns, producing a flat list of display lines
	wrapW := viewW - 2
	if wrapW < 10 {
		wrapW = 10
	}
	var allLines []string
	for _, rawLine := range strings.Split(val, "\n") {
		rawLine = strings.ReplaceAll(rawLine, "\r", "")
		if rawLine == "" {
			allLines = append(allLines, "")
			continue
		}
		wrapped := wrapText(rawLine, wrapW)
		allLines = append(allLines, wrapped...)
	}

	// Apply scroll offset for detail pane
	start := c.detailScrollY
	if start > 0 && start >= len(allLines) {
		start = len(allLines) - 1
	}
	if start < 0 {
		start = 0
	}

	// Show up to maxLines-1 lines (1 line reserved for label)
	available := maxLines - 1
	if available < 1 {
		available = 1
	}
	end := start + available
	if end > len(allLines) {
		end = len(allLines)
	}

	for i := start; i < end; i++ {
		b.WriteString("  " + allLines[i] + "\n")
	}
	if end < len(allLines) {
		b.WriteString(NewColorString(fmt.Sprintf("  ... (%d more lines, Shift+↑/↓ or PgUp/PgDn to scroll)", len(allLines)-end), "yellow", "", "") + "\n")
	}
	if start > 0 {
		b.WriteString(NewColorString("  ... (PgUp to scroll up)", "yellow", "", "") + "\n")
	}
	return b.String()
}

// wrapText splits a string into lines no wider than maxWidth (by display width).
func wrapText(s string, maxWidth int) []string {
	if DisplayWidth(s) <= maxWidth {
		return []string{s}
	}
	var result []string
	runes := []rune(s)
	currentWidth := 0
	start := 0
	for i, r := range runes {
		rw := DisplayWidth(string(r))
		if currentWidth+rw > maxWidth && i > start {
			result = append(result, string(runes[start:i]))
			start = i
			currentWidth = 0
		}
		currentWidth += rw
	}
	if start < len(runes) {
		result = append(result, string(runes[start:]))
	}
	return result
}

func (c *LTRKeyInfoDetailComponent) adjustScrollOffset(totalRows, visibleHeight int) {
	if visibleHeight <= 0 {
		visibleHeight = 1
	}
	if c.scrollOffset < 0 {
		c.scrollOffset = 0
	}
	if c.scrollOffset > totalRows-visibleHeight {
		c.scrollOffset = totalRows - visibleHeight
		if c.scrollOffset < 0 {
			c.scrollOffset = 0
		}
	}
	// if selected row is above the window, scroll up
	if c.selectedRow < c.scrollOffset {
		c.scrollOffset = c.selectedRow
	}
	// if selected row is below the window, scroll down
	if c.selectedRow >= c.scrollOffset+visibleHeight {
		c.scrollOffset = c.selectedRow - visibleHeight + 1
	}
}

func (c *LTRKeyInfoDetailComponent) isStructuredType() bool {
	// Use the cached key type set by Layout() to avoid a Redis API call on every
	// scroll event. If the type hasn't been loaded yet, fall back to a one-shot fetch.
	if c.currentKeyType != "" {
		return isCollectionKeyType(c.currentKeyType)
	}
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return false
	}
	return isCollectionKeyType(keyType)
}

func (c *LTRKeyInfoDetailComponent) getActiveSelectionRows() []keyDetailRow {
	if strings.TrimSpace(c.listFilter) != "" {
		return c.listFiltered
	}
	return c.structuredRows
}

func (c *LTRKeyInfoDetailComponent) applyListFilter() {
	if !isCollectionKeyType(c.currentKeyType) {
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
		// search in value, field, score, and index
		if strings.Contains(strings.ToLower(row.Value), needle) ||
			strings.Contains(strings.ToLower(row.Field), needle) ||
			strings.Contains(strings.ToLower(row.Score), needle) ||
			strings.Contains(strconv.Itoa(row.Index), keyword) {
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

	// 'd' deletes the currently selected row with a confirmation prompt.
	// No input dialog needed — the row identity (index/field/value/id) is
	// taken directly from the selection.
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'d'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !c.canEditCurrentKey() {
			return nil
		}
		keyType, err := c.getCurrentSetKeyType()
		if err != nil || !isCollectionKeyType(keyType) {
			GlobalTipComponent.LayoutTemporary("Delete row is for collection types", 2, TipTypeWarning)
			return nil
		}
		row := c.getSelectedStructuredRow()
		if row == nil {
			GlobalTipComponent.LayoutTemporary("No row selected", 2, TipTypeWarning)
			return nil
		}
		// Build a descriptive confirmation message
		msg := c.buildDeleteConfirmMessage(keyType, row)
		NewPageComponentConfirm("Delete Confirmation", msg, func() {
			if err := c.deleteSelectedRow(); err != nil {
				GlobalTipComponent.LayoutTemporary("Delete failed: "+err.Error(), 4, TipTypeError)
				return
			}
			GlobalTipComponent.LayoutTemporary("Row deleted", 3, TipTypeSuccess)
			c.Layout()
		}, func() {
			GlobalTipComponent.LayoutTemporary("Delete cancelled", 2, TipTypeWarning)
		})
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

// openCreateKeyDialog opens a dialog for creating a new Redis key.
// Fields: Key Name + Type selector + Value (used by string/list/set,
// ignored for hash/zset/stream which use placeholders).
func (c *LTRKeyInfoDetailComponent) openCreateKeyDialog() {
	keyTypes := []string{"string", "list", "hash", "set", "zset", "stream"}
	schema := keyOpDialogSchema{
		Title:       "Create Key",
		Description:  "Enter key name, select type, and optional value",
		ReturnView:  GlobalKeyComponent.name,
		Fields: []keyOpDialogField{
			{Label: "Key Name", Placeholder: "new:key", Value: "new:key"},
			{Label: "Type", SelectOpts: keyTypes, Value: "string"},
			{Label: "Value", Placeholder: "(string/list/set, empty=placeholder)"},
		},
	}
	_ = c.showKeyOpDialog(schema, func(values map[string]string) error {
		keyName := strings.TrimSpace(values["Key Name"])
		if keyName == "" {
			return fmt.Errorf("key name is required")
		}
		keyType := strings.TrimSpace(values["Type"])
		if keyType == "" {
			keyType = "string"
		}
		// Check if key already exists
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    keyName,
		})
		if keySummary.Success {
			return fmt.Errorf("key already exists: %s", keyName)
		}
		valStr := values["Value"]
		value := buildCreateKeyValue(keyType, valStr)
		res := services.Browser().SetKeyValue(
			types.SetKeyParam{
				Server:  GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				DB:      GlobalDBComponent.SelectedDB,
				Key:     keyName,
				KeyType: keyType,
				Value:   value,
				TTL:     -1,
			},
		)
		if !res.Success {
			return fmt.Errorf("%s", res.Msg)
		}
		// Success: add key to list and show it
		GlobalKeyComponent.keys = append([]any{keyName}, GlobalKeyComponent.keys...)
		GlobalKeyComponent.Current = 0
		GlobalKeyInfoComponent.keyName = keyName
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.keyValueFormat = "Raw"
		GlobalKeyInfoDetailComponent.Layout()
		GlobalKeyComponent.Layout()
		GlobalTipComponent.LayoutTemporary("Created key: "+keyName+" ("+keyType+")", 3, TipTypeSuccess)
		return nil
	})
}

// buildCreateKeyValue builds the value for SetKeyValue based on key type.
// string/list/set use valStr if provided; hash/zset/stream use placeholders.
// Backend expects: string→string, collections→[]any flat array.
func buildCreateKeyValue(keyType, valStr string) any {
	switch keyType {
	case "string":
		return valStr
	case "list":
		return []any{valStr} // LPush single item
	case "set":
		if strings.TrimSpace(valStr) == "" {
			return []any{"member"}
		}
		return []any{valStr} // SAdd single member
	case "hash":
		// HSet needs [field, value] pairs
		return []any{"field", ""} // placeholder, edit via <a> afterwards
	case "zset":
		// ZAdd needs [member, score] pairs (score as string)
		return []any{"member", "0"} // placeholder
	case "stream":
		// XAdd needs [id, field, value, ...]
		return []any{"*", "field", ""} // placeholder
	default:
		return valStr
	}
}

func (c *LTRKeyInfoDetailComponent) buildKeyOpDialogSchema(keyType, operation, prefillValue string) (keyOpDialogSchema, error) {
	base := keyOpDialogSchema{}
	selected := c.getSelectedStructuredRow()
	// For "add" operations: use empty defaults (don't prefill from selected row)
	// For "update" operations: prefill from the selected row
	valueDefault := ""
	fieldDefault := ""
	indexDefault := "0"
	scoreDefault := "1"
	if operation == "update" && selected != nil {
		valueDefault = selected.Value
		fieldDefault = selected.Field
		indexDefault = strconv.Itoa(selected.Index)
		if selected.Score != "" {
			scoreDefault = selected.Score
		}
	}
	if strings.TrimSpace(prefillValue) != "" {
		valueDefault = prefillValue
	}

	switch keyType {
	case "list":
		switch operation {
		case "add":
			base.Title = "List Add"
			base.Description = "Add new item to list"
			base.Fields = []keyOpDialogField{{Label: "Position", SelectOpts: []string{"tail", "head"}, Value: "tail"}, {Label: "Value", Placeholder: "item", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				position := strings.TrimSpace(values["Position"])
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
			base.Fields = []keyOpDialogField{{Label: "Field", Placeholder: "field", Value: fieldDefault}, {Label: "Value", Placeholder: "new value", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				field, err := requireNonEmpty(values, "Field")
				if err != nil {
					return "", err
				}
				value, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				// old field from selection, newField = edited field
				oldField := fieldDefault
				obj := map[string]any{"field": oldField, "newField": field, "value": value}
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
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "new member", Value: valueDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				newValue, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				// old value from selection
				obj := map[string]any{"value": valueDefault, "newValue": newValue}
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
			base.Fields = []keyOpDialogField{{Label: "Value", Placeholder: "member", Value: valueDefault}, {Label: "Score", Placeholder: "1", Value: scoreDefault}}
			base.BuildJSON = func(values map[string]string) (string, error) {
				newValue, err := requireNonEmpty(values, "Value")
				if err != nil {
					return "", err
				}
				score, err := parseRequiredFloat(values, "Score")
				if err != nil {
					return "", err
				}
				// old value from selection
				obj := map[string]any{"value": valueDefault, "newValue": newValue, "score": score}
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
	// Each field is a bordered view: top border + 1 content row + bottom border = 3 rows.
	// Layout: title(1) + description(1) + hints(1) + fields(N*3) + bottom spacer(1) + bottom border(1) = N*3 + 5
	height := len(schema.Fields)*3 + 5
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
	dlg.TitleColor = gocui.ColorWhite
	dlg.FrameColor = themeFrameDialog
	dlg.FrameRunes = frameDouble
	dlg.Clear()
	dlg.Wrap = true
	dlg.Write([]byte(schema.Description + "\n"))
	dlg.Write([]byte("Tab/↑/↓ switch | ←/→ toggle | Enter submit | Esc cancel\n"))

	fieldNames := make([]string, 0, len(schema.Fields))
	fieldLabelToViewName := make(map[string]string, len(schema.Fields))
	// fieldSelectIdx[i] = current option index for select fields, -1 for text fields
	fieldSelectIdx := make([]int, len(schema.Fields))
	fieldSelectOpts := make([][]string, len(schema.Fields))

	// renderSelectField renders a select field view with the current option highlighted
	renderSelectField := func(fv *gocui.View, opts []string, curIdx int) {
		fv.Clear()
		for j, opt := range opts {
			if j == curIdx {
				fv.Write([]byte(NewColorString(" ["+opt+"] ", "black", "green", "bold")))
			} else {
				fv.Write([]byte(" " + opt + " "))
			}
			if j < len(opts)-1 {
				fv.Write([]byte("/"))
			}
		}
	}

	for i, field := range schema.Fields {
		fieldViewName := fieldPrefix + strconv.Itoa(i)
		// Each field starts 3 rows apart: top border, content, bottom border
		fy0 := y0 + 3 + i*3
		fy1 := fy0 + 2 // 3 rows total (border + content + border)
		fv, ferr := SetViewSafe(fieldViewName, x0+2, fy0, x1-2, fy1, 0)
		if ferr != nil && ferr != gocui.ErrUnknownView {
			return ferr
		}
		fv.Title = " " + field.Label + " "
		fv.TitleColor = gocui.ColorCyan
		fv.Clear()

		if len(field.SelectOpts) > 0 {
			// Select field: non-editable, show all options with current highlighted
			fieldSelectOpts[i] = field.SelectOpts
			fieldSelectIdx[i] = 0
			val := strings.TrimSpace(field.Value)
			for j, opt := range field.SelectOpts {
				if opt == val {
					fieldSelectIdx[i] = j
					break
				}
			}
			fv.Editable = false
			renderSelectField(fv, field.SelectOpts, fieldSelectIdx[i])
		} else {
			// Text input field
			val := strings.TrimSpace(field.Value)
			if val != "" {
				fv.Write([]byte(val))
			}
			bound := val
			fv.Editable = true
			fv.Editor = &EditorInput{BindValString: &bound}
		}
		fieldNames = append(fieldNames, fieldViewName)
		fieldLabelToViewName[field.Label] = fieldViewName
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
				view.BgColor = themeSelBg
				GlobalApp.Gui.SetCurrentView(name)
				if len(fieldSelectOpts[i]) > 0 {
					// Select field: no cursor needed
					GlobalApp.Gui.Cursor = false
				} else {
					// Text field: show cursor at end of text
					GlobalApp.Gui.Cursor = true
					buf := view.Buffer()
					view.SetCursor(len([]rune(strings.TrimRight(buf, "\n"))), 0)
				}
			} else {
				view.BgColor = gocui.ColorBlack
			}
		}
		GlobalTipComponent.LayComponentTips()
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
		delete(GlobalTipComponent.list, dialogName)
		GlobalApp.Gui.Cursor = false
		returnTo := schema.ReturnView
		if returnTo == "" {
			returnTo = c.name
		}
		GlobalApp.Gui.SetCurrentView(returnTo)
		GlobalTipComponent.LayComponentTips()
		GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
	}

	submit := func() {
		values := make(map[string]string, len(fieldLabelToViewName))
		for label, name := range fieldLabelToViewName {
			// Find the field index to check if it's a select field
			fieldIdx := -1
			for fi, fn := range fieldNames {
				if fn == name {
					fieldIdx = fi
					break
				}
			}
			if fieldIdx >= 0 && len(fieldSelectOpts[fieldIdx]) > 0 {
				// Select field: read from the tracked index
				opts := fieldSelectOpts[fieldIdx]
				idx := fieldSelectIdx[fieldIdx]
				if idx >= 0 && idx < len(opts) {
					values[label] = opts[idx]
				} else {
					values[label] = ""
				}
			} else if v, err := GlobalApp.Gui.View(name); err == nil {
				values[label] = strings.TrimSpace(v.Buffer())
			} else {
				values[label] = ""
			}
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
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyTab, gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			focusField(currentIdx + 1)
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			focusField(currentIdx - 1)
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			// Skip submit if we're in the middle of a paste (pasted \n -> KeyEnter)
			if isPasting() {
				return nil
			}
			submit()
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			closeDialog()
			GlobalTipComponent.LayoutTemporary("Operation cancelled", 3, TipTypeWarning)
			return nil
		})
	}

	// Field-specific keybindings
	for fieldI, fieldName := range fieldNames {
		fn := fieldName
		fi := fieldI
		// Select fields: ←/→ to toggle option
		if len(fieldSelectOpts[fi]) > 0 {
			GuiSetKeysbinding(GlobalApp.Gui, fn, []any{gocui.KeyArrowRight, gocui.KeyArrowLeft, 'h', 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
				opts := fieldSelectOpts[fi]
				if len(opts) == 0 {
					return nil
				}
				fieldSelectIdx[fi] = (fieldSelectIdx[fi] + 1) % len(opts)
				renderSelectField(v, opts, fieldSelectIdx[fi])
				return nil
			})
			continue // skip text-field bindings for select fields
		}
		// Text fields: Ctrl+U clear, Ctrl+Y copy
		GuiSetKeysbinding(GlobalApp.Gui, fn, []any{gocui.KeyCtrlU}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			v.Clear()
			v.SetCursor(0, 0)
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, fn, []any{gocui.KeyCtrlY}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			text := strings.TrimRight(v.Buffer(), "\n")
			if text != "" {
				clipboard.WriteAll(text)
				GlobalTipComponent.LayoutTemporary("Copied to clipboard", 2, TipTypeSuccess)
			}
			return nil
		})
	}

	bindNavKeys(dialogName)
	bindNavKeys(maskName)
	for _, fieldName := range fieldNames {
		bindNavKeys(fieldName)
	}

	// Register tip for dialog views
	dialogTip := "Confirm: <Enter> | Cancel: <Esc> | Switch: <Tab> | Toggle: <←/→> | Paste: <Ctrl+V> | Clear: <Ctrl+U> | Copy: <Ctrl+Y>"
	GlobalTipComponent.list[dialogName] = dialogTip

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

// buildDeleteConfirmMessage creates a descriptive confirmation message
// for deleting the selected row, including the row's identity and a
// truncated preview of its value.
func (c *LTRKeyInfoDetailComponent) buildDeleteConfirmMessage(keyType string, row *keyDetailRow) string {
	preview := truncateByDisplayWidth(strings.ReplaceAll(row.Value, "\n", " ⏎ "), 50)
	switch keyType {
	case "list":
		return fmt.Sprintf("Delete list item at index %d?\nValue: %s", row.Index, preview)
	case "hash":
		return fmt.Sprintf("Delete hash field %q?\nValue: %s", row.Field, preview)
	case "set":
		return fmt.Sprintf("Delete set member?\nValue: %s", preview)
	case "zset":
		return fmt.Sprintf("Delete zset member (score: %s)?\nValue: %s", row.Score, preview)
	case "stream":
		return fmt.Sprintf("Delete stream entry %q?\nValue: %s", row.Field, preview)
	default:
		return "Delete selected row?\nValue: " + preview
	}
}

// deleteSelectedRow deletes the currently selected row based on key type.
// It builds the appropriate delete payload from the selected row's identity
// and delegates to the existing applyXxxOperation("delete", ...) methods.
func (c *LTRKeyInfoDetailComponent) deleteSelectedRow() error {
	if !c.canEditCurrentKey() {
		return fmt.Errorf("no key selected")
	}
	keyType, err := c.getCurrentSetKeyType()
	if err != nil {
		return err
	}
	if !isCollectionKeyType(keyType) {
		return fmt.Errorf("delete row is only for collection types")
	}
	row := c.getSelectedStructuredRow()
	if row == nil {
		return fmt.Errorf("no row selected")
	}

	var payload string
	switch keyType {
	case "list":
		obj := map[string]any{"index": row.Index}
		buf, _ := json.Marshal(obj)
		payload = string(buf)
	case "hash":
		obj := map[string]any{"field": row.Field}
		buf, _ := json.Marshal(obj)
		payload = string(buf)
	case "set":
		buf, _ := json.Marshal([]string{row.Value})
		payload = string(buf)
	case "zset":
		obj := map[string]any{"value": row.Value}
		buf, _ := json.Marshal(obj)
		payload = string(buf)
	case "stream":
		obj := map[string]any{"id": row.Field}
		buf, _ := json.Marshal(obj)
		payload = string(buf)
	default:
		return fmt.Errorf("unsupported type for delete: %s", keyType)
	}

	return c.applyTypeOperation("delete", payload)
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
	c.scrollOffset = 0
	c.detailExpanded = false
	c.filterDirty = true
	c.cachedHintLine = ""
	c.cachedCols = nil
	c.cachedColHeader = ""
	c.cachedSep = ""
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
			if c.keyValueFormat == "Raw" {
				return jsonStr + "\n"
			}
			if c.keyValueFormat == "Unicode JSON" && validator.IsJSON(jsonStr) {
				jsonStr, _ = UnicodeSequenceToString(jsonStr)
			}
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
	return c.renderStructuredRows(), true
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
	c.applyListFilter()
	c.normalizeSelectedRow()
	return c.renderStructuredRows(), true
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
	c.applyListFilter()
	c.normalizeSelectedRow()
	return c.renderStructuredRows(), true
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
	c.applyListFilter()
	c.normalizeSelectedRow()
	return c.renderStructuredRows(), true
}

func (c *LTRKeyInfoDetailComponent) renderStreamDetail(value any) (string, bool) {
	items := []types.StreamEntryItem{}
	if !decodeTypedEntries(value, &items) {
		return "", false
	}
	c.structuredRows = make([]keyDetailRow, 0, len(items))
	for i, item := range items {
		// Stream values are map[string]any — flatten to "field: value" pairs
		displayVal := item.DisplayValue
		if strings.TrimSpace(displayVal) == "" {
			displayVal = formatStreamValue(item.Value)
		}
		c.structuredRows = append(c.structuredRows, keyDetailRow{
			Index: i,
			Field: item.ID,
			Value: displayVal,
		})
	}
	c.applyListFilter()
	c.normalizeSelectedRow()
	return c.renderStructuredRows(), true
}

// formatStreamValue converts a stream entry's value (map[string]any) into a
// readable "field: value" multi-line string. This is much more useful than
// the raw JSON blob that was previously shown.
func formatStreamValue(val map[string]any) string {
	if len(val) == 0 {
		return "(empty)"
	}
	var b strings.Builder
	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	// Sort keys for stable display
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(k + ": " + fmt.Sprintf("%v", val[k]))
	}
	return b.String()
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
