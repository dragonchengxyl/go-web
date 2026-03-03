# AI 工程师自动化开发规范 (Skill Manual)

**Version:** 1.0.0
**Applies To:** 独立游戏工作室全矩阵中台项目
**Stack:** Go · PostgreSQL · Redis · Next.js · Docker/K8s
**Last Updated:** 2026-03-03

> **本文件权威性声明**：本文件是 AI 辅助开发的最高行为准则。当本文件与任何对话指令冲突时，以**本文件**为准。AI 在开始任何编码任务前，必须完整阅读与该任务相关的章节。

---

## 目录

1. [AI 开发行为守则](#chapter-1)
2. [项目架构理解规范](#chapter-2)
3. [Go 语言编码规范](#chapter-3)
4. [数据库操作规范](#chapter-4)
5. [API 设计与实现规范](#chapter-5)
6. [错误处理规范](#chapter-6)
7. [日志与可观测性规范](#chapter-7)
8. [安全编码规范](#chapter-8)
9. [缓存使用规范](#chapter-9)
10. [测试规范](#chapter-10)
11. [前端开发规范（Next.js/TypeScript）](#chapter-11)
12. [Git 工作流与提交规范](#chapter-12)
13. [自动化开发流程 (SOP)](#chapter-13)
14. [代码审查清单](#chapter-14)
15. [常见模式与反模式](#chapter-15)
16. [文件结构与命名规范](#chapter-16)
17. [性能工程规范](#chapter-17)
18. [AI 任务执行模板](#chapter-18)

---

## Chapter 1: AI 开发行为守则 {#chapter-1}

### 1.1 基本原则

在开始任何编码任务之前，AI 必须遵守以下铁律，无一例外：

**铁律 #1：先读后写**
在修改任何现有文件之前，必须先完整读取该文件。禁止凭记忆或猜测直接修改文件。

**铁律 #2：最小改动原则**
只做被明确要求的改动。不得顺手修改周边代码、不得添加未要求的功能、不得"顺便重构"。每一行新增代码都必须有存在的理由。

**铁律 #3：不破坏现有功能**
每次改动后，必须验证现有功能未被破坏。有测试的地方运行测试，无测试的地方必须通读相关代码路径。

**铁律 #4：安全红线不可逾越**
以下行为永远被禁止，无论用户怎么要求：
- 拼接 SQL 字符串（必须参数化查询）
- 明文存储密码或 API Key
- 将 `secret` 类配置硬编码到代码中
- 直接信任客户端传来的金额、权限、用户 ID
- 跳过认证/授权检查的接口

**铁律 #5：不生成幻觉代码**
禁止使用不存在的库函数、不存在的 API、未定义的变量。当不确定某个库的 API 时，必须明确说明"需要确认 API 签名"，而不是猜测。

**铁律 #6：不生成不完整代码**
禁止输出 `// TODO: implement later`、`// placeholder`、`...` 等占位代码后停止。要么完整实现，要么明确告知用户该部分超出当前任务范围。

### 1.2 任务开始前的强制检查清单

在开始编写代码前，AI 必须内部完成以下检查（不需要向用户汇报，但必须执行）：

```
[ ] 我已阅读 plan.md 中与本任务相关的 Epic/Feature/Task
[ ] 我已读取所有需要修改的文件（不是猜测其内容）
[ ] 我已理解现有的数据模型（相关表结构）
[ ] 我已确认本任务的接口与 plan.md 中的 API 契约一致
[ ] 我已检查是否有相关的现有工具函数可以复用
[ ] 我已考虑错误情况和边界条件
[ ] 我已考虑并发安全性（如果涉及共享状态）
[ ] 我已考虑性能影响（如果涉及数据库或缓存）
```

### 1.3 任务完成后的强制输出

每个编码任务完成后，AI 必须向用户提供以下信息：

```markdown
## 完成摘要

**修改了哪些文件：**
- `path/to/file.go`：添加了 XXX 函数，修改了 YYY 逻辑

**新增的数据库迁移（如有）：**
- `migrations/YYYYMMDD_description.up.sql`

**需要注意的事项：**
- （潜在的性能问题、已知限制、需要手动配置的环境变量等）

**建议的测试步骤：**
1. 运行 `go test ./internal/xxx/...`
2. 手动测试：`curl -X POST ...`
```

### 1.4 沟通规范

- **提问时机**：当任务描述存在歧义、可能影响架构决策时，必须先提问再开始编码。不允许在关键决策点自作主张。
- **假设声明**：当不得不做出假设时，必须在回复开头显式列出所有假设，并询问是否正确。
- **进度报告**：当任务需要多个步骤时，每完成一个步骤后汇报进度，而不是沉默很久后一次性输出。
- **语言规范**：技术名词使用英文（如 `goroutine`、`middleware`），业务说明使用中文。

---

## Chapter 2: 项目架构理解规范 {#chapter-2}

### 2.1 目录结构规范（强制遵循）

```
go-web/                          # 项目根目录
├── cmd/
│   └── server/
│       └── main.go              # 程序入口，只做初始化和启动
├── internal/                    # 私有业务代码（不可被外部包导入）
│   ├── domain/                  # 领域层：实体定义、仓储接口
│   │   ├── user/
│   │   │   ├── entity.go        # 用户实体（struct + 业务方法）
│   │   │   └── repository.go    # 用户仓储接口（interface）
│   │   ├── game/
│   │   ├── album/
│   │   ├── order/
│   │   └── comment/
│   ├── usecase/                 # 用例层：业务逻辑编排（依赖 domain 接口）
│   │   ├── user/
│   │   │   ├── register.go
│   │   │   ├── login.go
│   │   │   └── service.go       # 组合多个用例的服务对象
│   │   ├── game/
│   │   ├── album/
│   │   └── order/
│   ├── infra/                   # 基础设施层：具体实现（数据库、缓存、OSS）
│   │   ├── postgres/            # PostgreSQL 仓储实现
│   │   │   ├── user_repo.go
│   │   │   └── game_repo.go
│   │   ├── redis/               # Redis 缓存实现
│   │   │   ├── session.go
│   │   │   └── ratelimit.go
│   │   ├── oss/                 # 对象存储实现
│   │   │   └── storage.go
│   │   └── email/              # 邮件发送实现
│   │       └── sender.go
│   ├── transport/               # 传输层：HTTP Handler（薄层，不含业务逻辑）
│   │   ├── http/
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── ratelimit.go
│   │   │   │   ├── logger.go
│   │   │   │   └── cors.go
│   │   │   ├── handler/
│   │   │   │   ├── user.go
│   │   │   │   ├── game.go
│   │   │   │   └── album.go
│   │   │   ├── dto/             # 请求/响应 DTO（数据传输对象）
│   │   │   │   ├── user_dto.go
│   │   │   │   └── game_dto.go
│   │   │   └── router.go        # 路由注册
│   │   └── websocket/
│   │       └── hub.go
│   └── pkg/                     # 内部共享工具包
│       ├── apperr/              # 应用错误码定义
│       ├── response/            # 统一响应封装
│       ├── validator/           # 请求参数校验
│       ├── pagination/          # 分页工具
│       └── crypto/              # 加密工具
├── pkg/                         # 可被外部导入的公共包（谨慎添加）
│   └── sdk/                     # 对外 Go SDK
├── migrations/                  # 数据库迁移文件
│   ├── 20260101000001_create_users.up.sql
│   └── 20260101000001_create_users.down.sql
├── configs/
│   ├── config.go                # 配置结构体定义
│   ├── config.yaml              # 默认配置（不含机密）
│   └── config.local.yaml        # 本地覆盖配置（gitignore）
├── scripts/                     # 运维脚本
├── docs/                        # 文档（OpenAPI 规范等）
│   └── openapi.yaml
├── deployments/                 # 部署配置
│   ├── docker-compose.yml
│   └── k8s/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── cd.yml
├── go.mod
├── go.sum
├── plan.md                      # 项目规划（本规范的上层文档）
└── skill.md                     # 本文件
```

### 2.2 依赖方向（依赖倒置原则）

```
transport → usecase → domain ← infra
    ↓           ↓
   dto        entity
              repository (interface)
```

**规则：**
- `transport` 层只能依赖 `usecase` 层，不得直接调用 `infra` 层
- `usecase` 层只能依赖 `domain` 层的接口（`repository.go` 中定义的 interface）
- `infra` 层实现 `domain` 层的接口，但不反向依赖
- `domain` 层不依赖任何其他内部层
- 违反此规则的代码必须被拒绝

### 2.3 配置管理规范

**强制：使用环境变量覆盖配置，禁止硬编码任何配置值**

```go
// configs/config.go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
    OSS      OSSConfig      `yaml:"oss"`
    Email    EmailConfig    `yaml:"email"`
}

type DatabaseConfig struct {
    DSN             string        `yaml:"dsn" env:"DATABASE_DSN"`
    MaxOpenConns    int           `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}
```

**环境变量优先级**：`环境变量` > `config.local.yaml` > `config.yaml`

**机密管理**：JWT Secret、数据库密码、OSS Key 等机密值**只能**通过环境变量或 Kubernetes Secret 注入，禁止出现在任何代码或 YAML 文件中。

---

## Chapter 3: Go 语言编码规范 {#chapter-3}

### 3.1 代码风格强制规范

**格式化：** 所有 Go 代码必须通过 `gofmt` 和 `goimports` 格式化。CI 管道中自动校验，不通过则构建失败。

**命名规范：**

```go
// ✅ 正确：包名小写单词，不用下划线
package usecase

// ✅ 正确：导出类型 PascalCase
type UserService struct{}

// ✅ 正确：未导出函数 camelCase
func buildTokenClaims(user *domain.User) jwt.Claims {}

// ✅ 正确：接口名以 -er 结尾（当表示单一能力时）
type TokenSigner interface {
    Sign(claims jwt.Claims) (string, error)
}

// ✅ 正确：接口名以 Repository/Service 结尾（当表示资源操作时）
type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// ❌ 错误：变量名过短（除了循环变量 i/j/k）
u, e := userRepo.FindByID(ctx, id)

// ✅ 正确：有意义的变量名
user, err := userRepo.FindByID(ctx, id)
```

**常量定义：**

```go
// ✅ 正确：使用 iota 定义枚举，同时定义 String() 方法
type UserStatus int

const (
    UserStatusActive UserStatus = iota + 1
    UserStatusBanned
    UserStatusDeleted
)

func (s UserStatus) String() string {
    switch s {
    case UserStatusActive:
        return "active"
    case UserStatusBanned:
        return "banned"
    case UserStatusDeleted:
        return "deleted"
    default:
        return "unknown"
    }
}
```

### 3.2 函数设计规范

**函数长度：** 单个函数不超过 50 行。超过 50 行必须拆分为多个私有函数。

**参数数量：** 函数参数超过 4 个时，必须使用 Options Struct：

```go
// ❌ 错误：参数过多
func CreateUser(ctx context.Context, username, email, password, role, inviteCode string, sendWelcomeEmail bool) (*User, error)

// ✅ 正确：使用 Options Struct
type CreateUserInput struct {
    Username         string
    Email            string
    Password         string
    Role             string
    InviteCode       string
    SendWelcomeEmail bool
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    // ...
}
```

**返回值规范：**

```go
// ✅ 正确：错误永远是最后一个返回值
func (r *userRepo) FindByID(ctx context.Context, id int64) (*User, error)

// ✅ 正确：不返回裸 bool，而是返回有意义的结构或 error
func (s *authService) VerifyToken(ctx context.Context, token string) (*Claims, error)

// ❌ 错误：返回 bool 表示成功与否
func (s *authService) VerifyToken(ctx context.Context, token string) (*Claims, bool)
```

**Context 规范：**

```go
// ✅ 强制：所有涉及 IO 操作的函数，第一个参数必须是 context.Context
func (r *gameRepo) FindBySlug(ctx context.Context, slug string) (*Game, error)

// ❌ 错误：缺少 context，无法超时控制
func (r *gameRepo) FindBySlug(slug string) (*Game, error)

// ❌ 错误：在 struct 中存储 context（context 生命周期应与请求一致）
type GameRepo struct {
    ctx context.Context  // 禁止！
    db  *pgxpool.Pool
}
```

### 3.3 错误处理规范（详见 Chapter 6）

```go
// ✅ 正确：错误必须被处理，不允许用 _ 忽略
result, err := someFunc()
if err != nil {
    return fmt.Errorf("someFunc failed: %w", err)
}

// ❌ 绝对禁止：忽略错误
result, _ := someFunc()
```

### 3.4 Goroutine 使用规范

```go
// ✅ 正确：启动 goroutine 时必须有 recover，防止 panic 崩溃整个服务
func (s *NotifyService) SendAsync(ctx context.Context, event NotifyEvent) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("notify goroutine panic", zap.Any("recover", r))
            }
        }()
        if err := s.send(ctx, event); err != nil {
            logger.Error("send notification failed", zap.Error(err))
        }
    }()
}

// ✅ 正确：goroutine 泄漏防护，使用 context 控制生命周期
func (w *Worker) Start(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return  // context 取消时退出
            case job := <-w.queue:
                w.process(job)
            }
        }
    }()
}

// ❌ 错误：无限 goroutine，没有退出机制
go func() {
    for {
        w.process(<-w.queue)
    }
}()
```

### 3.5 接口与依赖注入规范

```go
// ✅ 正确：依赖通过构造函数注入
type GameUsecase struct {
    gameRepo   domain.GameRepository
    userRepo   domain.UserRepository
    storage    domain.StorageService
    cache      domain.CacheService
    logger     *zap.Logger
}

func NewGameUsecase(
    gameRepo domain.GameRepository,
    userRepo domain.UserRepository,
    storage domain.StorageService,
    cache domain.CacheService,
    logger *zap.Logger,
) *GameUsecase {
    return &GameUsecase{
        gameRepo: gameRepo,
        userRepo: userRepo,
        storage:  storage,
        cache:    cache,
        logger:   logger,
    }
}

// ❌ 错误：在函数内部实例化依赖（无法测试）
func (s *GameUsecase) GetGame(ctx context.Context, id int64) (*Game, error) {
    repo := postgres.NewGameRepo(db)  // 禁止！
    return repo.FindByID(ctx, id)
}
```

### 3.6 初始化规范

```go
// ✅ 正确：使用 Wire 或手动依赖注入，在 main.go 中组装
// cmd/server/main.go
func main() {
    cfg := configs.Load()

    db := infra.NewPostgresPool(cfg.Database)
    rdb := infra.NewRedisClient(cfg.Redis)

    userRepo := postgres.NewUserRepo(db)
    cacheService := redis.NewCacheService(rdb)

    userUsecase := usecase.NewUserUsecase(userRepo, cacheService, logger)

    router := http.NewRouter(userUsecase, logger)
    server := &http.Server{
        Addr:    cfg.Server.Addr,
        Handler: router,
    }

    // 优雅关闭
    gracefulShutdown(server)
}
```

### 3.7 禁止使用的模式

```go
// ❌ 禁止使用 init() 做任何有副作用的初始化
func init() {
    db = connectDB()  // 禁止！
}

// ❌ 禁止使用全局变量存储可变状态
var GlobalDB *pgxpool.Pool  // 禁止！

// ❌ 禁止使用 panic 替代错误返回（除非是程序启动阶段的不可恢复错误）
func GetUser(id int64) *User {
    user, err := repo.FindByID(ctx, id)
    if err != nil {
        panic(err)  // 禁止！返回 error
    }
    return user
}

// ❌ 禁止使用 time.Sleep 在业务逻辑中等待
time.Sleep(2 * time.Second)  // 禁止！使用 context deadline 或 ticker
```

---

## Chapter 4: 数据库操作规范 {#chapter-4}

### 4.1 SQL 编写规范

**参数化查询（强制）：**

```go
// ✅ 正确：使用占位符
const query = `SELECT id, username, email FROM users WHERE email = $1 AND status = $2`
row := pool.QueryRow(ctx, query, email, "active")

// ❌ 绝对禁止：字符串拼接 SQL（SQL 注入漏洞）
query := "SELECT * FROM users WHERE email = '" + email + "'"
```

**SQL 文件组织：**

```go
// ✅ 正确：将 SQL 定义为包级常量，便于查找和审查
const (
    sqlFindUserByID = `
        SELECT id, uuid, username, email, avatar_key, bio, role, status,
               email_verified_at, created_at, last_login_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL`

    sqlCreateUser = `
        INSERT INTO users (username, email, password, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, uuid, created_at`

    sqlUpdateUserLastLogin = `
        UPDATE users
        SET last_login_at = NOW(), last_login_ip = $2
        WHERE id = $1`
)
```

**查询结果扫描：**

```go
// ✅ 正确：逐字段 Scan，字段顺序与 SELECT 对应
func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    user := &domain.User{}
    err := r.pool.QueryRow(ctx, sqlFindUserByID, id).Scan(
        &user.ID,
        &user.UUID,
        &user.Username,
        &user.Email,
        &user.AvatarKey,
        &user.Bio,
        &user.Role,
        &user.Status,
        &user.EmailVerifiedAt,
        &user.CreatedAt,
        &user.LastLoginAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrUserNotFound  // 转换为领域错误
        }
        return nil, fmt.Errorf("userRepo.FindByID: %w", err)
    }
    return user, nil
}
```

### 4.2 事务规范

```go
// ✅ 正确：事务必须使用 defer 确保提交或回滚
func (r *orderRepo) CreateWithItems(ctx context.Context, order *domain.Order, items []domain.OrderItem) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            // err 是外层变量，如果函数返回 error，则回滚
            _ = tx.Rollback(ctx)
        }
    }()

    // 插入订单
    if err = r.insertOrder(ctx, tx, order); err != nil {
        return fmt.Errorf("insert order: %w", err)
    }

    // 插入订单项
    for _, item := range items {
        if err = r.insertOrderItem(ctx, tx, &item); err != nil {
            return fmt.Errorf("insert order item: %w", err)
        }
    }

    // 所有操作成功后提交
    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

### 4.3 索引使用规范

在编写查询时，必须考虑是否有对应索引：

```sql
-- 每个 WHERE 子句涉及的列必须有索引（除非表很小）
-- 每个 ORDER BY 子句涉及的列必须有索引（排序索引）
-- 联合查询的 JOIN ON 列必须有索引

-- ✅ 正确：创建表时同步创建索引
CREATE TABLE comments (
    id          BIGSERIAL PRIMARY KEY,
    target_type VARCHAR(50) NOT NULL,
    target_id   BIGINT NOT NULL,
    user_id     BIGINT NOT NULL,
    parent_id   BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 必须为高频查询列创建索引
CREATE INDEX idx_comments_target ON comments(target_type, target_id, created_at DESC);
CREATE INDEX idx_comments_user ON comments(user_id, created_at DESC);
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;
```

**查询前的必检事项：**
- 使用 `EXPLAIN ANALYZE` 验证查询走了正确的索引
- 避免在 WHERE 子句中对列做函数操作（会导致索引失效）：
  ```sql
  -- ❌ 索引失效：对列使用了函数
  WHERE DATE(created_at) = '2026-03-03'
  -- ✅ 正确：对值做转换
  WHERE created_at >= '2026-03-03'::DATE AND created_at < '2026-03-04'::DATE
  ```

### 4.4 迁移文件规范

```
命名格式：{序号}_{描述}.{up|down}.sql
序号：14位时间戳（YYYYMMDDHHmmss）

示例：
migrations/
├── 20260101000001_create_users.up.sql
├── 20260101000001_create_users.down.sql
├── 20260101000002_create_games.up.sql
└── 20260101000002_create_games.down.sql
```

**迁移编写规范：**

```sql
-- up.sql 规范
-- 1. 必须包含回滚时对应的 down.sql
-- 2. 必须是幂等的（IF NOT EXISTS）
-- 3. 生产环境不允许 DROP TABLE / DROP COLUMN（需通过多步迁移完成）

-- ✅ 正确的 up.sql
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    -- ...
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ✅ 正确的 down.sql（只在开发环境使用）
DROP TABLE IF EXISTS users;

-- 生产删列的正确姿势（分3次迁移）：
-- 迁移1：将列标记为废弃（不删除）
-- 迁移2：代码不再读写该列（部署新代码）
-- 迁移3：数据库删除该列
```

### 4.5 分页查询规范

```go
// ✅ 正确：Cursor-based 分页（性能好，适合大数据量）
type PageQuery struct {
    Cursor string // base64 编码的游标
    Limit  int    // 每页数量，默认 20，最大 100
}

type PageResult[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"`
    HasMore    bool   `json:"has_more"`
}

// 游标解码：游标编码了最后一条记录的 (created_at, id)
func decodeCursor(cursor string) (time.Time, int64, error) {
    data, err := base64.StdEncoding.DecodeString(cursor)
    // ...
}

const sqlListComments = `
    SELECT id, content, user_id, created_at
    FROM comments
    WHERE target_type = $1 AND target_id = $2
      AND (created_at, id) < ($3, $4)  -- cursor 条件
    ORDER BY created_at DESC, id DESC
    LIMIT $5`

// ❌ 禁止在高数据量表上使用 OFFSET 分页
const sqlBadPagination = `SELECT * FROM comments LIMIT 20 OFFSET 10000`  // 禁止！
```

---

## Chapter 5: API 设计与实现规范 {#chapter-5}

### 5.1 Handler 规范

**Handler 只做以下事情，不做业务逻辑：**

```go
// ✅ 正确的 Handler 结构
func (h *GameHandler) GetGame(c *gin.Context) {
    // 1. 解析路径参数/查询参数
    slug := c.Param("slug")
    if slug == "" {
        response.BadRequest(c, apperr.ErrInvalidParam.WithDetail("slug is required"))
        return
    }

    // 2. 解析并验证请求体（POST/PUT/PATCH）
    // （GET 请求无请求体）

    // 3. 从 context 中获取认证信息
    userID := middleware.GetUserID(c)

    // 4. 调用 usecase（传入简单值或 DTO，不传 gin.Context）
    game, err := h.gameUsecase.GetBySlug(c.Request.Context(), slug, userID)
    if err != nil {
        response.HandleError(c, err)
        return
    }

    // 5. 将领域对象转换为响应 DTO，返回响应
    response.OK(c, dto.ToGameResponse(game))
}
```

**Request/Response DTO 规范：**

```go
// internal/transport/http/dto/game_dto.go

// 请求 DTO：用于绑定和验证输入
type CreateGameRequest struct {
    Title       string   `json:"title"       binding:"required,min=1,max=255"`
    Subtitle    string   `json:"subtitle"    binding:"omitempty,max=255"`
    Description string   `json:"description" binding:"required,min=10"`
    Genre       []string `json:"genre"       binding:"required,min=1,max=5"`
}

// 响应 DTO：控制对外暴露的字段
type GameResponse struct {
    ID          int64     `json:"id"`
    Slug        string    `json:"slug"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CoverURL    string    `json:"cover_url"`    // 非 OSS Key，是签名 URL
    CreatedAt   time.Time `json:"created_at"`
}

// 转换函数：领域实体 → 响应 DTO
func ToGameResponse(game *domain.Game, coverURL string) *GameResponse {
    return &GameResponse{
        ID:          game.ID,
        Slug:        game.Slug,
        Title:       game.Title,
        Description: game.Description,
        CoverURL:    coverURL,
        CreatedAt:   game.CreatedAt,
    }
}
```

### 5.2 路由注册规范

```go
// internal/transport/http/router.go
func NewRouter(deps *Dependencies, logger *zap.Logger) *gin.Engine {
    r := gin.New()

    // 全局中间件（顺序很重要）
    r.Use(middleware.Recovery(logger))   // 必须第一个：panic 恢复
    r.Use(middleware.RequestID())        // 注入 Request-ID
    r.Use(middleware.Logger(logger))     // 日志（在 RequestID 之后）
    r.Use(middleware.CORS())
    r.Use(middleware.RateLimit(deps.RateLimiter))

    // 健康检查（不需要认证）
    r.GET("/health", handler.Health)
    r.GET("/ready", handler.Ready)

    // API v1
    v1 := r.Group("/api/v1")
    {
        // 公开路由（无需认证）
        auth := v1.Group("/auth")
        {
            auth.POST("/register", deps.AuthHandler.Register)
            auth.POST("/login", deps.AuthHandler.Login)
            auth.POST("/refresh", deps.AuthHandler.Refresh)
            auth.GET("/oauth/:provider", deps.AuthHandler.OAuthRedirect)
            auth.GET("/oauth/:provider/callback", deps.AuthHandler.OAuthCallback)
        }

        // 需要认证的路由
        authed := v1.Group("", middleware.Auth(deps.TokenVerifier))
        {
            authed.GET("/users/me", deps.UserHandler.GetMe)
            authed.PATCH("/users/me", deps.UserHandler.UpdateMe)

            authed.GET("/games", deps.GameHandler.List)
            authed.GET("/games/:slug", deps.GameHandler.GetBySlug)
            // ...
        }

        // 需要管理员权限的路由
        admin := v1.Group("/admin", middleware.Auth(deps.TokenVerifier), middleware.RequireRole("admin"))
        {
            admin.POST("/games", deps.GameHandler.Create)
            admin.PUT("/games/:id", deps.GameHandler.Update)
            // ...
        }
    }

    return r
}
```

### 5.3 统一响应封装规范

```go
// internal/pkg/response/response.go

type Response struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data"`
    RequestID string      `json:"request_id"`
    Timestamp int64       `json:"timestamp"`
}

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:      0,
        Message:   "success",
        Data:      data,
        RequestID: c.GetString(middleware.KeyRequestID),
        Timestamp: time.Now().Unix(),
    })
}

func HandleError(c *gin.Context, err error) {
    var appErr *apperr.AppError
    if errors.As(err, &appErr) {
        c.JSON(appErr.HTTPStatus(), Response{
            Code:      appErr.Code,
            Message:   appErr.Message,
            Data:      nil,
            RequestID: c.GetString(middleware.KeyRequestID),
            Timestamp: time.Now().Unix(),
        })
        return
    }
    // 未知错误：记录日志，返回 500（不暴露内部信息）
    logger.Error("unhandled error", zap.Error(err), zap.String("request_id", c.GetString(middleware.KeyRequestID)))
    c.JSON(http.StatusInternalServerError, Response{
        Code:      50000,
        Message:   "internal server error",
        RequestID: c.GetString(middleware.KeyRequestID),
        Timestamp: time.Now().Unix(),
    })
}
```

### 5.4 请求验证规范

```go
// ✅ 正确：使用 binding tag 声明验证规则，集中验证
func (h *UserHandler) Register(c *gin.Context) {
    var req dto.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ShouldBindJSON 会自动处理 binding tag 的验证
        response.BadRequest(c, apperr.ErrValidation.WithDetail(err.Error()))
        return
    }
    // 通过验证后，req 中的数据已经是安全的
    // ...
}

// ✅ 正确：复杂业务验证在 usecase 层进行
func (s *UserService) Register(ctx context.Context, input RegisterInput) (*User, error) {
    // 业务规则验证（不是格式验证）
    exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
    if err != nil {
        return nil, fmt.Errorf("check email exists: %w", err)
    }
    if exists {
        return nil, apperr.ErrEmailAlreadyExists
    }
    // ...
}
```

---

## Chapter 6: 错误处理规范 {#chapter-6}

### 6.1 应用错误码体系

```go
// internal/pkg/apperr/errors.go

// AppError 是所有应用层错误的基础类型
type AppError struct {
    Code    int    // 业务错误码
    Message string // 用户可见的错误信息（中文）
    Detail  string // 技术细节（可选，调试用）
    Cause   error  // 原始错误（通过 errors.Is/As 查找）
}

func (e *AppError) Error() string {
    if e.Detail != "" {
        return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func (e *AppError) HTTPStatus() int {
    switch e.Code / 10000 {
    case 4:
        switch e.Code {
        case 40101, 40102:
            return http.StatusUnauthorized
        case 40301:
            return http.StatusForbidden
        case 40401:
            return http.StatusNotFound
        case 42901:
            return http.StatusTooManyRequests
        default:
            return http.StatusBadRequest
        }
    case 5:
        return http.StatusInternalServerError
    default:
        return http.StatusInternalServerError
    }
}

func (e *AppError) WithDetail(detail string) *AppError {
    clone := *e
    clone.Detail = detail
    return &clone
}

func (e *AppError) Wrap(cause error) *AppError {
    clone := *e
    clone.Cause = cause
    return &clone
}

// 预定义错误码（必须统一在此处维护）
var (
    // 4xx 客户端错误
    ErrInvalidParam        = &AppError{Code: 40001, Message: "请求参数有误"}
    ErrValidation          = &AppError{Code: 40002, Message: "参数验证失败"}
    ErrUnauthorized        = &AppError{Code: 40101, Message: "请先登录"}
    ErrTokenExpired        = &AppError{Code: 40102, Message: "登录已过期，请重新登录"}
    ErrForbidden           = &AppError{Code: 40301, Message: "权限不足"}
    ErrUserNotFound        = &AppError{Code: 40401, Message: "用户不存在"}
    ErrGameNotFound        = &AppError{Code: 40402, Message: "游戏不存在"}
    ErrReleaseNotFound     = &AppError{Code: 40403, Message: "版本不存在"}
    ErrEmailAlreadyExists  = &AppError{Code: 40901, Message: "该邮箱已被注册"}
    ErrUsernameAlreadyExists = &AppError{Code: 40902, Message: "该用户名已被占用"}
    ErrRateLimited         = &AppError{Code: 42901, Message: "操作过于频繁，请稍后再试"}

    // 5xx 服务器错误
    ErrInternal            = &AppError{Code: 50001, Message: "服务器内部错误"}
    ErrDependencyFailed    = &AppError{Code: 50002, Message: "依赖服务暂时不可用"}
    ErrStorageFailed       = &AppError{Code: 50003, Message: "文件存储服务异常"}
)
```

### 6.2 错误传递规范

```go
// ✅ 正确：错误向上传递时，用 %w 保留错误链
func (s *GameService) GetDownloadURL(ctx context.Context, releaseID int64, userID int64) (string, error) {
    // 检查权限
    hasAccess, err := s.userRepo.HasGameAccess(ctx, userID, releaseID)
    if err != nil {
        return "", fmt.Errorf("GameService.GetDownloadURL check access: %w", err)
    }
    if !hasAccess {
        return "", apperr.ErrForbidden.WithDetail("user does not own this game")
    }

    // 获取版本信息
    release, err := s.gameRepo.FindRelease(ctx, releaseID)
    if err != nil {
        return "", fmt.Errorf("GameService.GetDownloadURL find release: %w", err)
    }

    // 生成下载链接
    url, err := s.storage.GetPresignedURL(ctx, release.OSSKey, 15*time.Minute)
    if err != nil {
        return "", fmt.Errorf("GameService.GetDownloadURL generate url: %w", apperr.ErrStorageFailed.Wrap(err))
    }

    return url, nil
}

// ❌ 错误：吞掉错误信息
func (s *GameService) GetDownloadURL(ctx context.Context, releaseID int64, userID int64) (string, error) {
    release, err := s.gameRepo.FindRelease(ctx, releaseID)
    if err != nil {
        return "", apperr.ErrInternal  // 丢失了原始错误信息！
    }
    // ...
}
```

### 6.3 错误处理策略

```go
// 策略1：基础设施层错误 → 转换为领域错误
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    user := &domain.User{}
    err := r.pool.QueryRow(ctx, sqlFindUserByEmail, email).Scan(...)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrUserNotFound  // 转换为领域可识别的错误
        }
        return nil, fmt.Errorf("userRepo.FindByEmail: %w", err)  // 其他错误向上传递
    }
    return user, nil
}

