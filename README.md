# Furry 同好社区平台

一个面向 Furry 爱好者的垂直社区平台，支持图文发帖、关注动态、即时通信、OSS 直传、WebSocket 实时推送、内容审核等功能。

## 项目架构

### 后端 (Go)
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + Redis
- **架构**: Clean Architecture（Domain → Usecase → Transport → Infra）
- **认证**: JWT + RBAC（7 个角色级别），注册 IP 限流（Redis）
- **存储**: Cloudflare R2 / 阿里云 OSS（支持前端直传 Policy）
- **实时通信**: WebSocket Hub（分布式 Redis Pub/Sub 多节点路由）
- **内容审核**: 阿里云内容安全异步审核管线

### 前端 (Next.js)
- **框架**: Next.js 14 App Router
- **语言**: TypeScript 5.x
- **样式**: Tailwind CSS
- **数据获取**: TanStack Query（含乐观更新）
- **实时**: 全局 WSContext（单连接 + 指数退避重连）
- **Monorepo**: Turborepo + pnpm workspace

## 快速启动

```bash
./dev.sh
```

脚本自动完成：Docker 基础设施 → 数据库迁移 → 后端 → 前端。

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080/api/v1 |
| WebSocket | ws://localhost:8080/ws/chat |
| MailHog | http://localhost:8025 |

**可选参数：**

```bash
./dev.sh --no-docker     # 跳过 Docker（已有本地 PG/Redis）
./dev.sh --backend-only  # 只启动后端
./dev.sh --stop          # 停止 Docker 基础设施
```

## 已实现功能

### 基础设施
- PostgreSQL 连接池 + Redis 缓存
- 全局中间件（JWT 认证、限流、CORS、结构化日志）
- Cloudflare R2 / 阿里云 OSS 文件存储，支持前端 OSS 直传签名 Policy
- WebSocket 分布式 Hub（Redis Pub/Sub，支持多节点部署）
- 每连接 Token Bucket 限流 + 单用户最多 5 路并发连接
- 数据库迁移管理（043 个迁移）

### 用户系统
- 注册 / 登录（JWT，Access + Refresh Token）
- RBAC 权限控制：`super_admin` / `admin` / `moderator` / `creator` / `supporter` / `member` / `guest`
- Token 黑名单（Redis）
- 注册 IP 限流（Redis，5 次/小时/IP）
- Furry 专属字段：`furry_name`（兽名）、`species`（物种）

### 社区核心
- **帖子**：图文发布（OSS 直传）、点赞（乐观更新）、可见性控制（public / followers_only / private）
- **内容标签**：`content_labels` JSONB，支持 `is_ai_generated` 等标注
- **内容审核**：发帖后异步调用阿里云内容安全，状态流转 `pending → approved / blocked`，前端实时展示审核遮罩
- **评论**：嵌套评论，多态关联（帖子等）
- **关注**：关注 / 取关，关注流 Feed
- **探索页**：按参与度评分排序（`like_count + comment_count×3 / 时间衰减`），支持标签过滤与 AI 内容过滤
- **即时通信**：私信会话，WebSocket 实时消息
- **通知**：点赞 / 评论 / 关注 / 打赏触发，WebSocket 实时推送，未读角标实时更新

### 创作者工具
- **打赏系统**：用户向创作者打赏，订单流转
- **赞助页**：月度目标进度、鸣谢名录、支付宝/微信收款码展示
- **创作者仪表盘**：帖子数、点赞数、评论数、粉丝数、打赏统计

### 社区运营
- **举报系统**：举报帖子 / 评论 / 用户，后端存储
- **屏蔽用户**：双向屏蔽，Feed 过滤
- **内容搜索**：PostgreSQL 全文搜索（帖子 + 用户）

### 前端页面
| 路由 | 功能 |
|------|------|
| `/` | 首页（热门帖子 + 社区特色展示） |
| `/feed` | 关注流（实时无限滚动） |
| `/explore` | 发现页（评分排序 + 标签/AI 过滤） |
| `/search` | 全文搜索（帖子 + 用户） |
| `/tags/[tag]` | 标签聚合页 |
| `/posts/create` | 发帖（OSS 直传 + AI 标签 + 草稿自动保存） |
| `/posts/[id]` | 帖子详情（乐观点赞） |
| `/users/[id]` | 用户主页 |
| `/users/[id]/followers` | 粉丝列表 |
| `/users/[id]/following` | 关注列表 |
| `/messages` | 会话列表（实时更新） |
| `/messages/[id]` | 聊天界面 |
| `/notifications` | 通知中心 |
| `/sponsor` | 赞助页 |
| `/creator` | 创作者仪表盘 |
| `/profile` | 个人资料 |
| `/settings` | 账号与隐私设置 |
| `/admin` | 管理后台入口 |

## 技术特性

- Clean Architecture，层间依赖倒置
- 参数化 SQL 查询（防 SQL 注入）
- 幂等性保证（点赞、关注去重）
- 统一错误响应（`response.Success` / `response.Error`）
- 请求限流（未认证 60/min，认证 200/min，管理员 1000/min）
- OSS 媒体 URL 白名单校验（服务端拒绝非授权域名）
- WebSocket 指数退避重连（1→2→4→8→16→30s），页面隐藏时暂停
- XSS 防御：帖子内容纯文本渲染，禁用 `dangerouslySetInnerHTML`
- OSS Policy/Signature 不写入 localStorage，不打印到 console

## 数据库迁移

```bash
# 执行所有迁移
migrate -path migrations -database "postgres://studio:password@localhost:5432/studio_db?sslmode=disable" up

# 或通过 Makefile
make migrate-up
```

迁移文件位于 `migrations/`，共 043 个，覆盖：用户、帖子（含 `moderation_status`、`content_labels`）、评论、点赞、关注、会话、通知、举报、屏蔽、群组、活动、AI 助手等表。

## 质量门禁

```bash
# 后端
go test ./...
go test -race ./...

# 前端
pnpm --filter web lint
pnpm --filter web type-check
pnpm --filter web build

# 一次性跑完本地门禁
make ci
```

## 许可证

MIT License
# deploy test
