# Furry 同好社区平台

一个以 Furry 社区为核心的全栈 monorepo。当前仓库已经落地社区内容、圈子、活动、私信通知、创作者赞助、音频作品、AI 助手、管理后台，以及一个持续扩展中的 `Hex Blitz` 游戏实验室。

默认开发形态是 `Next.js Web + Gin API + PostgreSQL + Redis + MailHog`，生产部署当前统一使用 `docker-compose.prod.yml`。

## 项目一句话

这不是一个单页 demo，而是一个已经具备后端分层、前端页面矩阵、数据库迁移、运维脚本和部署清单的中型工程。

当前架构更接近：

- 社区平台主站，承载内容、圈子、活动、私信、通知和创作者能力
- 单体 API 优先，按需要用 Redis Streams / gRPC 拆出通知、审核、统计和音频处理
- 同仓库维护 Web 前端、Go 后端、独立 worker/service、部署文件和 QA 工具

## 当前代码已经做了什么

### 社区主线

- 用户注册、登录、刷新 Token、退出登录、找回密码、重置密码、邮箱验证
- 用户资料编辑，包含 `furry_name`、`species`、头像、简介、站点链接等社区字段
- 帖子发布、编辑、删除、点赞、书签、举报、屏蔽
- 评论与回复、评论点赞、关注关系与关注流 `feed`
- 探索页、热门标签、全文搜索、个性化推荐

### 圈子与活动

- 圈子列表、详情、创建、加入/退出、成员管理
- 圈子帖子、公告、标签、高亮帖、置顶帖、管理面板
- 活动列表、详情、创建、报名、我的活动 / 我参与的活动

### 实时、通知与审核

- 私信会话、消息发送、已读状态
- `WebSocket` 单连接实时推送
- Redis Pub/Sub 分布式 WS Hub，支持多节点广播
- 通知中心、未读数查询与实时提醒
- Redis Streams 事件总线
- `notification-svc` 消费事件生成通知
- `moderation-svc` 消费帖子事件执行内容审核

### 创作者与音频

- 打赏订单创建
- 支付宝 / 微信支付接口接线，未配置时可回退到 mock 网关
- 赞助页、创作者统计页、赞助配置
- 音频任务创建、重试、发布
- 社区音频作品列表、详情、播放事件、点赞、收藏、评论、举报
- 独立 `audio-worker` 消费音频任务并执行本地处理

### AI 与管理后台

- AI 助手流式对话（SSE）
- 助手会话持久化、知识检索、媒体分析
- 助手知识索引后台同步与降级 fallback
- 管理后台仪表盘、用户管理、评论管理、帖子审核、举报处理、审计日志
- 管理后台 AI 助手设置与辅助工具
- 后台还覆盖圈子、活动、订单、音频作品和游戏概览等页面

### 游戏实验室

- `Hex Blitz` 房间列表、房间详情、排行榜、最近对局
- `Hex Blitz` 对局回放接口
- 游戏专用 WebSocket：`/ws/game/hex-blitz`
- 前端已接入实验室页面、排行榜和回放页

### 辅助模块与工程能力

- 音乐专辑与曲目接口
- 成就系统与排行榜
- `studio-cli` 健康检查、管理辅助、性能诊断、播种和 smoke 测试
- `/health`、`/ready`、`/metrics`、`/debug/pprof/*`
- Docker Compose、数据库迁移、运维脚本

## 运行形态

| 模式 | 组成 | 说明 |
| --- | --- | --- |
| 默认本地开发 | `apps/web` + `cmd/server` + PostgreSQL + Redis + MailHog | `./dev.sh` 一键拉起，主 API 已可独立承载大部分业务 |
| 事件驱动扩展 | 再加 `cmd/notification-svc` / `cmd/moderation-svc` / `cmd/stats-svc` | 用于验证异步通知、审核和 gRPC 拆分 |
| 音频处理扩展 | 再加 `cmd/audio-worker` | 消费音频任务、执行本地处理与重试 |
| 部署 | Docker Compose | 当前生产部署统一使用 `docker-compose.prod.yml` |

默认情况下，主 API 可以独立工作；统计服务支持本地 fallback，而通知、审核和音频处理在独立进程模式下更接近生产拓扑。

## 技术栈

### 后端

- Go 1.22
- Gin
- PostgreSQL + pgx
- Redis
- JWT + RBAC
- gRPC + Proto
- Prometheus 指标、`pprof`

### 前端

- Next.js 14 App Router
- React 18
- TypeScript
- TanStack Query
- Tailwind CSS
- pnpm workspace + Turborepo

### 基础能力

- Cloudflare R2 / 阿里云 OSS / 本地存储
- Redis Streams 事件总线
- Redis Pub/Sub + WebSocket 实时推送
- OpenAI-compatible / DeepSeek 风格 LLM 接口
- Embedding / Vision 能力接入

## 关键目录

| 目录 | 作用 |
| --- | --- |
| `cmd/server` | 主 API 入口，承载 HTTP、WebSocket、SSE、指标和大部分业务逻辑 |
| `cmd/audio-worker` | 消费音频任务并执行本地处理、重试 |
| `cmd/notification-svc` | 消费 Redis Streams 事件并生成通知 |
| `cmd/moderation-svc` | 消费帖子事件并执行内容审核 |
| `cmd/stats-svc` | 统计 gRPC 服务 |
| `cmd/seed-dev` | 本地开发数据播种 |
| `cmd/studio-cli` | 运维 / QA / 管理辅助工具 |
| `apps/web` | Next.js 前端 |
| `internal/domain` | 领域实体、仓储接口、权限模型 |
| `internal/usecase` | 用例层，负责业务编排 |
| `internal/infra` | PostgreSQL、Redis、OSS、支付、LLM、Streams 等基础设施实现 |
| `internal/transport` | HTTP、gRPC、WebSocket 传输层 |
| `migrations` | 数据库迁移 |
| `docs` | 架构、接口与工程说明 |
| `docs` | 架构、接口与部署说明 |

