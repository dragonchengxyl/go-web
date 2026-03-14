# Go Engineering Notes

这个文件是项目的面试讲稿底板，重点体现 Go 工程能力，而不是功能堆砌。

## 1. 可观测性

### 请求日志

HTTP 请求统一输出结构化字段：

- `request_id`
- `user_id`
- `role`
- `method`
- `route`
- `path`
- `status`
- `latency`
- `response_bytes`
- `client_ip`
- `user_agent`

代码位置：

- [internal/transport/http/middleware/logger.go](/home/chenlongting/go-web/internal/transport/http/middleware/logger.go)

### Prometheus 指标

当前重点指标：

- `http_request_duration_seconds`
- `http_requests_total`
- `http_slow_requests_total`
- `http_requests_in_flight`

代码位置：

- [internal/transport/http/middleware/metrics.go](/home/chenlongting/go-web/internal/transport/http/middleware/metrics.go)

### pprof

已接入 `/debug/pprof/*`：

- 开发环境直接可用
- `release` 模式下要求管理员 JWT

代码位置：

- [internal/transport/http/router.go](/home/chenlongting/go-web/internal/transport/http/router.go)

### 本地剖析建议

```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
go tool pprof http://localhost:8080/debug/pprof/heap
```

## 2. 并发模型

### HTTP 请求模型

- Gin 负责每请求一个 goroutine
- 请求上下文从 handler 透传到 usecase / repository
- 日志、限流、指标、恢复都通过中间件统一接入

### WebSocket

- 每个连接拆成 `readPump` / `writePump`
- Hub 负责用户维度的 fan-out
- 分布式模式通过 Redis Pub/Sub 做多节点消息路由
- 连接数与消息速率都有硬限制

关键代码：

- [internal/transport/ws/client.go](/home/chenlongting/go-web/internal/transport/ws/client.go)
- [internal/transport/ws/hub.go](/home/chenlongting/go-web/internal/transport/ws/hub.go)
- [internal/transport/ws/distributed_hub.go](/home/chenlongting/go-web/internal/transport/ws/distributed_hub.go)

### 后台 goroutine

- WebSocket hub 通过独立 `context` 驱动
- 订单过期取消任务按分钟轮询
- 通知、审核等异步链路通过 Redis Streams / PubSub 解耦
- 进程退出时统一走 graceful shutdown

关键代码：

- [cmd/server/main.go](/home/chenlongting/go-web/cmd/server/main.go)
- [cmd/notification-svc/main.go](/home/chenlongting/go-web/cmd/notification-svc/main.go)
- [cmd/moderation-svc/main.go](/home/chenlongting/go-web/cmd/moderation-svc/main.go)

## 3. 事务与幂等审计

### 已处理

- 直接会话创建：使用事务保证 `conversation` 与成员关系一起落库
- AI 助手会话写入：使用事务保证会话与消息的一致性
- 多种写路径依赖数据库唯一约束兜底

### 本轮补强

为避免并发请求导致“重复副作用但 SQL 被 `DO NOTHING` 吞掉”的问题，以下路径已在 repository 层检查 `RowsAffected()`，并把结果映射回领域错误：

- 关注：`user_follows`
- 帖子点赞：`post_likes`
- 评论点赞：`comment_likes`
- 圈子加成员：`group_members`

这样可以避免：

- 重复发 follow/like 事件
- 重复增加 like/member 计数
- 上层误以为写入成功

关键代码：

- [internal/infra/postgres/follow_repo.go](/home/chenlongting/go-web/internal/infra/postgres/follow_repo.go)
- [internal/infra/postgres/post_repo.go](/home/chenlongting/go-web/internal/infra/postgres/post_repo.go)
- [internal/infra/postgres/comment_repo.go](/home/chenlongting/go-web/internal/infra/postgres/comment_repo.go)
- [internal/infra/postgres/group_repo.go](/home/chenlongting/go-web/internal/infra/postgres/group_repo.go)

### 数据库唯一约束

迁移里已经有这些关键唯一约束：

- 用户：`username` / `email`
- 关注：`(follower_id, followee_id)`
- 帖子点赞：`(post_id, user_id)`
- 评论点赞：`(user_id, comment_id)`
- 举报：`(reporter_id, target_type, target_id)`
- 书签：`(user_id, target_type, target_id)`
- 圈子成员：`(group_id, user_id)`
- 活动参与：`(event_id, user_id)`
- 订单：`order_no` / `idempotency_key`

## 4. 性能基线

### HTTP 层

建议面试时直接展示：

- `/metrics` 中的请求量、慢请求和 in-flight 指标
- `/debug/pprof/profile`
- `/debug/pprof/heap`

### 数据库层

仓库里已经有数据库性能诊断脚本和查询分析器：

- [scripts/analyze_performance.sh](/home/chenlongting/go-web/scripts/analyze_performance.sh)
- [internal/infra/postgres/query_analyzer.go](/home/chenlongting/go-web/internal/infra/postgres/query_analyzer.go)

建议重点观测接口：

- `/api/v1/feed`
- `/api/v1/search`
- `/api/v1/conversations`
- `/api/v1/notifications`

## 5. 面试讲法

建议把项目讲成这三个主题：

1. 一个带实时链路和平台治理能力的 Go 社区后端
2. 一个以可观测性、幂等性和诊断能力为目标持续打磨的服务
3. 一个我能清楚解释并发模型、数据一致性和性能排查路径的工程项目
