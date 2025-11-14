package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/ethan/cce-client/internal/service"
)

// SystemTray 系统托盘管理器
type SystemTray struct {
	app            fyne.App
	mainWindow     fyne.Window
	serviceManager *service.Manager
	menu           *fyne.Menu
	icon           fyne.Resource
}

// NewSystemTray 创建系统托盘
func NewSystemTray(app fyne.App, mainWindow fyne.Window, serviceManager *service.Manager) *SystemTray {
	st := &SystemTray{
		app:            app,
		mainWindow:     mainWindow,
		serviceManager: serviceManager,
		icon:           getStatusIcon(service.StatusStopped),
	}

	// 设置状态变化回调
	serviceManager.SetStatusCallback(func(status service.Status) {
		st.updateIcon(status)
	})

	st.buildMenu()
	return st
}

// GetIcon 获取托盘图标
func (st *SystemTray) GetIcon() fyne.Resource {
	return st.icon
}

// GetMenu 获取托盘菜单
func (st *SystemTray) GetMenu() *fyne.Menu {
	return st.menu
}

// buildMenu 构建托盘菜单
func (st *SystemTray) buildMenu() {
	st.menu = fyne.NewMenu("CCE",
		fyne.NewMenuItem("打开主界面", func() {
			st.mainWindow.Show()
			st.mainWindow.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("启动服务", func() {
			st.startService()
		}),
		fyne.NewMenuItem("停止服务", func() {
			st.stopService()
		}),
		fyne.NewMenuItem("重启服务", func() {
			st.restartService()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() {
			st.quit()
		}),
	)
}

// startService 启动服务
func (st *SystemTray) startService() {
	if st.serviceManager.IsRunning() {
		log.Println("服务已在运行中")
		return
	}

	if err := st.serviceManager.Start(); err != nil {
		log.Printf("启动服务失败: %v", err)
		showError(st.mainWindow, "启动失败", err.Error())
	}
}

// stopService 停止服务
func (st *SystemTray) stopService() {
	if !st.serviceManager.IsRunning() {
		log.Println("服务未运行")
		return
	}

	if err := st.serviceManager.Stop(); err != nil {
		log.Printf("停止服务失败: %v", err)
		showError(st.mainWindow, "停止失败", err.Error())
	}
}

// restartService 重启服务
func (st *SystemTray) restartService() {
	if err := st.serviceManager.Restart(); err != nil {
		log.Printf("重启服务失败: %v", err)
		showError(st.mainWindow, "重启失败", err.Error())
	}
}

// quit 退出应用
func (st *SystemTray) quit() {
	// 停止服务
	if st.serviceManager.IsRunning() {
		log.Println("正在停止服务...")
		st.serviceManager.Stop()
	}

	// 退出应用
	st.app.Quit()
}

// updateIcon 更新托盘图标
func (st *SystemTray) updateIcon(status service.Status) {
	st.icon = getStatusIcon(status)
	// 注意：Fyne 目前不支持动态更新 system tray 图标
	// 需要在下次菜单打开时才会更新
	log.Printf("图标状态更新: %s", status)
}

// getStatusIcon 根据状态获取图标
func getStatusIcon(status service.Status) fyne.Resource {
	switch status {
	case service.StatusStopped:
		return theme.MediaStopIcon() // 灰色 - 已停止
	case service.StatusStarting:
		return theme.MediaPlayIcon() // 黄色 - 启动中
	case service.StatusRunning:
		return theme.ConfirmIcon() // 绿色 - 运行中
	case service.StatusUnhealthy:
		return theme.ErrorIcon() // 红色 - 异常
	default:
		return theme.QuestionIcon()
	}
}
