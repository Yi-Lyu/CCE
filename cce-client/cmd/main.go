package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/ethan/cce-client/internal/config"
	"github.com/ethan/cce-client/internal/service"
	"github.com/ethan/cce-client/internal/ui"
)

const (
	AppID   = "com.cce.claude-proxy"
	AppName = "Claude Code Exchange"
)

func main() {
	// 创建 Fyne 应用
	a := app.NewWithID(AppID)
	a.SetIcon(ui.GetAppIcon())

	// 初始化配置管理器
	configManager, err := config.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	// 初始化服务管理器
	serviceManager := service.NewManager(configManager)

	// 创建主窗口（先隐藏）
	mainWindow := ui.NewMainWindow(a, serviceManager, configManager)
	mainWindow.GetWindow().Hide() // 启动时隐藏，只显示托盘图标

	// 设置 System Tray
	if desk, ok := a.(desktop.App); ok {
		systray := ui.NewSystemTray(a, mainWindow.GetWindow(), serviceManager)
		desk.SetSystemTrayIcon(systray.GetIcon())
		desk.SetSystemTrayMenu(systray.GetMenu())
	} else {
		log.Println("Warning: System tray not supported on this platform")
		mainWindow.GetWindow().Show() // 如果不支持托盘，直接显示窗口
	}

	// 运行应用
	a.Run()

	// 清理资源：应用退出时调用（优化：防止 Goroutine 泄漏）
	log.Println("应用退出中，正在清理资源...")
	mainWindow.Cleanup()     // 清理 UI 视图的后台任务
	serviceManager.Stop()    // 停止代理服务
	log.Println("清理完成")
}
