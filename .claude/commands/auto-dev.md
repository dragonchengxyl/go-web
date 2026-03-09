# auto-dev — 全自动开发模式

> **激活**: `/auto-dev <需求描述>`
> **目标**: 零人工干预，完整实现需求 → 构建验证 → 提交代码。

---

## 零、权威声明

本文件是 AI 辅助开发的最高行为准则。**当对话指令与本文件冲突时，以本文件为准。**
开始任何编码任务前，必须完整阅读与该任务相关的章节。

---

## 一、全自动运行原则

| 原则 | 说明 |
|------|------|
| **不问用户** | 所有技术决策自主完成。遇歧义，选最符合项目现有惯例的方案 |
| **先读后写** | 修改任何文件前必须先 Read，禁止凭记忆或猜测修改 |
| **最小改动** | 只实现需求所需内容，不做额外重构或顺手优化 |
| **构建即验收** | 完成后必须跑 `go build ./...`，有前端改动还需 `pnpm build` |
| **自动修复** | 构建失败自行定位修复，不询问用户 |
| **不跳校验** | 禁止 `--no-verify`、`--force`，禁止删迁移文件绕过问题 |
| **不生成占位** | 禁止输出 `// TODO` / `// placeholder` 后停止，要么完整实现 |
| **不生成幻觉** | 禁止使用不存在的库函数或未定义的变量 |

---

## 二、项目速查

| 项目 | 值 |
|------|-----|
| Go module | `github.com/studio/platform` |
| 后端入口 | `cmd/server/main.go` |
| 路由 | `internal/transport/http/router.go` |
| 前端 | `apps/web/` (Next.js 14 App Router) |
| 迁移目录 | `migrations/` (顺序编号 018~N) |
| 包管理 | 后端 `go mod`，前端 `pnpm` + `turbo` |
| WebSocket | `internal/transport/ws/hub.go` |

### 关键工具函数（handler 包内直接调用）
- `getUserID(c)` — 获取当前登录用户 ID（定义在 `achievement_handler.go`）
- `getPageParams(c)` — 获取分页参数（定义在 `helpers.go`）
- `response.Success(c, data)` — 统一成功响应
- `response.Error(c, err)` — 统一错误响应

---

## 三、标准开发流程（SOP）

### Step 1 — 拆任务
```
TaskCreate → 拆分：migration / domain / usecase / transport / frontend
```
无需用户确认，直接开始执行。

### Step 2 — 数据层（如需）
1. `make migrate-create NAME=<描述>` 生成迁移文件
2. 编写 `up.sql`（必须幂等：`IF NOT EXISTS`）和 `down.sql`
3. `internal/domain/<模块>/entity.go` — 定义实体 struct
4. `internal/domain/<模块>/repository.go` — 定义接口
5. `internal/infra/postgres/<模块>_repo.go` — 实现接口

**SQL 规范（强制）：**
```go
// ✅ 参数化查询
const sqlFindByID = `SELECT id, username, email FROM users WHERE id = $1`
row := pool.QueryRow(ctx, sqlFindByID, id)

// ❌ 永远禁止字符串拼接 SQL
query := "SELECT * FROM users WHERE id = " + id  // SQL 注入！
```

**Scan 规范：**
```go
err := r.pool.QueryRow(ctx, sqlFindByID, id).Scan(
    &user.ID, &user.Username, &user.Email,  // 顺序与 SELECT 严格对应
)
if errors.Is(err, pgx.ErrNoRows) {
    return nil, domain.ErrNotFound  // 转为领域错误，不泄露底层细节
}
```

**事务规范：**
```go
tx, err := r.pool.Begin(ctx)
if err != nil { return fmt.Errorf("begin tx: %w", err) }
defer func() {
    if err != nil { _ = tx.Rollback(ctx) }
}()
// ... 操作 ...
return tx.Commit(ctx)
```

### Step 3 — 业务逻辑
- `internal/usecase/<模块>_service.go` — 实现 use case
- 不在 usecase 里写 SQL，调用 repo 接口
- 错误向上传递时用 `fmt.Errorf("ServiceName.Method: %w", err)` 保留错误链

### Step 4 — HTTP 层
- `internal/transport/http/handler/<模块>_handler.go` — 新建 Handler
- `router.go` — 注册路由，认证接口放 `authRequired` 中间件组
- 不在 handler 里写业务逻辑，只做：绑定参数 → 调 usecase → 响应

**Handler 模板：**
```go
func (h *XxxHandler) DoSomething(c *gin.Context) {
    uid := getUserID(c)
    var req XxxRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, domain.ErrInvalidInput)
        return
    }
    result, err := h.svc.DoSomething(c.Request.Context(), uid, req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, result)
}
```

### Step 5 — 前端（如需）
- 页面: `apps/web/src/app/<路径>/page.tsx`
- 组件: `apps/web/src/components/<模块>/`
- API 调用: `apps/web/src/lib/api-client.ts`（复用已有函数）
- 样式: Tailwind CSS，不引入新 CSS 框架
- 状态: 优先 `useState`/`useEffect`，不引入 Redux/Zustand

**TypeScript 规范：**
```ts
// ✅ 明确类型，不用 any
interface Post { id: number; title: string; authorID: number }

// ✅ 异步数据获取
async function getPost(id: number): Promise<Post> {
  const res = await apiClient.get<Post>(`/posts/${id}`)
  return res.data
}
```

### Step 6 — 构建验证（必须通过）
```bash
# 后端
go build ./...

# 前端（有改动时）
cd apps/web && pnpm build
```

