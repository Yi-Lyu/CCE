# 修复：端口占用导致启动失败

## 问题描述

**症状**：点击「启动服务」按钮后，没有任何反应或应用崩溃

**日志错误**：
```
listen tcp :27015: bind: address already in use
```

## 根本原因

1. **端口被占用**：27015 端口已被其他进程（通常是遗留的 claude-proxy 进程）占用
2. **缺少端口检查**：启动前没有检查端口是否可用
3. **遗留进程清理不完整**：停止服务后，进程可能没有完全退出

## 修复方案

### 1. 启动前端口检查

新增 `isPortInUse()` 方法，在启动服务前检查端口：

```go
// internal/service/manager.go:355-378
func (m *Manager) isPortInUse(port int) bool {
    // 方法 1: 健康检查（优先）
    client := &http.Client{Timeout: 1 * time.Second}
    url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
    resp, err := client.Get(url)
    if err == nil {
        resp.Body.Close()
        return true  // 端口有响应
    }

    // 方法 2: 尝试监听
    addr := fmt.Sprintf(":%d", port)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return true  // 无法监听，端口被占用
    }
    listener.Close()
    return false  // 端口空闲
}
```

### 2. 启动逻辑改进

在 `Start()` 方法中添加端口检查：

```go
// internal/service/manager.go:76-80
// 检查端口是否被占用
port := m.configManager.GetProxyPort()
if m.isPortInUse(port) {
    return fmt.Errorf("端口 %d 已被占用，请先停止其他占用该端口的进程，或在配置中修改端口号", port)
}
```

### 3. 遗留进程清理

新增 `killOrphanProcesses()` 方法，清理遗留的 claude-proxy 进程：

```go
// internal/service/manager.go:380-393
func (m *Manager) killOrphanProcesses() {
    // 使用 pkill 清理所有 claude-proxy 进程
    cmd := exec.Command("pkill", "-9", "claude-proxy")
    if err := cmd.Run(); err != nil {
        log.Printf("清理遗留进程: %v", err)
    } else {
        log.Println("已清理遗留的 claude-proxy 进程")
    }

    // 等待进程完全退出
    time.Sleep(500 * time.Millisecond)
}
```

### 4. 停止服务改进

在 `Stop()` 方法中调用遗留进程清理：

```go
// internal/service/manager.go:121-175
func (m *Manager) Stop() error {
    // ... 停止当前进程 ...

    // 清理可能的遗留进程
    m.killOrphanProcesses()

    return nil
}
```

### 5. 增强日志输出

添加详细的启动日志：

```go
log.Printf("准备启动代理服务...")
log.Printf("二进制路径: %s", binaryPath)
log.Printf("配置文件路径: %s", configPath)
```

## 错误提示改进

### 启动前
- **端口占用**：明确提示端口号和解决方案
- **配置缺失**：提示配置文件路径和可能的位置
- **二进制缺失**：提示二进制文件位置

### 启动后
- **启动成功**：显示 PID 和配置文件路径
- **健康检查失败**：提示等待超时但服务可能仍在启动
- **进程异常**：提示检查日志文件

## 使用场景

### 场景 1: 正常启动
```
用户: 点击「启动服务」
系统: 检查端口空闲 → 启动进程 → 等待就绪 → 显示「运行中」
```

### 场景 2: 端口被占用
```
用户: 点击「启动服务」
系统: 检测到端口 27015 被占用
系统: 显示错误「端口 27015 已被占用，请先停止其他占用该端口的进程...」
用户: 点击「停止服务」（清理遗留进程）
用户: 再次点击「启动服务」
系统: 端口空闲 → 启动成功
```

### 场景 3: 遗留进程
```
用户: 应用崩溃/强制退出后重新启动
系统: 检测到端口被占用
用户: 点击「停止服务」
系统: killOrphanProcesses() 清理所有 claude-proxy 进程
系统: 等待 500ms 确保进程退出
用户: 点击「启动服务」
系统: 启动成功
```

## 手动排查步骤

### 1. 检查端口占用
```bash
lsof -i :27015
# 输出示例:
# COMMAND     PID  USER   FD   TYPE NODE NAME
# claude-pr 85476 ethan   8u  IPv6      TCP *:27015 (LISTEN)
```

