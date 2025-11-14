package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/ethan/cce-client/internal/config"
	"github.com/ethan/cce-client/internal/service"
)

// ConfigView 配置编辑视图
type ConfigView struct {
	configManager  *config.Manager
	serviceManager *service.Manager
	tabs           *container.AppTabs
	selectedIndex  int // 追踪选中的服务索引
}

// NewConfigView 创建配置编辑视图
func NewConfigView(configManager *config.Manager, serviceManager *service.Manager) *ConfigView {
	return &ConfigView{
		configManager:  configManager,
		serviceManager: serviceManager,
		selectedIndex:  -1, // 初始化为未选中
	}
}

// Build 构建视图
func (cv *ConfigView) Build() fyne.CanvasObject {
	// 子标签页
	cv.tabs = container.NewAppTabs(
		container.NewTabItem("服务管理", cv.buildServicesTab()),
		container.NewTabItem("难度映射", cv.buildDifficultyTab()),
		container.NewTabItem("功能开关", cv.buildFeaturesTab()),
		container.NewTabItem("高级设置", cv.buildAdvancedTab()),
	)

	// 保存按钮
	saveBtn := widget.NewButton("保存配置并重启服务", func() {
		cv.saveConfig()
	})
	saveBtn.Importance = widget.HighImportance

	return container.NewBorder(
		nil,
		container.NewPadded(saveBtn),
		nil,
		nil,
		cv.tabs,
	)
}

// buildServicesTab 构建服务管理标签页
func (cv *ConfigView) buildServicesTab() fyne.CanvasObject {
	cfg := cv.configManager.GetConfig()

	// 服务列表
	servicesList := widget.NewList(
		func() int {
			return len(cfg.Services)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Service ID"),
				widget.NewLabel("Role"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			svc := cfg.Services[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s (%s)", svc.ID, svc.Name))
			box.Objects[1].(*widget.Label).SetText(svc.Role)
		},
	)

	// 设置选中回调
	servicesList.OnSelected = func(id widget.ListItemID) {
		cv.selectedIndex = int(id)
	}

	// 按钮
	addBtn := widget.NewButton("添加服务", func() {
		cv.showAddServiceDialog()
	})

	editBtn := widget.NewButton("编辑服务", func() {
		if cv.selectedIndex < 0 {
			showInfo(nil, "提示", "请先选择一个服务")
			return
		}
		cv.showEditServiceDialog(cv.selectedIndex)
	})

	deleteBtn := widget.NewButton("删除服务", func() {
		if cv.selectedIndex < 0 {
			showInfo(nil, "提示", "请先选择一个服务")
			return
		}
		cv.deleteService(cv.selectedIndex)
	})

	buttons := container.NewHBox(addBtn, editBtn, deleteBtn)

	return container.NewBorder(
		buttons,
		nil,
		nil,
		nil,
		servicesList,
	)
}

// buildDifficultyTab 构建难度映射标签页
func (cv *ConfigView) buildDifficultyTab() fyne.CanvasObject {
	cfg := cv.configManager.GetConfig()

	// 获取所有 executor 服务
	executors := []string{}
	for _, svc := range cfg.Services {
		if svc.Role == "executor" {
			executors = append(executors, svc.ID)
		}
	}

	if len(executors) == 0 {
		return widget.NewLabel("请先添加至少一个 executor 服务")
	}

	// 创建 5 个难度级别的下拉框
	form := &widget.Form{}
	selects := make(map[string]*widget.Select)

	for level := 1; level <= 5; level++ {
		levelStr := fmt.Sprintf("%d", level)
		currentValue := cfg.DifficultyMapping[levelStr]

		sel := widget.NewSelect(executors, func(value string) {
			cfg.DifficultyMapping[levelStr] = value
		})
		sel.SetSelected(currentValue)

		selects[levelStr] = sel
		form.Append(fmt.Sprintf("难度级别 %d", level), sel)
	}

	return container.NewVScroll(form)
}

// buildFeaturesTab 构建功能开关标签页
func (cv *ConfigView) buildFeaturesTab() fyne.CanvasObject {
	cfg := cv.configManager.GetConfig()

	fallbackCheck := widget.NewCheck("Evaluator Fallback（评估器失败时使用默认难度）", func(checked bool) {
		cfg.Features.EvaluatorFallback = checked
	})
	fallbackCheck.Checked = cfg.Features.EvaluatorFallback

	autoSwitchCheck := widget.NewCheck("Service Auto Switch（服务故障时自动切换）", func(checked bool) {
		cfg.Features.ServiceAutoSwitch = checked
	})
	autoSwitchCheck.Checked = cfg.Features.ServiceAutoSwitch

	loggingCheck := widget.NewCheck("Request Logging（记录请求日志）", func(checked bool) {
		cfg.Features.RequestLogging = checked
	})
	loggingCheck.Checked = cfg.Features.RequestLogging

	return container.NewVBox(
		widget.NewCard("功能开关", "", container.NewVBox(
			fallbackCheck,
			autoSwitchCheck,
			loggingCheck,
		)),
	)
}

// buildAdvancedTab 构建高级设置标签页
func (cv *ConfigView) buildAdvancedTab() fyne.CanvasObject {
	cfg := cv.configManager.GetConfig()

	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", cfg.Proxy.Port))

	logLevelSelect := widget.NewSelect([]string{"debug", "info", "warn", "error"}, func(value string) {
		cfg.Logging.Level = value
	})
	logLevelSelect.SetSelected(cfg.Logging.Level)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "代理端口", Widget: portEntry},
			{Text: "日志级别", Widget: logLevelSelect},
		},
	}

	return container.NewVScroll(form)
}