## 快速启动

### 依赖

- Docker / Docker Compose
- Go 1.22+
- pnpm 8+
- `golang-migrate`

### 一键启动

```bash
./dev.sh
```

脚本会完成：

- 检查依赖
- 启动 PostgreSQL、Redis、MailHog
- 执行数据库迁移
- 启动 Go API
- 启动 Next.js 前端

默认地址：

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:3000 |
| API | http://localhost:8080/api/v1 |
| 健康检查 | http://localhost:8080/health |
| 就绪检查 | http://localhost:8080/ready |
| 聊天 WebSocket | ws://localhost:8080/ws/chat |
| 游戏 WebSocket | ws://localhost:8080/ws/game/hex-blitz |
| MailHog | http://localhost:8025 |
| Prometheus Metrics | http://localhost:8080/metrics |

常用参数：

```bash
./dev.sh --no-docker
./dev.sh --backend-only
./dev.sh --frontend-only
./dev.sh --stop
./dev.sh --logs
```

## 手动开发流程

如果你想拆开启动：

```bash
make dev-setup
make infra-up
make migrate-up
make dev-backend
make dev-frontend
```

可选扩展进程：

```bash
go run ./cmd/stats-svc -config configs/config.local.yaml
go run ./cmd/notification-svc -config configs/config.local.yaml
go run ./cmd/moderation-svc -config configs/config.local.yaml
go run ./cmd/audio-worker -config configs/config.local.yaml
```

开发数据播种：

```bash
go run ./cmd/seed-dev -config configs/config.local.yaml
go run ./cmd/seed-dev -config configs/config.local.yaml -mode bulk -profile medium -namespace bulk
```

## 配置说明

- 主 API 默认读取 `configs/config.local.yaml`
- 前端使用 `apps/web/.env.local` 或 `apps/web/.env`
- `./dev.sh` 会自动补齐根目录 `.env` 和 `apps/web/.env`
- OSS 直传需要配置 `oss.*`
- AI 助手需要配置 `assistant.*`
- 支付接入需要配置 `payment.*`
- `moderation-svc` 需要有效的阿里云内容审核凭据
- `audio-worker` 读取 `audio.*` 和 `oss.allowed_hosts`

## 常用命令

```bash
# 后端
go test ./...
go test -race ./...
make build-all

# 前端
pnpm --filter web lint
pnpm --filter web type-check
pnpm --filter web build

# 全量本地门禁
make ci

# 构建并使用 studio-cli
make build-studio-cli
./bin/studio-cli health
./bin/studio-cli smoke
./bin/studio-cli seed demo
./bin/studio-cli seed bulk --profile medium --namespace bulk
./bin/studio-cli perf db
./bin/studio-cli pprof cpu --seconds 30
```

## 虚拟数据播种

当前仓库提供两种播种方式：

- `demo`
  - 小规模稳定样本，适合本地 smoke test 和面试演示
- `bulk`
  - 面向联调、截图、后台列表、搜索、分页和压测的多模块虚拟数据

`bulk` 当前覆盖：

- 用户、关注、圈子、圈子公告、活动、活动报名
- 帖子、点赞、评论、评论点赞、收藏
- 私信会话、消息、通知、举报
- 打赏订单、赞助配置、审计日志、行为事件
- 音频任务、音频作品、音频点赞
- AI 助手会话、消息、反馈、知识文档
- 专辑、曲目、成就、积分流水
- `Hex Blitz` / 斗地主对局与事件数据

推荐命令：

```bash
go run ./cmd/seed-dev -config configs/config.local.yaml -mode bulk -profile medium -namespace bulk
```

或：

```bash
./bin/studio-cli seed bulk --profile medium --namespace bulk
```

生产容器环境：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
SEED_PROFILE=small docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
SEED_PROFILE=medium SEED_NAMESPACE=bulk docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

可选 profile：

- `small`
  - 适合本地开发快速填充
- `medium`
  - 默认推荐，适合测试环境和后台联调
- `large`
  - 更高密度的数据样本，仍以控制在常见单库测试环境可接受体量为目标

说明：

- 所有 bulk 数据都带命名空间，默认密码仍为 `Passw0rd123`
- 使用相同 `namespace` 重跑时，UUID 主键类数据会保持稳定
- 命令结束后会输出当前数据库体量，便于你确认没有逼近容量上限

## 文档与部署

- 架构说明：[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- API 摘要：[`docs/API_CONTRACT.md`](docs/API_CONTRACT.md)
- Go 工程说明：[`docs/GO_ENGINEERING_NOTES.md`](docs/GO_ENGINEERING_NOTES.md)
- Docker 生产部署：[`docs/DOCKER_PROD_DEPLOY.md`](docs/DOCKER_PROD_DEPLOY.md)

部署相关文件：

- `docker-compose.yml`：本地开发基础设施（PostgreSQL / Redis / MailHog）
- `docker-compose.prod.yml`：当前实际使用的生产部署编排

## 许可证

MIT License
