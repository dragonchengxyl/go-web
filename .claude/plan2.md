# 待完善功能清单与实施计划 (plan2.md)

**Version:** 1.0.0
**基于代码分析日期:** 2026-03-06
**目标:** 将现有骨架代码推进到可上线状态

---

## 目录

1. [🔴 P0 — 阻塞性 Bug 与核心缺失](#p0)
2. [🟠 P1 — 重要功能缺失](#p1)
3. [🟡 P2 — 完整性与健壮性](#p2)
4. [🔵 P3 — 前端补全](#p3)
5. [⚪ P4 — 长期优化](#p4)
6. [实施顺序建议](#schedule)

---

## 🔴 P0 — 阻塞性 Bug 与核心缺失 {#p0}

> 这些问题会导致服务崩溃或核心流程完全不可用，必须优先修复。

---

### [P0-1] Coupon Service — `panic` 导致服务崩溃

- **文件:** `internal/usecase/coupon_service.go:144`
- **问题:** `GenerateRedeemCode()` 方法在 `rand.Read()` 失败时调用 `panic(err)`，会导致整个进程崩溃。
- **修复方案:** 将 `panic` 改为正常的 `error` 返回。

```go
// 当前（错误）
if _, err := rand.Read(b); err != nil {
    panic(err)
}

// 修复后
if _, err := rand.Read(b); err != nil {
    return "", fmt.Errorf("failed to generate random bytes: %w", err)
}
```

---

### [P0-2] Coupon Service — 过期判断逻辑错误

- **文件:** `internal/usecase/coupon_service.go:93`
- **问题:** 兑换码过期检查对比的是 `ExpiresAt` 与 `CreatedAt`，而不是与当前时间 `time.Now()` 对比，导致所有未过期的码被误判为过期（或永远不过期）。
- **修复方案:**

```go
// 当前（错误）
if rc.ExpiresAt != nil && rc.ExpiresAt.Before(rc.CreatedAt) {

// 修复后
if rc.ExpiresAt != nil && rc.ExpiresAt.Before(time.Now()) {
```

---

### [P0-3] Asset Handler — OSS 预签名 URL 是占位符

- **文件:** `internal/transport/http/handler/asset_handler.go` — `RequestDownload()` 末尾
- **问题:** 游戏下载接口返回硬编码的 `https://example.com/download/placeholder`，用户无法真实下载游戏。
- **修复方案:**
  1. 在 `configs/config.go` 的 `OSSConfig` 中确认字段完整（已有 Provider/Key/Bucket/Endpoint）。
  2. 在 `internal/infra/oss/` 目录下实现 `StorageService` 接口：
     - `aliyun.go` — 阿里云 OSS 实现
     - `s3.go` — AWS S3 / 兼容实现（备用）
  3. 在 `AssetService` 中注入 `StorageService`，新增 `GenerateDownloadURL(ctx, releaseID) (string, error)` 方法。
  4. Handler 中调用该方法替换占位符。

```go
// 目标实现
url, err := h.assetService.GenerateDownloadURL(c.Request.Context(), releaseID)
if err != nil {
    response.Error(c, err)
    return
}
response.Success(c, gin.H{
    "download_url": url,
    "expires_in":   900,
})
```

---

### [P0-4] Music Handler — 音频流媒体 URL 是占位符

- **文件:** `internal/transport/http/handler/music_handler.go` — `StreamTrack()` 末尾
- **问题:** 音频流接口返回硬编码的 `https://example.com/stream/placeholder`，音乐无法播放。
- **修复方案:**
  1. 复用 P0-3 中实现的 `StorageService`。
  2. 在 `MusicService` 中新增 `GenerateStreamURL(ctx, trackID) (string, error)` 方法。
  3. 根据音轨的 OSS 对象 Key 生成短期预签名 URL（建议有效期 1 小时）。

---

### [P0-5] 支付网关完全缺失

- **当前状态:** `order_service.go` 的 `PayOrder()` 方法仅将订单状态改为 `paid`，没有调用任何真实支付接口。前端 `checkout/page.tsx` 已有支付宝/微信/Stripe 选项 UI，但后端无对应实现。
- **需要创建的文件:**

```
internal/infra/payment/
  ├── interface.go          # PaymentGateway 接口定义
  ├── alipay/
  │   └── alipay.go         # 支付宝 SDK 封装
  ├── wechat/
  │   └── wechat.go         # 微信支付 SDK 封装
  └── mock/
      └── mock.go           # 测试用 Mock 实现
```

- **接口定义:**

```go
type PaymentGateway interface {
    // 发起支付，返回支付跳转URL或二维码内容
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error)
    // 查询支付状态
    QueryPayment(ctx context.Context, tradeNo string) (*PaymentStatus, error)
    // 验证异步回调签名
    VerifyCallback(ctx context.Context, params map[string]string) (bool, error)
    // 发起退款
    Refund(ctx context.Context, req RefundRequest) error
}
```

- **需要新增的数据库迁移:**
  - `payments` 表：`id`, `order_id`, `gateway`, `trade_no`, `amount_cents`, `status`, `raw_response`, `created_at`
  - `refunds` 表：`id`, `order_id`, `payment_id`, `amount_cents`, `reason`, `status`, `created_at`

- **需要新增的路由:**
  - `POST /api/v1/orders/:id/pay/alipay` — 发起支付宝支付
  - `POST /api/v1/orders/:id/pay/wechat` — 发起微信支付
  - `POST /api/v1/payment/callback/alipay` — 支付宝异步回调（无需 Auth）
  - `POST /api/v1/payment/callback/wechat` — 微信支付异步回调（无需 Auth）
  - `POST /api/v1/orders/:id/refund` — 申请退款（已认证用户）

- **需要新增的配置项** (`configs/config.go`):

```go
type Config struct {
    // ... 已有字段
    Email   EmailConfig   `mapstructure:"email"`
    Payment PaymentConfig `mapstructure:"payment"`
}

type EmailConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    From     string `mapstructure:"from"`
}

type PaymentConfig struct {
    Alipay AlipayConfig `mapstructure:"alipay"`
    Wechat WechatConfig `mapstructure:"wechat"`
}

type AlipayConfig struct {
    AppID      string `mapstructure:"app_id"`
    PrivateKey string `mapstructure:"private_key"`
    PublicKey  string `mapstructure:"public_key"`
    NotifyURL  string `mapstructure:"notify_url"`
    Sandbox    bool   `mapstructure:"sandbox"`
}

type WechatConfig struct {
    AppID     string `mapstructure:"app_id"`
    MchID     string `mapstructure:"mch_id"`
    APIKey    string `mapstructure:"api_key"`
    NotifyURL string `mapstructure:"notify_url"`
    Sandbox   bool   `mapstructure:"sandbox"`
}
```

---

## 🟠 P1 — 重要功能缺失 {#p1}

> 功能已有骨架但实现不完整，影响用户体验。

---

### [P1-1] Search Handler — 热门搜索词为硬编码

- **文件:** `internal/transport/http/handler/search_handler.go` — `GetPopularSearches()`
- **问题:** 返回写死的字符串数组，没有实际统计。
- **Router 状态:** `GetPopularSearches` 路由**未在 router.go 中注册**。
- **修复方案:**
  1. 在 Redis 中用 `ZIncrBy` 记录搜索词频率（Key: `studio:search:popular`）。
  2. 在 `search_service.go` 中新增 `GetPopularSearches(ctx, limit int) ([]string, error)` 方法。
  3. 在 router.go 中注册路由: `search.GET("/popular", searchHandler.GetPopularSearches)`
  4. 在 `SearchAll` / `SearchGames` 调用时顺带更新搜索词计数。

---

### [P1-2] Router — Analytics、Upload、Audit 路由未注册

- **文件:** `internal/transport/http/router.go`
- **问题:** `AnalyticsHandler`、`UploadHandler`、`AuditHandler` 已有实现文件，但均未在路由中注册，API 完全不可访问。
- **需要注册的路由:**

```go
// Analytics（需要注入 AnalyticsService 到 RouterConfig）
v1.POST("/events", analyticsHandler.TrackEvent)                         // 公开
admin.GET("/analytics/dashboard", analyticsHandler.GetDashboardMetrics) // 仅管理员
admin.GET("/analytics/funnel", analyticsHandler.GetConversionFunnel)
admin.GET("/analytics/retention", analyticsHandler.GetUserRetention)

// Upload（需要注入 StorageService）
uploadProtected := v1.Group("/upload")
uploadProtected.Use(authMiddleware.Authenticate())
uploadProtected.POST("/avatar", uploadHandler.UploadAvatar)
uploadProtected.POST("/game-asset", uploadHandler.UploadGameAsset)   // Admin only
uploadProtected.POST("/track", uploadHandler.UploadTrack)            // Admin only

// Search popular
search.GET("/popular", searchHandler.GetPopularSearches)
```

- **RouterConfig 需要补充的字段:**

```go
AnalyticsService *usecase.AnalyticsService
AuditService     *usecase.AuditService
EmailService     *usecase.EmailService
StorageService   oss.StorageService
```

---

### [P1-3] Email 配置未接入 Config 系统

- **文件:** `internal/pkg/email/sender.go`、`configs/config.go`
- **问题:** `email.SMTPConfig` 是独立结构体，没有从全局 `configs.Config` 中读取，导致 `cmd/server/main.go` 中无法统一配置。
- **修复方案:** 在 `configs/config.go` 中添加 `EmailConfig`（见 P0-5），并在 `main.go` 中用 `cfg.Email` 初始化 `email.Sender`。

---

### [P1-4] 前端 Cart — 优惠券逻辑未实现

- **文件:** `apps/web/src/app/cart/page.tsx:22`
- **问题:** `const discount = 0; // TODO: Implement coupon logic`，用户输入优惠券码后无任何效果。
- **修复方案:**
  1. 在 `api-client.ts` 中新增 `validateCoupon(code: string)` 方法，调用 `GET /coupons/validate?code=xxx`。
  2. Cart 页面新增「验证优惠券」按钮，调用 API 后展示折扣金额。
  3. 创建订单时将 `couponCode` 传入 `createOrder()`。

---

### [P1-5] 前端 Checkout — 无真实支付跳转

- **文件:** `apps/web/src/app/checkout/page.tsx`
- **问题:** UI 已有支付宝/微信/Stripe 三个选项，但点击支付后只调用 `apiClient.payOrder()`，没有跳转到支付页面或展示二维码。
- **修复方案（依赖 P0-5 完成后）:**
  1. 修改 `api-client.ts` 的 `payOrder()` 返回结构，增加 `pay_url`（支付宝跳转链接）或 `qr_code`（微信二维码数据）。
  2. Checkout 页面根据 `paymentMethod` 展示不同 UI：支付宝 → 跳转新窗口；微信 → 展示 QR 码弹窗。
  3. 轮询或 WebSocket 监听支付结果，成功后跳转订单详情页。

---

### [P1-6] api-client.ts — 缺少多个关键方法

- **文件:** `apps/web/src/lib/api-client.ts`
- **缺失方法:**

```typescript
// 优惠券
async validateCoupon(code: string): Promise<{ valid: boolean; discount: number }>
async redeemCode(code: string): Promise<void>

// 成就
async getUserAchievements(userId: string): Promise<Achievement[]>
async getMyAchievements(): Promise<Achievement[]>
async getMyPoints(): Promise<{ total: number; level: number }>

// 排行榜
async getLeaderboard(type?: 'all' | 'weekly'): Promise<LeaderboardEntry[]>

// 文件上传
async uploadAvatar(file: File): Promise<{ url: string }>
async uploadFile(endpoint: string, file: File): Promise<{ url: string }>

// 搜索
async getPopularSearches(): Promise<string[]>
```

---

## 🟡 P2 — 完整性与健壮性 {#p2}

> 系统能运行，但存在不健壮、不完整的地方。

---

### [P2-1] Order Service — 缺少退款状态流转

- **文件:** `internal/usecase/order_service.go`
- **问题:** 订单状态从 `paid` 到 `refunded` 没有任何实现。
- **需要新增:**

```go
func (s *OrderService) RefundOrder(ctx context.Context, orderID uuid.UUID, reason string) error {
    // 1. 检查订单状态必须为 paid 或 fulfilled
    // 2. 调用 PaymentGateway.Refund()
    // 3. 创建 refunds 表记录
    // 4. 更新订单状态为 refunded
    // 5. 触发退款成功邮件通知
}
```

---

### [P2-2] 缺少 Prometheus 监控中间件

- **文件:** `internal/transport/http/middleware/`（需新建 `metrics.go`）
- **问题:** plan.md Epic 1.3.2 要求 HTTP QPS、P99 延迟、错误率埋点，但中间件目录中没有该文件。
- **修复方案:**
  1. 引入 `github.com/prometheus/client_golang` 依赖。
  2. 新建 `middleware/metrics.go`，记录 `http_request_duration_seconds`、`http_requests_total`。
  3. 在 `router.go` 中注册 `GET /metrics` 端点（仅内网访问）。

---

### [P2-3] 定时任务未启动

- **文件:** `cmd/server/main.go`
- **问题:** `OrderService.CancelExpiredOrders()` 已实现，但没有 cron 任务调用它。其他定时任务（如 Analytics `RefreshDailyMetrics`）同样未启动。
- **修复方案:** 在 `main.go` 启动时开启后台 goroutine 或使用 `robfig/cron`：

```go
// 每分钟取消过期订单
c.AddFunc("@every 1m", func() {
    n, err := orderSvc.CancelExpiredOrders(ctx)
    if err != nil {
        logger.Error("cancel expired orders failed", zap.Error(err))
    } else if n > 0 {
        logger.Info("cancelled expired orders", zap.Int("count", n))
    }
})

// 每天凌晨刷新分析汇总视图
c.AddFunc("0 1 * * *", func() {
    analyticsSvc.RefreshDailyMetrics(ctx)
})
```

---

### [P2-4] 缺少数据库迁移

- **缺少的迁移文件:**

| 迁移编号 | 表名 | 说明 |
|---------|------|------|
| `20260306000015` | `payments` | 支付记录，含 `gateway`, `trade_no`, `raw_response` |
| `20260306000016` | `refunds` | 退款记录 |
| `20260306000017` | `notifications` | 站内通知 |
| `20260306000018` | `user_sessions` | 设备/会话管理（补充 Redis token store） |

---

### [P2-5] 缺少单元测试

- **当前状态:** 项目中只有 `game_service_test.go` 和 `crypto_test.go` 两个测试文件。
- **建议补充的测试:**

```
internal/usecase/coupon_service_test.go    # 覆盖验证、使用、过期逻辑
internal/usecase/order_service_test.go     # 覆盖创建、支付、取消、退款
internal/usecase/achievement_service_test.go
internal/infra/postgres/order_repo_test.go # 集成测试（需要测试数据库）
```

---

### [P2-6] Admin Handler 注入的 Service 不完整

- **文件:** `internal/transport/http/router.go:116`
- **问题:** `RouterConfig` 缺少 `AnalyticsService`、`AuditService`、`EmailService`、`OrderService`（管理员查看订单用），`AdminHandler` 无法正常处理所有管理请求。
- **修复方案:** 补充 RouterConfig 中的字段，并在 `main.go` 中注入对应 service 实例。

---

## 🔵 P3 — 前端补全 {#p3}

> 前端目前页面极少，大量功能页面缺失。

---

### [P3-1] 前端页面现状

当前已存在的页面（`apps/web/src/app/`）：

| 页面 | 状态 | 说明 |
|------|------|------|
| `/` (首页) | ✅ 基本完整 | Hero、Featured Games/Music 组件 |
| `/login` | ✅ 完整 | 登录表单 |
| `/register` | ✅ 完整 | 注册表单 |
| `/games` | ✅ 基础 | 游戏列表 |
| `/music` | ✅ 基础 | 专辑列表 |
| `/cart` | ⚠️ 不完整 | 优惠券逻辑缺失（见 P1-4） |
| `/checkout` | ⚠️ 不完整 | 支付跳转缺失（见 P1-5） |
| `/orders` | ✅ 基础 | 订单列表 |
| `/profile` | ✅ 基础 | 用户资料 |
| `/community` | ✅ 基础 | 评论/社区 |
| `/achievements` | ✅ 基础 | 成就列表 |
| `/leaderboard` | ✅ 基础 | 排行榜 |
| `/search` | ✅ 基础 | 搜索结果 |
| `/about` | ✅ 完整 | 关于页 |
| `/admin` | ⚠️ 入口页 | 只有管理员入口，无子页面 |

---

### [P3-2] 需要新建的前端页面

```
apps/web/src/app/
  games/[slug]/              # 游戏详情页（截图、描述、下载按钮、评论区）
  music/[slug]/              # 专辑详情页（音轨列表、播放器）
  admin/
    orders/                  # 订单管理（列表、详情、退款操作）
    products/                # 商品管理（CRUD、定价、折扣）
    coupons/                 # 优惠券管理（创建、批量生成兑换码）
    analytics/               # 数据分析（DAU图表、收入图表、转化漏斗）
    games/[id]/edit/         # 游戏编辑（含分支/版本管理）
    music/[id]/tracks/       # 音轨管理（上传、排序）
```

---

### [P3-3] 前端缺少 Game/Music 详情页

- **问题:** 游戏卡片 (`game-card.tsx`) 点击后没有跳转到详情页，`/games/[slug]` 路由不存在。
- **优先级:** 这是用户购买游戏的核心路径，**必须实现**。
- **详情页应包含:** 游戏截图轮播、详细介绍、系统要求、下载/购买按钮、用户评论区。

---

### [P3-4] 音乐播放器与 OSS 集成

- **文件:** `apps/web/src/components/music-player.tsx`
- **问题:** 播放器组件存在但播放 URL 来自 API 占位符（依赖 P0-4 修复）。
- **修复方案:**
  1. P0-4 完成后，播放器调用 `POST /tracks/:id/stream` 获取真实 URL。
  2. 实现播放列表状态管理（Context 或 Zustand）。
  3. 记录播放历史，支持上次播放继续。

---

## ⚪ P4 — 长期优化 {#p4}

> 不影响上线，但有利于系统长期稳定性和扩展性。

---

### [P4-1] 搜索升级到 Elasticsearch / Meilisearch

- **当前:** `search_service.go` 使用 PostgreSQL 全文搜索（`tsvector`）。
- **建议:** 在用户量增大后，迁移到专用搜索引擎，支持拼音搜索、模糊匹配、高亮。

---

### [P4-2] Redis Stream 消息队列接入

- **plan.md Task 1.2.4** 要求使用 Redis Stream 实现异步邮件、通知、下载日志。
- **当前:** 邮件发送使用 `go s.processPendingEmails()` goroutine，存在进程崩溃丢失任务的风险。
- **建议:** 实现 `studio:mq:email` Stream，消费者组模式保证至少一次投递。

---

### [P4-3] Cursor-based 分页替换 OFFSET 分页

- **技术债务 TD-001:** 评论、帖子列表使用 `OFFSET` 分页，高偏移量性能差。
- **计划:** M4 完成后迁移到 cursor-based 分页。

---

### [P4-4] 音频异步转码

- **技术债务 TD-002:** 音频上传后同步转码，阻塞 HTTP 接口。
- **计划:** 引入异步 Worker，上传后立即返回，转码完成后通知。

---

### [P4-5] 本地化/i18n（Epic 12）

- **当前:** 完全缺失，所有文案硬编码为中文。
- **建议:** 使用 `go-i18n` 库，前端使用 `next-intl`，支持中/英双语。

---

### [P4-6] 增加 OpenTelemetry 链路追踪

- **plan.md Task 1.1.4** 要求集成 OpenTelemetry 支持 Jaeger 可视化追踪。
- **当前:** 只有 `X-Request-ID` 传递，无分布式追踪。

---

## 实施顺序建议 {#schedule}

```
Sprint 1（本周）— P0 Bug 修复
  ✅ [P0-1] 修复 coupon_service.go 中的 panic
  ✅ [P0-2] 修复 coupon_service.go 中的过期判断
  ✅ [P0-3] 实现 OSS StorageService，接入游戏下载
  ✅ [P0-4] 实现音轨 OSS 流媒体 URL

Sprint 2（下周）— 支付系统
  □ [P0-5] 实现支付宝/微信支付网关
  □ [P0-5] 新增 payments/refunds 数据库迁移
  □ [P0-5] 新增支付回调路由
  □ [P2-1] 实现 RefundOrder 逻辑
  □ [P1-5] 前端 Checkout 支付跳转

Sprint 3 — 路由与配置补全
  □ [P1-1] 热门搜索 Redis 实现 + 注册路由
  □ [P1-2] 注册 Analytics/Upload/Audit 路由
  □ [P1-3] Email Config 接入全局配置
  □ [P2-3] 添加定时任务（过期订单取消、分析刷新）
  □ [P2-6] 完善 RouterConfig 依赖注入

Sprint 4 — 前端核心页面
  □ [P3-3] 游戏详情页 /games/[slug]
  □ [P3-3] 专辑详情页 /music/[slug]
  □ [P1-4] Cart 优惠券逻辑
  □ [P1-6] api-client.ts 补全缺失方法

Sprint 5 — 管理后台前端
  □ [P3-2] admin/orders 订单管理页
  □ [P3-2] admin/analytics 数据分析页
  □ [P3-2] admin/products 商品管理页
  □ [P3-2] admin/coupons 优惠券管理页

Sprint 6 — 健壮性
  □ [P2-2] Prometheus 监控中间件
  □ [P2-4] 补充数据库迁移（payments/refunds/notifications）
  □ [P2-5] 核心业务逻辑单元测试

后续（长期）
  □ [P4-1] 搜索引擎升级
  □ [P4-2] Redis Stream 消息队列
  □ [P4-3] Cursor 分页
  □ [P4-5] i18n 国际化
  □ [P4-6] OpenTelemetry 链路追踪
```

---

## 文件变更速查表

| 优先级 | 文件路径 | 变更类型 | 关联任务 |
|--------|---------|---------|---------|
| P0 | `internal/usecase/coupon_service.go` | 修复 Bug | P0-1, P0-2 |
| P0 | `internal/infra/oss/aliyun.go` | 新建 | P0-3, P0-4 |
| P0 | `internal/infra/oss/interface.go` | 新建 | P0-3, P0-4 |
| P0 | `internal/usecase/asset_service.go` | 修改 | P0-3 |
| P0 | `internal/usecase/music_service.go` | 修改 | P0-4 |
| P0 | `internal/transport/http/handler/asset_handler.go` | 修改 | P0-3 |
| P0 | `internal/transport/http/handler/music_handler.go` | 修改 | P0-4 |
| P0 | `internal/infra/payment/interface.go` | 新建 | P0-5 |
| P0 | `internal/infra/payment/alipay/alipay.go` | 新建 | P0-5 |
| P0 | `internal/infra/payment/wechat/wechat.go` | 新建 | P0-5 |
| P0 | `internal/transport/http/handler/payment_handler.go` | 新建 | P0-5 |
| P0 | `configs/config.go` | 修改（新增 Email/Payment block） | P0-5, P1-3 |
| P0 | `migrations/20260306000015_create_payments.*.sql` | 新建 | P0-5 |
| P0 | `migrations/20260306000016_create_refunds.*.sql` | 新建 | P0-5, P2-1 |
| P1 | `internal/transport/http/router.go` | 修改 | P1-1, P1-2 |
| P1 | `internal/transport/http/handler/search_handler.go` | 修改 | P1-1 |
| P1 | `apps/web/src/app/cart/page.tsx` | 修改 | P1-4 |
| P1 | `apps/web/src/app/checkout/page.tsx` | 修改 | P1-5 |
| P1 | `apps/web/src/lib/api-client.ts` | 修改 | P1-6 |
| P2 | `internal/usecase/order_service.go` | 修改 | P2-1 |
| P2 | `internal/transport/http/middleware/metrics.go` | 新建 | P2-2 |
| P2 | `cmd/server/main.go` | 修改 | P2-3, P2-6 |
| P3 | `apps/web/src/app/games/[slug]/page.tsx` | 新建 | P3-3 |
| P3 | `apps/web/src/app/music/[slug]/page.tsx` | 新建 | P3-3 |
| P3 | `apps/web/src/app/admin/orders/page.tsx` | 新建 | P3-2 |
| P3 | `apps/web/src/app/admin/analytics/page.tsx` | 新建 | P3-2 |
| P3 | `apps/web/src/app/admin/products/page.tsx` | 新建 | P3-2 |
| P3 | `apps/web/src/app/admin/coupons/page.tsx` | 新建 | P3-2 |
