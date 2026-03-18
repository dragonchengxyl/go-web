# Furry 同好社区平台

一个以 Furry 社区为核心的全栈 monorepo。当前代码已经实现了社区内容、群组、活动、私信、创作者打赏、AI 助手和管理后台；默认开发模式是 `Next.js Web + Gin API + PostgreSQL + Redis + MailHog`，并预留了通知、审核、统计三个可选扩展服务。

## 项目一句话

这不是一个单页 demo，而是一个已经具备完整后端分层、前端页面矩阵、数据库迁移、运维脚本和部署清单的中型工程。主线业务是社区平台，仓库内还保留了音乐内容域、成就体系和排行榜等辅助模块。

## 当前已经做了什么

### 社区主线

- 用户注册、登录、刷新 Token、退出登录、找回密码、邮箱验证
- 用户资料编辑，包含 `furry_name`、`species` 等社区字段
- 帖子发布、编辑、删除、点赞、书签、举报、屏蔽
- 评论和嵌套回复，关注关系与关注流 `feed`
- 探索页、热门标签、全文搜索、个性化推荐

### 群组与活动

- 群组列表、详情、创建、加入/退出、成员管理
- 群组帖子、公告、标签、高亮帖、置顶帖、管理面板
- 活动列表、详情、创建、报名、我的活动 / 我参与的活动

### 实时与消息

- 私信会话、消息发送、已读状态
- `WebSocket` 单连接实时推送
- Redis Pub/Sub 分布式 WS Hub，支持多节点广播

### 创作者与商业化

- 打赏订单创建
- 支付宝 / 微信支付接口接线，未配置时可回退到 mock 网关
- 赞助页、创作者统计页、赞助配置

### AI 与管理后台

- AI 助手流式对话（SSE）
- 助手会话持久化、知识检索、媒体分析
- 管理后台仪表盘、用户管理、评论管理、帖子审核、举报处理、审计日志、AI 助手设置

### 保留 / 辅助模块

- 音乐专辑与曲目接口
- 成就系统与排行榜
- `studio-cli` 健康检查、性能诊断、播种和 smoke 测试

## 运行形态

| 模式 | 组成 | 说明 |
| --- | --- | --- |
| 默认本地开发 | `apps/web` + `cmd/server` + PostgreSQL + Redis + MailHog | `./dev.sh` 一键拉起 |
| 事件驱动扩展 | 再加 `cmd/notification-svc` / `cmd/moderation-svc` / `cmd/stats-svc` | 用于验证异步通知、审核和 gRPC 拆分 |
| 部署 | Docker Compose / Kubernetes | 仓库内提供 `docker-compose.full.yml`、`docker-compose.ha.yml`、`k8s/` |

默认情况下，主 API 已经可以独立工作；其中统计服务支持本地 fallback，通知和审核则更适合在独立服务模式下验证完整事件链路。

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

- Cloudflare R2 / 阿里云 OSS
- Redis Streams 事件总线
- WebSocket 实时推送
- OpenAI-compatible / DeepSeek 风格 LLM 接口

## 关键目录

| 目录 | 作用 |
| --- | --- |
| `cmd/server` | 主 API 入口，承载 HTTP、WebSocket、SSE、指标和大部分业务逻辑 |
| `cmd/notification-svc` | 消费 Redis Streams 事件并生成通知 |
| `cmd/moderation-svc` | 消费帖子事件并执行内容审核 |
| `cmd/stats-svc` | 统计 gRPC 服务 |
| `cmd/studio-cli` | 运维 / QA 辅助工具 |
| `apps/web` | Next.js 前端 |
| `internal/domain` | 领域实体、仓储接口、权限模型 |
| `internal/usecase` | 用例层，负责业务编排 |
| `internal/infra` | PostgreSQL、Redis、OSS、支付、LLM、Streams 等基础设施实现 |
| `internal/transport` | HTTP、gRPC、WebSocket 传输层 |
| `migrations` | 数据库迁移 |
| `docs` | 架构、接口与工程说明 |
| `k8s` | Kubernetes 部署清单 |

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
| WebSocket | ws://localhost:8080/ws/chat |
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

可选扩展服务：

```bash
go run ./cmd/stats-svc -config configs/config.local.yaml
go run ./cmd/notification-svc -config configs/config.local.yaml
go run ./cmd/moderation-svc -config configs/config.local.yaml
```

## 配置说明

- 主 API 默认读取 `configs/config.local.yaml`
- 前端使用 `apps/web/.env.local` 或 `apps/web/.env`
- `./dev.sh` 会自动补齐根目录 `.env` 和 `apps/web/.env`
- OSS 直传需要配置 `oss.*`
- AI 助手需要配置 `assistant.*`
- `moderation-svc` 需要有效的阿里云内容审核凭据

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
./bin/studio-cli perf db
./bin/studio-cli pprof cpu --seconds 30
```

## 文档与部署

- 架构说明：[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- API 摘要：[`docs/API_CONTRACT.md`](docs/API_CONTRACT.md)
- Go 工程说明：[`docs/GO_ENGINEERING_NOTES.md`](docs/GO_ENGINEERING_NOTES.md)
- Kubernetes 部署：[`k8s/README.md`](k8s/README.md)

部署相关文件：

- `docker-compose.yml`：本地基础设施
- `docker-compose.full.yml`：完整前后端容器编排
- `docker-compose.ha.yml`：偏高可用部署样例

## 许可证

MIT License
