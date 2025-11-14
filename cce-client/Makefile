.PHONY: build run clean package install-deps prepare-binary

# 项目名称
APP_NAME = CCE
BINARY_NAME = cce-client
PROXY_BINARY = claude-proxy

# 路径
BUILD_DIR = build
RESOURCES_DIR = resources
PROXY_BUILD_DIR = ../proxy/build

# Go 参数
GO = go
GOFLAGS = -v

## 默认目标
all: build

## 安装依赖
install-deps:
	@echo "安装 Go 依赖..."
	$(GO) mod download
	@echo "安装 Fyne CLI 工具..."
	$(GO) install fyne.io/fyne/v2/cmd/fyne@latest
	@echo "依赖安装完成"

## 准备代理服务二进制
prepare-binary:
	@echo "准备 claude-proxy 二进制..."
	@if [ ! -f "$(PROXY_BUILD_DIR)/$(PROXY_BINARY)" ]; then \
		echo "错误: 未找到 claude-proxy 二进制文件"; \
		echo "请先编译代理服务: cd ../proxy && make build"; \
		exit 1; \
	fi
	@mkdir -p $(RESOURCES_DIR)
	@cp $(PROXY_BUILD_DIR)/$(PROXY_BINARY) $(RESOURCES_DIR)/
	@chmod +x $(RESOURCES_DIR)/$(PROXY_BINARY)
	@echo "二进制文件已复制到 $(RESOURCES_DIR)/"

## 构建开发版本
build: prepare-binary
	@echo "构建 $(APP_NAME) 客户端..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

## 运行开发版本
run: build
	@echo "启动 $(APP_NAME) 客户端..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## 打包 macOS .app 文件（手动方式，避免 fyne package 兼容性问题）
package: build
	@echo "打包 macOS 应用..."
	@mkdir -p $(APP_NAME).app/Contents/{MacOS,Resources}
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@chmod +x $(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@cp $(RESOURCES_DIR)/icon.png $(APP_NAME).app/Contents/Resources/
	@cp $(RESOURCES_DIR)/$(PROXY_BINARY) $(APP_NAME).app/Contents/Resources/
	@chmod +x $(APP_NAME).app/Contents/Resources/$(PROXY_BINARY)
	@if [ -f "$(PROXY_BUILD_DIR)/../configs/config.yaml" ]; then \
		cp $(PROXY_BUILD_DIR)/../configs/config.yaml $(APP_NAME).app/Contents/Resources/proxy-config.yaml; \
		echo "代理配置文件已复制"; \
	else \
		echo "警告: 未找到代理配置文件"; \
	fi
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(APP_NAME).app/Contents/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(APP_NAME).app/Contents/Info.plist
	@echo '<plist version="1.0"><dict>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleExecutable</key><string>$(APP_NAME)</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleIconFile</key><string>icon.png</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleIdentifier</key><string>com.cce.claude-proxy</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleName</key><string>$(APP_NAME)</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleDisplayName</key><string>Claude Code Exchange</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundlePackageType</key><string>APPL</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleShortVersionString</key><string>1.0.0</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>CFBundleVersion</key><string>1</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>LSMinimumSystemVersion</key><string>10.14</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>NSHighResolutionCapable</key><true/>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>LSUIElement</key><true/>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '	<key>NSHumanReadableCopyright</key><string>© 2025 CCE Team</string>' >> $(APP_NAME).app/Contents/Info.plist
	@echo '</dict></plist>' >> $(APP_NAME).app/Contents/Info.plist
	@echo "打包完成: $(APP_NAME).app （已配置隐藏 Dock 图标）"

## 打包并包含元数据
package-release: prepare-binary
	@echo "打包发布版本..."
	fyne package -os darwin -name "$(APP_NAME)" -icon resources/icon.png -release
	@echo "发布版本已打包"

## 清理构建产物
clean:
	@echo "清理构建产物..."
	rm -rf $(BUILD_DIR)
	rm -rf $(APP_NAME).app
	rm -f $(RESOURCES_DIR)/$(PROXY_BINARY)
	@echo "清理完成"

## 生成依赖图
deps-graph:
	$(GO) mod graph | modgraphviz | dot -Tpng -o deps.png

## 代码格式化
fmt:
	$(GO) fmt ./...

## 代码检查
lint:
	golangci-lint run ./...

## 显示帮助
help:
	@echo "可用的 Make 目标:"
	@echo "  make install-deps    - 安装所有依赖"
	@echo "  make prepare-binary  - 准备代理服务二进制"
	@echo "  make build          - 构建开发版本"
	@echo "  make run            - 运行开发版本"
	@echo "  make package        - 打包 macOS .app 文件"
	@echo "  make package-release- 打包发布版本"
	@echo "  make clean          - 清理构建产物"
	@echo "  make fmt            - 格式化代码"
	@echo "  make help           - 显示此帮助信息"
