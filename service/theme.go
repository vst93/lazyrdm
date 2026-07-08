package service

import "github.com/awesome-gocui/gocui"

// ── 主题配色 ──────────────────────────────────────────────
// 设计原则：
//   - OutputNormal 模式仅支持 8 色基本色，不使用 RGB
//   - 选中高亮用 Cyan（柔和、终端原生），不用 Blue（刺眼）
//   - 弹窗透明背景 + Cyan 边框，与终端融为一体
//   - 语义色：Green=成功/活跃, Yellow=警告, Red=错误, Cyan=提示/选中
// ── 边框样式 ──────────────────────────────────────────────
//   - Solid: 单实线框（┌─┐│└─┘）  用于主内容区
//   - Double: 双实线框（╔═╗║╚═╝）  用于弹窗区
//   - Dashed: 虚线框（┄┆┌┐└┘）   用于次要/临时区
// ──────────────────────────────────────────────────────────

const (
	// 边框颜色
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

// FrameRunes 样式 — gocui FrameRunes 格式: {顶边, 垂直边, 左上角, 底边, 左下角, 右下角}
// 主区域用单线，弹窗用双线，次要区域用虚线
var (
	frameSolid  = []rune{'─', '│', '┌', '┐', '└', '┘'} // 单线完整框 — 主内容区
	frameDouble = []rune{'═', '║', '╔', '╗', '╚', '╝'} // 双线完整框 — 弹窗
	frameDashed = []rune{'┄', '┆', '┌', '┐', '└', '┘'} // 虚线完整框 — 次要/临时
	frameHalfTR = []rune{'─', '│', '─', '┐', '─', '┘'} // 右上半框 — TTL 等右上角
	frameHalfTL = []rune{'─', '│', '┌', '─', '└', '─'} // 左上半框 — 行号等左上角
)

// NewColorString 选中色统一用 cyan（替代 blue）
// 为了向后兼容，保留 "blue" 映射但实际输出 cyan
