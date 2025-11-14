# Thinking Mode 配置说明

## 什么是 Thinking Mode？

Thinking Mode（思考模式）是 Claude Opus 4 的高级特性，允许模型在响应前进行更深入的思考和推理。当启用时，请求会包含 `thinking` 字段。

**示例请求**：
```json
{
  "model": "claude-opus-4-20250805",
  "messages": [...],
  "thinking": {
    "type": "enabled",
    "budget_tokens": 10000
  }
}
```

## 兼容性问题

**官方 Anthropic API** ✅ 完全支持 thinking 模式

**第三方 Claude 兼容 API** ❌ 通常不支持，会返回 400 错误：
- 智谱清言 (glm)
- 通义千问 (qwen)
- 其他第三方服务

**错误示例**：
```
400 Bad Request: messages.3.content.0.type: Expected `thinking` or `redacted_thinking`,
but found `text`. When `thinking` is enabled, a final `assistant` message must start
with a thinking block
```

## 解决方案：Per-Service 配置

代理服务支持为每个服务单独配置是否支持 thinking 模式。

### 配置方法

在 `config.yaml` 中，为每个服务添加 `supports_thinking` 字段：

```yaml
services:
  # 官方 Anthropic API - 支持 thinking
  - id: "official-opus"
    name: "Official Opus API"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "sk-ant-xxx"
    role: "executor"
    supports_thinking: true   # 默认值，可以省略

  # 第三方 API - 不支持 thinking
  - id: "zhipu-glm"
    name: "智谱清言"
    url: "https://api.example.com/v1/messages"
    api_key: "xxx.yyy"
    role: "executor"
    supports_thinking: false  # ⚠️ 必须显式设置为 false
```

### 配置规则

| 配置值 | 行为 | 适用场景 |
|--------|------|----------|
| `true` | 保留 `thinking` 字段 | 官方 Anthropic API |
| `false` | 移除 `thinking` 字段 | 第三方 Claude 兼容 API |
| 未配置 | 默认为 `true` | 向后兼容，安全默认值 |

### 推荐配置

```yaml
services:
  # ✅ 官方 API - 可以省略 supports_thinking（默认 true）
  - id: "anthropic-haiku"
    url: "https://api.anthropic.com/v1/messages"
    role: "executor"
    # supports_thinking: true  # 默认值，可省略

  # ✅ 第三方 API - 必须显式设置为 false
  - id: "third-party"
    url: "https://third-party-api.com/v1/messages"
    role: "executor"
    supports_thinking: false  # ⚠️ 重要：必须设置
```

## 工作原理

### 1. 配置加载
```go
// config/config.go
for i := range Cfg.Services {
    if !viper.IsSet(fmt.Sprintf("services.%d.supports_thinking", i)) {
        Cfg.Services[i].SupportsThinking = true  // 默认支持
    }
}
```

### 2. 请求清理
```go
// handler.go:createTargetRequest()
if !service.SupportsThinking {
    // 移除 thinking 字段
    sanitizedBody := sanitizeRequestForExecutor(body)
}
```

### 3. 日志记录
当 thinking 字段被移除时，会记录 debug 日志：
```
[DEBUG] 已清理请求体，移除了thinking字段 service=zhipu-glm
```

## 使用示例

### 场景 1：混合使用官方和第三方 API

```yaml
services:
  # Evaluator 使用官方 API（支持 thinking）
  - id: "evaluator"
    url: "https://api.anthropic.com/v1/messages"
    role: "evaluator"
    supports_thinking: true

  # 简单任务使用第三方 API（不支持 thinking）
  - id: "zhipu-haiku"
    url: "https://api.example.com/v1/messages"
    role: "executor"
    supports_thinking: false  # 自动移除 thinking

  # 复杂任务使用官方 API（支持 thinking）
  - id: "official-opus"
    url: "https://api.anthropic.com/v1/messages"
    role: "executor"
    supports_thinking: true

difficulty_mapping:
  "1": "zhipu-haiku"      # 简单任务 → 第三方（无 thinking）
  "5": "official-opus"    # 复杂任务 → 官方（有 thinking）
```

### 场景 2：全部使用第三方 API

