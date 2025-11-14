# CCE 客户端安装指南

本文档为没有 macOS 开发经验的用户提供详细的安装和使用指南。

## 前置要求

### 1. 安装 Homebrew（如果尚未安装）

Homebrew 是 macOS 的包管理器，用于安装开发工具。

```bash
# 打开终端（Terminal.app），运行以下命令：
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 按照屏幕提示完成安装
```

### 2. 安装 Go 语言环境

```bash
# 安装 Go
brew install go

# 验证安装
go version
# 应该显示: go version go1.21.x darwin/arm64（或 amd64）
```

### 3. 安装 Fyne 依赖

```bash
# 安装系统依赖
brew install pkg-config
brew install --cask glfw

# 安装 Fyne CLI 工具
go install fyne.io/fyne/v2/cmd/fyne@latest
```

## 构建 CCE 客户端

### 步骤 1: 克隆或下载项目

如果你已经有项目代码，跳到步骤 2。

```bash
# 克隆项目（替换为实际的 git 地址）
git clone https://github.com/yourusername/Claude-Code-Exchange.git
cd Claude-Code-Exchange
```

### 步骤 2: 切换到客户端分支

```bash
# 切换到 macOS 客户端分支
git checkout feature/macos-client
```

### 步骤 3: 编译代理服务

```bash
# 先编译代理服务（Go 后端）
cd proxy
make build

# 编译成功后，应该看到：
# build/claude-proxy（绿色可执行文件）

cd ..
```

### 步骤 4: 编译客户端

```bash
# 进入客户端目录
cd cce-client

# 安装 Go 依赖
make install-deps

# 准备代理服务二进制
make prepare-binary

# 构建客户端
make build

# 或者直接运行（会自动构建）
make run
```

### 步骤 5: 打包为 .app 文件（可选）

```bash
# 打包成 macOS 应用
make package

# 打包完成后，会在当前目录生成 CCE.app
```

## 安装和使用

### 方式 1: 直接运行二进制（开发模式）

```bash
cd cce-client
./build/cce-client
```

### 方式 2: 安装 .app 文件（推荐）

```bash
# 1. 打包应用
cd cce-client
make package

# 2. 将 CCE.app 拖到 Applications 文件夹
mv CCE.app /Applications/

# 3. 从启动台或 Applications 文件夹打开 CCE
```

## 首次配置

### 1. 启动应用

- 双击 CCE.app
- 或从启动台搜索 "CCE"
- 应用会在菜单栏（右上角）显示图标

### 2. 打开主界面

- 点击菜单栏的 CCE 图标
- 选择 "打开主界面"

### 3. 配置服务

进入"配置编辑"标签页：

#### a. 添加 Evaluator 服务（决策者）

```yaml
ID: evaluator-main
名称: 主决策者
URL: https://api.anthropic.com/v1/messages
API Key: sk-ant-api03-your-actual-key-here
角色: evaluator
支持 Thinking: 是
```

#### b. 添加 Executor 服务（执行者）

示例 1 - Haiku（便宜）:
```yaml
ID: haiku-service
名称: Haiku服务
URL: https://api.anthropic.com/v1/messages
API Key: sk-ant-api03-your-actual-key-here
角色: executor
支持 Thinking: 是
```

示例 2 - Opus（强大）:
```yaml
ID: opus-service
名称: Opus服务
URL: https://api.anthropic.com/v1/messages
API Key: sk-ant-api03-your-actual-key-here
角色: executor
支持 Thinking: 是
```

#### c. 配置难度映射

```
难度 1 → haiku-service（简单任务用便宜的服务）
难度 2 → haiku-service
难度 3 → haiku-service
难度 4 → opus-service（复杂任务用强大的服务）
难度 5 → opus-service
```

#### d. 保存配置

点击"保存配置并重启服务"按钮。

### 4. 启动服务

进入"服务控制"标签页，点击"启动服务"。

服务状态会变为：启动中 → 运行中（绿色）

### 5. 验证服务

在终端测试：

```bash
# 健康检查
curl http://127.0.0.1:27015/health

# 应该返回:
{"status":"healthy","time":"..."}
```

## 配置 Claude Code 客户端

要让 Claude Code 使用代理服务：

1. 打开 Claude Code 设置
2. 找到 API 设置
3. 将 API 端点修改为：`http://127.0.0.1:27015`
4. 保存设置

现在 Claude Code 的所有请求都会通过代理服务进行智能路由！

## 常见问题

### Q1: 提示 "go: command not found"

**答**: 需要安装 Go 语言环境
```bash
brew install go
```

### Q2: 提示 "未找到 claude-proxy 二进制文件"

**答**: 需要先编译代理服务
```bash
cd proxy
make build
cd ../cce-client
make prepare-binary
```

### Q3: 服务启动失败，端口被占用

**答**: 检查端口占用情况
```bash
# 查看 27015 端口
lsof -i :27015

# 如果有其他程序占用，杀掉进程：
kill -9 <PID>

# 或修改配置文件中的端口号
```

### Q4: 应用图标不显示

**答**: 当前使用临时图标，如需自定义：
1. 准备 512x512 的 PNG 图标
2. 保存为 `cce-client/resources/icon.png`
3. 重新打包：`make package`

### Q5: 无法在 "应用程序" 中找到 CCE

**答**: 确保已将 CCE.app 复制到 /Applications/ 目录：
```bash
sudo cp -r CCE.app /Applications/
```

### Q6: macOS 提示 "无法打开，因为无法验证开发者"

**答**: 右键点击 CCE.app → "打开" → 点击"打开"（首次需要）

或在终端运行：
```bash
xattr -cr /Applications/CCE.app
```

## 卸载

### 删除应用

```bash
# 删除应用
rm -rf /Applications/CCE.app

# 删除配置和日志
rm -rf ~/Library/Application\ Support/CCE/
```

### 删除开发环境（可选）

```bash
# 删除项目
cd ~
rm -rf Claude-Code-Exchange

# 卸载 Go（如果不再需要）
brew uninstall go
```

## 获取帮助

遇到问题？

1. 查看日志：`~/Library/Application Support/CCE/logs/`
2. 提交 Issue：https://github.com/yourusername/cce-client/issues
3. 查看项目文档：`cce-client/README.md`

## 下一步

- 📖 阅读 [README.md](README.md) 了解更多功能
- 🔧 查看配置示例：`proxy/configs/config.example.yaml`
- 📊 使用性能监控功能优化服务配置
- ⚙️ 启用开机自启动（功能开发中）
