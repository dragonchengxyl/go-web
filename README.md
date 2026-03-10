# Furry 同好社区平台

一个面向 Furry 爱好者的垂直社区平台，支持图文发帖、关注动态、即时通信、创作者变现等功能。

## 项目架构

### 后端 (Go)
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + Redis
- **架构**: Clean Architecture（Domain → Usecase → Transport → Infra）
- **认证**: JWT + RBAC（7 个角色级别）
- **存储**: Cloudflare R2 / 阿里云 OSS
- **实时通信**: WebSocket（gorilla/websocket）

### 前端 (Next.js)
- **框架**: Next.js 14 App Router
- **语言**: TypeScript 5.x
- **样式**: Tailwind CSS
- **数据获取**: TanStack Query
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
- Cloudflare R2 / 阿里云 OSS 文件存储
- WebSocket Hub（实时推送基础设施）
- 数据库迁移管理（029 个迁移）

### 用户系统
- 注册 / 登录（JWT，Access + Refresh Token）
- RBAC 权限控制：`super_admin` / `admin` / `moderator` / `creator` / `supporter` / `member` / `guest`
- Token 黑名单（Redis）
- Furry 专属字段：`furry_name`（兽名）、`species`（物种）

### 社区核心
- **帖子**：图文发布、点赞、可见性控制（public / followers_only / private）
- **评论**：嵌套评论，支持帖子 / 专辑 / 音轨多态关联
- **关注**：关注 / 取关，关注流（Feed）
- **即时通信**：私信会话，WebSocket 实时消息
- **通知**：点赞 / 评论 / 关注 / 打赏触发，WebSocket 实时推送

### 创作者工具
- **打赏系统**：用户向创作者打赏，订单流转
- **创作者仪表盘**：帖子数、点赞数、评论数、粉丝数、打赏统计

### 社区运营
- **举报系统**：举报帖子 / 评论 / 用户，后端存储
- **屏蔽用户**：双向屏蔽，Feed 过滤
- **内容搜索**：PostgreSQL 全文搜索（帖子 + 用户）

### 前端页面
| 路由 | 功能 |
|------|------|
| `/feed` | 关注流 |
| `/explore` | 发现页 |
| `/posts/create` | 发帖 |
| `/posts/[id]` | 帖子详情 |
| `/users/[id]` | 用户主页 |
| `/messages` | 会话列表 |
| `/messages/[id]` | 聊天界面 |
| `/notifications` | 通知中心 |
| `/creator` | 创作者仪表盘 |
| `/profile` | 个人资料 |
| `/admin` | 管理后台入口 |

## 技术特性

- Clean Architecture，层间依赖倒置
- 参数化 SQL 查询（防 SQL 注入）
- 软删除（评论 / 帖子）
- 幂等性保证（点赞、关注去重）
- 统一错误响应（`response.Success` / `response.Error`）
- 请求限流（未认证 60/min，认证 200/min，管理员 1000/min）

## 数据库迁移

```bash
# 执行所有迁移
migrate -path migrations -database "postgres://studio:password@localhost:5432/studio_db?sslmode=disable" up

# 或通过 Makefile
make migrate-up
```

迁移文件位于 `migrations/`，共 029 个，覆盖：用户、帖子、评论、点赞、关注、会话、通知、举报、屏蔽等表。

## 许可证

MIT License
