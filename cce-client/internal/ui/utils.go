package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

// showError 显示错误对话框
func showError(window fyne.Window, title, message string) {
	err := fmt.Errorf("%s: %s", title, message)
	dialog.ShowError(err, window)
}

// showInfo 显示信息对话框
func showInfo(window fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, window)
}

// showConfirm 显示确认对话框
func showConfirm(window fyne.Window, title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, window)
}

// GetAppIcon 获取应用图标
func GetAppIcon() fyne.Resource {
	// 使用 Fyne 内置图标作为临时方案
	return theme.ComputerIcon()
}
