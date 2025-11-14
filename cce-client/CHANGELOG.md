# 更新日志

本文档记录 CCE (Claude Code Exchange) macOS 客户端的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [未发布]

### 计划中
- 服务添加/编辑对话框
- JSON 日志解析和格式化
- 日志搜索和高亮功能
- 性能监控统计实现
- 开机自启动功能

---

## [0.1.0-alpha] - 2025-11-13

### 新增

#### 核心功能
- ✨ 实现服务管理器（启动/停止/重启代理服务）
- ✨ 实现配置管理器（YAML 配置加载/保存/验证）
- ✨ 集成 macOS System Tray（菜单栏图标和菜单）
- ✨ 创建主窗口界面（4 个 Tab 标签页）

#### 服务控制
- ✨ 服务启动/停止/重启按钮
- ✨ 服务状态监控（停止/启动中/运行中/异常）
- ✨ 健康检查机制（每 5 秒自动检查）
- ✨ 优雅停止（SIGTERM，30 秒超时）

#### 配置编辑
- ✨ 服务列表显示
- ✨ 难度映射配置（1-5 级下拉选择器）
- ✨ 功能开关（Evaluator Fallback、Service Auto Switch、Request Logging）
- ✨ 高级设置（代理端口、日志级别）
- ✨ 配置保存后自动重启服务

#### 日志查看
- ✨ 日志文件读取（最后 1000 行）
- ✨ 日志级别过滤（全部/debug/info/warn/error）
- ✨ 自动刷新（每 5 秒）
- ✨ 自动滚动到底部
- ✨ 清空显示按钮

#### 性能监控
- ✨ 服务状态显示
- ✨ 自动刷新框架（每 10 秒）

#### 开发工具
- ✨ Makefile 构建脚本
- ✨ FyneApp.toml 应用元数据
- ✨ 完整的项目文档（README、INSTALL、DEVELOPMENT）

### 修复

- 🐛 修复 Fyne List 组件选中状态访问问题（使用 OnSelected 回调）
- 🐛 修复按钮 Disabled 属性无法直接赋值问题（改用 Enable/Disable 方法）
- 🐛 移除未使用的 dialog 导入

### 技术细节

- 📦 使用 Fyne v2.5.4 作为 GUI 框架
- 📦 使用 gopkg.in/yaml.v3 处理配置文件
- 📦 支持 macOS arm64 和 amd64 架构
- 📦 内嵌 claude-proxy 二进制文件

### 已知问题

- ⚠️ System Tray 图标不能实时更新（Fyne 限制）
- ⚠️ 服务添加/编辑功能为占位实现
- ⚠️ 日志 JSON 解析未实现（显示原始文本）
- ⚠️ 性能监控统计功能未实现

### 文档

- 📝 添加 README.md（项目说明和使用指南）
- 📝 添加 INSTALL.md（详细安装步骤）
- 📝 添加 DEVELOPMENT.md（开发进度和规划）
- 📝 添加 CHANGELOG.md（本文件）

---

## 版本说明

### 版本命名规则

- **alpha**: 早期开发版本，功能不完整
- **beta**: 功能基本完整，待测试
- **rc**: 候选发布版本
- **stable**: 稳定版本

### 版本号格式

`主版本号.次版本号.修订号-阶段标识`

例如：`0.1.0-alpha`

- **主版本号**: 重大变更，可能不兼容
- **次版本号**: 新功能添加，向后兼容
- **修订号**: Bug 修复，向后兼容
- **阶段标识**: alpha、beta、rc（可选）

---

## 链接

- [GitHub 仓库](https://github.com/Yi-Lyu/Claude-Code-Exchange)
- [问题追踪](https://github.com/Yi-Lyu/Claude-Code-Exchange/issues)
- [Pull Requests](https://github.com/Yi-Lyu/Claude-Code-Exchange/pulls)