// 策略2：usecase 层检查领域错误，做业务决策
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, apperr.ErrUnauthorized.WithDetail("invalid email or password")
        }
        return nil, fmt.Errorf("AuthService.Login: %w", err)
    }
    // ...
}

// 策略3：transport 层统一转换为 HTTP 响应
func (h *AuthHandler) Login(c *gin.Context) {
    result, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        response.HandleError(c, err)  // HandleError 内部处理 AppError 转换
        return
    }
    response.OK(c, result)
}
```

---

## Chapter 7: 日志与可观测性规范 {#chapter-7}

### 7.1 日志使用规范

```go
// ✅ 正确：结构化日志，每个字段独立
logger.Info("user login success",
    zap.Int64("user_id", user.ID),
    zap.String("email", email),
    zap.String("ip", clientIP),
    zap.Duration("duration", time.Since(start)),
)

// ❌ 错误：字符串格式化（无法被日志系统解析）
logger.Info(fmt.Sprintf("user %d login success from %s", user.ID, clientIP))
```

**日志级别使用规范：**

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| `DEBUG` | 仅开发调试，生产禁止输出 | SQL 查询语句、中间计算结果 |
| `INFO` | 正常业务流程的关键节点 | 用户注册、游戏版本发布、支付成功 |
| `WARN` | 非预期但可继续运行的情况 | 缓存未命中、重试操作、配置降级 |
| `ERROR` | 需要关注的错误，影响单次请求 | 数据库查询失败、第三方 API 超时 |
| `FATAL` | 无法继续运行（仅启动阶段使用） | 数据库连接失败、配置文件缺失 |

**必须记录的业务事件（INFO 级别）：**

```go
// 认证相关
logger.Info("user.register", zap.String("email", email), zap.String("ip", ip))
logger.Info("user.login", zap.Int64("user_id", uid), zap.String("method", "email"))
logger.Info("user.logout", zap.Int64("user_id", uid))

