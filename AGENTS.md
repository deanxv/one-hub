# AGENTS.md — One Hub (one-api)

LLM API aggregation gateway. Go (Gin) backend + React (Vite/MUI) frontend.
Forked from songquanpeng/one-api. Multi-provider proxy for OpenAI, Claude, Gemini, etc.

## Build / Run / Test Commands

### Backend (Go)

```bash
# Full build (frontend + backend)
make all
# — or via Task runner —
task build

# Backend only (requires web/build/ to exist)
go build -o dist/one-api

# Hot-reload development (uses .air.toml)
air

# Run tests — all
go test ./...

# Run a single test by name
go test -run TestDingTalkSend ./common/notify/channel/...

# Run all tests in a package
go test ./providers/ali/...

# Format
task gofmt          # runs gofmt -s -w . && goimports -w .

# Lint
task golint         # runs golangci-lint run -v ./...
task lint           # gofmt + golint combined
task fmt            # gomod tidy + gofmt + golint
```

### Frontend (web/)

```bash
cd web
yarn install
yarn dev            # dev server (Vite)
yarn build          # production build → web/build/
yarn lint           # eslint
yarn lint:fix       # eslint --fix
yarn prettier       # prettier
```

### Docker

```bash
docker build -t one-api .
# or
task docker         # builds linux/amd64 + pushes image
```

## Project Structure

```
main.go              # Entry point, initialization sequence
controller/          # HTTP handlers (Gin), one file per resource
model/               # GORM models, DB queries, business logic
middleware/          # Gin middleware (auth, rate-limit, CORS, logging)
relay/               # API relay/proxy logic — core request forwarding
providers/           # Per-provider implementations (openai/, claude/, gemini/, etc.)
  base/              # Provider interfaces and base types
router/              # Route registration (api, relay, dashboard, web, mcp)
common/              # Shared utilities, config, logger, cache, redis, crypto
  config/            # Global configuration variables
  logger/            # Zap-based structured logging
  test/              # Test helpers (HTTP server mocking, chat checks)
types/               # Shared type definitions (OpenAI-compatible request/response)
web/                 # React frontend (Vite + MUI)
i18n/                # Internationalization files
mcp/                 # MCP server integration
metrics/             # Prometheus metrics
```

## Code Style — Go

### Formatting & Linting

- **gofmt + goimports** enforced. Tabs for Go files (Go standard).
- **golangci-lint** with: goimports, gofmt, govet, misspell, ineffassign,
  typecheck, whitespace, gocyclo, revive, unused.
- EditorConfig: UTF-8, LF line endings, 2-space indent for non-Go files.

### Import Order

Standard library, blank line, third-party, blank line, internal (`one-api/...`):

```go
import (
    "errors"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "one-api/common"
    "one-api/common/utils"
    "one-api/model"
)
```

### Naming Conventions

- **Packages**: lowercase, single word (`controller`, `model`, `relay`, `common`)
- **Exported functions**: PascalCase — `GetChannelsList`, `AddChannel`, `UserAuth`
- **Local variables**: camelCase — `channel`, `baseUrl`, `localChannel`
- **Structs/Interfaces**: PascalCase — `Channel`, `ProviderInterface`, `ChatInterface`
- **Constants**: PascalCase — `ChannelStatusEnabled`, `UserStatusDisabled`
- **File naming**: lowercase with hyphens — `channel-billing.go`, `rate-limit.go`

### Struct Tags

Always use multi-tag format: `json`, `form` (for query binding), `gorm`:

```go
Type   int     `json:"type" form:"type" gorm:"default:0"`
Key    string  `json:"key" form:"key" gorm:"type:text"`
Weight *uint   `json:"weight" gorm:"default:1"`
```

Use pointer types (`*string`, `*int64`, `*uint`) for nullable/optional fields.

### Error Handling

- Early return on error — `if err != nil { return }` pattern throughout.
- Simple errors: `errors.New("message")` — messages often in Chinese.
- Formatted errors: `fmt.Errorf("context: %v", err)`.
- GORM not-found: `errors.Is(err, gorm.ErrRecordNotFound)`.
- API error wrapping: `common.ErrorWrapper(err, "error_code", http.StatusBadRequest)`.
- Never swallow errors with empty catch blocks.

### API Response Format

Admin/dashboard endpoints use success/message/data:

```go
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "message": "",
    "data":    result,
})
```

Relay/proxy errors use OpenAI-compatible error format via `common.AbortWithMessage`
or `common.ErrorWrapper`.

### Logging

Use the custom logger (`one-api/common/logger`), backed by Zap:

```go
logger.SysLog("informational message")
logger.SysError("error: " + err.Error())
logger.LogError(c.Request.Context(), message)  // context-aware with request ID
```

Do NOT use `fmt.Println` or raw `log.Printf` for application logging.

### Database

- ORM: GORM. Global DB instance in `model` package.
- Must support MySQL, PostgreSQL, SQLite simultaneously.
- Check DB type: `common.UsingPostgreSQL`, `common.UsingSQLite` (both false = MySQL).
- Migrations in `model/migrate.go` using gormigrate.

### Caching / Redis

- Redis client: `github.com/redis/go-redis/v9`
- Check availability: `common.RedisEnabled`
- Always provide non-Redis fallback — Redis is optional.

### Configuration

- Uses `github.com/spf13/viper` for all config.
- Global config vars live in `common/config/` as exported Go variables.
- Config file: `config.yaml` (see `config.example.yaml` for reference).

## Code Style — Frontend (web/)

- **Framework**: React 18 + Vite + MUI v5
- **Icons**: `import { Icon } from '@iconify/react'`
- **Notifications**: `import { showError, showSuccess, showInfo } from 'utils/common'`
- **State**: Redux + react-redux
- **Routing**: react-router-dom v6
- **All UI changes must support MUI dark mode.**
- **Package manager**: Yarn

## Adding a New Provider

1. Create directory under `providers/<name>/`
2. Implement interfaces from `providers/base/interface.go` (e.g., `ChatInterface`,
   `EmbeddingInterface`)
3. Register in `providers/providers.go`
4. Add channel type constant in `common/config/`
5. Add relay handling if needed

## Testing Patterns

- Standard `testing` package + `github.com/stretchr/testify/assert`
- Test files use `_test` package suffix (e.g., `package ali_test`)
- Function naming: `TestXxxYyy(t *testing.T)`
- HTTP mocking via `common/test/server.go` — `test.NewTestServer()`
- Config in tests loaded via `viper.ReadInConfig()`
- Very few tests exist — most are integration tests requiring external services.

## Existing Rules (Copilot / Cursor)

**Copilot** (.github/copilot-instructions.md):
- Go code uses the Gin framework. Provide Gin-applicable examples.
- JS uses Vite React. Provide Vite React examples.
- Database uses GORM, supporting MySQL/PostgreSQL/SQLite simultaneously.
  Check `common.UsingPostgreSQL` / `common.UsingSQLite`.
- Redis via go-redis/v9. Check `common.RedisEnabled` before using.

**Cursor** (.cursor/rules/onehub.mdc):
- Backend: Gin. Frontend: React/Vite with MUI.
- All frontend changes must support MUI dark mode.
- Icons: `@iconify/react`. Notifications: `showError/showSuccess/showInfo` from `utils/common`.