### 2. 手动停止进程
```bash
# 方法 1: 通过 PID
kill -9 85476

# 方法 2: 通过进程名
pkill -9 claude-proxy

# 方法 3: 使用应用的停止功能
# 点击「停止服务」按钮
```

### 3. 验证端口空闲
```bash
lsof -i :27015
# 应该没有输出，说明端口已释放
```

### 4. 重新启动
```bash
# 在应用中点击「启动服务」
# 或使用命令行
./CCE.app/Contents/Resources/claude-proxy -config=...
```

## 测试验证

### 测试用例 1: 正常启动
```bash
# 1. 确保端口空闲
lsof -i :27015  # 无输出

# 2. 启动应用
open CCE.app

# 3. 点击「启动服务」
# 预期: 状态变为「运行中」

# 4. 验证
lsof -i :27015  # 显示 claude-proxy 监听
curl http://127.0.0.1:27015/health  # 返回 OK
```

### 测试用例 2: 端口已占用
```bash
# 1. 手动占用端口
nc -l 27015 &

# 2. 在应用中点击「启动服务」
# 预期: 显示错误「端口 27015 已被占用...」

# 3. 停止占用进程
killall nc

# 4. 再次点击「启动服务」
# 预期: 启动成功
```

### 测试用例 3: 遗留进程清理
```bash
# 1. 手动启动代理（模拟遗留进程）
./CCE.app/Contents/Resources/claude-proxy -config=... &

# 2. 在应用中点击「停止服务」
# 预期: 遗留进程被清理

# 3. 验证
ps aux | grep claude-proxy  # 无输出
lsof -i :27015  # 无输出

# 4. 点击「启动服务」
# 预期: 启动成功
```

## 日志示例

### 成功启动
```
2025-11-14 15:00:00 准备启动代理服务...
2025-11-14 15:00:00 二进制路径: /Users/ethan/.../claude-proxy
2025-11-14 15:00:00 配置文件路径: /Users/ethan/.../proxy-config.yaml
2025-11-14 15:00:00 代理服务已启动，PID: 12345
2025-11-14 15:00:01 代理服务已就绪
2025-11-14 15:00:01 服务状态变更: 运行中
```

### 端口被占用
```
2025-11-14 15:00:00 准备启动代理服务...
2025-11-14 15:00:00 错误: 端口 27015 已被占用，请先停止其他占用该端口的进程，或在配置中修改端口号
```

### 清理遗留进程
```
2025-11-14 15:00:00 停止服务...
2025-11-14 15:00:01 已清理遗留的 claude-proxy 进程
2025-11-14 15:00:01 服务状态变更: 已停止
```

## 性能影响

- **端口检查**：约 1 秒（健康检查超时 + 监听尝试）
- **遗留进程清理**：约 500 毫秒（pkill + 等待）
- **总启动延迟**：增加约 1-2 秒（可接受）

## 后续改进

### 短期
1. **状态提示优化**：在 UI 显示详细的启动进度
2. **配置端口修改**：UI 中直接修改端口号
3. **进程管理面板**：显示所有 claude-proxy 进程

### 长期
1. **自动端口选择**：如果默认端口被占用，自动选择其他端口
2. **进程守护**：监控进程健康，自动重启异常进程
3. **多实例支持**：支持同时运行多个代理实例（不同端口）

## 修复文件清单

| 文件 | 修改 | 说明 |
|------|-----|------|
| `internal/service/manager.go` | 新增 | `isPortInUse()` 端口检查 |
| `internal/service/manager.go` | 新增 | `killOrphanProcesses()` 进程清理 |
| `internal/service/manager.go` | 修改 | `Start()` 添加端口检查 |
| `internal/service/manager.go` | 修改 | `Stop()` 添加遗留进程清理 |
| `internal/service/manager.go` | 导入 | 添加 `net` 包 |

## 验证清单

- [x] 端口占用检测工作正常
- [x] 遗留进程清理功能有效
- [x] 错误提示清晰明确
- [x] 正常启动流程不受影响
- [x] 停止服务完全清理进程
- [x] 日志输出详细且有用

---

**修复状态**：✅ 已完成并验证
**修复时间**：2025-11-14
**影响范围**：所有服务启动/停止操作
**关联问题**：端口占用、进程管理、错误处理
