package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ethan/cce-client/internal/config"
	"github.com/ethan/cce-client/internal/service"
)

// MainWindow 主窗口
type MainWindow struct {
	window         fyne.Window
	serviceManager *service.Manager
	configManager  *config.Manager

	// UI 组件
	tabs          *container.AppTabs
	controlView   *ControlView
	configView    *ConfigView
	logsView      *LogsView
	monitorView   *MonitorView
}

// NewMainWindow 创建主窗口
func NewMainWindow(app fyne.App, serviceManager *service.Manager, configManager *config.Manager) *MainWindow {
	w := app.NewWindow("Claude Code Exchange - 配置管理")

	mw := &MainWindow{
		window:         w,
		serviceManager: serviceManager,
		configManager:  configManager,
	}

	mw.buildUI()

	// 设置窗口大小（macOS 推荐尺寸）
	w.Resize(fyne.NewSize(1000, 700))

	// 窗口关闭时隐藏而不是退出
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	return mw
}

// GetWindow 获取 Fyne 窗口对象
func (mw *MainWindow) GetWindow() fyne.Window {
	return mw.window
}

// Cleanup 清理资源（应用退出时调用）
func (mw *MainWindow) Cleanup() {
	// 停止所有视图的后台任务
	if mw.logsView != nil {
		mw.logsView.Stop()
	}
	if mw.monitorView != nil {
		mw.monitorView.Stop()
	}
}

// buildUI 构建 UI
func (mw *MainWindow) buildUI() {
	// 创建各个视图
	mw.controlView = NewControlView(mw.serviceManager, mw.configManager)
	mw.configView = NewConfigView(mw.configManager, mw.serviceManager)
	mw.logsView = NewLogsView(mw.configManager)
	mw.monitorView = NewMonitorView(mw.serviceManager, mw.configManager)

	// 创建 Tab 容器
	mw.tabs = container.NewAppTabs(
		container.NewTabItem("服务控制", mw.controlView.Build()),
		container.NewTabItem("配置编辑", mw.configView.Build()),
		container.NewTabItem("日志查看", mw.logsView.Build()),
		container.NewTabItem("性能监控", mw.monitorView.Build()),
	)

	mw.tabs.SetTabLocation(container.TabLocationTop)

	mw.window.SetContent(mw.tabs)
}

// ControlView 服务控制视图
type ControlView struct {
	serviceManager  *service.Manager
	configManager   *config.Manager
	statusLabel     *widget.Label
	startBtn        *widget.Button
	stopBtn         *widget.Button
	restartBtn      *widget.Button
	progressBar     *widget.ProgressBarInfinite
	progressVisible bool
}

// NewControlView 创建服务控制视图
func NewControlView(serviceManager *service.Manager, configManager *config.Manager) *ControlView {
	cv := &ControlView{
		serviceManager: serviceManager,
		configManager:  configManager,
	}

	// 设置状态回调
	serviceManager.SetStatusCallback(func(status service.Status) {
		cv.updateStatus(status)
	})

	return cv
}

// Build 构建视图
func (cv *ControlView) Build() fyne.CanvasObject {
	// 状态显示
	cv.statusLabel = widget.NewLabel("状态: " + cv.serviceManager.GetStatus().String())
	cv.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 按钮（移除弹窗通知，状态变化已通过进度条和标签显示）
	cv.startBtn = widget.NewButton("启动服务", func() {
		if err := cv.serviceManager.Start(); err != nil {
			// 启动失败时显示在状态标签中
			cv.statusLabel.SetText("启动失败: " + err.Error())
			cv.statusLabel.Refresh()
		}
		// 成功时会通过状态回调自动更新 UI
	})

	cv.stopBtn = widget.NewButton("停止服务", func() {
		if err := cv.serviceManager.Stop(); err != nil {
			cv.statusLabel.SetText("停止失败: " + err.Error())
			cv.statusLabel.Refresh()
		}
	})

	cv.restartBtn = widget.NewButton("重启服务", func() {
		if err := cv.serviceManager.Restart(); err != nil {
			cv.statusLabel.SetText("重启失败: " + err.Error())
			cv.statusLabel.Refresh()
		}
	})

	// 更新按钮状态
	cv.updateButtonStates(cv.serviceManager.GetStatus())

	// 服务信息
	port := cv.configManager.GetProxyPort()
	infoText := widget.NewLabel(fmt.Sprintf("代理端口: 127.0.0.1:%d", port))

	// 进度指示器（初始隐藏）
	cv.progressBar = widget.NewProgressBarInfinite()
	cv.progressBar.Hide()

	// 布局（优化间距和样式）
	// 使用 NewGridWithColumns 使按钮均匀分布，更符合 macOS 风格
	buttons := container.NewGridWithColumns(3,
		cv.startBtn,
		cv.stopBtn,
		cv.restartBtn,
	)

	// 状态卡片
	statusCard := widget.NewCard("服务状态", "", container.NewVBox(
		cv.statusLabel,
		infoText,
		cv.progressBar, // 添加进度条
	))

	// 控制卡片
	controlCard := widget.NewCard("服务控制", "", container.NewVBox(
		buttons,
		widget.NewSeparator(),
		widget.NewLabel("提示：服务启动后将监听在 127.0.0.1:27015（可在配置中修改）"),
	))

	content := container.NewVBox(
		statusCard,
		controlCard,
	)

	// 使用 NewPadded 增加边距，符合 macOS 设计规范
	return container.NewPadded(content)
}

// updateStatus 更新状态显示
// 此方法从 goroutine 中调用，需要确保在主线程更新UI
func (cv *ControlView) updateStatus(status service.Status) {
	// 添加调试日志
	fmt.Printf("[UI] 收到状态更新: %s\n", status)

	// Fyne UI 更新需要在主线程，使用 goroutine 调度
	go func() {
		cv.statusLabel.SetText("状态: " + status.String())
		cv.statusLabel.Refresh()
		cv.updateButtonStates(status)

		// 根据状态显示/隐藏进度条
		if status == service.StatusStarting {
			cv.progressBar.Start()
			cv.progressBar.Show()
		} else {
			cv.progressBar.Stop()
			cv.progressBar.Hide()
		}

		fmt.Printf("[UI] UI 已更新，按钮状态: 启动=%v, 停止=%v\n", cv.startBtn.Disabled(), cv.stopBtn.Disabled())
	}()
}

// updateButtonStates 更新按钮状态
func (cv *ControlView) updateButtonStates(status service.Status) {
	// 修复：启动中、运行中、异常 状态都应该禁用启动按钮
	isRunning := status == service.StatusStarting || status == service.StatusRunning || status == service.StatusUnhealthy

	fmt.Printf("[按钮] 状态=%s, isRunning=%v\n", status, isRunning)

	if isRunning {
		fmt.Println("[按钮] 禁用启动按钮，启用停止/重启按钮")
		cv.startBtn.Disable()
		cv.stopBtn.Enable()
		cv.restartBtn.Enable()
	} else {
		fmt.Println("[按钮] 启用启动按钮，禁用停止/重启按钮")
		cv.startBtn.Enable()
		cv.stopBtn.Disable()
		cv.restartBtn.Disable()
	}

	// 刷新按钮显示
	cv.startBtn.Refresh()
	cv.stopBtn.Refresh()
	cv.restartBtn.Refresh()
}
