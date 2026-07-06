package service

import "github.com/awesome-gocui/gocui"

// ── 主题配色 ──────────────────────────────────────────────
// 设计原则：
//   - OutputNormal 模式仅支持 8 色基本色，不使用 RGB
//   - 选中高亮用 Cyan（柔和、终端原生），不用 Blue（刺眼）
//   - 弹窗透明背景 + Cyan 边框，与终端融为一体
//   - 语义色：Green=成功/活跃, Yellow=警告, Red=错误, Cyan=提示/选中
// ──────────────────────────────────────────────────────────

const (
	// 边框
	themeFrameActive   = gocui.ColorGreen
	themeFrameInactive = gocui.ColorDefault
	themeFrameDialog   = gocui.ColorCyan

	// 文字
	themeTextBright = gocui.ColorWhite | gocui.AttrBold
	themeTextNormal = gocui.ColorDefault
	themeTextDim    = gocui.ColorCyan
	themeTextHint   = gocui.ColorCyan

	// 选中高亮（列表项）
	themeSelFg = gocui.ColorBlack
	themeSelBg = gocui.ColorCyan

	// 弹窗
	themeDialogFg = gocui.ColorWhite
	themeDialogBg = gocui.ColorDefault

	// 输入框
	themeInputFg = gocui.ColorBlack | gocui.AttrBold
	themeInputBg = gocui.ColorWhite

	// 语义
	themeSuccess = gocui.ColorGreen | gocui.AttrBold
	themeWarning = gocui.ColorYellow | gocui.AttrBold
	themeError   = gocui.ColorRed | gocui.AttrBold

	// 行号
	themeLineNum = gocui.ColorCyan

	// 搜索/格式指示器
	themeIndicatorBg = gocui.ColorCyan
	themeIndicatorFg = gocui.ColorBlack
)

// NewColorString 选中色统一用 cyan（替代 blue）
// 为了向后兼容，保留 "blue" 映射但实际输出 cyan
func themeSelColorString(text string) string {
	return NewColorString(text, "black", "cyan", "bold")
}
