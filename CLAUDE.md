# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **intelligent HTTP/HTTPS proxy service** for Claude API requests that implements smart routing based on task complexity. The proxy intercepts Claude API calls, evaluates them using an AI service (Claude Haiku), and routes them to cost-appropriate backend services based on a 1-5 difficulty scale.

**Technology Stack**: Go 1.21+, Gin web framework, Viper (config), Zap (logging)

## Development Commands

```bash
# Build
make build              # Build binary to build/claude-proxy
make build-all          # Cross-platform builds

# Run
make run                # Run with default config (configs/config.yaml)
make run-test           # Run with test config (configs/config.test.yaml)
./claude-proxy -config=path/to/config.yaml  # Custom config

# Test
make test               # Run unit tests
make mock               # Run mock evaluator server (port 8081)
make test-api           # Run integration tests against mock

# Maintenance
make deps               # Download/update dependencies
make clean              # Clean build artifacts
```

## Architecture

### Request Flow

1. **Client** → Proxy (127.0.0.1:27015, configurable)
2. **Proxy** → Evaluator Service (Claude Haiku) for task complexity assessment
3. **Evaluator** returns difficulty score (1-5)
4. **Proxy** maps difficulty to target service via `difficulty_mapping` config
5. **Proxy** → Target Executor Service (forwards request with proper auth headers)
6. **Target Service** → Proxy → Client (streams or returns response)

### Core Components

- **Server** (`internal/proxy/server.go`): Gin-based HTTP server with middleware chain (Logger → CORS → Recovery → Proxy handler). Listens on port 27015 by default.

- **Handler** (`internal/proxy/handler.go`): Core proxy logic. Handles both streaming (SSE) and non-streaming responses. Buffers request body for replay. Manages configurable timeouts (default 30 minutes). Implements Warmup broadcast mechanism and per-service thinking mode sanitization.

- **Evaluator Client** (`internal/evaluator/client.go`): Calls Claude API to assess task complexity. Implements intelligent user intent extraction to filter out system-reminder tags, tool_result blocks, and command outputs. Maintains user context/history for smarter evaluations. Uses exponential backoff retry (3 attempts). Supports configurable prompt templates with variable substitution. Expects JSON-structured responses but has fallback parsing.

- **Config Manager** (`internal/config/config.go`): Loads YAML config with Viper. Validates services and mappings. Provides service lookup by ID. Supports environment variable overrides with `CLAUDE_PROXY_` prefix.

- **Logger** (`internal/logger/logger.go`): Zap-based structured logging. Writes to console + daily rotating files (`logs/YYYY-MM-DD/claude-proxy-YYYY-MM-DD.log`).

### Directory Structure

```
proxy/
├── cmd/main.go                    # Entry point, server initialization
├── internal/                      # Private application code
│   ├── config/                   # Configuration loading & validation
│   ├── evaluator/                # AI evaluation client
│   ├── logger/                   # Logging utilities
│   ├── models/                   # Shared data structures
│   └── proxy/                    # HTTP server & request handling
├── configs/                       # YAML configuration files
│   ├── config.example.yaml       # Template configuration
│   └── config.test.yaml          # Test environment config
├── docs/                          # Documentation
│   ├── EVALUATOR_CONFIG.md       # Evaluator configuration guide
│   └── THINKING_MODE.md          # Thinking mode compatibility guide
├── tests/                         # Test utilities and mocks
└── logs/                          # Log output directory
```

## Configuration

**Critical Files**:
- `configs/config.yaml` - Production config (not in repo, create from example)
- `configs/config.example.yaml` - Template with all available options
- `configs/config.test.yaml` - Test environment setup

**Key Configuration Sections**:

1. **Proxy Settings**: Configure timeouts and port
   - `port`: Proxy listening port (default: 27015)
   - `read_timeout`: HTTP read timeout in seconds (default: 1800 = 30 minutes)
   - `write_timeout`: HTTP write timeout in seconds (default: 1800 = 30 minutes)
   - `idle_timeout`: Connection idle timeout in seconds (default: 300 = 5 minutes)
   - `request_timeout`: Forwarding request timeout in seconds (default: 1800 = 30 minutes)
   - `evaluator_timeout`: Evaluator assessment timeout in seconds (default: 30)

2. **Services**: Define evaluator and executor services
   - `id`: Unique identifier for routing
   - `name`: Human-readable name
   - `url`: API endpoint
   - `api_key`: Authentication token
   - `role`: "evaluator" or "executor"
   - `supports_thinking`: Whether the service supports Claude's thinking mode (default: `true`)
     - Set to `false` for third-party Claude-compatible APIs (e.g., 智谱清言, 通义千问)

3. **Difficulty Mapping**: Maps complexity scores (1-5) to service IDs
   ```yaml
   difficulty_mapping:
     1: "haiku"    # Simplest tasks
     5: "opus"     # Most complex tasks
   ```

