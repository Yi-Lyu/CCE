# 修复：启动服务崩溃问题

## 问题描述

**症状**：点击「启动服务」按钮后，应用崩溃

**原因**：服务管理器启动代理服务时，传递了**错误的配置文件路径**

## 根本原因分析

### 问题代码（修复前）

```go
// internal/service/manager.go:83
configPath := m.configManager.GetConfigPath()
cmd := exec.Command(binaryPath, "-config="+configPath)
```

### 问题分析

1. **配置混淆**：
   - `m.configManager.GetConfigPath()` 返回的是**客户端配置文件**路径：
     `~/Library/Application Support/CCE/config.yaml`
   - 但代理服务需要的是**代理服务配置文件**路径：
     `proxy/configs/config.yaml`

2. **配置格式不匹配**：
   - **客户端配置**：非常简单，只有基本设置
   - **代理服务配置**：复杂配置，包含服务列表、难度映射、evaluator 配置等

3. **启动失败**：
   - 代理服务读取错误的配置文件
   - 配置解析失败或找不到配置文件
   - 进程崩溃或启动失败

## 修复方案

### 1. 添加专门的配置路径查找方法

新增 `getProxyConfigPath()` 方法，按优先级查找代理配置文件：

```go
// internal/service/manager.go:293-335
func (m *Manager) getProxyConfigPath() (string, error) {
    // 方法 1: 用户配置目录（优先级最高）
    // ~/Library/Application Support/CCE/proxy-config.yaml

    // 方法 2: 开发环境
    // ../proxy/configs/config.yaml

    // 方法 3: 应用包内（生产环境）
    // CCE.app/Contents/Resources/proxy-config.yaml

    // 方法 4: 已知绝对路径（开发备用）
    // /Users/ethan/code/Claude-Code-Exchange/proxy/configs/config.yaml
}
```

### 2. 修改启动逻辑

```go
// internal/service/manager.go:82-89
// 获取代理服务配置文件路径
configPath, err := m.getProxyConfigPath()
if err != nil {
    return fmt.Errorf("无法找到代理配置文件: %w", err)
}

// 启动进程
cmd := exec.Command(binaryPath, "-config="+configPath)
```

### 3. 打包时复制配置文件

更新 Makefile，在打包时自动复制代理配置文件：

```makefile
# Makefile:62-67
@if [ -f "$(PROXY_BUILD_DIR)/../configs/config.yaml" ]; then \
    cp $(PROXY_BUILD_DIR)/../configs/config.yaml $(APP_NAME).app/Contents/Resources/proxy-config.yaml; \
    echo "代理配置文件已复制"; \
else \
    echo "警告: 未找到代理配置文件"; \
fi
```

## 修复后的文件结构

### 开发环境
```
cce-client/
├── CCE.app/
│   └── Contents/
│       └── Resources/
│           ├── claude-proxy           # 代理服务二进制
│           ├── proxy-config.yaml      # 代理配置文件 ✅
│           └── icon.png

proxy/
└── configs/
    └── config.yaml                    # 原始代理配置
```

### 运行时查找顺序

1. **用户配置**（可自定义）：
   `~/Library/Application Support/CCE/proxy-config.yaml`

2. **开发环境**（相对路径）：
   `../proxy/configs/config.yaml`

3. **应用包内**（打包后）：
   `CCE.app/Contents/Resources/proxy-config.yaml`

4. **已知路径**（开发备用）：
   `/Users/ethan/code/Claude-Code-Exchange/proxy/configs/config.yaml`

## 测试验证

### 1. 验证配置文件已打包

```bash
ls -lh CCE.app/Contents/Resources/
# 应显示:
# - claude-proxy (13M)
# - icon.png (97K)
# - proxy-config.yaml (3.6K) ✅
```

### 2. 验证配置内容正确

```bash
head -30 CCE.app/Contents/Resources/proxy-config.yaml
# 应显示完整的代理配置，包括:
# - proxy 设置
# - services 列表
# - difficulty_mapping
# - evaluator 配置
```

### 3. 测试启动服务

```bash
# 1. 关闭之前的应用实例
killall CCE 2>/dev/null

# 2. 启动新版本
open CCE.app

# 3. 在应用中点击「启动服务」
# 应该成功启动，不再崩溃
```

### 4. 验证代理服务运行

```bash
# 检查进程
ps aux | grep claude-proxy

# 检查端口监听
lsof -i :27015

# 测试健康检查
curl http://127.0.0.1:27015/health
```

## 日志输出

修复后，启动服务时会在控制台输出配置文件路径：

```
使用应用包配置: /Users/ethan/code/Claude-Code-Exchange/cce-client/CCE.app/Contents/Resources/proxy-config.yaml
代理服务已启动，PID: 12345
代理服务已就绪
```

## 后续改进建议

### 短期改进
1. **配置同步**：让客户端 UI 编辑的配置实时写入 `proxy-config.yaml`
2. **配置验证**：启动前验证代理配置文件的完整性
3. **错误提示**：配置文件缺失时给出明确的用户提示

### 长期改进
1. **统一配置管理**：
   - 客户端配置 = 代理配置
   - UI 直接编辑代理配置
   - 无需维护两份配置

2. **配置迁移工具**：
   - 首次运行时引导配置
   - 支持从旧配置迁移

3. **配置模板**：
   - 提供多种预设配置
   - 一键切换开发/生产环境

## 修复文件清单

| 文件 | 修改类型 | 说明 |
|------|---------|------|
| `internal/service/manager.go` | 修改 | 添加 `getProxyConfigPath()` 方法 |
| `internal/service/manager.go` | 修改 | 修改 `Start()` 方法使用正确配置 |
| `Makefile` | 修改 | 打包时复制代理配置文件 |
| `CCE.app/Contents/Resources/` | 新增 | 包含 `proxy-config.yaml` |

## 验证清单

- [x] 代理配置文件正确打包到 `.app` 中
- [x] 配置文件内容正确（包含所有必需字段）
- [x] 启动服务不再崩溃
- [x] 代理服务成功监听端口 27015
- [x] 健康检查通过
- [x] 日志输出正确的配置文件路径

## 使用说明

### 开发环境
```bash
cd cce-client
make clean
make package
open CCE.app
# 点击「启动服务」应该成功
```

### 自定义配置
如需自定义配置，复制配置文件到用户目录：
```bash
mkdir -p ~/Library/Application\ Support/CCE
cp proxy/configs/config.yaml ~/Library/Application\ Support/CCE/proxy-config.yaml
# 编辑 ~/Library/Application Support/CCE/proxy-config.yaml
```

应用会优先使用用户配置目录的配置文件。

---

**修复状态**：✅ 已完成并验证
**修复时间**：2025-11-14
**影响范围**：所有启动服务的操作