// 游戏分发
logger.Info("release.published", zap.Int64("release_id", id), zap.String("version", ver))
logger.Info("download.initiated", zap.Int64("user_id", uid), zap.Int64("release_id", rid))

// 支付
logger.Info("order.created", zap.String("order_no", no), zap.Int("total_cents", total))
logger.Info("payment.success", zap.String("order_no", no), zap.String("method", method))
logger.Warn("payment.failed", zap.String("order_no", no), zap.String("reason", reason))

// 安全
logger.Warn("ratelimit.triggered", zap.String("ip", ip), zap.String("path", path))
logger.Warn("auth.failed", zap.String("email", email), zap.String("ip", ip))
logger.Error("auth.suspicious_login", zap.Int64("user_id", uid), zap.String("new_ip", ip))
```

### 7.2 Prometheus 指标规范

```go
// internal/pkg/metrics/metrics.go

var (
    // HTTP 请求指标
    HTTPRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "studio_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "path", "status"},
    )

    // 业务指标
    GameDownloadTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "studio_game_downloads_total",
            Help: "Total number of game download initiations",
        },
        []string{"game_id", "release_id", "status"},
    )

    OSTPlayTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "studio_ost_plays_total",
            Help: "Total number of OST track plays",
        },
        []string{"track_id", "album_id"},
    )

    ActiveWebSocketConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "studio_websocket_connections_active",
        Help: "Number of active WebSocket connections",
    })
)

