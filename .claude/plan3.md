# Plan 3: 构建完整的 Furry 爱好者社区平台

> 基于当前已完成的重构基础，系统性地完善成一个真正可用的 Furry 社区。
> 当前日期：2026-03-10

---

## 现状速览

### 已完成（基础骨架）
- 后端：Post / Follow / Chat / Tip 四大领域完整实现
- WebSocket 实时通信基础设施
- 前端页面框架：feed、explore、posts、messages、notifications、profile

### 核心缺口
| 领域 | 问题 |
|------|------|
| 通知系统 | 后端零实现，前端是空壳 |
| 用户主页 | 无公开个人主页（`/users/:id`），无 Furry 特有字段展示 |
| 帖子详情 | 无评论区渲染、无分享、无打赏入口 |
| 探索页 | 无 Tag 筛选，无推荐算法 |
| 创作者工具 | 无图文混排上传，发帖页不支持图片 |
| 遗留代码 | `game_handler.go`、`product_handler.go`、`coupon_handler.go` 等残留 |
| 成就文案 | "下载游戏" 等旧文案未清理 |
| 社区运营 | 无举报、无屏蔽、无管理员审核队列 |

---

## Phase 1 — 核心用户体验打通（最高优先）

> 目标：让用户真实可用，能发帖、看动态、互动、私信

### 1.1 公开个人主页 `/users/:id`
**后端**
- `GET /api/v1/users/:id` 返回公开 profile（已有端点，确认字段含 `furry_name`, `species`）
- 确认 `GET /api/v1/users/:id/posts` 已实现

**前端** 新建 `apps/web/src/app/users/[id]/page.tsx`
- 头像 + 昵称 + Furry 名 + 物种 + 简介
- 关注/被关注数，关注按钮（已登录时显示）
- Tab：动态 / 关注者 / 关注中
- 未登录用户可查看 public 内容

### 1.2 帖子详情页完善 `/posts/:id`
**当前状态**：只渲染 PostCard，无评论

**新增**
- 评论列表（`GET /api/v1/posts/:id/comments`，后端已有评论系统，迁移024已支持 post 类型）
- 发表评论（登录状态）
- 打赏按钮 → 触发 Tip 弹窗（接入 `tip_handler`）
- 分享按钮（复制链接）

### 1.3 发帖页支持图片上传
**当前状态**：只有文字，没有图片

**改造** `apps/web/src/app/posts/create/page.tsx`
- 图片选择（最多 9 张）
- 调用 `/api/v1/upload` 分片上传到 R2
- 上传进度条
- 预览 + 删除

**新增组件** `apps/web/src/components/ui/image-upload.tsx`

### 1.4 通知系统（端到端）
**后端** 新建 `internal/domain/notification/entity.go`
```
type Notification struct {
    ID         string
    UserID     string   // 接收者
    ActorID    string   // 触发者
    Type       string   // like / comment / follow / tip / system
    TargetID   string   // post_id / comment_id 等
    TargetType string
    IsRead     bool
    CreatedAt  time.Time
}
```
- 新建 `internal/infra/postgres/notification_repo.go`
- 新建 `internal/usecase/notification_service.go`
- 新建 `internal/transport/http/handler/notification_handler.go`
- 在点赞、评论、关注、打赏的 usecase 中写入通知
- 通过 WebSocket Hub 推送实时通知

**数据库** 新增 migration 027：`create_notifications` 表

**后端路由**
```
GET  /api/v1/notifications        # 列表（分页）
POST /api/v1/notifications/read   # 标记已读（批量/全部）
GET  /api/v1/notifications/unread-count
```

**前端** 完善 `apps/web/src/app/notifications/page.tsx`
- 按时间倒序展示，区分类型图标
- 点击跳转到对应内容
- Header 铃铛显示未读红点（WebSocket 实时更新）

### 1.5 个人中心 Furry 字段
**改造** `apps/web/src/app/profile/page.tsx`
- 编辑表单新增 `furry_name`（兽名）、`species`（物种）字段
- 头像上传（调用上传接口）
- 把"下载游戏、购买内容"等旧成就文案改为"发帖、评论、关注他人"

---

## Phase 2 — 社区发现与内容生态

> 目标：用户能找到感兴趣的人和内容

### 2.1 探索页强化 `/explore`
**后端**
- `GET /api/v1/explore?tag=xxx` 支持 Tag 筛选（在 SQL 中 `tags @> ARRAY['xxx']`）
- `GET /api/v1/tags/hot` 热门标签（基于最近 7 天帖子 tag 统计）
- `GET /api/v1/users/recommended` 推荐创作者（粉丝数排行，简单实现）

**前端** 改造 `apps/web/src/app/explore/page.tsx`
- 顶部热门 Tag 横向滚动列表，点击筛选
- "推荐创作者" 侧边栏（桌面端）/ 顶部卡片（移动端）
- 瀑布流布局（媒体内容较多时）

### 2.2 搜索功能前端接入
**当前状态**：`search_handler.go` 已实现，前端 `/search` 页面待确认

**改造** `apps/web/src/app/search/page.tsx`
- 搜索框联动 URL query 参数
- 分 Tab：帖子 / 用户 / 标签
- 高亮关键词

### 2.3 标签体系
- 帖子 Tag 可点击 → 跳转到 `/explore?tag=xxx`
- 创建帖子时 Tag 自动补全（基于热门标签）

### 2.4 关注流优化
- 首次进入 feed 未关注任何人时，引导用户去探索页
- 无限滚动改为 Intersection Observer 自动触发加载（代替手动点击"加载更多"）