```yaml
services:
  - id: "evaluator"
    url: "https://api.example.com/v1/messages"
    role: "evaluator"
    supports_thinking: false  # 第三方 API

  - id: "executor-1"
    url: "https://api.example.com/v1/messages"
    role: "executor"
    supports_thinking: false  # 第三方 API

  - id: "executor-2"
    url: "https://dashscope.aliyuncs.com/api/anthropic/v1/messages"
    role: "executor"
    supports_thinking: false  # 通义千问
```

### 场景 3：全部使用官方 API

```yaml
services:
  - id: "evaluator"
    url: "https://api.anthropic.com/v1/messages"
    role: "evaluator"
    # supports_thinking: true  # 默认，可省略

  - id: "haiku"
    url: "https://api.anthropic.com/v1/messages"
    role: "executor"
    # supports_thinking: true  # 默认，可省略
```

## 验证配置

### 1. 检查配置加载
```bash
cd proxy
./build/claude-proxy -config=configs/config.yaml
```

查看启动日志，确认服务配置正确加载。

### 2. 观察日志
```bash
tail -f logs/*/claude-proxy-*.log | grep -E "清理|thinking"
```

如果看到 "已清理请求体，移除了thinking字段"，说明配置生效。

### 3. 测试请求
使用 Claude Code 发送请求，观察：
- ✅ 官方 API：正常工作，保留 thinking
- ✅ 第三方 API：正常工作，移除 thinking
- ❌ 未配置：可能出现 400 错误

## 故障排查

### 问题：第三方 API 返回 400 错误

**症状**：
```
400 Bad Request: Expected `thinking` or `redacted_thinking`, but found `text`
```

**原因**：未设置 `supports_thinking: false`

**解决**：
```yaml
- id: "problematic-service"
  url: "https://third-party-api.com/v1/messages"
  role: "executor"
  supports_thinking: false  # 添加这一行
```

### 问题：日志中没有看到 "清理请求体" 信息

**检查清单**：
1. 确认日志级别设置为 `debug`：
   ```yaml
   logging:
     level: "debug"
   ```
2. 确认服务的 `supports_thinking` 为 `false`
3. 确认该服务确实被使用（检查难度映射）

### 问题：想要对所有服务禁用 thinking

**不推荐**：这种做法会影响官方 API 的高级特性

**替代方案**：只对第三方 API 禁用：
```yaml
services:
  - id: "official"
    supports_thinking: true   # 保留官方特性
  - id: "third-party"
    supports_thinking: false  # 仅第三方禁用
```

## 最佳实践

1. **明确配置**：即使是默认值也建议写出来，便于维护
   ```yaml
   supports_thinking: true   # 明确表示支持
   ```

2. **注释说明**：添加注释说明 API 类型
   ```yaml
   supports_thinking: false  # 智谱清言不支持 thinking
   ```

3. **测试验证**：新增服务后，先用小请求测试
   ```bash
   curl -X POST http://127.0.0.1:27015/v1/messages \
     -H "Content-Type: application/json" \
     -d '{"model":"test","messages":[{"role":"user","content":"hello"}]}'
   ```

4. **监控日志**：启用 debug 日志观察 thinking 是否被正确处理
   ```yaml
   logging:
     level: "debug"
   ```

5. **文档记录**：在配置文件中记录每个服务是否支持 thinking
   ```yaml
   # config.yaml
   services:
     - id: "service-1"
       name: "Service Name"
       supports_thinking: false  # 不支持：智谱清言 API
   ```

## 未来扩展

如果需要过滤更多不兼容字段，可以扩展 `sanitizeRequestForExecutor` 函数：

```go
// handler.go
func (h *Handler) sanitizeRequestForExecutor(body []byte) ([]byte, error) {
    var reqMap map[string]interface{}
    json.Unmarshal(body, &reqMap)

    // 移除不兼容字段
    delete(reqMap, "thinking")         // 现有
    delete(reqMap, "future_feature")   // 未来可能添加

    return json.Marshal(reqMap)
}
```

并在 Service 结构体中添加对应的配置项。

## 参考

- **Claude API 文档**: https://docs.anthropic.com/
- **Thinking Mode 说明**: Claude Opus 4 特性文档
- **代理服务文档**: [CLAUDE.md](../CLAUDE.md)
- **配置示例**: [config.example.yaml](../configs/config.example.yaml)

---

**总结**：通过 `supports_thinking` 配置项，您可以灵活控制每个服务是否接收 thinking 字段，实现官方 API 和第三方 API 的完美兼容。
