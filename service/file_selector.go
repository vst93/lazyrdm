package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jroimartin/gocui"
)

type FileSelector struct {
	g           *gocui.Gui
	currentDir  string
	files       []os.FileInfo
	selectedIdx int
	name        string
	filtExt     string
}

func NewFileSelector(filtExt string) *FileSelector {
	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return &FileSelector{currentDir: dir, name: "file_selector", filtExt: filtExt}
}

func (fs *FileSelector) readDir() error {
	files, err := ioutil.ReadDir(fs.currentDir)
	if err != nil {
		return err
	}
	fs.files = files
	// 过滤文件类型
	if fs.filtExt != "" {
		filtered := make([]os.FileInfo, 0)
		for _, f := range files {
			if f.IsDir() || filepath.Ext(f.Name()) == "."+fs.filtExt {
				filtered = append(filtered, f)
			}
		}
		fs.files = filtered
	}
	fs.selectedIdx = 0 // 重置选中索引
	return nil
}

func (fs *FileSelector) Layout(g *gocui.Gui) error {
	fs.g = g
	maxX, maxY := g.Size()

	// 创建标题视图
	if v, err := g.SetView("title", 0, 0, maxX-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " File Selector"
		v.Wrap = true
		v.Frame = true
		fmt.Fprintf(v, "Current path: %s", fs.currentDir)
	}

	// 创建文件列表视图
	if v, err := g.SetView(fs.name, 0, 3, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// v.Highlight = true
		v.Title = " File List "
		if fs.filtExt != "" {
			v.Title += fmt.Sprintf(" Filter: %s", fs.filtExt)
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fs.renderFiles(v)
		if _, err := g.SetCurrentView(fs.name); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileSelector) renderFiles(v *gocui.View) {
	v.Clear()
	for i, f := range fs.files {
		prefix := " "
		theItem := f.Name()
		if i == fs.selectedIdx {
			prefix = ">"
			theItem = NewColorString(theItem, "white", "green", "blod")

		}

		// 添加目录标记
		suffix := ""
		if f.IsDir() {
			suffix = "/"
		}
		theItem = fmt.Sprintf("%s%s", theItem, suffix)
		fmt.Fprintf(v, "%s %s\n", prefix, theItem)
	}
}

func (fs *FileSelector) moveSelection(dy int) error {
	fs.selectedIdx += dy
	if fs.selectedIdx < 0 {
		fs.selectedIdx = 0
	} else if fs.selectedIdx >= len(fs.files) {
		fs.selectedIdx = len(fs.files) - 1
	}

	v, err := fs.g.View(fs.name)
	if err != nil {
		return err
	}
	fs.renderFiles(v)
	return nil
}

func (fs *FileSelector) selectItem() error {
	if len(fs.files) == 0 {
		return nil
	}

	selected := fs.files[fs.selectedIdx]
	if selected.IsDir() {
		// 进入子目录
		fs.currentDir = filepath.Join(fs.currentDir, selected.Name())
		if err := fs.readDir(); err != nil {
			return err
		}

		// 更新标题
		if v, err := fs.g.View("title"); err == nil {
			v.Clear()
			fmt.Fprintf(v, "Current path: %s", fs.currentDir)
		}

		// 更新文件列表
		if v, err := fs.g.View(fs.name); err == nil {
			fs.renderFiles(v)
		}
	} else {
		// 选择文件
		selectedPath := filepath.Join(fs.currentDir, selected.Name())
		v, err := fs.g.View("title")
		if err != nil {
			return err
		}
		v.Clear()
		fmt.Fprintf(v, "Selected file: %s", selectedPath)
	}
	return nil
}

func (fs *FileSelector) goBack() error {
	// 返回上级目录
	if fs.currentDir == filepath.Dir(fs.currentDir) {
		return nil // 已经是最顶层
	}

	fs.currentDir = filepath.Dir(fs.currentDir)
	if err := fs.readDir(); err != nil {
		return err
	}

	// 更新视图
	if v, err := fs.g.View("title"); err == nil {
		v.Clear()
		fmt.Fprintf(v, "当前目录: %s", fs.currentDir)
	}
	if v, err := fs.g.View(fs.name); err == nil {
		fs.renderFiles(v)
	}
	return nil
}

func (fs *FileSelector) Keybindings(g *gocui.Gui) error {
	// 文件列表导航
	GuiSetKeysbinding(GlobalApp.Gui, fs.name, []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return fs.moveSelection(-1)
	})

	GuiSetKeysbinding(GlobalApp.Gui, fs.name, []any{gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return fs.moveSelection(1)
	})

	// 选择项目
	GuiSetKeysbinding(GlobalApp.Gui, fs.name, []any{gocui.KeySpace, gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return fs.selectItem()
	})

	// 返回上级目录
	GuiSetKeysbinding(GlobalApp.Gui, fs.name, []any{gocui.KeyArrowLeft, gocui.KeyBackspace, gocui.KeyBackspace2}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return fs.goBack()
	})

	// 退出程序
	// if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone,
	// 	func(g *gocui.Gui, v *gocui.View) error {
	// 		return gocui.ErrQuit
	// 	}); err != nil {
	// 	return err
	// }

	return nil
}

func (fs *FileSelector) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Navigate", "↑/↓"},
		{"Back", "<-/<Backspace>"},
		{"Select", "->/<Enter>/<Space>"},
		{"Cancel", "<Esc>"},
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

func SelectFile(fileExt string) {
	fs := NewFileSelector(fileExt)
	if err := fs.readDir(); err != nil {
		log.Fatal(err)
	}
	fs.Layout(GlobalApp.Gui)
	fs.Keybindings(GlobalApp.Gui)
	GlobalTipComponent.Layout(fs.KeyMapTip())
}
