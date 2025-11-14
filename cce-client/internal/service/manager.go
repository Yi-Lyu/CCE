package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ethan/cce-client/internal/config"
)

// Status 代理服务状态
type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "已停止"
	case StatusStarting:
		return "启动中"
	case StatusRunning:
		return "运行中"
	case StatusUnhealthy:
		return "运行异常"
	default:
		return "未知"
	}
}

// Manager 服务管理器
type Manager struct {
	configManager     *config.Manager
	process           *os.Process
	status            Status
	statusMu          sync.RWMutex
	healthCtx         context.Context
	healthCancel      context.CancelFunc
	statusListeners   []func(Status) // 状态监听器列表
	statusListenersMu sync.RWMutex   // 监听器列表锁
}

// NewManager 创建服务管理器
func NewManager(configManager *config.Manager) *Manager {
	return &Manager{
		configManager: configManager,
		status:        StatusStopped,
	}
}

// AddStatusListener 添加状态监听器
func (m *Manager) AddStatusListener(listener func(Status)) {
	m.statusListenersMu.Lock()
	defer m.statusListenersMu.Unlock()
	m.statusListeners = append(m.statusListeners, listener)
}

// SetStatusCallback 设置状态变化回调（已废弃，使用 AddStatusListener）
// 为了向后兼容保留此方法
func (m *Manager) SetStatusCallback(callback func(Status)) {
	m.AddStatusListener(callback)
}

// Start 启动代理服务
func (m *Manager) Start() error {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	if m.status != StatusStopped {
		return fmt.Errorf("服务已在运行中")
	}

	// 检查端口是否被占用
	port := m.configManager.GetProxyPort()
	if m.isPortInUse(port) {
		return fmt.Errorf("端口 %d 已被占用，请先停止其他占用该端口的进程，或在配置中修改端口号", port)
	}

	// 获取二进制文件路径
	binaryPath, err := m.getBinaryPath()
	if err != nil {
		return fmt.Errorf("无法找到代理服务二进制: %w", err)
	}

	// 获取代理服务配置文件路径
	configPath, err := m.getProxyConfigPath()
	if err != nil {
		return fmt.Errorf("无法找到代理配置文件: %w", err)
	}

	log.Printf("启动代理服务 (PID 将在启动后显示)")

	// 启动进程
	cmd := exec.Command(binaryPath, "-config="+configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	m.process = cmd.Process
	m.setStatus(StatusStarting)

	// 等待服务就绪
	go m.waitForReady()

	// 启动健康检查
	m.startHealthCheck()

	log.Printf("代理服务已启动，PID: %d", m.process.Pid)
	return nil
}

// Stop 停止代理服务
func (m *Manager) Stop() error {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	if m.status == StatusStopped {
		// 即使状态是已停止，也检查是否有遗留进程
		m.killOrphanProcesses()
		return nil
	}

	// 停止健康检查
	if m.healthCancel != nil {
		m.healthCancel()
		m.healthCancel = nil
	}

	if m.process == nil {
		m.setStatus(StatusStopped)
		m.killOrphanProcesses()
		return nil
	}

	// 发送 SIGTERM 优雅停止
	if err := m.process.Signal(syscall.SIGTERM); err != nil {
		// 尝试强制杀死
		m.process.Kill()
	}

	// 等待进程退出（最多 5 秒）
	done := make(chan error, 1)
	go func() {
		_, err := m.process.Wait()
		done <- err
	}()

	select {
	case <-done:
		log.Println("代理服务已停止")
	case <-time.After(5 * time.Second):
		log.Println("停止超时，强制终止进程")
		m.process.Kill()
	}

	m.process = nil
	m.setStatus(StatusStopped)

	// 清理可能的遗留进程
	m.killOrphanProcesses()

	return nil
}

// Restart 重启代理服务
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("停止服务失败: %w", err)
	}

	// 等待 1 秒确保端口释放
	time.Sleep(1 * time.Second)

	return m.Start()
}

// GetStatus 获取当前状态
func (m *Manager) GetStatus() Status {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.status
}

// IsRunning 是否在运行
func (m *Manager) IsRunning() bool {
	status := m.GetStatus()
	return status == StatusRunning || status == StatusUnhealthy
}

// setStatus 设置状态（内部方法，需要持有锁）
func (m *Manager) setStatus(status Status) {
	m.status = status
	log.Printf("服务状态变更: %s", status)

	// 通知所有监听器
	m.statusListenersMu.RLock()
	listeners := make([]func(Status), len(m.statusListeners))
	copy(listeners, m.statusListeners)
	m.statusListenersMu.RUnlock()

	// 触发所有监听器
	for _, listener := range listeners {
		listener(status)
	}
}

