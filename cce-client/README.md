# CCE (Claude Code Exchange) - macOS 客户端

这是 Claude Code Exchange 智能代理服务的 macOS 图形化客户端，用于管理和配置代理服务。

## 功能特性

- ✅ **菜单栏集成**: 运行在 macOS 右上角任务栏，方便快捷
- ✅ **服务控制**: 启动、停止、重启代理服务
- ✅ **图形化配置**: 可视化编辑服务列表、难度映射、功能开关
- ✅ **日志查看**: 实时查看代理服务日志
- ✅ **性能监控**: 监控请求统计和难度分布（开发中）
- 🚧 **开机自启动**: 支持开机自动启动（计划中）

## 系统要求

- macOS 11.0 (Big Sur) 或更高版本
- Apple Silicon (arm64) 或 Intel (amd64) 处理器

## 开发环境准备

### 1. 安装 Go

```bash
# 使用 Homebrew 安装
brew install go

# 验证安装
go version  # 应该显示 go1.21 或更高版本
```

### 2. 安装 Fyne 依赖

```bash
# 安装必要的系统库
brew install pkg-config
brew install --cask glfw

# 安装 Fyne CLI 工具
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### 3. 克隆项目并编译代理服务

```bash
# 进入项目根目录
cd /path/to/Claude-Code-Exchange

# 先编译代理服务
cd proxy
make build
cd ..

# 现在可以构建客户端了
cd cce-client
```

## 构建和运行

### 开发模式

```bash
# 安装依赖
make install-deps

# 准备代理服务二进制
make prepare-binary

# 构建并运行
make run
```

### 生产打包

```bash
# 打包成 .app 文件
make package

# 打包完成后，会生成 CCE.app
# 将其拖到 Applications 文件夹即可安装
```

## 目录结构

```
cce-client/
├── cmd/
│   └── main.go                   # 应用入口
├── internal/
│   ├── service/                  # 服务管理
│   │   └── manager.go           # 进程启动/停止/健康检查
│   ├── config/                   # 配置管理
│   │   └── manager.go           # 配置加载/保存/验证
│   └── ui/                       # UI 组件
│       ├── systray.go           # System Tray 菜单
│       ├── main_window.go       # 主窗口
│       ├── config_view.go       # 配置编辑视图
│       ├── logs_view.go         # 日志查看器
│       ├── monitor_view.go      # 性能监控
│       └── utils.go             # 工具函数
├── resources/                    # 资源文件
│   ├── icon.png                 # 应用图标
│   └── claude-proxy             # 内嵌的代理服务（构建时生成）
├── FyneApp.toml                  # Fyne 应用元数据
├── go.mod                        # Go 模块定义
├── Makefile                      # 构建脚本
└── README.md                     # 本文件
```

## 配置文件位置

客户端的配置文件保存在：
```
~/Library/Application Support/CCE/config.yaml
```

日志文件保存在：
```
~/Library/Application Support/CCE/logs/
```

## 使用说明

### 首次启动

1. 启动 CCE.app
2. 首次启动会在菜单栏显示图标（灰色 = 已停止）
3. 点击图标 → "打开主界面"
4. 在"配置编辑"标签页中配置你的服务：
   - 添加 evaluator 服务（用于评估任务难度）
   - 添加 executor 服务（用于执行任务）
   - 配置难度映射（1-5 级对应的服务）
5. 点击"保存配置并重启服务"
6. 在"服务控制"标签页中点击"启动服务"

### 日常使用

- **启动服务**: 菜单栏图标 → "启动服务"
- **查看状态**: 主界面 → "服务控制" 标签页
- **修改配置**: 主界面 → "配置编辑" 标签页 → 修改后保存
- **查看日志**: 主界面 → "日志查看" 标签页
- **性能监控**: 主界面 → "性能监控" 标签页

### 菜单栏图标含义

- ⚫ **灰色**: 服务已停止
- 🟡 **黄色**: 服务启动中
- 🟢 **绿色**: 服务正常运行
- 🔴 **红色**: 服务运行异常

## 故障排查

### 问题：启动失败，提示"未找到 claude-proxy 二进制文件"

**解决方案**:
```bash
cd ../proxy
make build
cd ../cce-client
make prepare-binary
```

### 问题：端口被占用

**解决方案**:
1. 检查是否有其他程序占用 27015 端口：
   ```bash
   lsof -i :27015
   ```
2. 如果需要，修改配置文件中的端口号

### 问题：服务启动后立即变为异常状态

**解决方案**:
1. 查看日志文件：`~/Library/Application Support/CCE/logs/`
2. 检查配置文件是否正确
3. 确认 API Key 是否有效

## 开发计划

### Phase 1: 基础功能 ✅
- [x] 服务启动/停止/重启
- [x] 健康状态监控
- [x] System Tray 集成
- [x] 基础 UI 框架

### Phase 2: 配置管理 🚧
- [x] 配置加载/保存
- [x] 配置验证
- [x] 图形化编辑器框架
- [ ] 服务添加/编辑/删除对话框
- [ ] 配置模板导入/导出

### Phase 3: 日志和监控 🚧
- [x] 日志文件读取
- [x] 日志实时刷新
- [ ] 日志 JSON 解析和格式化
- [ ] 日志搜索和高亮
- [ ] 性能统计实现
- [ ] 难度分布图表

### Phase 4: 高级功能 📅
- [ ] 开机自启动
- [ ] 通知中心集成
- [ ] 多语言支持
- [ ] 自动更新检查
- [ ] 配置备份/恢复

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 联系方式

- GitHub: https://github.com/yourusername/cce-client
- Email: your.email@example.com
