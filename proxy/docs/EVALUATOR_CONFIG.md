# 决策者配置说明

本文档介绍如何配置和自定义决策者（Evaluator）的行为，包括复杂度评估的prompt模板。

## 配置项说明

在 `config.yaml` 中的 `evaluator` 部分包含以下配置项：

```yaml
evaluator:
  # 评估使用的模型
  model: "claude-3-haiku-20240307"

  # 最大Token数（评估响应通常很短）
  max_tokens: 100

  # 是否包含历史上下文
  include_history: true

  # 历史上下文的最大轮数
  max_history_rounds: 3

  # Prompt模板（支持变量替换）
  prompt_template: |
    你的自定义prompt...
```

### 配置项详解

#### 1. `model` (string)
- **说明**: 用于评估任务复杂度的Claude模型
- **默认值**: `"claude-3-haiku-20240307"`
- **推荐**: 使用Haiku以获得快速响应和较低成本
- **可选值**: 任何Claude模型名称

#### 2. `max_tokens` (int)
- **说明**: 评估响应的最大token数
- **默认值**: `100`
- **说明**: 复杂度评估只需返回一个JSON对象，通常不需要太多tokens

#### 3. `include_history` (bool)
- **说明**: 是否在评估时包含用户的历史请求上下文
- **默认值**: `true`
- **作用**: 帮助evaluator更好地理解用户的任务进展
- **建议**: 保持开启以获得更准确的评估

#### 4. `max_history_rounds` (int)
- **说明**: 包含的历史上下文最大轮数
- **默认值**: `3`
- **范围**: 建议1-5轮
- **说明**: 太多历史可能导致评估偏差，太少可能缺乏上下文

#### 5. `prompt_template` (string)
- **说明**: 评估任务复杂度的prompt模板
- **格式**: 多行字符串，支持变量替换
- **变量**: 见下文"可用变量"部分

## Prompt模板变量

在 `prompt_template` 中，你可以使用以下变量（使用 `{{.VarName}}` 格式）：

| 变量 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `{{.Model}}` | string | 用户请求使用的模型名称 | `claude-sonnet-4-5-20250929` |
| `{{.MessageCount}}` | int | 用户请求中的消息数量 | `25` |
| `{{.CurrentTask}}` | string | 当前步骤的任务描述（自动提取） | `创建auth.py文件` |
| `{{.HistoryContext}}` | string | 用户历史请求上下文（如果启用） | `最近的请求历史...` |

### 任务提取逻辑

`{{.CurrentTask}}` 变量会智能提取用户的真实意图，自动过滤以下辅助内容：
- `<system-reminder>` 系统提醒
- `<tool_result>` 工具调用结果
- `<command-name>` 命令输出
- Tool相关的提示信息

这确保evaluator看到的是纯净的用户任务，而不是Claude Code的内部信息。

## 自定义Prompt示例

### 示例1: 简洁版本（推荐用于测试）

```yaml
evaluator:
  prompt_template: |
    评估任务复杂度（1-5级），聚焦当前步骤。

    任务: {{.CurrentTask}}

    标准：1-简单 2-基础 3-中等 4-复杂 5-很复杂

    返回: {"difficulty_level": 数字}
```

### 示例2: 详细版本（推荐用于生产）

```yaml
evaluator:
  prompt_template: |
    你是一个任务复杂度评估专家。请分析当前这一步具体任务的复杂度。

    重要：评估【当前步骤】而非整体项目！

    当前任务: {{.CurrentTask}}
    请求模型: {{.Model}}
    消息数: {{.MessageCount}}{{.HistoryContext}}

    评估标准：
    1级 - 简单查询、单文件操作
    2级 - 基础编码、配置修改
    3级 - 多文件开发、模块编写
    4级 - 架构设计、复杂重构
    5级 - 系统设计、大型项目

    返回JSON: {"difficulty_level": 1-5之间的整数}
```

### 示例3: 英文版本

