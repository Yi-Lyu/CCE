<div align="center">

# 🚀 Claude Code Exchange (CCE)

**An Intelligent Proxy for Claude API with Smart Cost Optimization**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey?style=flat&logo=apple)](https://www.apple.com/macos/)

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

**Claude Code Exchange (CCE)** is an intelligent HTTP/HTTPS proxy service that optimizes your Claude API usage through smart request routing. It analyzes incoming requests, evaluates their complexity using AI, and automatically routes them to the most cost-effective Claude model based on a 1-5 difficulty scale.

### Why CCE?

- **💰 Cost Savings**: Automatically use cheaper models for simple tasks, premium models only when needed
- **⚡ Performance**: Built-in warmup mechanism and streaming support for minimal latency
- **🎯 Precision Routing**: AI-powered task complexity evaluation ensures optimal model selection
- **🖥️ User-Friendly**: Native macOS GUI for easy management and monitoring
- **🔧 Flexible**: Fully configurable routing rules and service mappings

### How It Works

```
┌─────────────┐         ┌──────────────┐         ┌─────────────────┐
│   Client    │────────▶│  CCE Proxy   │────────▶│   Evaluator     │
│ (Claude AI) │         │  (Port 27015)│         │ (Claude Haiku)  │
└─────────────┘         └──────────────┘         └─────────────────┘
                               │                          │
                               │                 ┌────────▼────────┐
                               │                 │ Difficulty: 1-5 │
                               │                 └─────────────────┘
                               │
                    ┌──────────┴──────────┐
                    ▼                     ▼
            ┌──────────────┐      ┌──────────────┐
            │ Simple Tasks │      │ Complex Tasks│
            │ (Haiku/Sonnet)│     │ (Opus/Sonnet)│
            └──────────────┘      └──────────────┘
```

## ✨ Features

### Core Capabilities

- **🧠 Intelligent Request Routing**
  - AI-powered task complexity evaluation
  - Intent extraction from conversation context
  - Automatic model selection based on difficulty

- **💸 Cost Optimization**
  - Route simple tasks to cheaper models (Haiku)
  - Reserve premium models (Opus) for complex operations
  - Configurable difficulty-to-model mapping

- **🖥️ macOS Native Client**
  - Menu bar integration
  - Real-time proxy status
  - Configuration management
  - Request history and analytics

- **⚡ Performance Features**
  - Warmup broadcast to all services
  - Full Server-Sent Events (SSE) streaming
  - Configurable timeouts (up to 30 minutes)
  - Connection pooling and reuse

- **🔧 Developer-Friendly**
  - Comprehensive logging with Zap
  - Structured configuration with Viper
  - RESTful status endpoint
  - Extensive testing support

## 📋 System Requirements

| Component | Requirement |
|-----------|-------------|
| **macOS** | 10.15 (Catalina) or later |
| **Architecture** | Apple Silicon (M1/M2/M3) or Intel |
| **Go** | 1.21+ (for development) |
| **Memory** | 512MB minimum |
| **Disk Space** | 100MB |
| **Network** | Internet connection for API access |

## 📥 Installation

### Method 1: Pre-built Release (Recommended)

**Download the latest release:**

1. Visit the [Releases](https://github.com/Yi-Lyu/cce/releases) page
2. Download the DMG for your system:
   - **Apple Silicon (M1/M2/M3)**: `CCE-vX.X.X-arm64.dmg`
   - **Intel Mac**: `CCE-vX.X.X-amd64.dmg`
   - **Universal**: `CCE-vX.X.X-universal.dmg` (works on both)
3. Open the DMG and drag CCE to Applications
4. Follow the first-launch instructions below

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/Yi-Lyu/cce.git
cd cce

# Quick build for your architecture
./build-mac-release-opensource.sh 1.0.0

# Or build individual components
cd proxy && make build      # Build proxy server
cd ../cce-client && make package  # Build GUI client
```

### ⚠️ Important: First Launch Security

**CCE is an unsigned open-source application.** macOS will show security warnings on first launch. This is normal for open-source software without an Apple Developer certificate.

**Choose one of these methods:**

#### Option 1: Right-Click to Open (Easiest)
1. Navigate to Applications folder
2. **Right-click** (or Control-click) on CCE.app
3. Select **"Open"** from the menu
4. Click **"Open"** in the security dialog

#### Option 2: System Settings (macOS 13+)
1. Try opening CCE (it will be blocked)
2. Go to **System Settings → Privacy & Security**
3. Find "CCE was blocked" message
4. Click **"Open Anyway"**

#### Option 3: Terminal Command
```bash
# Remove quarantine attribute
xattr -cr /Applications/CCE.app
```

After completing any method once, CCE will open normally in the future.

## 🚀 Quick Start

### 1. Configure API Keys

Launch CCE and configure your Claude API keys:

```yaml
# Edit proxy/configs/config.yaml
services:
  - id: "haiku"
    name: "Claude 3 Haiku"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "your-haiku-api-key"
    role: "evaluator"

  - id: "sonnet"
    name: "Claude 3.5 Sonnet"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "your-sonnet-api-key"
    role: "executor"
```

### 2. Start the Proxy

Via GUI:
- Click the CCE menu bar icon
- Select "Start Proxy"
- The icon will turn green when active

Via Command Line:
```bash
cd proxy
./claude-proxy -config=configs/config.yaml
```

### 3. Configure Your Client

Point your Claude client to the proxy:

```bash
# Set proxy endpoint
export CLAUDE_API_BASE_URL="http://127.0.0.1:27015"

# Or configure in your application
# Endpoint: http://127.0.0.1:27015/v1/messages
```

### 4. Verify It's Working

```bash
# Check proxy status
curl http://127.0.0.1:27015/status

# Send a test request (see examples in /docs)
```

## ⚙️ Configuration

### Configuration File Structure

CCE uses YAML for configuration. Start with the example:

```bash
cp proxy/configs/config.example.yaml proxy/configs/config.yaml
```

### Key Configuration Sections

#### 1. Proxy Settings

```yaml
proxy:
  port: 27015                 # Proxy listening port
  request_timeout: 1800       # Request timeout (seconds)
  read_timeout: 1900          # Read timeout (seconds)
  write_timeout: 1900         # Write timeout (seconds)
  evaluator_timeout: 30       # Evaluator timeout (seconds)
```

#### 2. Service Definitions

```yaml
services:
  # Evaluator: Analyzes request complexity
  - id: "haiku"
    name: "Claude 3 Haiku"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "${HAIKU_API_KEY}"  # Environment variable
    role: "evaluator"
    supports_thinking: true

  # Executors: Handle actual requests
  - id: "sonnet"
    name: "Claude 3.5 Sonnet"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "${SONNET_API_KEY}"
    role: "executor"
    supports_thinking: true

  - id: "opus"
    name: "Claude 3 Opus"
    url: "https://api.anthropic.com/v1/messages"
    api_key: "${OPUS_API_KEY}"
    role: "executor"
    supports_thinking: true
```

#### 3. Difficulty Mapping

Map complexity scores (1-5) to service IDs:

```yaml
difficulty_mapping:
  1: "haiku"      # Very simple tasks
  2: "haiku"      # Simple tasks
  3: "sonnet"     # Moderate complexity
  4: "sonnet"     # Complex tasks
  5: "opus"       # Very complex tasks
```

#### 4. Evaluator Configuration

```yaml
evaluator:
  model: "claude-3-haiku-20240307"
  max_tokens: 100
  temperature: 0
  max_history_rounds: 3
  prompt_template: |
    Analyze this task and rate its complexity from 1-5:
    {{.CurrentTask}}

    Context: {{.HistoryContext}}

    Return ONLY a number 1-5.
```

#### 5. Logging

```yaml
logging:
  level: "info"              # debug, info, warn, error
  output: "logs"             # Log directory
  max_size: 100              # MB per log file
  max_backups: 10            # Number of backups
  max_age: 30                # Days to keep logs
```

### Environment Variables

All configuration values support environment variable substitution:

```bash
# Set API keys via environment
export HAIKU_API_KEY="your-haiku-key"
export SONNET_API_KEY="your-sonnet-key"
export OPUS_API_KEY="your-opus-key"

# Override configuration via env vars
export CLAUDE_PROXY_PORT=8080
export CLAUDE_PROXY_REQUEST_TIMEOUT=3600
```

## 🛠️ Development

### Project Structure

```
cce/
├── proxy/                      # Go proxy service
│   ├── cmd/                   # Main entry point (main.go)
│   ├── internal/              # Internal packages
│   │   ├── proxy/            # Proxy server & handler
│   │   ├── evaluator/        # Task complexity evaluator
│   │   ├── config/           # Configuration management
│   │   └── models/           # Data structures
│   ├── configs/              # Configuration files
│   │   ├── config.example.yaml
│   │   └── config.test.yaml
│   └── Makefile              # Build commands
│
├── cce-client/               # macOS GUI application
│   ├── cmd/                  # Main entry point
│   ├── internal/             # Client logic
│   │   ├── app/             # Application core
│   │   ├── ui/              # User interface
│   │   └── proxy/           # Proxy management
│   ├── resources/            # App icons & assets
│   └── Makefile              # Build commands
│
├── scripts/                  # Build & release automation
│   ├── sign-app.sh          # Code signing
│   ├── create-dmg.sh        # DMG creation
│   └── generate-release-notes.sh
│
├── build-mac-release.sh          # Full release script
├── build-mac-release-opensource.sh  # Simplified build
└── CLAUDE.md                 # Project documentation
```

### Technology Stack

| Component | Technologies |
|-----------|-------------|
| **Proxy Server** | Go 1.21+, Gin, Viper, Zap |
| **GUI Client** | Go, Fyne (native macOS UI) |
| **API** | Claude API v1 (compatible) |
| **Config** | YAML, Environment Variables |
| **Logging** | Structured logging with Zap |
| **Build** | Make, Shell scripts |

### Development Commands

#### Proxy Server

```bash
cd proxy

# Development
make run                # Run with default config
make run-test           # Run with test config
make build              # Build binary
make test               # Run tests with coverage

# Testing
make mock               # Start mock evaluator (port 8081)
make test-api           # Run integration tests

# Utilities
make fmt                # Format code
make lint               # Run linter
make clean              # Clean build artifacts
```

#### GUI Client

```bash
cd cce-client

# Development
make run                # Run the app
make build              # Build binary
make package            # Create .app bundle
make test               # Run tests

# Utilities
make clean              # Clean build artifacts
```

### Building Releases

#### Quick Build (Open Source)

```bash
# Auto-detect architecture and build
./build-mac-release-opensource.sh 1.0.0

# Output: releases/v1.0.0/CCE-v1.0.0-{arch}.dmg
```

#### Advanced Build (With Signing)

```bash
# Unsigned build
./build-mac-release.sh --version 1.0.0

# Signed build (requires Apple Developer cert)
./build-mac-release.sh \
  --version 1.0.0 \
  --sign \
  --developer-id "Developer ID Application: Your Name (TEAM_ID)"

# Full release with notarization
./build-mac-release.sh \
  --version 1.0.0 \
  --sign \
  --notarize \
  --developer-id "Developer ID Application: Your Name" \
  --team-id "TEAM_ID" \
  --apple-id "you@email.com" \
  --app-password "app-specific-password"
```

#### Build Options

| Flag | Description | Default |
|------|-------------|---------|
| `-v, --version` | Release version (e.g., 1.0.0) | Required |
| `-a, --arch` | Architecture (arm64/amd64/universal) | Auto-detect |
| `-s, --sign` | Enable code signing | false |
| `-n, --notarize` | Submit for notarization | false |
| `-d, --developer-id` | Apple Developer ID | - |
| `-o, --output` | Output directory | releases |

## 🧪 Testing

### Unit Tests

```bash
# Test proxy server
cd proxy
make test                    # Run all tests with coverage
go test ./internal/proxy -v  # Test specific package

# Test GUI client
cd cce-client
go test ./...               # Run all tests
```

### Integration Tests

```bash
# Terminal 1: Start mock evaluator
cd proxy
make mock

# Terminal 2: Run integration tests
make test-api
```

### Manual Testing

```bash
# 1. Start proxy with test config
cd proxy
make run-test

# 2. Send test request
curl -X POST http://127.0.0.1:27015/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 100,
    "messages": [{
      "role": "user",
      "content": "Hello, Claude!"
    }]
  }'

# 3. Check status
curl http://127.0.0.1:27015/status
```

## 📦 Deployment

### Local Deployment

1. **Configure** your API keys in `proxy/configs/config.yaml`
2. **Build** the release: `./build-mac-release-opensource.sh 1.0.0`
3. **Install** the DMG from `releases/v1.0.0/`
4. **Launch** CCE from Applications

### GitHub Releases

#### Manual Release

```bash
# 1. Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. Build release artifacts
./build-mac-release-opensource.sh 1.0.0

# 3. Generate release notes
./scripts/generate-release-notes.sh \
  --version 1.0.0 \
  --output release-notes.md

# 4. Create GitHub release and upload DMG files
# Go to: https://github.com/Yi-Lyu/cce/releases/new
```

#### Automated Release (GitHub Actions)

Create `.github/workflows/release.yml`:

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-release:
    runs-on: macos-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build release
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          ./build-mac-release-opensource.sh $VERSION

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: releases/**/*.dmg
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 🔧 Troubleshooting

### Application Issues

#### "CCE can't be opened because it is from an unidentified developer"

**Solution:**
- Right-click on CCE.app and select "Open"
- Or: System Settings → Privacy & Security → Click "Open Anyway"
- Or: Run `xattr -cr /Applications/CCE.app` in Terminal

#### App crashes on startup

**Solution:**
1. Check logs: `~/Library/Logs/CCE/`
2. Remove corrupted config: `rm ~/Library/Application\ Support/CCE/config.yaml`
3. Reinstall the application

### Proxy Issues

#### "Connection refused" errors

**Checklist:**
- [ ] Proxy is running (check menu bar icon)
- [ ] Port 27015 is not in use: `lsof -i :27015`
- [ ] Firewall isn't blocking the proxy
- [ ] Check logs: `tail -f proxy/logs/*/proxy.log`

#### Requests timing out

**Solutions:**
- Increase timeout in config:
  ```yaml
  proxy:
    request_timeout: 3600  # 1 hour
  ```
- Check API key validity
- Verify network connectivity to Claude API

#### "Service not found" errors

**Solutions:**
- Verify service ID in difficulty mapping
- Check service configuration in `config.yaml`
- Ensure all required API keys are set

### Build Issues

#### Go version mismatch

```bash
# Check Go version
go version  # Should be 1.21+

# Install correct version via Homebrew
brew install go@1.21
```

#### Missing dependencies

```bash
cd proxy
make deps  # Download dependencies
go mod tidy  # Clean up go.mod
```

#### DMG creation fails

```bash
# Install required tools
brew install create-dmg

# Clean and retry
make clean
./build-mac-release-opensource.sh 1.0.0
```

### Configuration Issues

#### API keys not working

**Checklist:**
- [ ] Keys are properly quoted in YAML
- [ ] Environment variables are exported
- [ ] No extra spaces in key strings
- [ ] Keys have correct permissions

#### Evaluator always returns same difficulty

**Solutions:**
- Review evaluator prompt template
- Check evaluator logs: `proxy/logs/*/evaluator.log`
- Increase `max_history_rounds` for more context
- Verify evaluator API key is valid

### Getting Help

1. **Check Documentation:**
   - `CLAUDE.md` - Project overview
   - `proxy/README.md` - Proxy details
   - `cce-client/README.md` - Client details

2. **Enable Debug Logging:**
   ```yaml
   logging:
     level: "debug"
   ```

3. **Open an Issue:**
   - Include logs from `proxy/logs/`
   - Describe steps to reproduce
   - Mention your macOS and Go versions

## 📚 Documentation

- **[CLAUDE.md](CLAUDE.md)** - Comprehensive project guide for developers
- **[proxy/README.md](proxy/README.md)** - Proxy server documentation
- **[cce-client/README.md](cce-client/README.md)** - GUI client documentation
- **[configs/config.example.yaml](proxy/configs/config.example.yaml)** - Configuration reference

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Development Setup

1. **Fork and clone:**
   ```bash
   git clone https://github.com/Yi-Lyu/cce.git
   cd cce
   ```

2. **Install dependencies:**
   ```bash
   cd proxy && make deps
   cd ../cce-client && go mod download
   ```

3. **Create a feature branch:**
   ```bash
   git checkout -b feature/amazing-feature
   ```

### Development Workflow

1. **Make your changes**
   - Follow Go best practices
   - Add tests for new features
   - Update documentation

2. **Test your changes:**
   ```bash
   cd proxy && make test
   cd ../cce-client && make test
   ```

3. **Format and lint:**
   ```bash
   cd proxy && make fmt && make lint
   ```

4. **Commit and push:**
   ```bash
   git commit -m 'feat: add amazing feature'
   git push origin feature/amazing-feature
   ```

5. **Open a Pull Request**

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions or changes
- `chore:` Maintenance tasks

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write clear, descriptive comments

## 📄 License

This project is licensed under the **MIT License** - see below for details:

```
MIT License

Copyright (c) 2025 Ethan (Yi-Lyu)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

See [LICENSE](LICENSE) file for full details.

## 🙏 Acknowledgments

This project is built with excellent open-source tools:

- **[Gin](https://github.com/gin-gonic/gin)** - High-performance HTTP web framework
- **[Fyne](https://fyne.io/)** - Cross-platform GUI toolkit for Go
- **[Viper](https://github.com/spf13/viper)** - Configuration management
- **[Zap](https://github.com/uber-go/zap)** - Blazing fast, structured logging
- **[Claude API](https://anthropic.com/claude)** - AI-powered language models

Special thanks to all contributors who have helped improve this project!

## 💬 Support & Community

- **Issues:** [GitHub Issues](https://github.com/Yi-Lyu/cce/issues)
- **Discussions:** [GitHub Discussions](https://github.com/Yi-Lyu/cce/discussions)
- **Documentation:** [Project Wiki](https://github.com/Yi-Lyu/cce/wiki)

## 🗺️ Roadmap

- [ ] Windows support
- [ ] Linux support
- [ ] Web-based dashboard
- [ ] Advanced analytics and metrics
- [ ] Custom evaluator plugins
- [ ] Docker deployment option
- [ ] Kubernetes support

---

<div align="center">

**Made with ❤️ by Ethan**

[⭐ Star on GitHub](https://github.com/Yi-Lyu/cce) • [🐛 Report Bug](https://github.com/Yi-Lyu/cce/issues) • [💡 Request Feature](https://github.com/Yi-Lyu/cce/issues)

</div>