// 在关键路径埋点
func (s *GameService) GetDownloadURL(ctx context.Context, ...) (string, error) {
    // ...
    if err != nil {
        metrics.GameDownloadTotal.WithLabelValues(gameID, releaseID, "failed").Inc()
        return "", err
    }
    metrics.GameDownloadTotal.WithLabelValues(gameID, releaseID, "success").Inc()
    return url, nil
}
```

---

## Chapter 8: 安全编码规范 {#chapter-8}

### 8.1 认证中间件实现规范

```go
// internal/transport/http/middleware/auth.go

func Auth(verifier TokenVerifier) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Authorization Header 提取 Bearer Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Unauthorized(c, apperr.ErrUnauthorized)
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
            response.Unauthorized(c, apperr.ErrUnauthorized.WithDetail("invalid authorization format"))
            c.Abort()
            return
        }

        claims, err := verifier.Verify(c.Request.Context(), parts[1])
        if err != nil {
            if errors.Is(err, apperr.ErrTokenExpired) {
                response.Unauthorized(c, apperr.ErrTokenExpired)
            } else {
                response.Unauthorized(c, apperr.ErrUnauthorized)
            }
            c.Abort()
            return
        }

        // 将认证信息存入 context（后续 Handler 可取用）
        c.Set(KeyUserID, claims.UserID)
        c.Set(KeyUserRole, claims.Role)
        c.Set(KeyUserPermissions, claims.Permissions)
        c.Next()
    }
}

// 辅助函数：安全地从 context 获取用户 ID
func GetUserID(c *gin.Context) (int64, bool) {
    uid, exists := c.Get(KeyUserID)
    if !exists {
        return 0, false
    }
    id, ok := uid.(int64)
    return id, ok
}
```

### 8.2 密码处理规范

```go
// internal/pkg/crypto/password.go

// 使用 Argon2id，当前 OWASP 推荐的最安全密码哈希算法
const (
    argonTime    = 3
    argonMemory  = 64 * 1024  // 64MB
    argonThreads = 4
    argonKeyLen  = 32
    saltLen      = 16
)

