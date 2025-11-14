# Claude Proxy Makefile

# 变量定义
BINARY_NAME=claude-proxy
MAIN_PATH=cmd/main.go
BUILD_DIR=build
LOGS_DIR=logs

# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 版本信息
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.VERSION=$(VERSION) -X main.BUILD_TIME=$(BUILD_TIME)"

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help:
	@echo "Claude Proxy 构建工具"
	@echo ""
	@echo "可用命令："
	@echo "  make build      - 构建二进制文件"
	@echo "  make run        - 运行代理服务"
	@echo "  make run-local  - 使用本地配置运行"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理构建文件"
	@echo "  make deps       - 下载依赖"
	@echo "  make mock       - 运行模拟服务器"
	@echo "  make test-api   - 测试代理 API"
	@echo "  make all        - 清理、测试并构建"

# 构建
.PHONY: build
build:
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行
.PHONY: run
run:
	@echo "启动代理服务..."
	@mkdir -p $(LOGS_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.yaml

# 使用本地配置运行
.PHONY: run-local
run-local:
	@echo "使用本地配置启动代理服务..."
	@if [ ! -f configs/config.local.yaml ]; then \
		echo "错误: configs/config.local.yaml 不存在"; \
		echo "请运行: cp configs/config.yaml configs/config.local.yaml"; \
		echo "然后编辑 config.local.yaml 填入您的 API 信息"; \
		exit 1; \
	fi
	@mkdir -p $(LOGS_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.local.yaml

# 使用测试配置运行
.PHONY: run-test
run-test:
	@echo "使用测试配置启动代理服务..."
	@mkdir -p $(LOGS_DIR)/test
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.test.yaml

# 测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"

# 下载依赖
.PHONY: deps
deps:
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "依赖下载完成"

# 运行模拟服务器
.PHONY: mock
mock:
	@echo "启动模拟决策者服务..."
	$(GOCMD) run tests/mock_evaluator_server.go

# 测试 API
.PHONY: test-api
test-api:
	@echo "测试代理 API..."
	@./tests/test_proxy.sh

# 完整构建流程
.PHONY: all
all: clean test build

# 跨平台构建
.PHONY: build-all
build-all:
	@echo "构建多平台版本..."
	@mkdir -p $(BUILD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "多平台构建完成"

# 安装到系统
.PHONY: install
install: build
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "卸载 $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"