// saveConfig 保存配置
func (cv *ConfigView) saveConfig() {
	// 验证配置
	if err := cv.configManager.Validate(); err != nil {
		showError(nil, "配置验证失败", err.Error())
		return
	}

	// 保存配置
	if err := cv.configManager.Save(); err != nil {
		showError(nil, "保存配置失败", err.Error())
		return
	}

	showInfo(nil, "保存成功", "配置已保存，正在重启服务...")

	// 重启服务
	if cv.serviceManager.IsRunning() {
		if err := cv.serviceManager.Restart(); err != nil {
			showError(nil, "重启服务失败", err.Error())
		} else {
			showInfo(nil, "重启成功", "服务已重启")
		}
	}
}

// showAddServiceDialog 显示添加服务对话框
func (cv *ConfigView) showAddServiceDialog() {
	// 创建一个新的空服务
	trueVal := true
	newService := &config.Service{
		SupportsThinking: &trueVal,
	}
	cv.showServiceDialog(newService, true)
}

// showEditServiceDialog 显示编辑服务对话框
func (cv *ConfigView) showEditServiceDialog(index int) {
	cfg := cv.configManager.GetConfig()
	if index < 0 || index >= len(cfg.Services) {
		return
	}
	// 复制服务以便编辑
	svc := cfg.Services[index]
	cv.showServiceDialog(&svc, false)
}

// showServiceDialog 显示服务编辑对话框
func (cv *ConfigView) showServiceDialog(svc *config.Service, isNew bool) {
	// 创建表单输入框
	idEntry := widget.NewEntry()
	idEntry.SetText(svc.ID)
	if !isNew {
		idEntry.Disable() // 编辑时不允许修改ID
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(svc.Name)

	urlEntry := widget.NewEntry()
	urlEntry.SetText(svc.URL)

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(svc.APIKey)

	roleSelect := widget.NewSelect([]string{"evaluator", "executor"}, func(value string) {
		svc.Role = value
	})
	if svc.Role != "" {
		roleSelect.SetSelected(svc.Role)
	} else {
		roleSelect.SetSelected("executor")
	}

	thinkingCheck := widget.NewCheck("", func(checked bool) {
		svc.SupportsThinking = &checked
	})
	if svc.SupportsThinking != nil {
		thinkingCheck.Checked = *svc.SupportsThinking
	} else {
		thinkingCheck.Checked = true
	}

	// 创建表单项
	formItems := []*widget.FormItem{
		{Text: "服务 ID", Widget: idEntry},
		{Text: "服务名称", Widget: nameEntry},
		{Text: "API URL", Widget: urlEntry},
		{Text: "API Key", Widget: apiKeyEntry},
		{Text: "角色", Widget: roleSelect},
		{Text: "支持 Thinking", Widget: thinkingCheck},
	}

	// 保存回调
	onSave := func(confirmed bool) {
		if !confirmed {
			return
		}

		// 验证输入
		if idEntry.Text == "" || nameEntry.Text == "" || urlEntry.Text == "" || apiKeyEntry.Text == "" {
			showError(nil, "输入错误", "所有字段都必须填写")
			return
		}

		// 更新服务配置
		svc.ID = idEntry.Text
		svc.Name = nameEntry.Text
		svc.URL = urlEntry.Text
		svc.APIKey = apiKeyEntry.Text
		svc.Role = roleSelect.Selected

		cfg := cv.configManager.GetConfig()
		if isNew {
			// 检查ID是否重复
			for _, existingSvc := range cfg.Services {
				if existingSvc.ID == svc.ID {
					showError(nil, "添加失败", "服务ID已存在: "+svc.ID)
					return
				}
			}
			// 添加新服务
			cfg.Services = append(cfg.Services, *svc)
		} else {
			// 更新现有服务
			for i, existingSvc := range cfg.Services {
				if existingSvc.ID == svc.ID {
					cfg.Services[i] = *svc
					break
				}
			}
		}

		// 刷新列表
		cv.tabs.Refresh()
		showInfo(nil, "成功", "服务已保存（请记得点击保存配置按钮）")
	}

	// 显示对话框
	title := "添加服务"
	if !isNew {
		title = "编辑服务: " + svc.ID
	}

	d := dialog.NewForm(title, "保存", "取消", formItems, onSave, nil)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// deleteService 删除服务
func (cv *ConfigView) deleteService(index int) {
	cfg := cv.configManager.GetConfig()
	if index < 0 || index >= len(cfg.Services) {
		return
	}

	svc := cfg.Services[index]
	confirmMsg := fmt.Sprintf("确定要删除服务 %s (%s) 吗？\n\n警告：删除后需要手动调整难度映射配置", svc.ID, svc.Name)

	showConfirm(nil, "确认删除", confirmMsg, func(confirmed bool) {
		if confirmed {
			// 删除服务
			cfg.Services = append(cfg.Services[:index], cfg.Services[index+1:]...)
			cv.selectedIndex = -1

			// 刷新列表
			cv.tabs.Refresh()
			showInfo(nil, "删除成功", "服务已删除（请记得点击保存配置按钮）")
		}
	})
}