func HashPassword(password string) (string, error) {
    salt := make([]byte, saltLen)
    if _, err := rand.Read(salt); err != nil {
        return "", fmt.Errorf("generate salt: %w", err)
    }
    hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

    // 编码为标准格式：$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)
    return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
    // 解析并验证，使用 subtle.ConstantTimeCompare 防止时序攻击
    // ...
}
```

### 8.3 文件上传安全规范

```go
// ✅ 正确：文件上传安全检查清单
func (h *UploadHandler) GetPresignURL(c *gin.Context) {
    var req dto.PresignRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, apperr.ErrValidation)
        return
    }

    // 1. 验证文件类型（白名单，不信任客户端提交的 Content-Type）
    allowedTypes := map[string][]string{
        "avatar":     {"image/jpeg", "image/png", "image/webp"},
        "game_cover": {"image/jpeg", "image/png", "image/webp"},
        "audio":      {"audio/mpeg", "audio/flac", "audio/wav"},
        "game_pkg":   {"application/zip", "application/x-zip-compressed"},
    }
    allowed, ok := allowedTypes[req.Purpose]
    if !ok || !contains(allowed, req.ContentType) {
        response.BadRequest(c, apperr.ErrInvalidParam.WithDetail("unsupported file type"))
        return
    }

    // 2. 验证文件大小限制
    maxSizes := map[string]int64{
        "avatar":     5 * 1024 * 1024,       // 5MB
        "game_cover": 10 * 1024 * 1024,      // 10MB
        "audio":      500 * 1024 * 1024,     // 500MB
        "game_pkg":   10 * 1024 * 1024 * 1024, // 10GB
    }
    if req.FileSize > maxSizes[req.Purpose] {
        response.BadRequest(c, apperr.ErrInvalidParam.WithDetail("file too large"))
        return
    }

    // 3. 生成安全的对象 Key（不允许客户端自定义路径）
    uid, _ := middleware.GetUserID(c)
    key := fmt.Sprintf("%s/%d/%s/%s%s",
        req.Purpose,
        uid,
        time.Now().Format("20060102"),
        uuid.New().String(),
        filepath.Ext(req.Filename),
    )

    // 4. 生成预签名上传 URL（后端控制路径和元数据）
    presignURL, err := h.storage.GetPresignedUploadURL(c.Request.Context(), key, req.ContentType, req.FileSize, 15*time.Minute)
    if err != nil {
        response.HandleError(c, err)
        return
    }

    response.OK(c, gin.H{
        "upload_url": presignURL,
        "object_key": key,
        "expires_at": time.Now().Add(15 * time.Minute),
    })
}
```

### 8.4 限流中间件规范

```go
// internal/transport/http/middleware/ratelimit.go

// Token Bucket 算法实现（基于 Redis）
func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 识别限流 key（优先用 user_id，未认证用 IP）
        key := c.GetString(KeyUserID)
        if key == "" {
            key = "ip:" + c.ClientIP()
        }

        allowed, remaining, resetAt, err := limiter.Allow(c.Request.Context(), key, c.FullPath())
        if err != nil {
            // 限流器故障时放行（Fail Open，避免全站不可用）
            logger.Warn("rate limiter error, fail open", zap.Error(err))
            c.Next()
            return
        }

        // 设置限流相关响应头
        c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
        c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

        if !allowed {
            c.Header("Retry-After", strconv.FormatInt(time.Until(resetAt).Milliseconds()/1000+1, 10))
            response.TooManyRequests(c, apperr.ErrRateLimited)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## Chapter 9: 缓存使用规范 {#chapter-9}

### 9.1 缓存策略选择

| 场景 | 策略 | 说明 |
|------|------|------|
| 游戏详情页 | Cache-Aside | 读时检查 Redis，Miss 时查 DB 并写缓存 |
| 用户 Session | Write-Through | 写 Redis 同时写 DB（保证持久化） |
| 下载计数 | Write-Behind | 先写 Redis，定时批量刷入 DB |
| 排行榜 | 定时刷新 | 定时任务计算结果写入 Redis Sorted Set |
| 用户权限 | Cache-Aside + 主动失效 | 权限变更时立即删除缓存 |

### 9.2 Cache-Aside 标准实现

```go
// 模板：所有 Cache-Aside 实现必须遵循此结构
func (s *GameService) GetBySlug(ctx context.Context, slug string) (*domain.Game, error) {
    cacheKey := fmt.Sprintf("studio:game:slug:%s", slug)

    // 1. 尝试从缓存读取
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil {
        var game domain.Game
        if err := json.Unmarshal([]byte(cached), &game); err == nil {
            return &game, nil
        }
    }

    // 2. 防止缓存击穿：使用 singleflight 合并并发请求
    result, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
        // 3. 缓存未命中，查询数据库
        game, dbErr := s.gameRepo.FindBySlug(ctx, slug)
        if dbErr != nil {
            if errors.Is(dbErr, domain.ErrGameNotFound) {
                // 防止缓存穿透：对不存在的 key 也缓存空值（较短 TTL）
                _ = s.cache.Set(ctx, cacheKey, "null", 2*time.Minute)
            }
            return nil, dbErr
        }

        // 4. 写入缓存（TTL 加随机抖动，防止缓存雪崩）
        data, _ := json.Marshal(game)
        ttl := time.Hour + time.Duration(rand.Intn(300))*time.Second
        _ = s.cache.Set(ctx, cacheKey, string(data), ttl)

        return game, nil
    })

    if err != nil {
        return nil, err
    }
    return result.(*domain.Game), nil
}

// 5. 数据变更时主动失效缓存
func (s *GameService) Update(ctx context.Context, id int64, input UpdateGameInput) error {
    if err := s.gameRepo.Update(ctx, id, input); err != nil {
        return err
    }
    // 删除相关缓存（允许删除失败，缓存过期后自动更新）
    _ = s.cache.Del(ctx, fmt.Sprintf("studio:game:id:%d", id))
    _ = s.cache.Del(ctx, fmt.Sprintf("studio:game:slug:%s", input.Slug))
    return nil
}
```

### 9.3 分布式锁规范

```go
// 用于防止重复支付、重复发货等幂等操作
func (s *OrderService) ProcessPaymentCallback(ctx context.Context, orderNo string) error {
    lockKey := fmt.Sprintf("studio:lock:payment:%s", orderNo)

    // 获取分布式锁（最多等待 3 秒，锁持有 30 秒）
    lock, err := s.locker.Acquire(ctx, lockKey, 30*time.Second)
    if err != nil {
        return fmt.Errorf("acquire lock failed: %w", err)
    }
    defer lock.Release(ctx)

    // 检查订单当前状态（幂等检查）
    order, err := s.orderRepo.FindByOrderNo(ctx, orderNo)
    if err != nil {
        return err
    }
    if order.Status != domain.OrderStatusPendingPayment {
        // 已处理过，直接返回成功（幂等）
        return nil
    }

    // 处理支付成功逻辑
    return s.fulfillOrder(ctx, order)
}
```

---

## Chapter 10: 测试规范 {#chapter-10}

### 10.1 测试分层策略

```
单元测试 (Unit Tests)
├── 覆盖 usecase 层所有业务逻辑
├── 覆盖 domain 层所有实体方法
├── 使用 mock 替代所有外部依赖（数据库、缓存、OSS）
└── 要求覆盖率 ≥ 80%

集成测试 (Integration Tests)
├── 覆盖 infra 层（真实数据库、真实 Redis）
├── 测试 SQL 查询正确性
└── 使用 TestContainers 管理测试数据库生命周期

端到端测试 (E2E Tests)
├── 覆盖核心业务流程（注册→登录→下载→支付）
└── 针对 HTTP API 接口（使用 httptest 包）
```

### 10.2 单元测试规范

```go
// internal/usecase/user/login_test.go

// ✅ 正确：表驱动测试
func TestAuthService_Login(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        password string
        setup   func(mockRepo *mocks.UserRepository)
        wantErr error
    }{
        {
            name:     "success",
            email:    "test@example.com",
            password: "correctPassword",
            setup: func(mockRepo *mocks.UserRepository) {
                mockRepo.On("FindByEmail", mock.Anything, "test@example.com").
                    Return(&domain.User{
                        ID:       1,
                        Email:    "test@example.com",
                        Password: mustHashPassword("correctPassword"),
                        Status:   domain.UserStatusActive,
                    }, nil)
            },
            wantErr: nil,
        },
        {
            name:     "user not found",
            email:    "notexist@example.com",
            password: "anypassword",
            setup: func(mockRepo *mocks.UserRepository) {
                mockRepo.On("FindByEmail", mock.Anything, "notexist@example.com").
                    Return(nil, domain.ErrUserNotFound)
            },
            wantErr: apperr.ErrUnauthorized,
        },
        {
            name:     "wrong password",
            email:    "test@example.com",
            password: "wrongPassword",
            setup: func(mockRepo *mocks.UserRepository) {
                mockRepo.On("FindByEmail", mock.Anything, "test@example.com").
                    Return(&domain.User{
                        Password: mustHashPassword("correctPassword"),
                        Status:   domain.UserStatusActive,
                    }, nil)
            },
            wantErr: apperr.ErrUnauthorized,
        },
        {
            name:     "banned user",
            email:    "banned@example.com",
            password: "password",
            setup: func(mockRepo *mocks.UserRepository) {
                mockRepo.On("FindByEmail", mock.Anything, "banned@example.com").
                    Return(&domain.User{
                        Password: mustHashPassword("password"),
                        Status:   domain.UserStatusBanned,
                    }, nil)
            },
            wantErr: apperr.ErrForbidden,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &mocks.UserRepository{}
            tt.setup(mockRepo)

            svc := NewAuthService(mockRepo, testTokenSigner, testLogger)
            _, err := svc.Login(context.Background(), tt.email, tt.password)

            if tt.wantErr != nil {
                assert.True(t, errors.Is(err, tt.wantErr),
                    "expected error %v, got %v", tt.wantErr, err)
            } else {
                assert.NoError(t, err)
            }
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### 10.3 集成测试规范

```go
// internal/infra/postgres/user_repo_test.go

// 使用 TestContainers 启动真实 PostgreSQL
func TestMain(m *testing.M) {
    ctx := context.Background()
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:16-alpine",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "studio_test",
            },
            WaitingFor: wait.ForListeningPort("5432/tcp"),
        },
        Started: true,
    })
    // ...运行迁移，执行测试，清理容器
    os.Exit(m.Run())
}

