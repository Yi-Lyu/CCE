# CCE 客户端开发进度文档

> **最后更新**: 2025-11-13
> **当前版本**: v0.1.0-alpha
> **当前分支**: feature/macos-client
> **状态**: 基础框架已完成，可编译运行

---

## 📋 目录

- [当前开发状态](#当前开发状态)
- [已完成功能](#已完成功能)
- [已知问题](#已知问题)
- [待开发功能](#待开发功能)
- [开发路线图](#开发路线图)
- [快速开始](#快速开始)
- [开发环境配置](#开发环境配置)
- [代码结构说明](#代码结构说明)
- [调试技巧](#调试技巧)

---

## 📊 当前开发状态

### ✅ Phase 1: 基础框架 (已完成 100%)

**完成时间**: 2025-11-13

| 功能模块 | 状态 | 完成度 | 备注 |
|---------|------|--------|------|
| 项目结构搭建 | ✅ 完成 | 100% | 符合 Go 标准项目布局 |
| Service Manager | ✅ 完成 | 100% | 进程管理、健康检查 |
| Config Manager | ✅ 完成 | 100% | YAML 配置加载/保存/验证 |
| System Tray 集成 | ✅ 完成 | 100% | 菜单栏图标、右键菜单 |
| 主窗口框架 | ✅ 完成 | 100% | 4 个 Tab 标签页 |
| 服务控制界面 | ✅ 完成 | 100% | 启动/停止/重启按钮 |
| 配置编辑器框架 | ✅ 完成 | 80% | 基础 UI 完成，对话框待实现 |
| 日志查看器框架 | ✅ 完成 | 60% | 基础读取完成，解析待完善 |
| 性能监控框架 | ✅ 完成 | 30% | UI 框架完成，统计功能待实现 |
| 构建脚本 | ✅ 完成 | 100% | Makefile 完善 |
| 文档 | ✅ 完成 | 100% | README、INSTALL 已编写 |

**第一次运行测试**:
- ✅ 编译成功
- ✅ 客户端启动正常
- ✅ 菜单栏图标显示
- ✅ 配置自动生成
- ✅ UI 界面正常

---

## ✅ 已完成功能

### 1. 核心服务管理

**文件**: `internal/service/manager.go` (327 行)

**功能清单**:
- [x] 启动代理服务（exec.Command）
- [x] 停止代理服务（SIGTERM 优雅停止）
- [x] 重启服务（停止 + 启动）
- [x] 进程状态监控（4 种状态：停止/启动中/运行中/异常）
- [x] 健康检查（HTTP GET /health，每 5 秒）
- [x] 自动查找二进制文件（支持开发环境和生产环境）
- [x] 状态变化回调机制

**状态枚举**:
```go
StatusStopped    // 已停止（灰色图标）
StatusStarting   // 启动中（黄色图标）
StatusRunning    // 运行中（绿色图标）
StatusUnhealthy  // 运行异常（红色图标）
```

**已测试场景**:
- ✅ 从停止状态启动服务
- ✅ 健康检查成功（状态变为运行中）
- ⚠️ 服务崩溃检测（待测试）
- ⚠️ 端口占用处理（待测试）

---

### 2. 配置管理系统

**文件**: `internal/config/manager.go` (260 行)

**功能清单**:
- [x] YAML 配置加载（gopkg.in/yaml.v3）
- [x] 配置保存（格式化输出）
- [x] 配置验证（服务、难度映射）
- [x] 默认配置生成
- [x] 配置文件路径管理（macOS 标准路径）

**配置结构**:
```yaml
proxy:              # 代理服务配置
  port: 27015
  timeouts: ...
services:           # 服务列表
  - evaluator       # 至少 1 个
  - executor(s)     # 至少 1 个
difficulty_mapping: # 难度映射 (1-5)
evaluator:          # 评估器配置
features:           # 功能开关
logging:            # 日志配置
```

**验证规则**:
- [x] 至少 1 个 evaluator 服务
- [x] 至少 1 个 executor 服务
- [x] 服务 ID 唯一性
- [x] 难度映射完整性（1-5 级）
- [x] 映射的服务 ID 存在性

**配置路径**:
- macOS: `~/Library/Application Support/CCE/config.yaml`
- 日志: `~/Library/Application Support/CCE/logs/`

---

### 3. System Tray 菜单栏集成

**文件**: `internal/ui/systray.go` (106 行)

**功能清单**:
- [x] 菜单栏图标显示
- [x] 图标状态变色（根据服务状态）
- [x] 右键菜单（6 个选项）
- [x] 打开主界面
- [x] 启动/停止/重启服务
- [x] 退出应用（优雅关闭）

**菜单结构**:
```
CCE
├── 打开主界面
├── ────────────
├── 启动服务
├── 停止服务
├── 重启服务
├── ────────────
└── 退出
```

**图标状态映射**:
| 服务状态 | 图标 | 颜色 | Fyne Icon |
|---------|------|------|-----------|
| 已停止 | ⚫ | 灰色 | MediaStopIcon |
| 启动中 | 🟡 | 黄色 | MediaPlayIcon |
| 运行中 | 🟢 | 绿色 | ConfirmIcon |
| 异常 | 🔴 | 红色 | ErrorIcon |

**已知限制**:
- ⚠️ Fyne 目前不支持动态更新 system tray 图标（需要重新打开菜单才能看到图标变化）

---

### 4. 主窗口界面

**文件**: `internal/ui/main_window.go` (165 行)

**布局结构**:
```
主窗口 (900x600)
├── Tab: 服务控制
│   ├── 状态显示
│   └── 控制按钮（启动/停止/重启）
├── Tab: 配置编辑
│   ├── 服务管理
│   ├── 难度映射
│   ├── 功能开关
│   └── 高级设置
├── Tab: 日志查看
│   ├── 日志过滤
│   └── 实时刷新
└── Tab: 性能监控
    ├── 请求统计
    └── 难度分布
```

**交互特性**:
- [x] 窗口关闭时隐藏（不退出应用）
- [x] 从托盘菜单可重新打开
- [x] Tab 切换流畅
- [x] 按钮状态自动更新

---

### 5. 配置编辑器

**文件**: `internal/ui/config_view.go` (203 行)

**已实现**:
- [x] 服务列表显示
- [x] 难度映射下拉选择器
- [x] 功能开关（3 个 Checkbox）
- [x] 高级设置（端口、日志级别）
- [x] 保存按钮（自动重启服务）

**未实现（占位）**:
- ⚠️ 添加服务对话框（显示"功能开发中"）
- ⚠️ 编辑服务对话框（显示"功能开发中"）
- ⚠️ 删除服务确认对话框（显示"功能开发中"）

**待完善功能**:
```go
// TODO: 实现这些方法
func (cv *ConfigView) showAddServiceDialog()
func (cv *ConfigView) showEditServiceDialog(index int)
func (cv *ConfigView) deleteService(index int)
```

---

### 6. 日志查看器

**文件**: `internal/ui/logs_view.go` (155 行)

**已实现**:
- [x] 读取日志文件（最后 1000 行）
- [x] 日志级别过滤（全部/debug/info/warn/error）
- [x] 自动刷新（每 5 秒）
- [x] 自动滚动到底部
- [x] 手动刷新按钮
- [x] 清空显示按钮

**未实现**:
- ⚠️ JSON 日志解析（当前显示原始文本）
- ⚠️ 日志格式化（彩色、对齐）
- ⚠️ 关键词搜索
- ⚠️ 关键词高亮
- ⚠️ 日志导出功能

**待优化**:
```go
// TODO: 实现 JSON 解析
type LogEntry struct {
    Time      string `json:"time"`
    Level     string `json:"level"`
    Msg       string `json:"msg"`
    UserID    string `json:"user_id,omitempty"`
    SessionID string `json:"session_id,omitempty"`
    Duration  int    `json:"duration_ms,omitempty"`
}

// TODO: 实现精确过滤
func containsLevel(line, level string) bool {
    // 当前实现不准确，需要解析 JSON
}
```

---

### 7. 性能监控

**文件**: `internal/ui/monitor_view.go` (106 行)

**已实现**:
- [x] 服务状态显示
- [x] 自动刷新（每 10 秒）
- [x] 手动刷新按钮

**未实现（占位）**:
- ⚠️ 请求统计（总数、成功/失败）
- ⚠️ 难度分布图表
- ⚠️ 平均响应延迟
- ⚠️ 性能趋势图

**待实现功能**:
```go
// TODO: 解析日志获取统计数据
type Statistics struct {
    TotalRequests   int
    SuccessCount    int
    FailureCount    int
    AvgDuration     float64
    DifficultyDist  map[int]int // 1-5 级分布
}

// TODO: 实现数据收集
func (mv *MonitorView) collectStats() *Statistics
```

---

## ⚠️ 已知问题

### 编译相关

1. **CGO 依赖** ⚠️
   - **问题**: Fyne 依赖 CGO，在某些环境下可能编译失败
   - **影响**: 需要安装 pkg-config 和相关系统库
   - **解决**: 按 INSTALL.md 安装依赖
   - **优先级**: 中

2. **编译警告** ℹ️
   ```
   ld: warning: ignoring duplicate libraries: '-lobjc'
   ```
   - **问题**: macOS 链接器警告
   - **影响**: 不影响功能，可忽略
   - **优先级**: 低

### 功能相关

3. **图标不更新** ⚠️
   - **问题**: System Tray 图标状态变化不实时更新
   - **原因**: Fyne v2.5.4 的限制
   - **影响**: 需要重新打开菜单才能看到图标变化
   - **解决方案**: 等待 Fyne 更新或寻找替代方案
   - **优先级**: 中
   - **参考**: https://github.com/fyne-io/fyne/issues/xxxx

4. **日志级别过滤不准确** ⚠️
   - **问题**: `containsLevel()` 函数实现过于简单
   - **原因**: 未解析 JSON 格式日志
   - **影响**: 过滤可能不准确
   - **优先级**: 高

5. **端口占用检测缺失** ⚠️
   - **问题**: 启动前未检查端口是否被占用
   - **影响**: 启动失败时错误信息不明确
   - **解决方案**: 添加端口检测逻辑
   - **优先级**: 高

### UI/UX 相关

6. **临时图标** ℹ️
   - **问题**: 使用 Fyne 内置图标，不够专业
   - **影响**: 用户体验
   - **解决方案**: 设计自定义图标
   - **优先级**: 低

7. **对话框占位** ⚠️
   - **问题**: 服务添加/编辑功能未实现
   - **影响**: 只能手动编辑 YAML 文件
   - **优先级**: 高

8. **窗口大小固定** ℹ️
   - **问题**: 窗口大小固定为 900x600
   - **影响**: 在小屏幕上可能不适配
   - **解决方案**: 支持自适应布局
   - **优先级**: 低

---

## 🚧 待开发功能

### Phase 2A: 配置编辑器完善 (高优先级)

**预计工作量**: 2-3 天

#### 2.1 服务添加对话框
- [ ] 创建表单对话框（Fyne Dialog）
- [ ] 字段验证（ID、URL、API Key）
- [ ] 角色选择（evaluator/executor）
- [ ] supports_thinking 开关
- [ ] 保存后刷新列表

**实现文件**: `internal/ui/service_dialog.go` (新建)

```go
// TODO: 实现
type ServiceDialog struct {
    window fyne.Window
    form   *widget.Form
    // ...
}

func ShowAddServiceDialog(window fyne.Window, onSave func(service *config.Service))
```

#### 2.2 服务编辑对话框
- [ ] 加载现有服务数据
- [ ] 编辑表单
- [ ] 保存更新

#### 2.3 服务删除确认
- [ ] 确认对话框
- [ ] 检查依赖（难度映射是否使用）
- [ ] 删除后更新配置

#### 2.4 配置导入/导出
- [ ] 导出为 YAML 文件
- [ ] 从文件导入配置
- [ ] 配置模板库（官方 API、第三方 API）

---

### Phase 2B: 日志功能增强 (中优先级)

**预计工作量**: 2-3 天

#### 2.5 JSON 日志解析
- [ ] 实现 LogEntry 结构体
- [ ] JSON 解析逻辑
- [ ] 错误处理（解析失败时显示原始文本）

**实现文件**: `internal/logger/parser.go` (新建)

```go
type LogEntry struct {
    Time      string
    Level     string
    Msg       string
    UserID    string
    SessionID string
    Duration  int
    Extra     map[string]interface{}
}

func ParseLogLine(line string) (*LogEntry, error)
```

#### 2.6 日志格式化显示
- [ ] 彩色日志（级别不同颜色）
- [ ] 列对齐（时间、级别、消息）
- [ ] 字段折叠（可展开查看详细信息）

#### 2.7 日志搜索
- [ ] 搜索输入框
- [ ] 关键词高亮
- [ ] 正则表达式支持
- [ ] 搜索历史

#### 2.8 日志导出
- [ ] 导出为文本文件
- [ ] 导出为 JSON 文件
- [ ] 按时间范围导出
- [ ] 按级别过滤导出

---

### Phase 3: 性能监控实现 (中优先级)

**预计工作量**: 2-3 天

#### 3.1 数据收集
- [ ] 解析日志文件获取统计数据
- [ ] 请求计数（总数、成功、失败）
- [ ] 难度分布统计（1-5 级）
- [ ] 响应延迟统计（平均、P50、P95、P99）

**实现文件**: `internal/logger/stats.go` (新建)

```go
type Statistics struct {
    TimeRange       TimeRange
    TotalRequests   int
    SuccessCount    int
    FailureCount    int
    AvgDuration     float64
    P50Duration     float64
    P95Duration     float64
    P99Duration     float64
    DifficultyDist  map[int]int
    HourlyTrend     []HourlyStats
}

func CollectStatistics(logsPath string, timeRange TimeRange) (*Statistics, error)
```

#### 3.2 图表绘制
- [ ] 难度分布饼图（Fyne Canvas）
- [ ] 请求趋势折线图（按小时）
- [ ] 响应延迟柱状图
- [ ] 实时刷新动画

#### 3.3 /status 接口集成
- [ ] 调用代理服务 /status 接口
- [ ] 显示当前配置信息
- [ ] 显示服务健康状态

---

### Phase 4: 高级功能 (低优先级)

**预计工作量**: 3-4 天

#### 4.1 开机自启动
- [ ] macOS LaunchAgent 配置
- [ ] plist 文件生成
- [ ] launchctl 命令集成
- [ ] 开关按钮（启用/禁用）

**实现文件**: `internal/autostart/launchd.go` (待实现)

```go
func EnableAutoStart() error {
    // 创建 ~/Library/LaunchAgents/com.cce.claude-proxy.plist
}

func DisableAutoStart() error {
    // 删除 plist 文件，卸载服务
}

func IsAutoStartEnabled() bool
```

**plist 模板**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "...">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.cce.claude-proxy</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Applications/CCE.app/Contents/MacOS/CCE</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
```

#### 4.2 通知中心集成
- [ ] macOS 通知权限请求
- [ ] 服务状态变化通知
- [ ] 错误告警通知
- [ ] 通知点击跳转

#### 4.3 多语言支持
- [ ] i18n 框架集成
- [ ] 英文翻译
- [ ] 语言切换按钮

#### 4.4 自动更新
- [ ] 检查 GitHub Releases
- [ ] 版本比较
- [ ] 下载新版本
- [ ] 自动安装

#### 4.5 配置备份/恢复
- [ ] 自动备份配置文件
- [ ] 配置历史记录（最近 10 个）
- [ ] 一键恢复

---

## 🗺️ 开发路线图

### v0.1.0-alpha (当前版本) ✅
**发布日期**: 2025-11-13
**状态**: 已完成

- ✅ 基础框架搭建
- ✅ 服务启动/停止/重启
- ✅ System Tray 集成
- ✅ 基础 UI 界面
- ✅ 配置加载/保存
- ✅ 文档完善

**里程碑**: 可编译运行，基础功能可用

---

### v0.2.0-alpha (下一版本) 🚧
**预计发布**: 2025-11-20
**状态**: 规划中

**目标**: 完善配置编辑器和日志功能

- [ ] 服务添加/编辑/删除对话框
- [ ] JSON 日志解析和格式化
- [ ] 日志搜索和高亮
- [ ] 端口占用检测
- [ ] 错误信息优化

**里程碑**: 日常使用完全可用

---

### v0.3.0-beta 📅
**预计发布**: 2025-11-27
**状态**: 规划中

**目标**: 实现性能监控

- [ ] 请求统计
- [ ] 难度分布图表
- [ ] 响应延迟统计
- [ ] /status 接口集成

**里程碑**: 功能完整

---

### v1.0.0 🎯
**预计发布**: 2025-12-10
**状态**: 规划中

**目标**: 生产就绪

- [ ] 开机自启动
- [ ] 通知中心集成
- [ ] 自动更新
- [ ] 完整测试覆盖
- [ ] 用户手册
- [ ] 视频教程

**里程碑**: 正式版发布

---

## 🚀 快速开始

### 首次克隆后的设置

```bash
# 1. 克隆仓库
git clone https://github.com/Yi-Lyu/Claude-Code-Exchange.git
cd Claude-Code-Exchange

# 2. 切换到客户端分支
git checkout feature/macos-client

# 3. 安装 Go 依赖（macOS）
brew install go
brew install pkg-config

# 4. 进入客户端目录
cd cce-client

# 5. 下载 Go 模块
go mod download

# 6. 编译代理服务
cd ../proxy
make build
cd ../cce-client

# 7. 准备二进制文件
make prepare-binary

# 8. 编译客户端
make build

# 9. 运行客户端
make run
```

### 日常开发流程

```bash
# 1. 更新代码
git pull origin feature/macos-client

# 2. 修改代码
# ... 编辑文件 ...

# 3. 编译测试
make build
./build/cce-client

# 4. 提交更改
git add .
git commit -m "feat: 添加新功能"
git push
```

---

## 🛠️ 开发环境配置

### macOS (推荐)

**系统要求**:
- macOS 11.0 (Big Sur) 或更高
- Apple Silicon (arm64) 或 Intel (amd64)

**依赖安装**:
```bash
# 安装 Homebrew（如果没有）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go
brew install go

# 安装 Fyne 依赖
brew install pkg-config

# 安装 Fyne CLI（可选，用于打包）
go install fyne.io/fyne/v2/cmd/fyne@latest
```

**IDE 推荐**:
- GoLand（商业）
- VS Code + Go 插件（免费）

**VS Code 配置**:
```json
// .vscode/settings.json
{
    "go.toolsManagement.autoUpdate": true,
    "go.lintTool": "golangci-lint",
    "go.formatTool": "goimports",
    "go.useLanguageServer": true
}
```

---

### Linux (实验性支持)

**注意**: Fyne 支持 Linux，但 System Tray 功能可能受限。

```bash
# Ubuntu/Debian
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install golang gcc libXcursor-devel libXrandr-devel \
                 mesa-libGL-devel libXi-devel libXinerama-devel \
                 libXxf86vm-devel
```

---

### Windows (不支持)

当前版本不支持 Windows，因为：
1. System Tray API 依赖 macOS
2. 配置路径使用 macOS 标准路径
3. 未测试 Windows 编译

如需支持 Windows，需要：
- [ ] 适配 Windows 系统托盘 API
- [ ] 修改配置文件路径
- [ ] 测试编译和运行

---

## 📁 代码结构说明

### 目录树
```
cce-client/
├── cmd/
│   └── main.go                # 应用入口（70 行）
│
├── internal/
│   ├── service/
│   │   └── manager.go         # 服务管理（327 行）
│   │                          # - 启动/停止/重启
│   │                          # - 健康检查
│   │                          # - 状态监控
│   │
│   ├── config/
│   │   └── manager.go         # 配置管理（260 行）
│   │                          # - YAML 加载/保存
│   │                          # - 配置验证
│   │
│   ├── ui/
│   │   ├── systray.go         # System Tray（106 行）
│   │   ├── main_window.go     # 主窗口（165 行）
│   │   ├── config_view.go     # 配置编辑器（203 行）
│   │   ├── logs_view.go       # 日志查看器（155 行）
│   │   ├── monitor_view.go    # 性能监控（106 行）
│   │   └── utils.go           # 工具函数（27 行）
│   │
│   ├── logger/                # 日志解析（待实现）
│   └── autostart/             # 开机自启动（待实现）
│
├── resources/
│   ├── claude-proxy           # 内嵌的代理服务二进制
│   └── README.md              # 资源说明
│
├── go.mod                      # Go 模块定义
├── go.sum                      # 依赖锁定文件
├── Makefile                    # 构建脚本
├── FyneApp.toml               # Fyne 应用元数据
│
├── README.md                   # 项目说明
├── INSTALL.md                  # 安装指南
└── DEVELOPMENT.md             # 本文件
```

**总代码量**: ~1,400 行（不含注释和空行）

---

### 核心模块职责

#### 1. cmd/main.go
**职责**: 应用初始化和组装

```go
func main() {
    // 1. 创建 Fyne 应用
    // 2. 初始化配置管理器
    // 3. 初始化服务管理器
    // 4. 创建主窗口
    // 5. 设置 System Tray
    // 6. 运行应用
    // 7. 清理资源
}
```

**依赖**:
- Fyne App
- Config Manager
- Service Manager
- UI 组件

---

#### 2. internal/service/manager.go
**职责**: 代理服务生命周期管理

**关键方法**:
```go
func (m *Manager) Start() error        // 启动服务
func (m *Manager) Stop() error         // 停止服务
func (m *Manager) Restart() error      // 重启服务
func (m *Manager) GetStatus() Status   // 获取状态
func (m *Manager) IsRunning() bool     // 是否运行中

// 内部方法
func (m *Manager) checkHealth() bool              // 健康检查
func (m *Manager) startHealthCheck()              // 启动健康监控
func (m *Manager) waitForReady()                  // 等待服务就绪
func (m *Manager) getBinaryPath() (string, error) // 查找二进制
```

**设计模式**:
- Observer 模式（状态变化回调）
- Singleton 模式（单例 Manager）

---

#### 3. internal/config/manager.go
**职责**: 配置文件管理和验证

**关键方法**:
```go
func (m *Manager) Load() error       // 加载配置
func (m *Manager) Save() error       // 保存配置
func (m *Manager) Validate() error   // 验证配置
func (m *Manager) GetConfig() *Config
func (m *Manager) SetConfig(config *Config)
func (m *Manager) GetProxyPort() int
func (m *Manager) GetLogsPath() string
```

**数据结构**:
```go
type Config struct {
    Proxy             ProxyConfig
    Services          []Service
    DifficultyMapping map[string]string
    Evaluator         EvaluatorConfig
    Features          Features
    Logging           LoggingConfig
}
```

---

#### 4. internal/ui/systray.go
**职责**: System Tray 菜单栏集成

**关键方法**:
```go
func NewSystemTray(...) *SystemTray
func (st *SystemTray) GetIcon() fyne.Resource
func (st *SystemTray) GetMenu() *fyne.Menu
func (st *SystemTray) updateIcon(status service.Status)

// 内部回调
func (st *SystemTray) startService()
func (st *SystemTray) stopService()
func (st *SystemTray) restartService()
func (st *SystemTray) quit()
```

---

#### 5. internal/ui/*_view.go
**职责**: 各个 Tab 页面的 UI 逻辑

**ControlView** (服务控制):
```go
func (cv *ControlView) Build() fyne.CanvasObject
func (cv *ControlView) updateStatus(status service.Status)
func (cv *ControlView) updateButtonStates()
```

**ConfigView** (配置编辑):
```go
func (cv *ConfigView) Build() fyne.CanvasObject
func (cv *ConfigView) buildServicesTab() fyne.CanvasObject
func (cv *ConfigView) buildDifficultyTab() fyne.CanvasObject
func (cv *ConfigView) buildFeaturesTab() fyne.CanvasObject
func (cv *ConfigView) buildAdvancedTab() fyne.CanvasObject
func (cv *ConfigView) saveConfig()
```

**LogsView** (日志查看):
```go
func (lv *LogsView) Build() fyne.CanvasObject
func (lv *LogsView) reloadLogs()
func (lv *LogsView) readLastLines(filepath string, n int)
func (lv *LogsView) startAutoRefresh()
```

**MonitorView** (性能监控):
```go
func (mv *MonitorView) Build() fyne.CanvasObject
func (mv *MonitorView) updateStats()
func (mv *MonitorView) startAutoRefresh()
```

---

## 🐛 调试技巧

### 1. 查看运行时日志

```bash
# 直接运行（查看 stdout）
./build/cce-client

# 或使用 make
make run
```

### 2. 检查配置文件

```bash
# 查看配置
cat ~/Library/Application\ Support/CCE/config.yaml

# 编辑配置
vi ~/Library/Application\ Support/CCE/config.yaml

# 删除配置（重新生成默认配置）
rm ~/Library/Application\ Support/CCE/config.yaml
```

### 3. 检查服务状态

```bash
# 健康检查
curl http://127.0.0.1:27015/health

# 详细状态
curl http://127.0.0.1:27015/status | jq .

# 检查进程
ps aux | grep claude-proxy
```

### 4. 查看日志文件

```bash
# 实时查看日志
tail -f ~/Library/Application\ Support/CCE/logs/claude-proxy-*.log

# 查看最近 100 行
tail -100 ~/Library/Application\ Support/CCE/logs/claude-proxy-*.log

# 搜索错误
grep -i error ~/Library/Application\ Support/CCE/logs/*.log
```

### 5. 调试编译问题

```bash
# 查看详细编译输出
go build -v -x -o build/cce-client ./cmd

# 检查依赖
go mod verify
go mod graph

# 清理缓存
go clean -cache -modcache -i -r
```

### 6. 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行
dlv debug ./cmd/main.go

# 设置断点
(dlv) break main.main
(dlv) continue
```

### 7. 内存泄漏检测

```bash
# 启用 pprof
go build -o build/cce-client ./cmd
GODEBUG=gctrace=1 ./build/cce-client

# 或使用 pprof
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 📝 代码规范

### Go 代码风格

遵循 [Effective Go](https://go.dev/doc/effective_go)：

1. **命名规范**:
   - 包名：小写，单个单词（如 `service`）
   - 导出函数：大写开头（如 `NewManager`）
   - 私有函数：小写开头（如 `checkHealth`）
   - 常量：驼峰式（如 `StatusRunning`）

2. **注释规范**:
   ```go
   // Manager 服务管理器（导出类型必须有注释）
   type Manager struct {
       // ...
   }

   // Start 启动代理服务（导出方法必须有注释）
   func (m *Manager) Start() error {
       // ...
   }
   ```

3. **错误处理**:
   ```go
   // 好的做法
   if err := doSomething(); err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }

   // 避免
   if err := doSomething(); err != nil {
       return err // 缺少上下文信息
   }
   ```

4. **格式化**:
   ```bash
   # 格式化代码
   make fmt

   # 或手动
   go fmt ./...
   goimports -w .
   ```

---

### Commit 规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**示例**:
```bash
feat(ui): 实现服务添加对话框

- 添加 ServiceDialog 组件
- 实现表单验证
- 集成到配置编辑器

Closes #123
```

---

## 📚 参考资源

### 官方文档

- **Fyne 文档**: https://docs.fyne.io
- **Fyne API**: https://pkg.go.dev/fyne.io/fyne/v2
- **Go 官方文档**: https://go.dev/doc

### 相关项目

- **Fyne 示例**: https://github.com/fyne-io/examples
- **System Tray 示例**: https://github.com/fyne-io/systray

### 工具

- **在线图标生成**: https://appicon.co
- **YAML 验证**: https://www.yamllint.com
- **JSON 格式化**: https://jsonlint.com

---

## 🤝 贡献指南

### 开发流程

1. **创建功能分支**:
   ```bash
   git checkout -b feature/service-dialog
   ```

2. **开发和测试**:
   ```bash
   # 编写代码
   # 运行测试
   make test

   # 格式化代码
   make fmt
   ```

3. **提交更改**:
   ```bash
   git add .
   git commit -m "feat(ui): 实现服务添加对话框"
   ```

4. **推送分支**:
   ```bash
   git push origin feature/service-dialog
   ```

5. **创建 Pull Request**

---

## 📞 联系方式

- **GitHub Issues**: https://github.com/Yi-Lyu/Claude-Code-Exchange/issues
- **作者**: Ethan
- **Email**: your.email@example.com

---

## 📄 许可证

MIT License

---

**最后更新**: 2025-11-13
**下次审查**: 2025-11-20