```yaml
evaluator:
  prompt_template: |
    Evaluate the complexity of the CURRENT STEP (not the entire project).

    Task: {{.CurrentTask}}
    Model: {{.Model}}{{.HistoryContext}}

    Levels:
    1 - Simple query, single file
    2 - Basic coding, config changes
    3 - Multi-file, module development
    4 - Architecture design, complex refactor
    5 - System design, large project

    Return JSON: {"difficulty_level": <number 1-5>}
```

## 最佳实践

### 1. Prompt设计原则

- **明确性**: 清楚说明要评估"当前步骤"而非"整体项目"
- **简洁性**: 避免过长的prompt，增加token消耗
- **标准化**: 提供清晰的1-5级评估标准
- **格式要求**: 明确指定JSON响应格式

### 2. 历史上下文使用

```yaml
# 推荐：启用历史但限制轮数
evaluator:
  include_history: true
  max_history_rounds: 3  # 最近3轮即可

# 场景：快速评估，无需上下文
evaluator:
  include_history: false
```

### 3. 模型选择

```yaml
# 推荐：使用Haiku（快速、便宜）
evaluator:
  model: "claude-3-haiku-20240307"

# 高精度需求：使用Sonnet
evaluator:
  model: "claude-3-5-sonnet-20241022"
```

### 4. 测试你的Prompt

修改配置后，可以通过以下方式测试：

```bash
# 重启代理服务
make run

# 观察日志中evaluator的评估结果
tail -f logs/YYYY-MM-DD/claude-proxy-YYYY-MM-DD.log | grep "difficulty"
```

## 常见问题

### Q1: 为什么所有任务都被评估为难度1？

**原因**: Prompt可能没有强调"当前步骤"，导致evaluator看到的任务描述不准确。

**解决**:
1. 确保prompt包含"评估当前这一步"的明确说明
2. 检查 `{{.CurrentTask}}` 提取的内容是否正确（查看日志）
3. 调整prompt中的评估标准描述

### Q2: 如何查看evaluator实际收到的prompt？

**方法**: 启用debug日志

```yaml
logging:
  level: "debug"
```

然后查看日志文件，搜索"evaluator prompt"或类似关键词。

### Q3: 变量不生效怎么办？

**检查清单**:
1. 变量格式是否正确：`{{.VarName}}` （注意大小写）
2. 变量名是否在支持列表中（见上文表格）
3. 重启服务使配置生效

### Q4: 可以添加自定义变量吗？

**当前版本**: 不支持自定义变量

**扩展方法**: 如需添加新变量，需要修改 `evaluator/client.go` 中的 `buildEvaluationPrompt` 函数：

```go
templateData := map[string]interface{}{
    "Model":          model,
    "MessageCount":   messageCount,
    "CurrentTask":    currentTask,
    "HistoryContext": contextInfo,
    "YourCustomVar":  yourValue, // 添加你的变量
}
```

## 配置验证

启动服务时会自动验证配置，如果配置有误会显示错误信息：

```bash
make run

# 成功：
2025-01-13 10:00:00 INFO  配置加载成功

# 失败：
2025-01-13 10:00:00 ERROR 配置验证失败: evaluator.prompt_template 不能为空
```

## 进阶配置

### 动态调整评估策略

你可以根据不同场景使用不同的配置文件：

```bash
# 开发环境（快速评估）
./claude-proxy -config=configs/config.dev.yaml

# 生产环境（精确评估）
./claude-proxy -config=configs/config.prod.yaml

# 测试环境（调试模式）
./claude-proxy -config=configs/config.test.yaml
```

### 多语言Prompt

如果你的用户使用不同语言，可以创建多个配置文件：

```
configs/
  config.zh.yaml  # 中文prompt
  config.en.yaml  # 英文prompt
  config.ja.yaml  # 日文prompt
```

## 参考资源

- [Claude Prompt工程指南](https://docs.anthropic.com/claude/docs/prompt-engineering)
- [项目主配置文件](../configs/config.example.yaml)
- [Evaluator源码](../internal/evaluator/client.go)

---

如有问题或建议，欢迎提交Issue或PR！