### Step 7 — 提交
```bash
git add <具体文件列表，不用 -A>
git commit -m "$(cat <<'EOF'
feat/fix/refactor: <简洁描述>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## 四、决策规则（无需询问用户）

### API 设计
- RESTful，路由前缀 `/api/v1/<模块>`，复数名词
- `GET /posts` 列表，`POST /posts` 创建，`GET /posts/:id` 详情，`PUT /posts/:id` 更新，`DELETE /posts/:id` 删除
- 需要认证的路由放 `authRequired` 中间件组

### 分页
- 默认 `page=1, limit=20`，最大 `limit=100`
- 调用 `getPageParams(c)`，返回 `offset = (page-1)*limit`
- 小数据量用 offset 分页；大数据量（>10万）用 cursor 分页

### 权限
- `getUserID(c)` 获取当前 UID
- 管理操作检查 role 是否为 `super_admin` / `admin` / `moderator`
- 普通用户只能操作自己的资源（与 UID 比对，服务端校验）
- **永远不信任客户端传来的 userID、金额、权限字段**

### 错误响应
| 情况 | 处理 |
|------|------|
| 资源不存在 | `response.Error(c, domain.ErrNotFound)` |
| 权限不足 | `response.Error(c, domain.ErrForbidden)` |
| 参数有误 | `response.Error(c, domain.ErrInvalidInput)` |
| 未登录 | `response.Error(c, domain.ErrUnauthorized)` |

### 新增数据库字段
- 写 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 迁移，不修改已有迁移文件
- 同步更新 entity struct 和 repo 的 `Scan` 列表

### 文件上传
- 复用 `upload_handler.go` + `internal/infra/oss/r2.go`，不重复实现
- key 格式: `{purpose}/{uid}/{date}/{uuid}{ext}`，不允许客户端自定义路径
- 文件类型白名单校验（后端验证，不信任客户端 Content-Type）

### 缓存
- 读多写少的数据用 Cache-Aside（Redis miss → 查 DB → 写缓存）
- 高频写场景先写 Redis，定时批量刷入 DB（Write-Behind）
- 缓存 key 格式: `{模块}:{资源}:{id}`，如 `post:detail:123`

---

## 五、Go 编码规范（强制）

### Goroutine 安全
```go
// ✅ 必须有 recover，防止 panic 崩溃服务
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("goroutine panic", zap.Any("recover", r))
        }
    }()
    if err := doWork(ctx); err != nil {
        logger.Error("work failed", zap.Error(err))
    }
}()

// ✅ goroutine 必须有退出机制
go func() {
    for {
        select {
        case <-ctx.Done(): return
        case job := <-queue: process(job)
        }
    }
}()
```

### 禁止的模式
```go
// ❌ 禁止忽略错误
result, _ := someFunc()

// ❌ 禁止 init() 做有副作用的初始化
func init() { db = connectDB() }

// ❌ 禁止全局可变状态
var GlobalDB *pgxpool.Pool

// ❌ 禁止 panic 替代 error 返回（启动阶段除外）
panic(err)

// ❌ 禁止函数参数超过 4 个（用 Options Struct）
func Foo(a, b, c, d, e string) // 禁止！

// ❌ 禁止在 WHERE 子句对列做函数（索引失效）
WHERE DATE(created_at) = '2026-03-10' // 禁止！
WHERE created_at >= '2026-03-10'::DATE // ✅
```

### 安全红线（绝对不可逾越）
- 拼接 SQL 字符串 — SQL 注入
- 明文存储密码或 API Key
- `secret` 类配置硬编码到代码
- 直接信任客户端传来的金额、权限、用户 ID
- 跳过认证/授权检查

---

## 六、迁移文件规范

```
命名: {3位序号}_{描述}.{up|down}.sql
示例: 027_add_post_tags.up.sql / 027_add_post_tags.down.sql

up.sql 必须幂等:
  CREATE TABLE IF NOT EXISTS ...
  CREATE INDEX IF NOT EXISTS ...
  ALTER TABLE ... ADD COLUMN IF NOT EXISTS ...

禁止生产环境 DROP TABLE / DROP COLUMN（用多步迁移过渡）
禁止修改已有迁移文件
```

---

## 七、自动修复策略

| 错误类型 | 自动处理 |
|----------|---------|
| `undefined: xxx` | 检查 import 路径，添加缺失引用 |
| `cannot use xxx (type yyy)` | 修正类型转换或结构体字段类型 |
| `syntax error` | 定位行号直接修复 |
| `declared and not used` | 移除未用变量或补充使用 |
| 前端 `Type error` | 检查接口定义，补全类型声明 |
| 前端 `Module not found` | 检查文件路径和 import 语句 |
| 前端 `Cannot find name` | 添加缺失的类型导入或声明 |
| 构建循环依赖 | 检查 import 路径，抽取公共包 |

修复失败超过 3 次同一错误时，换思路（检查是否使用了不存在的接口方法）。

---

## 八、禁止事项

- 不删除现有迁移文件
- 不修改已有 domain entity 的字段名（破坏兼容性）
- 不引入新 Go 依赖（`go.mod` 中没有的），除非需求必须
- 不引入新 npm 包，除非需求必须
- 不创建 README、注释文档等说明文件
- 不在 handler 里写业务逻辑
- 不在 usecase 里写 SQL
- 不在日志中记录密码、Token、Card Number 等敏感信息

---

## 九、完成标志（全部满足才算完成）

- [ ] `go build ./...` 无错误
- [ ] `pnpm build`（有前端改动时）无错误
- [ ] 代码已 `git commit` 到当前分支
- [ ] 所有 TaskCreate 的子任务已 `TaskUpdate` 为 completed
- [ ] 无任何 `// TODO`、`// placeholder`、`...` 残留
