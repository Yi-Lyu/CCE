package main

import (
	"flag"
	"fmt"
	"os"
	
	"github.com/ethan/claude-proxy/internal/config"
	"github.com/ethan/claude-proxy/internal/logger"
	"github.com/ethan/claude-proxy/internal/proxy"
)

var (
	configFile = flag.String("config", "./configs/config.yaml", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	VERSION = "1.0.0"
	BANNER  = `
   _____ _                 _        _____                     
  / ____| |               | |      |  __ \                    
 | |    | | __ _ _   _  __| | ___  | |__) | __ _____  ___   _ 
 | |    | |/ _  | | | |/ _  |/ _ \ |  ___/ '__/ _ \ \/ / | | |
 | |____| | (_| | |_| | (_| |  __/ | |   | | | (_) >  <| |_| |
  \_____|_|\__,_|\__,_|\__,_|\___| |_|   |_|  \___/_/\_\\__, |
                                                          __/ |
  智能代理服务 v%s                                        |___/ 
`
)

func main() {
	flag.Parse()
	
	if *version {
		fmt.Printf("Claude Proxy version %s\n", VERSION)
		os.Exit(0)
	}
	
	fmt.Printf(BANNER, VERSION)
	fmt.Println()
	
	// 加载配置
	fmt.Println("正在加载配置...")
	if err := config.LoadConfig(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("配置加载成功")
	
	// 初始化日志
	fmt.Println("正在初始化日志...")
	if err := logger.InitLogger(&config.Cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()
	fmt.Println("日志初始化成功")
	
	// 打印配置摘要
	printConfigSummary()
	
	// 创建并启动服务器
	server := proxy.NewServer()
	if err := server.Start(); err != nil {
		logger.LogError("服务器启动失败", err)
		os.Exit(1)
	}
}

// printConfigSummary 打印配置摘要
func printConfigSummary() {
	fmt.Println("\n=== 配置摘要 ===")
	fmt.Printf("代理端口: %d\n", config.Cfg.Proxy.Port)
	fmt.Printf("服务数量: %d\n", len(config.Cfg.Services))
	
	// 打印服务列表
	fmt.Println("\n配置的服务:")
	for _, svc := range config.Cfg.Services {
		fmt.Printf("  - %s (%s): %s [角色: %s]\n", svc.ID, svc.Name, svc.URL, svc.Role)
	}
	
	// 打印难度映射
	fmt.Println("\n难度等级映射:")
	for i := 1; i <= 5; i++ {
		serviceID, ok := config.Cfg.DifficultyMapping[fmt.Sprintf("%d", i)]
		if ok {
			service, _ := config.GetServiceByID(serviceID)
			if service != nil {
				fmt.Printf("  - 等级 %d -> %s (%s)\n", i, service.Name, serviceID)
			}
		}
	}
	
	// 打印功能开关
	fmt.Println("\n功能开关:")
	fmt.Printf("  - 决策者备选: %v\n", config.Cfg.Features.EvaluatorFallback)
	fmt.Printf("  - 服务自动切换: %v\n", config.Cfg.Features.ServiceAutoSwitch)
	fmt.Printf("  - 请求日志记录: %v\n", config.Cfg.Features.RequestLogging)
	
	fmt.Println("\n按 Ctrl+C 停止服务器")
	fmt.Println("================\n")
}
