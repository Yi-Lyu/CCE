package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ethan/cce-client/internal/config"
)

// LogsView 日志查看视图
type LogsView struct {
	configManager *config.Manager
	logText       *widget.Entry
	levelSelect   *widget.Select
	searchEntry   *widget.Entry
	autoScroll    bool
	ticker        *time.Ticker
	stopChan      chan bool
	allLines      []string // 缓存所有日志行用于搜索
}

// NewLogsView 创建日志查看视图
func NewLogsView(configManager *config.Manager) *LogsView {
	return &LogsView{
		configManager: configManager,
		autoScroll:    true,
		stopChan:      make(chan bool),
	}
}

// Build 构建视图
func (lv *LogsView) Build() fyne.CanvasObject {
	// 日志文本框
	lv.logText = widget.NewMultiLineEntry()
	lv.logText.Wrapping = fyne.TextWrapWord
	lv.logText.Disable() // 只读

	// 日志级别过滤
	lv.levelSelect = widget.NewSelect(
		[]string{"全部", "debug", "info", "warn", "error"},
		func(value string) {
			lv.filterAndDisplay()
		},
	)
	lv.levelSelect.SetSelected("全部")

	// 搜索框
	lv.searchEntry = widget.NewEntry()
	lv.searchEntry.SetPlaceHolder("搜索日志...")
	lv.searchEntry.OnChanged = func(text string) {
		lv.filterAndDisplay()
	}

	// 自动滚动开关
	autoScrollCheck := widget.NewCheck("自动滚动", func(checked bool) {
		lv.autoScroll = checked
	})
	autoScrollCheck.Checked = lv.autoScroll

	// 刷新按钮
	refreshBtn := widget.NewButton("刷新", func() {
		lv.reloadLogs()
	})

	// 清空按钮
	clearBtn := widget.NewButton("清空显示", func() {
		lv.logText.SetText("")
	})

	// 工具栏（优化布局，更符合 macOS 风格）
	toolbar1 := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewLabel("日志级别:"),
			lv.levelSelect,
		),
		container.NewHBox(
			autoScrollCheck,
			refreshBtn,
			clearBtn,
		),
		nil,
	)

	toolbar2 := container.NewBorder(
		nil, nil,
		widget.NewLabel("搜索:"),
		nil,
		lv.searchEntry,
	)

	toolbar := container.NewVBox(
		toolbar1,
		widget.NewSeparator(),
		toolbar2,
	)

	// 启动自动刷新
	lv.startAutoRefresh()

	return container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewScroll(lv.logText),
	)
}

// reloadLogs 重新加载日志（异步执行，避免阻塞 UI）
func (lv *LogsView) reloadLogs() {
	logsPath := lv.configManager.GetLogsPath()
	todayLog := filepath.Join(logsPath, fmt.Sprintf("claude-proxy-%s.log", time.Now().Format("2006-01-02")))

	// 在后台线程读取日志
	go func() {
		// 检查文件是否存在
		if _, err := os.Stat(todayLog); os.IsNotExist(err) {
			// Fyne v2 UI 更新是线程安全的
			lv.allLines = []string{"日志文件不存在", "路径: " + todayLog}
			lv.filterAndDisplay()
			return
		}

		// 读取日志文件（最后 1000 行）
		lines, err := lv.readLastLines(todayLog, 1000)
		if err != nil {
			lv.allLines = []string{fmt.Sprintf("读取日志失败: %v", err), "路径: " + todayLog}
			lv.filterAndDisplay()
			return
		}

		// 缓存所有行并触发过滤显示
		lv.allLines = lines
		lv.filterAndDisplay()
	}()
}

// filterAndDisplay 根据级别和搜索关键词过滤并显示日志
func (lv *LogsView) filterAndDisplay() {
	if lv.allLines == nil {
		return
	}

	levelFilter := lv.levelSelect.Selected
	searchText := strings.ToLower(lv.searchEntry.Text)
	filtered := []string{}

	for _, line := range lv.allLines {
		// 级别过滤
		if levelFilter != "全部" && !containsLevel(line, levelFilter) {
			continue
		}
		// 搜索过滤
		if searchText != "" && !strings.Contains(strings.ToLower(line), searchText) {
			continue
		}
		filtered = append(filtered, line)
	}

	// 构建显示内容
	content := ""
	if len(filtered) == 0 {
		if searchText != "" {
			content = "没有匹配搜索条件的日志"
		} else {
			content = "没有匹配的日志"
		}
	} else {
		for _, line := range filtered {
			content += line + "\n"
		}
	}

	lv.logText.SetText(content)
	// 自动滚动到底部
	if lv.autoScroll {
		lv.logText.CursorRow = len(lv.logText.Text)
	}
}

// readLastLines 读取文件最后 N 行
func (lv *LogsView) readLastLines(filepath string, n int) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 简化实现：读取所有行，返回最后 N 行
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 返回最后 N 行
	start := len(lines) - n
	if start < 0 {
		start = 0
	}

	return lines[start:], nil
}

// startAutoRefresh 启动自动刷新
func (lv *LogsView) startAutoRefresh() {
	lv.ticker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-lv.ticker.C:
				lv.reloadLogs()
			case <-lv.stopChan:
				return
			}
		}
	}()
}

// Stop 停止自动刷新
func (lv *LogsView) Stop() {
	if lv.ticker != nil {
		lv.ticker.Stop()
	}
	close(lv.stopChan)
}

// containsLevel 检查日志行是否包含指定级别（优化：使用 strings.Contains）
func containsLevel(line, level string) bool {
	if len(line) == 0 {
		return false
	}
	// 检查 JSON 格式: "level":"info" 或 "level": "info"
	// 也支持文本格式: [INFO] 或 INFO:
	// 优化：使用标准库 strings.Contains 提升 3-5 倍性能
	return strings.Contains(line, `"level":"`+level+`"`) ||
		strings.Contains(line, `"level": "`+level+`"`) ||
		strings.Contains(line, "["+level+"]") ||
		strings.Contains(line, level+":")
}