4. **Evaluator Configuration**: Customize complexity assessment behavior
   - `model`: Model used for evaluation (default: "claude-3-haiku-20240307")
   - `max_tokens`: Maximum tokens for evaluation response (default: 100)
   - `include_history`: Whether to include user conversation history (default: true)
   - `max_history_rounds`: Maximum number of historical rounds to include (default: 3)
   - `prompt_template`: Customizable prompt template with variable substitution
     - Variables: `{{.Model}}`, `{{.MessageCount}}`, `{{.CurrentTask}}`, `{{.HistoryContext}}`
   - See `docs/EVALUATOR_CONFIG.md` for detailed configuration guide

5. **Features**: Toggle proxy behaviors
   - `evaluator_fallback`: Use fallback service when evaluator unavailable (default: false)
   - `service_auto_switch`: Dynamic service switching on failure (default: false)
   - `request_logging`: Enable request/response logging (default: true)

6. **Logging**: Control log verbosity and output paths
   - `level`: Log level (debug, info, warn, error)
   - `output_path`: Directory for log files

**Environment Variables**: Override config with `CLAUDE_PROXY_` prefix (e.g., `CLAUDE_PROXY_PROXY_PORT=8080`)

## Implementation Details

### Warmup Preheating Mechanism (Broadcast Mode)
The proxy implements a broadcast-style warmup mechanism to optimize prompt caching across all executor services.

**How it works**:
1. **Detection**: When Claude Code sends a Warmup request (message text = "Warmup"), the proxy detects it using `models.IsWarmupRequest()`
2. **Broadcast**: The request is concurrently sent to **all** executor services (role="executor")
3. **Caching**: Each service establishes prompt cache with ephemeral cache_control markers
4. **Response**: Returns the first successful response to the client
5. **Logging**: Records success/failure statistics for all services

**Benefits**:
- All backend services are preheated, not just one
- Subsequent requests of any difficulty level benefit from cached prompts
- Reduces latency for all task complexity levels
- No additional evaluator cost (Warmup bypasses evaluation)

**Implementation**:
- Located in `handler.go:handleWarmupRequest()` (handler.go:155-315)
- Uses goroutines with `sync.WaitGroup` for concurrent requests
- 10-second timeout per service
- Requires at least one successful response
- Supports both streaming and non-streaming responses

**Performance**:
- Total warmup time ≈ slowest service response (typically 2-3 seconds)
- Cost: N executor API calls per warmup (where N = number of executors)
- Success rate logged for monitoring

### Intelligent Complexity Assessment
The evaluator implements smart user intent extraction to accurately assess task complexity by filtering out auxiliary content.

**Problem Solved**: Early versions would evaluate complexity based on the entire conversation history, including system reminders, tool results, and command outputs. This caused nearly all tasks to be rated as difficulty level 1 because the evaluator couldn't identify the actual user task.

**Solution - Intent Extraction** (`evaluator/client.go`):
1. **extractUserIntent()**: Intelligently extracts the real user task from messages
   - Filters out `<system-reminder>` tags
   - Filters out `<tool_result>` blocks and tool use outputs
   - Filters out command execution results
   - Filters out "User has answered your questions" prompts
   - Returns only genuine user-written text

2. **isAuxiliaryContent()**: Detects non-task content
   - Identifies system-generated messages
   - Recognizes tool invocation results
   - Detects file operation confirmations

3. **extractRecentContext()**: Fallback context extractor
   - Used when extractUserIntent returns empty
   - Collects last N user messages
   - Provides conversation flow summary

**Benefits**:
- **Accurate**: Evaluator focuses on actual user intent, not conversation noise
- **Contextual**: Still maintains conversation history for better assessment
- **Robust**: Handles complex multi-turn conversations with many tool calls
- **Configurable**: Prompt template system allows customization

**Prompt Template System** (`evaluator/client.go:renderTemplate`):
- Supports variable substitution: `{{.Model}}`, `{{.MessageCount}}`, `{{.CurrentTask}}`, `{{.HistoryContext}}`
- Configurable via `evaluator.prompt_template` in config.yaml
- Emphasizes "current step" evaluation vs. "overall project" complexity
- See `docs/EVALUATOR_CONFIG.md` for template customization guide

### Request Sanitization (Per-Service Configuration)
The proxy supports per-service configuration to control whether thinking mode should be stripped from requests, ensuring compatibility with third-party Claude-compatible APIs.

**Configuration**:
Each service in `config.yaml` can specify `supports_thinking`:
```yaml
services:
  - id: "official-api"
    url: "https://api.anthropic.com/v1/messages"
    role: "executor"
    supports_thinking: true   # Official API supports thinking (default)

  - id: "third-party-api"
    url: "https://api.example.com/v1/messages"
    role: "executor"
    supports_thinking: false  # Third-party API doesn't support thinking
```

**What gets removed** (when `supports_thinking: false`):
- `thinking` field - Extended thinking mode (Claude Opus 4 feature) that may not be supported by third-party APIs