---

## Phase 3 — 创作者工具与变现

> 目标：让内容创作者愿意在此平台持续创作

### 3.1 创作者主页增强
**新增** `/users/:id` 页面
- "支持创作者"打赏按钮（触发金额选择弹窗）
- 收到的打赏总额展示（需创作者同意公开）

### 3.2 创作者数据仪表盘
**新建** `apps/web/src/app/creator/page.tsx`
- 近 30 天：帖子数、点赞数、新增粉丝、打赏收入
- 帖子列表（按互动量排序）
- 粉丝增长折线图（简单实现，用 recharts）

**后端** 新增 `GET /api/v1/creator/stats` 端点

### 3.3 内容可见性完善
- `followers_only` 内容在 explore 页对非粉丝展示模糊预览 + 关注提示
- 私密帖子仅自己可见（后端已支持，前端处理 403）

### 3.4 音乐内容与帖子联动
- 发帖时可关联一首曲目（从已有音乐系统选取）
- PostCard 显示关联音乐条（点击可播放）

---

## Phase 4 — 社区健康与运营

> 目标：维持社区氛围，防止垃圾内容

### 4.1 举报系统
**后端** 新建 `internal/domain/report/` 领域
- 举报帖子 / 评论 / 用户
- 举报类型：色情、广告、骚扰、违规

**数据库** migration 028：`reports` 表

**前端**
- PostCard 三点菜单 → "举报"选项
- 举报弹窗（选择原因 + 可选补充说明）

### 4.2 屏蔽用户
- `POST /api/v1/users/:id/block` / `DELETE` 取消屏蔽
- 屏蔽后在 feed 和 explore 中过滤该用户内容

**数据库** migration 029：`user_blocks` 表

### 4.3 管理员审核队列
**改造** `apps/web/src/app/admin/` 页面（已存在）
- 举报内容审核列表
- 快速操作：删除内容 / 警告用户 / 封禁用户

### 4.4 内容审核基础
- 发帖时关键词过滤（黑名单，配置在 Redis）
- 图片大小/格式限制前端校验

---

## Phase 5 — 质量与技术债清理

> 目标：可维护、可扩展、性能稳定

### 5.1 清理遗留代码
| 文件 | 处理方式 |
|------|---------|
| `game_handler.go` | 删除（或保留空文件注释说明已移除） |
| `product_handler.go` | 删除 |
| `coupon_handler.go` | 删除 |
| `branch_handler.go` | 删除 |
| `release_handler.go` | 删除 |
| `asset_handler.go` | 评估是否复用于媒体资产 |
| `admin_handler.go.bak` | 删除 |

同步清理 `router.go` 中对应路由注册。

### 5.2 前端组件库补全
优先新增：
- `Avatar` 组件（头像 + fallback 首字母）
- `InfiniteScroll` 组件（Intersection Observer）
- `ImageGallery` 组件（帖子多图展示，支持点击放大）
- `TipModal` 组件（打赏金额选择）
- `ConfirmDialog` 组件（删除确认等操作）
- `Toast` 通知（替代当前 `alert()`）

### 5.3 数据库迁移验证
确认 018-026 已在生产/开发 DB 执行，补充任何缺失索引：
- `notifications(user_id, is_read, created_at)`
- `posts(tags)` GIN 索引（支持 Tag 搜索）

### 5.4 后端单元测试
重点覆盖：
- `post_service.go`：创建、点赞、Feed 流逻辑
- `follow_service.go`：关注/取关防重复
- `notification_service.go`：事件写入

### 5.5 性能优化
- Feed 流引入 Redis 缓存（关注列表、最新帖子 ID 列表）
- 帖子详情页 SSR（Next.js Server Component，对 SEO 有帮助）
- 图片懒加载（PostCard 中 img 加 `loading="lazy"`）

---

## 执行顺序建议

```
Phase 1（必须先做，用户基本可用）
  → 1.4 通知系统（依赖 WebSocket，是后续功能基础）
  → 1.1 公开个人主页（社区核心）
  → 1.2 帖子详情评论区
  → 1.3 发帖图片上传
  → 1.5 个人中心 Furry 字段

Phase 2（探索与增长）
  → 2.1 探索页 Tag 筛选
  → 2.2 搜索前端接入
  → 2.4 关注流 UX 优化

Phase 3（变现，可与 Phase 2 并行）
  → 3.2 创作者仪表盘
  → 3.1 打赏入口

Phase 4（运营，上线前必须）
  → 4.1 举报系统
  → 4.2 屏蔽用户

Phase 5（持续进行）
  → 5.1 清理遗留代码（随时可做）
  → 5.2 组件库补全（穿插 Phase 1-2 进行）
```

---

## 数据库迁移计划

| 编号 | 内容 |
|------|------|
| 027  | `create_notifications` 表 |
| 028  | `create_reports` 表 |
| 029  | `create_user_blocks` 表 |
| 030  | `posts` 表增加 GIN 索引（tags 列） |

---

## 关键设计决策

1. **通知实时推送**：复用现有 WebSocket Hub，新增消息类型 `notification`，避免引入额外长连接
2. **Tag 搜索**：PostgreSQL `text[]` 列 + GIN 索引，简单高效，不引入 Elasticsearch
3. **推荐算法**：初期用粉丝数排行，后期可迭代为协同过滤
4. **图片存储**：沿用 Cloudflare R2，分片上传已有基础设施
5. **前端状态管理**：沿用现有 `@tanstack/react-query`，不引入新状态库
