# 配置文件说明

## 如何使用配置文件

1. **首次使用**：
   ```bash
   # 复制示例配置文件
   cp config.yaml config.local.yaml

   # 编辑 config.local.yaml，填入您的真实 API 信息
   vim config.local.yaml
   ```

2. **运行代理服务**：
   ```bash
   # 使用本地配置文件运行
   ./claude-proxy -config=configs/config.local.yaml

   # 或使用 make 命令
   make run-local
   ```

## 配置文件说明

- `config.yaml` - 示例配置文件，包含占位符，可安全提交到版本控制
- `config.local.yaml` - 您的本地配置文件（包含真实 API keys），已被 .gitignore 忽略
- `config.example.yaml` - 完整的配置示例，展示所有可用选项
- `config.test.yaml` - 测试环境配置

## 安全注意事项

⚠️ **重要**：
- 永远不要将包含真实 API keys 的配置文件提交到 Git
- 始终使用 `*.local.yaml` 格式命名您的本地配置文件
- 这些文件已被 `.gitignore` 自动忽略

## 服务配置

示例配置包含 3 个服务：
1. **evaluator-main** - 决策者服务，评估任务复杂度
2. **easy-executor** - 执行简单任务（难度 1-3）
3. **harder-executor** - 执行复杂任务（难度 4-5）

您可以根据需要添加更多服务或调整难度映射。