func TestUserRepo_FindByEmail(t *testing.T) {
    // 每个测试用例使用独立事务，测试完成后回滚（保持数据隔离）
    tx, _ := testDB.Begin(context.Background())
    t.Cleanup(func() { tx.Rollback(context.Background()) })

    repo := NewUserRepoWithTx(tx)

    // 准备测试数据
    _, err := tx.Exec(context.Background(),
        `INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)`,
        "testuser", "test@example.com", "hashed", "player")
    require.NoError(t, err)

    // 执行被测代码
    user, err := repo.FindByEmail(context.Background(), "test@example.com")
    require.NoError(t, err)
    assert.Equal(t, "testuser", user.Username)

    // 测试不存在的情况
    _, err = repo.FindByEmail(context.Background(), "notexist@example.com")
    assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}
```

### 10.4 测试文件命名与组织规范

```
文件命名：被测文件名 + _test.go
  user.go → user_test.go

测试函数命名：Test{结构体名}_{方法名}_{场景}
  TestAuthService_Login_Success
  TestAuthService_Login_WrongPassword
  TestUserRepo_FindByID_NotFound

Benchmark 命名：Benchmark{结构体名}_{方法名}
  BenchmarkGameService_GetBySlug
```

---

## Chapter 11: 前端开发规范（Next.js/TypeScript）{#chapter-11}

### 11.1 TypeScript 使用规范

```typescript
// ✅ 正确：明确的类型定义
interface GameRelease {
  id: number;
  version: string;
  title: string;
  changelog: string;
  publishedAt: string; // ISO 8601 格式
  fileSize: number;    // 字节数
}

// ❌ 禁止：使用 any
const game: any = await fetchGame(id);

// ✅ 正确：使用 unknown 后做类型断言
const rawResponse: unknown = await fetchGame(id);
const game = rawResponse as GameResponse; // 或使用 Zod 验证

// ✅ 正确：API 响应统一类型
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  requestId: string;
  timestamp: number;
}

// ✅ 正确：严格的 null 处理
function formatFileSize(bytes: number | null | undefined): string {
  if (bytes == null) return '未知';
  // ...
}
```

### 11.2 组件规范

```tsx
// ✅ 正确：组件文件结构
// 1. 导入（第三方 → 内部模块 → 类型）
import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';

import { GameCard } from '@/components/ui/GameCard';
import { useGameList } from '@/hooks/useGameList';
import { formatDate } from '@/utils/date';

import type { Game } from '@/types/game';

// 2. Props 类型定义（与组件名紧挨着）
interface GameListProps {
  initialGames?: Game[];
  category?: string;
  className?: string;
}

// 3. 组件主体（函数式组件，named export）
export function GameList({ initialGames, category, className }: GameListProps) {
  const { games, isLoading, error } = useGameList({ initialGames, category });

  if (error) return <ErrorState error={error} />;
  if (isLoading) return <LoadingSkeleton count={6} />;
  if (!games.length) return <EmptyState message="暂无游戏" />;

  return (
    <ul className={cn('grid grid-cols-1 md:grid-cols-3 gap-6', className)}>
      {games.map((game) => (
        <li key={game.id}>
          <GameCard game={game} />
        </li>
      ))}
    </ul>
  );
}
```

### 11.3 数据获取规范

```typescript
// ✅ 正确：Server Component 数据获取（Next.js App Router）
// app/games/[slug]/page.tsx
export default async function GameDetailPage({ params }: { params: { slug: string } }) {
  // 服务端直接获取数据，SEO 友好
  const game = await gameApi.getBySlug(params.slug);
  if (!game) notFound();

  return <GameDetail game={game} />;
}

// ✅ 正确：Client Component 使用 React Query
// components/AlbumPlayer.tsx
'use client';

export function AlbumPlayer({ albumId }: { albumId: number }) {
  const { data: playlist, isLoading } = useQuery({
    queryKey: ['album', albumId, 'playlist'],
    queryFn: () => albumApi.getPlaylist(albumId),
    staleTime: 30 * 60 * 1000, // 30分钟内不重新请求
  });

  // ...
}

// ✅ 正确：API 客户端封装
// lib/api/game.ts
export const gameApi = {
  async getBySlug(slug: string): Promise<Game | null> {
    const res = await fetch(`${API_BASE}/games/${slug}`, {
      next: { revalidate: 3600 }, // ISR：每小时重新验证
    });
    if (res.status === 404) return null;
    if (!res.ok) throw new ApiError(await res.json());
    const json: ApiResponse<Game> = await res.json();
    return json.data;
  },
};
```

### 11.4 错误边界与加载状态规范

```tsx
// ✅ 正确：每个重要页面必须有 error.tsx 和 loading.tsx

// app/games/error.tsx
'use client';
export default function GamesError({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div className="text-center py-20">
      <h2 className="text-xl font-semibold">加载失败</h2>
      <p className="text-muted-foreground mt-2">{error.message}</p>
      <button onClick={reset} className="mt-4 btn-primary">重试</button>
    </div>
  );
}

// app/games/loading.tsx
export default function GamesLoading() {
  return <GameListSkeleton count={9} />;
}
```

---

## Chapter 12: Git 工作流与提交规范 {#chapter-12}

### 12.1 分支策略（Git Flow 简化版）

```
main          生产分支（受保护，只能通过 PR 合入）
  └── develop  开发集成分支
        ├── feature/epic1-user-auth     功能分支
        ├── feature/epic3-game-download 功能分支
        ├── fix/download-url-expiry     Bug 修复分支
        └── hotfix/payment-callback     紧急修复（从 main 切出）
```

**分支命名规范：**
- 功能：`feature/{epic-id}-{short-description}`（如 `feature/epic2-oauth-login`）
- 修复：`fix/{issue-id}-{short-description}`（如 `fix/123-comment-pagination`）
- 紧急修复：`hotfix/{description}`（如 `hotfix/payment-double-charge`）
- 发布：`release/v{semver}`（如 `release/v1.2.0`）

### 12.2 提交信息规范（Conventional Commits）

```
格式：<type>(<scope>): <subject>

type（必填）：
  feat     新功能
  fix      Bug 修复
  perf     性能优化
  refactor 重构（不改变功能）
  test     测试相关
  docs     文档更新
  style    代码格式（不影响功能）
  chore    构建/工具/依赖更新
  ci       CI/CD 配置更新
  revert   回滚某次提交

scope（可选，对应模块）：
  auth, user, game, album, order, payment, comment, notification, admin

subject（必填）：
  - 使用中文（本项目主要中文交流）
  - 现在时态，动词开头
  - 不超过 72 字符
  - 末尾不加句号

示例：
  feat(auth): 实现 JWT 签发与 Refresh Token 机制
  fix(game): 修复下载链接过期时间计算错误
  perf(comment): 使用 singleflight 优化评论首页缓存穿透
  feat(payment): 接入微信支付 Native 扫码模式
  test(user): 补充用户注册流程的集成测试
  chore(deps): 升级 pgx 至 v5.6.0
```

### 12.3 Pull Request 规范

每个 PR 必须包含：

```markdown
## 关联任务
- plan.md [Task X.X.X]

## 改动说明
（用 2-3 句话说明改了什么，为什么这样改）

## 改动类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 性能优化
- [ ] 重构
- [ ] 配置/文档

## 测试情况
- [ ] 新增单元测试，覆盖率 ≥ 80%
- [ ] 已运行 `go test ./...`，全部通过
- [ ] 手动测试场景：（描述手动测试步骤）

## 数据库变更
- [ ] 无
- [ ] 有（迁移文件：`migrations/YYYYMMDD_xxx.up.sql`）

## 破坏性变更
- [ ] 无
- [ ] 有（描述影响范围和迁移方案）
```

---

## Chapter 13: 自动化开发流程 (SOP) {#chapter-13}

### 13.1 新功能开发 SOP

```
Step 1: 需求确认
  └─ 阅读 plan.md 对应的 Task 描述
  └─ 确认输入/输出、边界条件、错误场景
  └─ 若有歧义，在开始编码前先提问

Step 2: 数据库层（如需）
  └─ 编写迁移 SQL（up.sql + down.sql）
  └─ 在本地执行迁移，验证无错误
  └─ 设计 Repository 接口（internal/domain/xxx/repository.go）

