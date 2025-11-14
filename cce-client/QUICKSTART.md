# CCE 客户端快速开始指南

> 5 分钟快速上手 CCE macOS 客户端开发

---

## 📥 克隆代码

```bash
# 1. 克隆仓库
git clone https://github.com/Yi-Lyu/Claude-Code-Exchange.git
cd Claude-Code-Exchange

# 2. 切换到客户端开发分支
git checkout feature/macos-client
```

---

## ⚙️ 环境配置（一次性）

### 方法 1: 使用脚本（推荐）

```bash
# 运行自动安装脚本（如果提供）
./cce-client/scripts/setup.sh
```

### 方法 2: 手动安装

```bash
# 安装 Homebrew（如果没有）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go
brew install go

# 安装 Fyne 依赖
brew install pkg-config

# 验证安装
go version  # 应该显示 go1.21 或更高
```

---

## 🔨 编译和运行

```bash
# 进入客户端目录
cd cce-client

# 下载 Go 依赖
go mod download

# 编译代理服务（必需）
cd ../proxy
make build
cd ../cce-client

# 准备二进制文件
make prepare-binary

# 编译客户端
make build

# 运行客户端
make run
```

**预期结果**:
- ✅ 菜单栏出现 CCE 图标
- ✅ 点击图标可以打开主界面
- ✅ 配置文件自动生成到 `~/Library/Application Support/CCE/`

---

## 🎮 首次使用

### 1. 打开主界面

点击菜单栏的 CCE 图标 → 选择 **"打开主界面"**

### 2. 配置 API Key

进入 **"配置编辑"** 标签页：

1. 找到服务列表中的 API Key 字段
2. 将 `sk-your-api-key-here` 替换为真实的 Claude API Key
3. 点击 **"保存配置并重启服务"**

### 3. 启动服务

进入 **"服务控制"** 标签页：

1. 点击 **"启动服务"** 按钮
2. 等待状态变为 **"运行中"**（图标变绿）

### 4. 验证服务

打开终端，测试健康检查：

```bash
curl http://127.0.0.1:27015/health
```

应该返回：
```json
{"status":"healthy","time":"..."}
```

---

## 📂 项目结构（重要文件）

```
cce-client/
├── cmd/main.go              # 入口文件
├── internal/
│   ├── service/manager.go   # 服务管理（修改这里）
│   ├── config/manager.go    # 配置管理（修改这里）
│   └── ui/                  # UI 组件（修改这里）
│       ├── systray.go
│       ├── main_window.go
│       ├── config_view.go
│       ├── logs_view.go
│       └── monitor_view.go
├── Makefile                 # 构建命令
├── go.mod                   # Go 依赖
├── DEVELOPMENT.md          # 📖 开发文档（详细）
├── TODO.md                 # ✅ 待办事项
└── CHANGELOG.md            # 📝 变更记录
```

---

## 🛠️ 常用命令

```bash
# 编译
make build

# 运行
make run

# 清理
make clean

# 打包 .app
make package

# 格式化代码
make fmt

# 运行测试（待添加）
make test
```

---

## 🐛 遇到问题？

### 问题 1: 提示 "go: command not found"

**解决**:
```bash
brew install go
```

### 问题 2: 提示 "未找到 claude-proxy 二进制"

**解决**:
```bash
cd ../proxy
make build
cd ../cce-client
make prepare-binary
```

### 问题 3: 编译错误 "package fyne.io/fyne/v2: cannot find package"

**解决**:
```bash
go mod download
go mod tidy
```

### 问题 4: 服务启动失败

**解决**:
1. 检查 API Key 是否正确
2. 查看日志：`~/Library/Application Support/CCE/logs/`
3. 检查端口占用：`lsof -i :27015`

---

## 📖 进一步阅读

- **完整文档**: [DEVELOPMENT.md](DEVELOPMENT.md) - 开发进度、架构、调试技巧
- **安装指南**: [INSTALL.md](INSTALL.md) - 详细的安装步骤
- **项目说明**: [README.md](README.md) - 项目介绍和使用指南
- **待办事项**: [TODO.md](TODO.md) - 待实现功能清单
- **变更记录**: [CHANGELOG.md](CHANGELOG.md) - 版本历史

---

## 🚀 开始开发

### 1. 创建功能分支

```bash
git checkout -b feature/your-feature-name
```

### 2. 修改代码

参考 [DEVELOPMENT.md](DEVELOPMENT.md) 中的代码结构说明

### 3. 测试

```bash
make build
./build/cce-client
```

### 4. 提交更改

```bash
git add .
git commit -m "feat: 添加某某功能"
```

### 5. 推送分支

```bash
git push origin feature/your-feature-name
```

---

## 🎯 下一步

1. **查看待办事项**: 打开 [TODO.md](TODO.md) 查看待实现功能
2. **了解架构**: 阅读 [DEVELOPMENT.md](DEVELOPMENT.md)
3. **开始编码**: 从高优先级任务开始
4. **提交代码**: 遵循 Conventional Commits 规范

---

## 💡 提示

- **IDE 推荐**: VS Code + Go 插件 或 GoLand
- **调试**: 使用 `dlv debug ./cmd/main.go`
- **日志**: 查看 `~/Library/Application Support/CCE/logs/`
- **配置**: `~/Library/Application Support/CCE/config.yaml`

---

## 🤝 获取帮助

- **GitHub Issues**: https://github.com/Yi-Lyu/Claude-Code-Exchange/issues
- **查看文档**: [DEVELOPMENT.md](DEVELOPMENT.md) 有详细的调试技巧
- **联系作者**: your.email@example.com

---

**快速开始完成！** 🎉

现在你可以开始开发了！遇到问题先查看 [DEVELOPMENT.md](DEVELOPMENT.md)。
