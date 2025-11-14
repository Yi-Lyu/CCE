# Claude 智能代理服务

一个智能的 HTTP/HTTPS 代理服务，专门用于 Claude API 请求的智能路由。通过决策者服务评估任务难度，自动将请求路由到最合适的服务端点，优化成本和性能。

## 核心特性

- 🧠 **智能路由**：基于任务复杂度（1-5级）自动选择最合适的服务
- 💰 **成本优化**：简单任务使用便宜的服务，复杂任务使用高级服务
- 🔄 **流式支持**：完整支持 Claude API 的流式响应
- 📊 **上下文感知**：基于用户历史请求优化决策
- 🔧 **灵活配置**：支持多服务配置、故障转移等高级特性
- 📝 **请求日志**：详细的请求日志用于分析和调试

## 工作原理

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│ Claude Code │────▶│   代理服务   │────▶│  决策者服务   │
└─────────────┘     └─────────────┘     └──────────────┘
                           │                      │
                           │                      ▼
                           │              难度评估 (1-5)
                           │                      │
                           ▼                      │
                    ┌──────────────┐             │
                    │  目标服务     │◀────────────┘
                    └──────────────┘
```

1. Claude Code 发送请求到代理服务（127.0.0.1:27015）
2. 代理提取用户上下文，发送到决策者服务评估难度
3. 根据难度等级映射，转发请求到对应的目标服务
4. 实时转发响应（支持流式）回 Claude Code

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- 至少一个可用的 Claude API 服务端点

### 安装

```bash
# 克隆仓库并切换到代理分支
git clone https://github.com/ethan/Claude-Code-Exchange.git
cd Claude-Code-Exchange
git checkout feature/intelligent-proxy
cd proxy

# 安装依赖
go mod download

# 编译
go build -o claude-proxy cmd/main.go
```

### 配置

1. 复制示例配置文件：
```bash
cp configs/config.example.yaml configs/config.yaml
```

2. 编辑 `configs/config.yaml`，配置您的服务：

```yaml
services:
  # 决策者服务（必需）
  - id: "evaluator-1"
    name: "决策者服务"
    url: "https://your-evaluator-api.com/v1/messages"
    api_key: "your_evaluator_api_key"
    role: "evaluator"
  
  # 执行服务（根据需要配置多个）
  - id: "simple-service"
    name: "简单任务服务"
    url: "https://simple-api.com/v1/messages"
    api_key: "your_simple_api_key"
    role: "executor"

# 难度映射
difficulty_mapping:
  "1": "simple-service"
  "2": "simple-service"
  "3": "medium-service"
  "4": "complex-service"
  "5": "complex-service"
```

### 运行

```bash
# 使用默认配置运行
./claude-proxy

# 指定配置文件
./claude-proxy -config=/path/to/config.yaml

# 查看版本
./claude-proxy -version
```

### 配置 Claude Code

在 Claude Code 中，将 API 端点设置为：
```
http://127.0.0.1:27015/api/v1/messages
```

## 配置说明

### 服务配置

每个服务需要配置：
- `id`：唯一标识符
- `name`：服务名称（用于日志）
- `url`：完整的 API 端点 URL
- `api_key`：Bearer token 认证密钥
- `role`：服务角色（`evaluator` 或 `executor`）

### 功能开关

- `evaluator_fallback`：决策者服务不可用时使用默认难度（默认：false）
- `service_auto_switch`：目标服务不可用时自动切换（默认：false）
- `request_logging`：记录详细请求日志（默认：true）

## 决策者服务接口

决策者服务需要实现以下接口：

**请求：** `POST /evaluate-difficulty`
```json
{
  "original_request": {
    "model": "claude-3-opus-20240229",
    "messages": [...],
    "metadata": {...}
  },
  "user_context": {
    "user_id": "...",
    "session_id": "...",
    "request_history": [...]
  }
}
```

**响应：**
```json
{
  "difficulty_level": 3,  // 1-5
  "reasoning": "包含复杂的多步骤任务"
}
```

## 测试

### 运行单元测试

```bash
go test ./...
```

### 运行模拟服务器

```bash
# 启动模拟决策者服务
go run tests/mock_evaluator_server.go

# 在另一个终端运行测试脚本
./tests/test_proxy.sh
```

## 监控和日志

- 健康检查：`GET http://127.0.0.1:27015/health`
- 状态信息：`GET http://127.0.0.1:27015/status`
- 日志文件：`./logs/claude-proxy-YYYY-MM-DD.log`

## 开发路线图

- [ ] 支持更多认证方式
- [ ] 添加请求缓存机制
- [ ] 实现服务健康检查
- [ ] 添加 Prometheus 监控指标
- [ ] 开发 Mac 客户端（Menu Bar App）

## 注意事项

1. 确保所有配置的服务都兼容 Claude API 协议
2. 决策者服务的响应时间会影响整体延迟
3. 建议在生产环境中启用请求日志以便故障排查
4. 流式响应需要客户端支持 Server-Sent Events

## 故障排查

### 代理无法启动
- 检查端口 27015 是否被占用
- 验证配置文件格式是否正确
- 确保至少配置了一个决策者服务

### 请求失败
- 检查目标服务的 URL 和 API Key 是否正确
- 查看日志文件了解详细错误信息
- 使用状态端点验证服务配置

### 决策者服务超时
- 增加决策者服务的超时时间
- 启用 `evaluator_fallback` 使用默认难度

## 许可证

MIT License