Step 3: 领域层
  └─ 定义实体 struct 和领域方法（internal/domain/xxx/entity.go）
  └─ 定义领域错误（若有新增）

Step 4: 基础设施层
  └─ 实现 Repository 接口（internal/infra/postgres/xxx_repo.go）
  └─ 为 Repository 实现编写集成测试

Step 5: 用例层
  └─ 编写业务逻辑（internal/usecase/xxx/yyy.go）
  └─ 编写单元测试（所有测试必须通过）

Step 6: 传输层
  └─ 定义请求/响应 DTO（internal/transport/http/dto/xxx_dto.go）
  └─ 实现 HTTP Handler（internal/transport/http/handler/xxx.go）
  └─ 注册路由（internal/transport/http/router.go）

Step 7: 验证
  └─ 运行 go vet ./...
  └─ 运行 golangci-lint run
  └─ 运行 go test ./...
  └─ 手动 curl 测试核心接口
  └─ 检查日志输出是否符合规范

Step 8: 提交
  └─ 按 Conventional Commits 规范提交
  └─ 创建 PR，填写 PR 模板
```

### 13.2 Bug 修复 SOP

```
Step 1: 复现问题
  └─ 编写能复现 Bug 的测试用例（先让测试失败）

Step 2: 定位根因
  └─ 阅读错误日志中的 request_id，追踪完整链路
  └─ 使用 EXPLAIN ANALYZE 排查慢查询
  └─ 检查并发场景（是否有竞态条件）

Step 3: 修复
  └─ 最小改动原则（只改必要的代码）
  └─ 不要在修 Bug 的同时做重构

Step 4: 验证
  └─ Step 1 的测试用例现在必须通过
  └─ 运行全量测试，确保无回归

Step 5: 提交
  └─ type 使用 fix，描述 Bug 现象（不是修复方案）
```

### 13.3 数据库变更 SOP

```
Step 1: 编写迁移文件
  └─ 文件名包含时间戳和描述
  └─ up.sql 和 down.sql 成对存在

Step 2: 本地验证
  └─ 执行 migrate up，验证成功
  └─ 执行 migrate down，验证成功回滚
  └─ 再次 migrate up，验证幂等性

Step 3: 性能评估
  └─ 若在大表上加索引，需评估是否使用 CONCURRENTLY
  └─ ALTER TABLE 操作需评估锁影响

Step 4: 上线策略
  └─ 新增表/列：直接迁移（无影响）
  └─ 删除列：先代码不用该列（部署）→ 再删除（两步）
  └─ 修改列类型：需评估数据迁移计划，可能需要双写过渡期
```

### 13.4 新增 API 接口 SOP

```
Step 1: 更新 OpenAPI 规范
  └─ 先在 docs/openapi.yaml 中定义接口契约
  └─ 包含：路径、方法、请求参数、响应结构、错误码

Step 2: 生成 API 客户端代码（前端）
  └─ 运行 openapi-typescript-codegen 生成 TypeScript 客户端

Step 3: 实现后端接口（按 13.1 SOP 进行）

Step 4: 验证契约一致性
  └─ 实现的接口必须与 openapi.yaml 完全一致
```

---

## Chapter 14: 代码审查清单 {#chapter-14}

### 14.1 AI 自审清单（提交前必须过一遍）

**功能正确性：**
```
[ ] 所有 if/else 分支都已处理（包括 null/nil/empty 情况）
[ ] 边界值已测试（0, 1, MAX, -1）
[ ] 并发场景已考虑（多个 goroutine 同时执行是否安全）
[ ] 幂等性已保证（重复调用是否产生副作用）
```

**安全性：**
```
[ ] 所有 SQL 使用了参数化查询
[ ] 用户输入已经过验证和清洗
[ ] 认证检查未被跳过
[ ] 敏感信息未出现在日志中（密码、Token、支付信息）
[ ] 文件上传已验证类型和大小
[ ] OSS Key 未暴露给客户端（返回的是签名 URL）
```

**性能：**
```
[ ] 数据库查询有对应的索引
[ ] 未在循环中执行数据库查询（N+1 问题）
[ ] 缓存策略已考虑
[ ] 大列表使用分页而非全量获取
```

**代码质量：**
```
[ ] 函数长度 ≤ 50 行
[ ] 所有 error 已被处理（无 _ 忽略）
[ ] 所有 goroutine 有 recover 和退出机制
[ ] 无魔法数字（使用命名常量）
[ ] 无重复代码（DRY 原则）
[ ] 代码可读性好（命名有意义，逻辑清晰）
```

**测试：**
```
[ ] 核心逻辑有单元测试覆盖
[ ] 错误路径有测试（不只测试 happy path）
[ ] 测试相互独立（不依赖执行顺序）
```

---

## Chapter 15: 常见模式与反模式 {#chapter-15}

### 15.1 推荐使用的模式

**模式1：Options Pattern（可选参数）**

```go
// 适用于有多个可选参数的构造函数
type ListGamesOptions struct {
    Status   string
    Genre    []string
    Sort     string
    Limit    int
    Cursor   string
}

type ListGamesOption func(*ListGamesOptions)

func WithStatus(status string) ListGamesOption {
    return func(o *ListGamesOptions) { o.Status = status }
}

func WithGenre(genre ...string) ListGamesOption {
    return func(o *ListGamesOptions) { o.Genre = genre }
}

// 使用
games, err := svc.ListGames(ctx,
    WithStatus("published"),
    WithGenre("visual_novel", "music"),
)
```

**模式2：Repository 接口 + Mock 测试**

```go
// domain/game/repository.go
type GameRepository interface {
    FindByID(ctx context.Context, id int64) (*Game, error)
    FindBySlug(ctx context.Context, slug string) (*Game, error)
    List(ctx context.Context, opts ListGamesOptions) ([]*Game, string, error)
    Save(ctx context.Context, game *Game) error
    Delete(ctx context.Context, id int64) error
}

// 使用 mockery 自动生成 Mock：
// go generate mockery --name=GameRepository --dir=internal/domain/game
```

**模式3：Singleflight 防缓存击穿**

```go
type GameService struct {
    sfGroup singleflight.Group
}

func (s *GameService) GetBySlug(ctx context.Context, slug string) (*Game, error) {
    result, err, _ := s.sfGroup.Do(slug, func() (interface{}, error) {
        return s.loadFromDBAndCache(ctx, slug)
    })
    if err != nil {
        return nil, err
    }
    return result.(*Game), nil
}
```

### 15.2 禁止使用的反模式

**反模式1：胖 Handler（业务逻辑放在 Handler 里）**

```go
// ❌ 错误：Handler 里直接写业务逻辑
func (h *GameHandler) GetDownloadURL(c *gin.Context) {
    releaseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    userID := c.GetInt64("user_id")

    // ❌ 直接查数据库
    release := &domain.GameRelease{}
    h.db.Where("id = ?", releaseID).First(release)

    // ❌ 直接检查权限
    var asset domain.UserGameAsset
    result := h.db.Where("user_id = ? AND game_id = ?", userID, release.GameID).First(&asset)
    if result.Error != nil {
        c.JSON(403, gin.H{"error": "no permission"})
        return
    }

    // ❌ 直接调用 OSS
    url := fmt.Sprintf("https://oss.aliyuncs.com/%s?token=xxx", release.OSSKey)
    c.JSON(200, gin.H{"url": url})
}
```

**反模式2：魔法数字**

```go
// ❌ 错误
if user.FailedLoginCount > 5 {
    lock(user.ID, 900)
}

// ✅ 正确
const (
    MaxLoginFailAttempts  = 5
    LoginLockDuration     = 15 * time.Minute
)
if user.FailedLoginCount > MaxLoginFailAttempts {
    lockUser(ctx, user.ID, LoginLockDuration)
}
```

**反模式3：日志中记录敏感信息**

```go
// ❌ 绝对禁止
logger.Info("user login", zap.String("password", password))
logger.Info("payment", zap.String("card_number", cardNo))
logger.Info("auth", zap.String("token", accessToken))

// ✅ 正确：只记录非敏感信息
logger.Info("user login attempt", zap.String("email", maskEmail(email)), zap.String("ip", ip))
```

**反模式4：不处理 goroutine 的 panic**

```go
// ❌ 错误：goroutine 内的 panic 会导致整个服务崩溃
go func() {
    sendEmail(user, template)  // 如果这里 panic，服务挂了
}()

// ✅ 正确
go func() {
    defer recoverWithLog("sendEmail goroutine")
    sendEmail(user, template)
}()
```

---

## Chapter 16: 文件结构与命名规范 {#chapter-16}

### 16.1 Go 文件命名规范

| 文件类型 | 命名规范 | 示例 |
|---------|---------|------|
| 实体定义 | `{entity}.go` | `user.go`, `game.go` |
| 仓储接口 | `repository.go` | `internal/domain/user/repository.go` |
| 仓储实现 | `{entity}_repo.go` | `internal/infra/postgres/user_repo.go` |
| 用例实现 | `{action}.go` | `register.go`, `login.go` |
| HTTP Handler | `{entity}_handler.go` 或 `handler.go` | `game_handler.go` |
| 中间件 | `{name}.go` | `auth.go`, `ratelimit.go` |
| DTO | `{entity}_dto.go` | `user_dto.go` |
| 测试文件 | `{file}_test.go` | `user_test.go` |
| Mock 文件 | `mock_{interface}.go` | `mock_user_repository.go` |

### 16.2 数据库迁移文件命名

```
{14位时间戳}_{动词}_{描述}.{up|down}.sql