// waitForReady 等待服务就绪（优化：减少调试日志）
func (m *Manager) waitForReady() {
	log.Println("等待服务就绪...")
	maxAttempts := 60 // 最多等待 30 秒
	for i := 0; i < maxAttempts; i++ {
		// 只在每 10 次检查时记录一次日志（减少 90% 日志量）
		if i%10 == 0 && i > 0 {
			log.Printf("正在等待服务启动... (%d/%d 秒)", i/2, maxAttempts/2)
		}
		if m.checkHealth() {
			m.statusMu.Lock()
			m.setStatus(StatusRunning)
			m.statusMu.Unlock()
			log.Println("代理服务已就绪")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("警告：服务启动超时，但仍在运行中")
	m.statusMu.Lock()
	m.setStatus(StatusUnhealthy)
	m.statusMu.Unlock()
}

// startHealthCheck 启动健康检查（优化：降低频率，减少日志）
func (m *Manager) startHealthCheck() {
	m.healthCtx, m.healthCancel = context.WithCancel(context.Background())

	go func() {
		// 优化：将检查间隔从 5 秒增加到 10 秒，减少 50% 开销
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.healthCtx.Done():
				return
			case <-ticker.C:
				if m.checkHealth() {
					m.statusMu.Lock()
					// 只在状态变化时通知（已优化，避免重复通知）
					if m.status != StatusRunning {
						log.Println("健康检查恢复正常")
						m.setStatus(StatusRunning)
					}
					m.statusMu.Unlock()
				} else {
					m.statusMu.Lock()
					if m.status == StatusRunning {
						log.Println("警告：健康检查失败，服务可能异常")
						m.setStatus(StatusUnhealthy)
					}
					m.statusMu.Unlock()
				}
			}
		}
	}()
}

// checkHealth 执行健康检查（优化：只在失败时记录日志）
func (m *Manager) checkHealth() bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// 获取端口号
	port := m.configManager.GetProxyPort()
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	resp, err := client.Get(url)
	if err != nil {
		// 只在失败时记录详细日志
		log.Printf("[健康检查] 失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	success := resp.StatusCode == http.StatusOK
	if !success {
		// 只在失败时记录详细日志
		log.Printf("[健康检查] 状态码异常: %d", resp.StatusCode)
	}
	return success
}

// getBinaryPath 获取代理服务二进制路径
func (m *Manager) getBinaryPath() (string, error) {
	// 方法 1: 检查 resources 目录（开发环境）
	devPath := filepath.Join("resources", "claude-proxy")
	if _, err := os.Stat(devPath); err == nil {
		return filepath.Abs(devPath)
	}

	// 方法 2: 检查应用包内（生产环境）
	execPath, err := os.Executable()
	if err == nil {
		// macOS .app 结构: CCE.app/Contents/MacOS/CCE
		// 资源路径: CCE.app/Contents/Resources/claude-proxy
		appPath := filepath.Dir(filepath.Dir(execPath))
		resourcePath := filepath.Join(appPath, "Resources", "claude-proxy")
		if _, err := os.Stat(resourcePath); err == nil {
			return resourcePath, nil
		}
	}

	// 方法 3: 检查 proxy/build 目录（开发环境备用）
	proxyPath := filepath.Join("..", "proxy", "build", "claude-proxy")
	if _, err := os.Stat(proxyPath); err == nil {
		return filepath.Abs(proxyPath)
	}

	return "", fmt.Errorf("未找到 claude-proxy 二进制文件")
}

// getProxyConfigPath 获取代理服务配置文件路径
func (m *Manager) getProxyConfigPath() (string, error) {
	// 方法 1: 检查用户配置目录（优先级最高）
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, "Library", "Application Support", "CCE", "proxy-config.yaml")
		if _, err := os.Stat(userConfigPath); err == nil {
			return userConfigPath, nil
		}
	}

	// 方法 2: 检查 proxy/configs 目录（开发环境）
	devConfigPath := filepath.Join("..", "proxy", "configs", "config.yaml")
	if absPath, err := filepath.Abs(devConfigPath); err == nil {
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	// 方法 3: 检查应用包内（生产环境）
	execPath, err := os.Executable()
	if err == nil {
		// macOS .app 结构: CCE.app/Contents/MacOS/CCE
		// 配置路径: CCE.app/Contents/Resources/proxy-config.yaml
		appPath := filepath.Dir(filepath.Dir(execPath))
		resourceConfigPath := filepath.Join(appPath, "Resources", "proxy-config.yaml")
		if _, err := os.Stat(resourceConfigPath); err == nil {
			return resourceConfigPath, nil
		}
	}

	// 方法 4: 尝试绝对路径（如果在已知位置开发）
	knownPath := "/Users/ethan/code/Claude-Code-Exchange/proxy/configs/config.yaml"
	if _, err := os.Stat(knownPath); err == nil {
		return knownPath, nil
	}

	return "", fmt.Errorf("未找到代理配置文件 (proxy-config.yaml 或 proxy/configs/config.yaml)")
}

// isPortInUse 检查端口是否被占用
func (m *Manager) isPortInUse(port int) bool {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		// 端口有响应，说明已被占用
		return true
	}

	// 尝试监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// 无法监听，说明端口被占用
		return true
	}
	listener.Close()
	return false
}

// killOrphanProcesses 清理遗留的 claude-proxy 进程
func (m *Manager) killOrphanProcesses() {
	// 使用 pkill 查找并杀死所有 claude-proxy 进程
	cmd := exec.Command("pkill", "-9", "claude-proxy")
	cmd.Run() // 忽略错误，pkill 如果没找到进程会返回错误

	// 等待一小段时间确保进程完全退出
	time.Sleep(500 * time.Millisecond)
}
