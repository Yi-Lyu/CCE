<div align="center">

# 🚀 Claude Code Exchange (CCE)

**Claude Code 智能代理 - 基于任务复杂度的 AI 模型路由系统**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey?style=flat&logo=apple)](https://www.apple.com/macos/)

[功能特性](#-功能特性) • [安装指南](#-安装指南) • [快速开始](#-快速开始) • [配置说明](#-配置说明) • [开发文档](#-开发指南)

**[📖 English Documentation](docs/README_EN.md)**

</div>

---

## 📖 项目简介

**Claude Code Exchange (CCE)** 是一个为 [Claude Code](https://claude.ai/code) 设计的智能本地代理服务。它能够拦截 Claude Code 发出的 API 请求，通过 AI 评估任务复杂度（1-5 级），并智能地将请求分发到最合适的 AI 模型服务。

### 🎯 核心价值

CCE 解决了使用 Claude Code 时的两大痛点：

1. **成本优化** - 不是所有任务都需要昂贵的顶级模型
2. **灵活选择** - 支持官方 API、第三方中转站、国内开源模型

### 💡 为什么选择 CCE？

- **💰 显著降低成本** - 简单任务自动使用便宜的模型，复杂任务才用强力模型
- **🌏 支持国内模型** - 可接入智谱 GLM、月之暗面 Kimi、MiniMax 等国内优秀模型
- **🔄 第三方 API 友好** - 兼容 Anthropic API 中转站，突破网络限制
- **⚡ 智能路由** - AI 自动评估任务难度，无需手动选择
- **🖥️ macOS 原生客户端** - 菜单栏应用，开箱即用
- **🔧 完全可配置** - 自由定义难度映射和服务端点

### 🎬 工作原理

```
┌─────────────────┐          ┌──────────────────┐          ┌─────────────────┐
│   Claude Code   │─────────▶│   CCE 代理服务    │─────────▶│   AI 评估器     │
│  (本地运行)      │          │  (127.0.0.1:27015)│          │  (Claude Haiku) │
└─────────────────┘          └──────────────────┘          └─────────────────┘
                                     │                              │
                                     │                     ┌────────▼────────┐
                                     │                     │ 任务复杂度: 1-5  │
                                     │                     └─────────────────┘
                                     │
                      ┌──────────────┴──────────────┐
                      ▼                             ▼
              ┌──────────────┐              ┌──────────────┐
              │  简单任务 1-2 │              │  复杂任务 3-5 │
              │ 国内模型 API  │              │  Claude API  │
              │ GLM/Kimi等   │              │ Sonnet/Opus  │
              └──────────────┘              └──────────────┘
```

### 🌟 推荐配置策略

| 难度等级 | 任务类型 | 推荐模型 | 成本 |
|---------|---------|---------|------|
| **1** | 简单查询、语法检查 | 智谱 GLM / Kimi | 💰 |
| **2** | 代码解释、文档编写 | 智谱 GLM / MiniMax | 💰 |
| **3** | 代码重构、调试 | Claude 3.5 Sonnet | 💰💰 |
| **4** | 架构设计、算法实现 | Claude 3.5 Sonnet | 💰💰 |
| **5** | 复杂系统设计 | Claude 3 Opus | 💰💰💰 |

> 💡 **提示**: 根据统计，约 60% 的 Claude Code 任务为简单任务（难度 1-2），使用此配置可节省约 50% 的 API 成本！

## ✨ 功能特性

### 核心功能

- **🧠 智能任务评估**
  - AI 驱动的复杂度分析（1-5 级）
  - 自动提取任务意图
  - 上下文感知评估
  - 可自定义评估提示词

- **🔀 灵活的端点支持**
  - ✅ Anthropic 官方 API
  - ✅ 第三方 Anthropic API 中转站
  - ✅ 智谱 GLM (ChatGLM)
  - ✅ 月之暗面 Kimi
  - ✅ MiniMax
  - ✅ 任何兼容 Claude API 格式的服务

- **💸 成本优化策略**
  - 基于难度的自动路由
  - 可自定义难度-模型映射
  - 详细的请求日志和分析
  - 成本节省统计

- **🖥️ macOS 原生应用**
  - 菜单栏常驻图标
  - 一键启动/停止代理
  - 实时状态显示
  - 图形化配置界面
  - 请求历史查看

- **⚡ 性能特性**
  - 服务预热机制
  - 完整的 SSE 流式支持
  - 可配置超时（支持长任务）
  - 连接池和复用
  - 并发请求处理

- **🔧 开发者友好**
  - 结构化日志（Zap）
  - YAML 配置管理（Viper）
  - 完整的测试覆盖
  - RESTful 状态端点
  - 详细的错误信息

## 📋 系统要求

| 组件 | 要求 |
|------|------|
| **操作系统** | macOS 10.15 (Catalina) 或更高版本 |
| **架构** | Apple Silicon (M1/M2/M3) 或 Intel |
| **开发环境** | Go 1.21+ (仅开发时需要) |
| **内存** | 512MB 以上 |
| **磁盘空间** | 100MB |
| **网络** | 需要访问 AI API 服务 |

## 📥 安装指南

### 方式一：预编译版本（推荐）

**下载最新版本：**

1. 访问 [Releases 页面](https://github.com/Yi-Lyu/cce/releases)
2. 根据你的 Mac 架构下载对应的 DMG：
   - **Apple Silicon (M1/M2/M3)**: `CCE-vX.X.X-arm64.dmg`
   - **Intel Mac**: `CCE-vX.X.X-amd64.dmg`
   - **通用版**: `CCE-vX.X.X-universal.dmg` (两种架构都支持)
3. 打开 DMG，将 CCE 拖入应用程序文件夹
4. 按照下方的首次启动指南操作

### 方式二：从源码编译

```bash
# 克隆仓库
git clone https://github.com/Yi-Lyu/cce.git
cd cce

# 快速构建（自动检测架构）
./build-mac-release-opensource.sh 1.0.0

# 或者分别构建各组件
cd proxy && make build          # 构建代理服务
cd ../cce-client && make package  # 构建 GUI 客户端
```

### ⚠️ 重要：首次启动安全设置

**CCE 是开源软件，未经过 Apple 签名。** macOS 会在首次启动时显示安全警告，这是正常现象。

**信任应用的方法（选择其一）：**

#### 方法 1：右键打开（最简单）
1. 打开应用程序文件夹
2. **右键点击**（或 Control + 点击）CCE.app
3. 选择菜单中的 **"打开"**
4. 在弹出的对话框中点击 **"打开"**

#### 方法 2：系统设置（macOS 13+）
1. 尝试正常打开 CCE（会被阻止）
2. 前往 **系统设置 → 隐私与安全性**
3. 找到关于 CCE 被阻止的提示
4. 点击 **"仍要打开"**

#### 方法 3：终端命令（技术用户）
```bash
# 移除隔离属性
xattr -cr /Applications/CCE.app
```

完成以上任一操作后，CCE 将可以正常启动，后续无需重复操作。

## 🚀 快速开始

### 第 1 步：配置 API 密钥

启动 CCE 后，编辑配置文件设置 API 密钥：

```yaml
# 编辑 proxy/configs/config.yaml
services:
  # 评估器：用于分析任务复杂度
  - id: "haiku"
    name: "Claude 3 Haiku"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "your-haiku-api-key"
    role: "evaluator"
    supports_thinking: true

  # 执行器：简单任务（推荐国内模型）
  - id: "glm"
    name: "智谱 GLM-4"
    url: "https://open.bigmodel.cn/api/paas/v4/chat/completions"
    api_key: "your-glm-api-key"
    role: "executor"
    supports_thinking: false

  - id: "kimi"
    name: "月之暗面 Kimi"
    url: "https://api.moonshot.cn/v1/chat/completions"
    api_key: "your-kimi-api-key"
    role: "executor"
    supports_thinking: false

  # 执行器：复杂任务（推荐 Claude）
  - id: "sonnet"
    name: "Claude 3.5 Sonnet"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "your-sonnet-api-key"
    role: "executor"
    supports_thinking: true

  - id: "opus"
    name: "Claude 3 Opus"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "your-opus-api-key"
    role: "executor"
    supports_thinking: true
```

### 第 2 步：配置难度映射

根据你的需求和预算设置任务路由策略：

```yaml
# 推荐配置：成本优化型
difficulty_mapping:
  1: "glm"      # 简单任务 → 智谱 GLM
  2: "kimi"     # 简单任务 → Kimi
  3: "sonnet"   # 中等任务 → Claude Sonnet
  4: "sonnet"   # 复杂任务 → Claude Sonnet
  5: "opus"     # 超复杂任务 → Claude Opus

# 或者：性能优先型（全部使用 Claude）
# difficulty_mapping:
#   1: "haiku"
#   2: "haiku"
#   3: "sonnet"
#   4: "sonnet"
#   5: "opus"

# 或者：极致省钱型（尽量使用国内模型）
# difficulty_mapping:
#   1: "glm"
#   2: "glm"
#   3: "kimi"
#   4: "sonnet"
#   5: "sonnet"
```

### 第 3 步：启动代理服务

**通过 GUI：**
- 点击菜单栏的 CCE 图标
- 选择 "启动代理"
- 图标变绿表示服务已启动

**通过命令行：**
```bash
cd proxy
./claude-proxy -config=configs/config.yaml
```

### 第 4 步：配置 Claude Code

在 Claude Code 设置中将 API 端点指向 CCE：

```bash
# 设置代理端点
export CLAUDE_API_BASE_URL="http://127.0.0.1:27015"

# 或在应用程序中配置
# API 端点: http://127.0.0.1:27015/v1/messages
```

### 第 5 步：验证运行

```bash
# 检查代理状态
curl http://127.0.0.1:27015/status

# 发送测试请求
curl -X POST http://127.0.0.1:27015/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 🎉 完成！

现在 Claude Code 的所有请求都会通过 CCE 智能路由到最合适的模型！

## ⚙️ 配置说明

### 配置文件结构

CCE 使用 YAML 格式配置。从示例开始：

```bash
cp proxy/configs/config.example.yaml proxy/configs/config.yaml
```

### 主要配置项

#### 1. 代理服务设置

```yaml
proxy:
  port: 27015                 # 代理监听端口
  request_timeout: 1800       # 请求超时（秒）
  read_timeout: 1900          # 读取超时（秒）
  write_timeout: 1900         # 写入超时（秒）
  evaluator_timeout: 30       # 评估器超时（秒）
```

#### 2. 服务端点定义

```yaml
services:
  # 每个服务包含以下字段：
  - id: "service-id"              # 唯一标识符
    name: "服务显示名称"            # 友好名称
    url: "https://api.example.com"  # API 端点 URL
    api_key: "${API_KEY}"          # API 密钥（支持环境变量）
    role: "executor"               # 角色：evaluator 或 executor
    supports_thinking: true        # 是否支持 thinking 字段
```

#### 3. 难度映射配置

```yaml
difficulty_mapping:
  1: "service-id-for-level-1"
  2: "service-id-for-level-2"
  3: "service-id-for-level-3"
  4: "service-id-for-level-4"
  5: "service-id-for-level-5"
```

#### 4. 评估器配置

```yaml
evaluator:
  model: "claude-3-haiku-20240307"
  max_tokens: 100
  temperature: 0
  max_history_rounds: 3         # 保留的历史对话轮数
  prompt_template: |            # 评估提示词模板
    分析以下任务的复杂度，返回 1-5 的数字：
    {{.CurrentTask}}

    历史上下文：{{.HistoryContext}}

    只返回一个数字 1-5。
```

#### 5. 日志配置

```yaml
logging:
  level: "info"              # 日志级别：debug, info, warn, error
  output: "logs"             # 日志目录
  max_size: 100              # 单个日志文件最大大小（MB）
  max_backups: 10            # 保留的日志文件数
  max_age: 30                # 日志保留天数
```

### 环境变量支持

所有配置值都支持环境变量替换：

```bash
# 设置 API 密钥
export HAIKU_API_KEY="your-haiku-key"
export GLM_API_KEY="your-glm-key"
export KIMI_API_KEY="your-kimi-key"
export SONNET_API_KEY="your-sonnet-key"

# 覆盖配置
export CLAUDE_PROXY_PORT=8080
export CLAUDE_PROXY_REQUEST_TIMEOUT=3600
```

### 国内模型接入示例

#### 智谱 GLM

```yaml
- id: "glm"
  name: "智谱 GLM-4"
  url: "https://open.bigmodel.cn/api/paas/v4/chat/completions"
  api_key: "${GLM_API_KEY}"
  role: "executor"
  supports_thinking: false
```

#### 月之暗面 Kimi

```yaml
- id: "kimi"
  name: "Kimi"
  url: "https://api.moonshot.cn/v1/chat/completions"
  api_key: "${KIMI_API_KEY}"
  role: "executor"
  supports_thinking: false
```

#### MiniMax

```yaml
- id: "minimax"
  name: "MiniMax"
  url: "https://api.minimax.chat/v1/text/chatcompletion"
  api_key: "${MINIMAX_API_KEY}"
  role: "executor"
  supports_thinking: false
```

## 🛠️ 开发指南

### 项目结构

```
cce/
├── proxy/                      # Go 代理服务
│   ├── cmd/                   # 主入口 (main.go)
│   ├── internal/              # 内部包
│   │   ├── proxy/            # 代理服务器 & 处理器
│   │   ├── evaluator/        # 任务复杂度评估器
│   │   ├── config/           # 配置管理
│   │   └── models/           # 数据结构
│   ├── configs/              # 配置文件
│   │   ├── config.example.yaml
│   │   └── config.test.yaml
│   └── Makefile              # 构建命令
│
├── cce-client/               # macOS GUI 应用
│   ├── cmd/                  # 主入口
│   ├── internal/             # 客户端逻辑
│   │   ├── app/             # 应用核心
│   │   ├── ui/              # 用户界面
│   │   └── proxy/           # 代理管理
│   ├── resources/            # 应用图标和资源
│   └── Makefile              # 构建命令
│
├── scripts/                  # 构建和发布自动化脚本
│   ├── sign-app.sh          # 代码签名
│   ├── create-dmg.sh        # DMG 创建
│   └── generate-release-notes.sh
│
├── docs/                     # 文档目录
│   └── README_EN.md         # 英文文档
│
├── build-mac-release.sh          # 完整发布脚本
├── build-mac-release-opensource.sh  # 简化构建脚本
├── CLAUDE.md                 # 开发者指南
└── LICENSE                   # MIT 许可证
```

### 技术栈

| 组件 | 技术 |
|------|------|
| **代理服务器** | Go 1.21+, Gin, Viper, Zap |
| **GUI 客户端** | Go, Fyne (原生 macOS UI) |
| **API 兼容** | Claude API v1 格式 |
| **配置** | YAML, 环境变量 |
| **日志** | 结构化日志 (Zap) |
| **构建** | Make, Shell 脚本 |

### 开发命令

#### 代理服务器

```bash
cd proxy

# 开发
make run                # 使用默认配置运行
make run-test           # 使用测试配置运行
make build              # 构建二进制文件
make test               # 运行测试（带覆盖率）

# 测试
make mock               # 启动模拟评估器（端口 8081）
make test-api           # 运行集成测试

# 工具
make fmt                # 格式化代码
make lint               # 运行 linter
make clean              # 清理构建产物
```

#### GUI 客户端

```bash
cd cce-client

# 开发
make run                # 运行应用
make build              # 构建二进制文件
make package            # 创建 .app 包
make test               # 运行测试

# 工具
make clean              # 清理构建产物
```

### 构建发布版本

#### 快速构建（开源）

```bash
# 自动检测架构并构建
./build-mac-release-opensource.sh 1.0.0

# 输出: releases/v1.0.0/CCE-v1.0.0-{arch}.dmg
```

#### 高级构建（带签名）

```bash
# 未签名构建
./build-mac-release.sh --version 1.0.0

# 已签名构建（需要 Apple Developer 证书）
./build-mac-release.sh \
  --version 1.0.0 \
  --sign \
  --developer-id "Developer ID Application: Your Name (TEAM_ID)"

# 完整发布（包含公证）
./build-mac-release.sh \
  --version 1.0.0 \
  --sign \
  --notarize \
  --developer-id "Developer ID Application: Your Name" \
  --team-id "TEAM_ID" \
  --apple-id "you@email.com" \
  --app-password "app-specific-password"
```

## 🧪 测试

### 单元测试

```bash
# 测试代理服务器
cd proxy
make test                    # 运行所有测试（带覆盖率）
go test ./internal/proxy -v  # 测试特定包

# 测试 GUI 客户端
cd cce-client
go test ./...               # 运行所有测试
```

### 集成测试

```bash
# 终端 1：启动模拟评估器
cd proxy
make mock

# 终端 2：运行集成测试
make test-api
```

### 手动测试

```bash
# 1. 启动代理（测试配置）
cd proxy
make run-test

# 2. 发送测试请求
curl -X POST http://127.0.0.1:27015/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 100,
    "messages": [{
      "role": "user",
      "content": "你好，Claude！"
    }]
  }'

# 3. 检查状态
curl http://127.0.0.1:27015/status
```

## 📦 部署

### 本地部署

1. **配置** API 密钥到 `proxy/configs/config.yaml`
2. **构建** 发布版：`./build-mac-release-opensource.sh 1.0.0`
3. **安装** `releases/v1.0.0/` 中的 DMG
4. **启动** 应用程序文件夹中的 CCE

### GitHub Releases

#### 手动发布

```bash
# 1. 打标签
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. 构建发布产物
./build-mac-release-opensource.sh 1.0.0

# 3. 生成发布说明
./scripts/generate-release-notes.sh \
  --version 1.0.0 \
  --output release-notes.md

# 4. 创建 GitHub Release 并上传 DMG 文件
# 前往: https://github.com/Yi-Lyu/cce/releases/new
```

#### 自动发布（GitHub Actions）

创建 `.github/workflows/release.yml`：

```yaml
name: 构建和发布

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-release:
    runs-on: macos-latest
    steps:
      - name: 检出代码
        uses: actions/checkout@v4

      - name: 设置 Go 环境
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: 构建发布版
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          ./build-mac-release-opensource.sh $VERSION

      - name: 创建 Release
        uses: softprops/action-gh-release@v1
        with:
          files: releases/**/*.dmg
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 🔧 故障排除

### 应用问题

#### "CCE 无法打开，因为来自身份不明的开发者"

**解决方案：**
- 右键点击 CCE.app，选择"打开"
- 或：系统设置 → 隐私与安全性 → 点击"仍要打开"
- 或：终端运行 `xattr -cr /Applications/CCE.app`

#### 应用启动时崩溃

**解决方案：**
1. 查看日志：`~/Library/Logs/CCE/`
2. 删除损坏的配置：`rm ~/Library/Application\ Support/CCE/config.yaml`
3. 重新安装应用

### 代理问题

#### "连接被拒绝"错误

**检查清单：**
- [ ] 代理正在运行（检查菜单栏图标）
- [ ] 端口 27015 未被占用：`lsof -i :27015`
- [ ] 防火墙未阻止代理
- [ ] 查看日志：`tail -f proxy/logs/*/proxy.log`

#### 请求超时

**解决方案：**
- 增加配置中的超时时间：
  ```yaml
  proxy:
    request_timeout: 3600  # 1 小时
  ```
- 检查 API 密钥是否有效
- 验证网络连接到 AI API 服务

#### "未找到服务"错误

**解决方案：**
- 验证难度映射中的服务 ID
- 检查 `config.yaml` 中的服务配置
- 确保所有必需的 API 密钥已设置

### 配置问题

#### API 密钥无效

**检查清单：**
- [ ] 密钥在 YAML 中正确引用
- [ ] 环境变量已导出
- [ ] 密钥字符串中没有多余空格
- [ ] 密钥具有正确的权限

#### 评估器总是返回相同难度

**解决方案：**
- 检查评估器提示词模板
- 查看评估器日志：`proxy/logs/*/evaluator.log`
- 增加 `max_history_rounds` 以获取更多上下文
- 验证评估器 API 密钥有效

### 获取帮助

1. **查看文档：**
   - `CLAUDE.md` - 项目概览
   - `proxy/README.md` - 代理服务详情
   - `cce-client/README.md` - 客户端详情

2. **启用调试日志：**
   ```yaml
   logging:
     level: "debug"
   ```

3. **提交 Issue：**
   - 包含 `proxy/logs/` 中的日志
   - 描述复现步骤
   - 说明你的 macOS 和 Go 版本

## 📚 文档

- **[CLAUDE.md](CLAUDE.md)** - 开发者综合指南
- **[proxy/README.md](proxy/README.md)** - 代理服务器文档
- **[cce-client/README.md](cce-client/README.md)** - GUI 客户端文档
- **[configs/config.example.yaml](proxy/configs/config.example.yaml)** - 配置参考
- **[📖 English Documentation](docs/README_EN.md)** - 英文文档

## 🤝 贡献指南

欢迎贡献！以下是参与项目的方法：

### 开发环境设置

1. **Fork 并克隆：**
   ```bash
   git clone https://github.com/Yi-Lyu/cce.git
   cd cce
   ```

2. **安装依赖：**
   ```bash
   cd proxy && make deps
   cd ../cce-client && go mod download
   ```

3. **创建功能分支：**
   ```bash
   git checkout -b feature/amazing-feature
   ```

### 开发流程

1. **进行更改**
   - 遵循 Go 最佳实践
   - 为新功能添加测试
   - 更新文档

2. **测试更改：**
   ```bash
   cd proxy && make test
   cd ../cce-client && make test
   ```

3. **格式化和 lint：**
   ```bash
   cd proxy && make fmt && make lint
   ```

4. **提交并推送：**
   ```bash
   git commit -m 'feat: 添加某个很棒的功能'
   git push origin feature/amazing-feature
   ```

5. **提交 Pull Request**

### 提交信息规范

我们遵循 [约定式提交](https://www.conventionalcommits.org/zh-hans/)：

- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档变更
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建/工具相关

### 代码风格

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化
- 提交前运行 `golangci-lint`
- 编写清晰的注释

## 📄 许可证

本项目采用 **MIT 许可证** 开源，详见 [LICENSE](LICENSE) 文件。

```
MIT License

Copyright (c) 2025 Ethan (Yi-Lyu)

特此免费授予任何获得本软件副本和相关文档文件（"软件"）的人不受限制地处置该软件的权利，
包括不受限制地使用、复制、修改、合并、发布、分发、再许可和/或出售该软件副本，
以及允许拥有软件副本的人员进行上述行为，但须符合以下条件：

上述版权声明和本许可声明应包含在该软件的所有副本或实质成分中。

本软件按"原样"提供，不提供任何形式的明示或暗示的保证，包括但不限于对适销性、
特定用途适用性和非侵权性的保证。在任何情况下，作者或版权持有人都不对任何索赔、
损害或其他责任负责，无论这些追责来自合同、侵权或其它行为中，还是产生于、
源于或有关于本软件以及本软件的使用或其它处置。
```

## 🙏 致谢

本项目使用了以下优秀的开源工具构建：

- **[Gin](https://github.com/gin-gonic/gin)** - 高性能 HTTP Web 框架
- **[Fyne](https://fyne.io/)** - Go 跨平台 GUI 工具包
- **[Viper](https://github.com/spf13/viper)** - 配置管理
- **[Zap](https://github.com/uber-go/zap)** - 高性能结构化日志
- **[Claude API](https://anthropic.com/claude)** - AI 语言模型

特别感谢所有为本项目做出贡献的开发者！

## 💬 支持与社区

- **Issues:** [GitHub Issues](https://github.com/Yi-Lyu/cce/issues)
- **Discussions:** [GitHub Discussions](https://github.com/Yi-Lyu/cce/discussions)
- **Documentation:** [项目 Wiki](https://github.com/Yi-Lyu/cce/wiki)

## 🗺️ 路线图

- [ ] Windows 平台支持
- [ ] Linux 平台支持
- [ ] Web 管理控制台
- [ ] 高级分析和指标
- [ ] 自定义评估器插件
- [ ] Docker 部署选项
- [ ] Kubernetes 支持
- [ ] 更多国内模型支持
  - [ ] 阿里通义千问
  - [ ] 百度文心一言
  - [ ] 腾讯混元

## 🌟 Star History

如果这个项目对你有帮助，请给我们一个 Star ⭐️

[![Star History Chart](https://api.star-history.com/svg?repos=Yi-Lyu/cce&type=Date)](https://star-history.com/#Yi-Lyu/cce&Date)

---

<div align="center">

**用 ❤️ 构建 by Ethan**

[⭐ Star on GitHub](https://github.com/Yi-Lyu/cce) • [🐛 报告 Bug](https://github.com/Yi-Lyu/cce/issues) • [💡 功能建议](https://github.com/Yi-Lyu/cce/issues)

</div>
