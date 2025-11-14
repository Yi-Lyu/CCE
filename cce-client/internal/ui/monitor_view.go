package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ethan/cce-client/internal/config"
	"github.com/ethan/cce-client/internal/service"
)

// MonitorView 性能监控视图
type MonitorView struct {
	serviceManager *service.Manager
	configManager  *config.Manager
	statusLabel    *widget.Label
	requestsLabel  *widget.Label
	difficultyInfo *widget.Label
	ticker         *time.Ticker
	stopChan       chan bool
}

// NewMonitorView 创建性能监控视图
func NewMonitorView(serviceManager *service.Manager, configManager *config.Manager) *MonitorView {
	return &MonitorView{
		serviceManager: serviceManager,
		configManager:  configManager,
		stopChan:       make(chan bool),
	}
}

// Build 构建视图
func (mv *MonitorView) Build() fyne.CanvasObject {
	// 服务状态
	mv.statusLabel = widget.NewLabel("服务状态: 未知")

	// 请求统计（占位）
	mv.requestsLabel = widget.NewLabel("总请求数: -")

	// 难度分布（占位）
	mv.difficultyInfo = widget.NewLabel("难度分布统计功能开发中...")

	// 刷新按钮
	refreshBtn := widget.NewButton("立即刷新", func() {
		mv.updateStats()
	})

	// 布局
	content := container.NewVBox(
		widget.NewCard("服务状态", "", mv.statusLabel),
		widget.NewCard("请求统计", "", mv.requestsLabel),
		widget.NewCard("难度分布", "", mv.difficultyInfo),
		refreshBtn,
		widget.NewLabel("\n提示：性能监控功能需要解析日志文件，当前为简化版实现"),
	)

	// 启动自动刷新
	mv.startAutoRefresh()

	return container.NewPadded(content)
}

// updateStats 更新统计信息（从 goroutine 中调用）
func (mv *MonitorView) updateStats() {
	// 获取数据并更新 UI（Fyne v2 UI 更新是线程安全的）
	status := mv.serviceManager.GetStatus()
	mv.statusLabel.SetText(fmt.Sprintf("服务状态: %s", status))

	// TODO: 实现日志解析和统计
	mv.requestsLabel.SetText("总请求数: - （功能开发中）")
	mv.difficultyInfo.SetText("难度分布统计功能开发中...\n" +
		"计划实现：\n" +
		"- 1 级难度: XX 次 (XX%)\n" +
		"- 2 级难度: XX 次 (XX%)\n" +
		"- ...\n" +
		"- 平均响应时间: XX ms")
}

// startAutoRefresh 启动自动刷新
func (mv *MonitorView) startAutoRefresh() {
	mv.ticker = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-mv.ticker.C:
				mv.updateStats()
			case <-mv.stopChan:
				return
			}
		}
	}()

	// 立即更新一次
	mv.updateStats()
}

// Stop 停止自动刷新
func (mv *MonitorView) Stop() {
	if mv.ticker != nil {
		mv.ticker.Stop()
	}
	close(mv.stopChan)
}
