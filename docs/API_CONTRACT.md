# API Contract

面试时优先讲这个文件，不需要逐个接口背诵实现。

## 基础约定

- 基础路径：`/api/v1`
- 鉴权方式：`Authorization: Bearer <access_token>`
- 成功响应：统一由 `response.Success` 返回，核心字段为 `code/message/data/request_id/timestamp`
- 失败响应：统一由 `response.Error` 返回，业务错误码定义在 [internal/pkg/apperr/codes.go](/home/chenlongting/go-web/internal/pkg/apperr/codes.go)
- 请求追踪：每个请求都有 `X-Request-ID`，服务端会回写到响应头和结构化日志中

## 分页约定

- 大多数列表接口使用 `page` + `page_size`
- 默认页大小通常为 `20`
- 服务层会限制上限，避免单次查询拉取过多数据
- 返回体通常至少包含：

```json
{
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "size": 20
  }
}
```

不同模块字段名会有差异，例如 `posts/comments/groups/notifications`，但分页语义一致。

## 鉴权分层

- 公开接口：搜索、探索、帖子详情、公开圈子/活动、排行榜、赞助页
- 登录后接口：发帖、关注、评论、私信、通知、书签、举报、屏蔽
- 管理接口：`/api/v1/admin/*`
- WebSocket：`/ws/chat`，沿用 JWT 鉴权

## 关键业务接口

### 认证

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`

### 社区

- `GET /api/v1/feed`
- `GET /api/v1/explore`
- `POST /api/v1/posts`
- `POST /api/v1/posts/:id/like`
- `POST /api/v1/comments`
- `POST /api/v1/users/:id/follow`

### 实时

- `GET /ws/chat`
- `GET /api/v1/conversations`
- `POST /api/v1/conversations/:id/messages`
- `GET /api/v1/notifications`

### 群组与活动

- `GET /api/v1/groups`
- `POST /api/v1/groups`
- `POST /api/v1/groups/:id/join`
- `GET /api/v1/events`
- `POST /api/v1/events`

## WebSocket 事件

- `chat`
- `notification`
- `ping`
- `pong`

服务端会限制：

- 单用户最多 5 条连接
- 每连接使用 token bucket 做消息速率控制
- 浏览器来源必须匹配 `allow_origins`

## 调试与诊断

- 指标：`/metrics`
- 健康检查：`/health`、`/ready`
- 性能剖析：`/debug/pprof/*`

`pprof` 在 `release` 模式下需要管理员 JWT，避免生产环境裸露诊断面。
