# 开发环境配置完成

本文档记录了开发环境的配置完成状态和使用说明。

## ✅ 已完成配置

### 1. 代理服务配置 (`../proxy/configs/config.yaml`)

已创建完整的开发环境配置：

- **超时配置**：所有超时设置为 30 分钟，支持长时间任务
- **服务配置**：
  - **Evaluator（决策者）**：智谱清言 API
  - **Haiku 执行器**：智谱清言 API（难度 1-2）
  - **Sonnet 执行器**：智谱清言 API（难度 3-4）
  - **Opus 执行器**：qcode API（难度 5）
- **Thinking 模式**：
  - 智谱清言服务：`supports_thinking: false`
  - qcode 服务：`supports_thinking: true`
- **日志级别**：`debug`（开发模式）
- **Fallback**：已启用 `evaluator_fallback: true`

### 2. macOS 客户端打包（隐藏 Dock 图标）

已配置 `LSUIElement: true` 来隐藏 Dock 图标，只在菜单栏显示：

- **Info.plist**：包含 `LSUIElement` 配置
- **打包方式**：手动创建 .app 包结构（避免 fyne package 兼容性问题）
- **Makefile**：已更新 `make package` 命令

## 🚀 如何使用

### 启动代理服务（开发模式）

```bash
cd /Users/ethan/code/Claude-Code-Exchange/proxy
make run  # 使用 configs/config.yaml
```

### 运行 macOS 客户端

#### 方式 1：开发模式（会显示 Dock 图标）
```bash
cd /Users/ethan/code/Claude-Code-Exchange/cce-client
make run
```

#### 方式 2：.app 包模式（隐藏 Dock 图标）
```bash
cd /Users/ethan/code/Claude-Code-Exchange/cce-client
make package  # 打包成 CCE.app
open CCE.app  # 运行应用，只显示菜单栏图标
```

## 🧪 验证代理功能

### 1. 检查代理服务状态

```bash
# 启动代理服务
cd /Users/ethan/code/Claude-Code-Exchange/proxy
make run

# 在另一个终端检查状态
curl http://127.0.0.1:27015/status
```

### 2. 测试简单请求（难度 1-2）

应该路由到智谱清言的 haiku-service：

```bash
curl -X POST http://127.0.0.1:27015/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: test-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "messages": [{
      "role": "user",
      "content": "Say hello"
    }]
  }'
```

### 3. 测试复杂请求（难度 5）

应该路由到 qcode 的 opus-service：

```bash
curl -X POST http://127.0.0.1:27015/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: test-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 2000,
    "messages": [{
      "role": "user",
      "content": "Design a complete microservices architecture for an e-commerce platform with 10+ services"
    }]
  }'
```

### 4. 检查日志

```bash
# 实时查看代理服务日志
tail -f /Users/ethan/code/Claude-Code-Exchange/proxy/logs/$(date +%Y-%m-%d)/claude-proxy-$(date +%Y-%m-%d).log
```

## 📊 配置说明

### 难度映射规则

| 难度等级 | 任务类型 | 路由服务 | API 提供商 |
|---------|---------|---------|-----------|
| 1-2 | 简单查询、基础问答 | haiku-service | 智谱清言 |
| 3-4 | 代码编写、数据分析 | sonnet-service | 智谱清言 |
| 5 | 系统设计、大型重构 | opus-service | qcode |

### Thinking 模式兼容性

- **智谱清言 API**：不支持 `thinking` 字段，代理会自动移除
- **qcode API**：支持 `thinking` 字段，完整转发

### 超时配置

所有超时均设置为开发友好的值：

- `read_timeout`: 1800 秒（30 分钟）
- `write_timeout`: 1800 秒（30 分钟）
- `request_timeout`: 1800 秒（30 分钟）
- `evaluator_timeout`: 30 秒

## ⚠️ 注意事项

1. **首次启动**：确保代理服务先启动，然后再启动客户端
2. **Dock 图标**：只有使用 `.app` 包启动才会隐藏 Dock 图标
3. **日志级别**：开发环境使用 `debug` 级别，生产环境建议改为 `info`
4. **API 密钥**：配置中已包含测试 API 密钥，请勿提交到公开仓库

## 🔧 故障排查

### 代理服务无法启动

```bash
# 检查端口是否被占用
lsof -i :27015

# 杀死占用端口的进程
kill -9 <PID>
```

### 客户端无法连接代理

1. 确认代理服务正在运行：`curl http://127.0.0.1:27015/health`
2. 检查防火墙设置
3. 查看代理服务日志

### Dock 图标仍然显示

确保使用 `.app` 包启动，而不是直接运行二进制文件：

```bash
# ❌ 错误方式（会显示 Dock 图标）
./build/cce-client

# ✅ 正确方式（隐藏 Dock 图标）
open CCE.app
```

## 📝 开发配置文件位置

- 代理配置：`/Users/ethan/code/Claude-Code-Exchange/proxy/configs/config.yaml`
- 客户端配置：`/Users/ethan/code/Claude-Code-Exchange/cce-client/FyneApp.toml`
- 应用包：`/Users/ethan/code/Claude-Code-Exchange/cce-client/CCE.app`

## 🎯 下一步

配置已完成，现在可以：

1. 启动代理服务验证路由功能
2. 运行 `.app` 包验证界面和 Dock 隐藏
3. 使用 Claude Code 连接到代理（`http://127.0.0.1:27015`）进行端到端测试