**How it works**:
1. Each service has a `supports_thinking` flag (default: `true`)
2. When forwarding requests, the proxy checks `service.SupportsThinking`
3. If `false`, removes incompatible fields using `sanitizeRequestForExecutor()` (handler.go:437-456)
4. Re-serializes the cleaned request and forwards it

**Benefits**:
- **Flexible**: Configure per-service, not globally
- **Explicit**: Clear documentation in config which APIs support thinking
- **Safe**: Defaults to supporting thinking (official API behavior)
- **Compatible**: Works with any third-party Claude-compatible API (智谱清言, 通义千问, etc.)
- **No client changes**: Claude Code client works without modification

**Implementation**:
- Service config: `config.go` (models/config.go:30-37)
- Default value handling: `config.go` (config/config.go:56-73)
- Request cleaning: `handler.go:createTargetRequest()` (handler.go:458-505)
- Logs debug message when thinking is stripped

**Documentation**: See `docs/THINKING_MODE.md` for comprehensive guide on thinking mode configuration and troubleshooting

### Configurable Timeout System
The proxy implements comprehensive timeout configuration to handle long-running tasks without disconnection.

**Problem Solved**: Early versions had hardcoded 5-minute timeouts, causing disconnections during long-running tasks (e.g., complex code generation taking 10-30 minutes).

**Configuration** (`proxy/internal/models/config.go:ProxyConfig`):
- `read_timeout`: Server read timeout (default: 1800s = 30 minutes)
- `write_timeout`: Server write timeout (default: 1800s = 30 minutes)
- `idle_timeout`: Connection idle timeout (default: 300s = 5 minutes)
- `request_timeout`: HTTP client timeout for forwarding requests (default: 1800s = 30 minutes)
- `evaluator_timeout`: Evaluator assessment timeout (default: 30s)

**Implementation**:
- Server timeouts: `server.go:Start()` (server.go:149-156)
- Client timeouts: `handler.go:handleNormalProxy()` and `handler.go:handleStreamingProxy()` (handler.go:327, 372)
- Evaluator timeout: `handler.go:handleProxyRequest()` (handler.go:118)
- Logs timeout configuration at startup for debugging

**Benefits**:
- **Long-running tasks**: Supports tasks taking up to 30 minutes by default
- **Customizable**: Adjust timeouts per deployment via config
- **No disconnections**: Claude Code clients stay connected during complex operations
- **Efficient**: Evaluator has shorter timeout (30s) for fast decisions

**Example Configuration**:
```yaml
proxy:
  port: 27015
  read_timeout: 1800       # 30 minutes
  write_timeout: 1800      # 30 minutes
  idle_timeout: 300        # 5 minutes
  request_timeout: 1800    # 30 minutes for long tasks
  evaluator_timeout: 30    # 30 seconds
```

### Context Management
The evaluator maintains per-user conversation history to make better complexity assessments. User IDs are extracted from metadata using pattern: `user_{hash}_account__session_{id}`.

### Streaming Support
The proxy properly handles Server-Sent Events (SSE) for streaming responses. It uses `gin.Context.Stream()` with proper flushing and forwards the stream line-by-line to clients.

### Error Handling
- Evaluator failures trigger retry logic with exponential backoff (3 attempts)
- Missing difficulty mappings use fallback service if configured
- Request body is buffered once and reused for both evaluation and forwarding
- Graceful shutdown with 30-second timeout for in-flight requests

### API Endpoints
- `/api/v1/messages` (and variations) - Main proxy endpoint for Claude API requests
- `/health` - Health check endpoint
- `/status` - Detailed status with service configuration

### Difficulty Extraction
The evaluator response parser supports multiple formats:
1. JSON with `difficulty` field (preferred)
2. Plain text with numeric value (1-5)
3. Fallback to default difficulty if parsing fails

The evaluator service MUST be Claude-compatible and all executor services must support the same Claude API contract.

## Development Notes

- Default proxy port is 27015 (configurable via `proxy.port`)
- Request timeout is **30 minutes** for long-running operations (configurable via `proxy.request_timeout`)
- Server timeouts (read/write) are also **30 minutes** by default (configurable)
- The project is currently on `feature/intelligent-proxy` branch
- Main branch is `main` (use for PRs)
- All internal packages follow Go standard project layout
- Use structured logging (Zap) for all log messages, not fmt.Println
- See `docs/EVALUATOR_CONFIG.md` for evaluator customization
- See `docs/THINKING_MODE.md` for third-party API compatibility

## Recent Major Updates

### Version: Nov 13, 2025 (Commit 382b868)
1. ✅ **Configurable Timeouts**: 30-minute defaults for long-running tasks
2. ✅ **Intelligent Evaluation**: Smart user intent extraction filters auxiliary content
3. ✅ **Template System**: Customizable evaluator prompts with variable substitution
4. ✅ **Thinking Mode Compatibility**: Per-service configuration for third-party APIs
5. ✅ **Warmup Broadcasting**: Concurrent warmup to all executor services
6. ✅ **Enhanced Logging**: Added LogDebug() function for detailed debugging