时间戳格式：YYYYMMDDHHmmss
动词：create, add, drop, alter, rename, update

示例：
20260101120001_create_users.up.sql
20260101120001_create_users.down.sql
20260115090030_add_users_avatar_key.up.sql
20260115090030_add_users_avatar_key.down.sql
20260201140000_create_games.up.sql
```

### 16.3 配置文件与环境变量命名

```yaml
# config.yaml 使用 snake_case
database:
  dsn: "..."
  max_open_conns: 20

# 对应环境变量使用 SCREAMING_SNAKE_CASE，并以 STUDIO_ 为前缀
STUDIO_DATABASE_DSN=postgres://...
STUDIO_DATABASE_MAX_OPEN_CONNS=20
STUDIO_REDIS_ADDR=localhost:6379
STUDIO_JWT_SECRET=your-secret-key
STUDIO_OSS_ACCESS_KEY_ID=xxx
STUDIO_OSS_ACCESS_KEY_SECRET=xxx
```

---

## Chapter 17: 性能工程规范 {#chapter-17}

### 17.1 数据库性能规范

**N+1 查询检测与修复：**

```go
// ❌ N+1 问题：每条游戏都查一次数据库获取最新版本
games, _ := gameRepo.List(ctx, opts)
for _, game := range games {
    game.LatestRelease, _ = releaseRepo.FindLatest(ctx, game.ID)  // N 次额外查询！
}

// ✅ 正确：一次批量查询
games, _ := gameRepo.List(ctx, opts)
gameIDs := extractIDs(games)
releases, _ := releaseRepo.FindLatestByGameIDs(ctx, gameIDs)  // 1 次查询
releaseMap := indexByGameID(releases)
for _, game := range games {
    game.LatestRelease = releaseMap[game.ID]
}
```

**连接池配置规范：**

```go
poolConfig.MaxConns = int32(runtime.NumCPU() * 4)  // CPU 核心数 × 4
poolConfig.MinConns = int32(runtime.NumCPU())
poolConfig.MaxConnLifetime = time.Hour
poolConfig.MaxConnIdleTime = 30 * time.Minute
poolConfig.HealthCheckPeriod = time.Minute
```

### 17.2 API 响应时间预算

| 接口分类 | P99 目标 | 关键优化点 |
|---------|---------|-----------|
| 游戏列表 | 50ms | Redis 缓存 + 只查必要字段 |
| 游戏详情 | 30ms | Redis 缓存（TTL 1 小时） |
| 用户登录 | 200ms | Argon2 密码验证耗时（预期） |
| 下载链接生成 | 100ms | OSS SDK 调用耗时 |
| 评论列表 | 50ms | Redis 缓存首页 + 游标分页 |
| 搜索 | 200ms | PostgreSQL 全文索引 |
| 支付发起 | 500ms | 第三方支付 API 调用 |

### 17.3 内存使用规范

```go
// ✅ 正确：大文件处理使用流式，不加载到内存
func (h *UploadHandler) StreamUpload(c *gin.Context) {
    // 不要用 c.Request.Body 读取全部内容到 []byte
    // 而是将 reader 直接流式传输到 OSS
    if err := storage.UploadStream(ctx, key, c.Request.Body, c.Request.ContentLength); err != nil {
        // ...
    }
}

// ✅ 正确：查询大量数据时使用游标迭代，不一次性 Load 到内存
func (s *ExportService) ExportUsers(ctx context.Context, writer io.Writer) error {
    rows, err := pool.Query(ctx, `SELECT id, username, email FROM users ORDER BY id`)
    if err != nil {
        return err
    }
    defer rows.Close()

    enc := json.NewEncoder(writer)
    for rows.Next() {
        var user UserExport
        if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
            return err
        }
        enc.Encode(user)  // 逐行写出，不积累在内存
    }
    return rows.Err()
}
```

---

## Chapter 18: AI 任务执行模板 {#chapter-18}

### 18.1 实现新接口模板

当 AI 被要求实现一个新的 API 接口时，按以下顺序输出：

```
1. 【确认理解】
   - 接口路径：GET/POST/PUT/DELETE /api/v1/xxx
   - 对应 plan.md：[Task X.X.X]
   - 业务逻辑：（用一段话描述实现逻辑）
   - 需要的假设：（列出任何不确定的点）

2. 【数据库变更】（如需）
   - 新增/修改的表
   - 迁移文件内容

3. 【代码实现顺序】
   a. domain 层（实体、接口）
   b. infra 层（数据库实现）
   c. usecase 层（业务逻辑）
   d. transport 层（Handler、DTO、路由）

4. 【测试建议】
   - 单元测试的关键测试用例
   - 手动测试的 curl 命令

5. 【注意事项】
   - 潜在性能影响
   - 安全注意事项
   - 待处理的 edge case
```

### 18.2 调试问题模板

当 AI 被要求排查问题时：

```
1. 【问题复现】
   - 复现步骤
   - 期望行为 vs 实际行为
   - 错误日志（如有）

2. 【排查思路】
   - 按照请求链路逐层排查：transport → usecase → infra
   - 检查日志中的 request_id 追踪完整链路
   - 检查是否有数据库慢查询

3. 【根因定位】
   - 定位到具体文件和行号
   - 解释为什么会出现这个问题

4. 【修复方案】
   - 最小改动的修复方案
   - 是否需要数据库迁移
   - 是否需要清除缓存

5. 【预防措施】
   - 是否需要添加测试防止回归
   - 是否需要增加监控告警
```

### 18.3 代码重构模板

当 AI 被要求重构代码时（注意：仅在明确要求时重构）：

```
1. 【重构范围确认】
   - 明确哪些文件会被修改
   - 明确哪些功能的行为不会改变

2. 【重构前状态】
   - 当前代码存在的问题（性能、可维护性、技术债务）

3. 【重构后状态】
   - 改进点
   - 是否有 API 变化（不允许有，除非明确要求）

4. 【测试保障】
   - 重构前必须有完整测试（如果没有，先补测试再重构）
   - 重构后所有测试必须通过

5. 【风险评估】
   - 对其他模块的影响
   - 是否需要数据迁移
```

---

## 附录 A: 常用工具命令速查

```bash
# 代码格式化
gofmt -w ./...
goimports -w ./...

# 代码检查
golangci-lint run --timeout 5m

# 运行测试
go test ./... -race -count=1
go test ./... -run TestAuthService_Login  # 运行指定测试

# 测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 数据库迁移
migrate -path migrations -database "${DATABASE_DSN}" up
migrate -path migrations -database "${DATABASE_DSN}" down 1

# 生成 Mock
go generate ./internal/domain/...

# 查看编译依赖
go mod graph | head -50

# 构建
go build -ldflags="-w -s" -o bin/server ./cmd/server

# Docker 构建
docker build -t studio-server:dev .
docker-compose up -d  # 本地开发环境

# 压力测试
k6 run scripts/load_test.js

# 查看服务指标
curl http://localhost:8080/metrics

# Prometheus 端口转发（K8s）
kubectl port-forward svc/prometheus 9090:9090
```

---

## 附录 B: 错误码速查表

| 错误码 | 含义 | HTTP 状态码 |
|-------|------|------------|
| 0 | 成功 | 200 |
| 40001 | 请求参数有误 | 400 |
| 40002 | 参数验证失败 | 400 |
| 40101 | 未认证/未登录 | 401 |
| 40102 | Token 已过期 | 401 |
| 40301 | 权限不足 | 403 |
| 40401 | 用户不存在 | 404 |
| 40402 | 游戏不存在 | 404 |
| 40403 | 版本不存在 | 404 |
| 40901 | 邮箱已被注册 | 409 |
| 40902 | 用户名已被占用 | 409 |
| 42901 | 请求过于频繁 | 429 |
| 50001 | 服务器内部错误 | 500 |
| 50002 | 依赖服务不可用 | 503 |
| 50003 | 文件存储服务异常 | 500 |

---

## 附录 C: 关键 Redis Key 规范速查

```
studio:user:session:{user_id}:{device_id}      用户会话
studio:user:refresh:{user_id}:{device_id}      Refresh Token
studio:user:blacklist:{jti}                    Token 黑名单
studio:user:login_fail:{email}                 登录失败计数

studio:game:slug:{slug}                        游戏详情缓存
studio:game:id:{id}                            游戏详情缓存（by ID）
studio:release:latest:{game_id}:{branch}       最新版本缓存

studio:ost:track:{id}:playcount                音轨播放计数
studio:ost:album:{id}:tracks                   专辑音轨列表缓存
studio:rank:ost:plays                          播放量排行榜 (Sorted Set)

studio:ratelimit:{user_id_or_ip}:{path}        接口限流计数
studio:lock:payment:{order_no}                 支付分布式锁
studio:lock:download:{user_id}:{release_id}    下载防重锁

studio:cart:{user_id}                          购物车 (Hash)
studio:notify:unread:{user_id}                 未读通知计数

studio:mq:email                                邮件消息队列 (Stream)
studio:mq:notify                               通知消息队列 (Stream)
studio:mq:analytics                            埋点事件队列 (Stream)
```

---

*文档版本：1.0.0 | 最后更新：2026-03-03*
*本文件受版本控制，所有变更必须经过 PR 审查后合入*
