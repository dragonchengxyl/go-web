# 🗺️ Frontend Vision & Implementation Blueprint
## A+B 混合路线 — 极速同步 × 实时优先

**生成日期**: 2026-03-10
**方向决议**: A（极速同步）+ B（实时优先）混合
**交付周期**: ~3 周
**核心目标**: 前后端完全对齐 + WebSocket 升级为应用底座，驱动实时体验飞轮。

---

## 🔭 战略方向决议

### 放弃的路线

| 方向 | 放弃原因 |
|------|----------|
| 纯 A（极速同步） | 单纯同步无技术护城河，用户体验仍停留在轮询时代 |
| 纯 B（实时优先） | 跳过后端同步会导致数据层不一致，TypeScript 类型错误潜伏 |
| 纯 C（内容生态） | 依赖 pgvector 接口尚未前端暴露，时机不成熟 |

### 选定路线的核心逻辑

```
Week 1 (A): 先还债 — 补齐 TypeScript 类型、切换 OSS 直传、
            暴露审核状态、完善赞助页
Week 2-3 (B): 再投资 — 抽取全局 WSContext、乐观更新、
              实时通知、草稿保存
```

**A 是 B 的前提**：如果 Post 接口不包含 `moderation_status`，审核通过的 WS 推送就无法更新 UI。

---

## ⚖️ 核心技术选型

### T-1: WebSocket 状态管理方案

| 方案 | 优点 | 缺点 | 决策 |
|------|------|------|------|
| Redux / Zustand store | 全局可见 | 引入新依赖，over-engineering | ❌ |
| Context + useReducer | 零依赖，可组合 | 需手写 | ✅ **选定** |
| 每页面独立 WS 连接 | 简单 | 多连接浪费，状态割裂 | ❌ |

**决策**：`WSContext` (React Context + useReducer) 挂载在 `layout.tsx`，单连接全局共享，`subscribe/unsubscribe` 事件分发模式。

### T-2: 图片上传方案

| 方案 | 优点 | 缺点 | 决策 |
|------|------|------|------|
| 服务器中转 `/upload/image` | 已实现，简单 | 占服务器带宽，速度慢 | ❌ 废弃 |
| OSS 直传 `/upload/oss-policy` | 零服务器带宽，快 | 需 XHR 进度事件 | ✅ **选定** |
| 第三方上传 SDK | 功能齐全 | 引入新依赖 | ❌ |

**决策**：后端返回签名 Policy → 前端 `XMLHttpRequest` 直传 OSS（支持 `upload.onprogress`）。

### T-3: 乐观更新策略

React Query `useMutation` 的 `onMutate` → `onError` rollback 模式，不引入额外库。

---

## 🛡️ 威胁建模与防线

| 威胁 | 攻击向量 | 防线 |
|------|----------|------|
| OSS Policy 泄露 | 客户端日志/DevTools | 不 console.log Policy/Signature；Policy 5 分钟过期 |
| WS 洪泛重连 | 服务重启时大量客户端同时重连 | 指数退避：1→2→4→8→16→30s cap |
| XSS via 帖子内容 | 恶意帖子注入 `<script>` | 禁用 `dangerouslySetInnerHTML`，使用 `whitespace-pre-wrap` text 渲染 |
| 大文件绕过 | 前端绕过改包 | 客户端+服务端双重校验（前端 10MB 硬限，OSS Policy `content-length-range`） |
| 客户端伪造审核状态 | 修改 `moderation_status` 前端参数 | 服务端以数据库为准，`moderation_status` 从不作为发帖 API 入参 |
| Token 过期后 WS 持续连接 | 长会话 Token 自然过期 | WS close code 4001 → 触发 Token Refresh 后重连 |

---

## 📝 Issue 级别执行清单 (Phase 2 → Phase 5 严格排序)

> 严格按编号顺序。Phase 2 (Infra) 是所有业务 Issue 的前置依赖。

---

### Phase 2: 基础设施 (Infrastructure)

---

#### [INFRA-A01] TypeScript 接口全量同步

**优先级**: P0 — 所有其他 Issue 的类型前提
**文件**: `apps/web/src/lib/api-client.ts`

**实现要点**:

```typescript
// 新增枚举
export type ModerationStatus = 'pending' | 'approved' | 'blocked'

// 更新 Post 接口
export interface Post {
  id: string
  author_id: string
  title?: string
  content: string
  media_urls?: string[]
  tags?: string[]
  content_labels?: Record<string, boolean>  // ← 新增: { is_ai_generated: true }
  visibility: 'public' | 'followers_only' | 'private'
  moderation_status: ModerationStatus       // ← 新增
  like_count: number
  comment_count: number
  is_pinned: boolean
  created_at: string
  updated_at: string
  author_username?: string
  author_avatar_key?: string
  is_liked_by_me?: boolean
}

// 更新 createPost 方法签名
async createPost(data: {
  title?: string
  content: string
  media_urls?: string[]
  tags?: string[]
  visibility?: string
  is_ai_generated?: boolean  // ← 新增
})

// 新增 OSS Policy 接口
export interface OSSUploadPolicy {
  host: string
  OSSAccessKeyId: string
  policy: string
  signature: string
  expire: number
  dir: string
}

async getOSSPolicy(purpose?: string): Promise<OSSUploadPolicy>
```

---

#### [INFRA-A02] OSS 直传 Hook

**优先级**: P0 — 替换所有图片上传路径
**文件新增**: `apps/web/src/hooks/use-oss-upload.ts`

**实现逻辑**:

```
1. 调用 apiClient.getOSSPolicy(purpose)
2. 客户端校验: 文件类型白名单 + 10MB 大小限制
3. 构造 FormData: { key: dir+uuid+ext, OSSAccessKeyId, policy, signature, success_action_status: 200, Content-Type }
4. XMLHttpRequest POST 到 policy.host，监听 upload.onprogress → 更新进度状态
5. 返回: { upload, progress, error }
```

**接口**:
```typescript
function useOSSUpload(): {
  upload: (file: File, purpose?: string) => Promise<string>  // 返回对象 URL
  progress: number   // 0-100
  uploading: boolean
  error: string | null
}
```

**安全要点**:
- Policy/Signature 不写入 localStorage，不打印到 console
- key 格式: `{purpose}/{uid}/{date}/{uuid}{ext}`（uid 来自后端 Policy dir 字段）
- 文件类型校验: `['image/jpeg','image/png','image/gif','image/webp']`

---

#### [INFRA-B01] 全局 WSContext

**优先级**: P0 — B 方向所有功能的底座
**文件新增**: `apps/web/src/contexts/ws-context.tsx`

**接口设计**:

```typescript
interface WSContextValue {
  status: 'connecting' | 'connected' | 'disconnected'
  subscribe: (type: string, handler: (payload: unknown) => void) => () => void
  // subscribe 返回 unsubscribe 函数，用于 useEffect cleanup
}

// 使用示例
const { subscribe } = useWS()
useEffect(() => {
  return subscribe('notification', (payload) => {
    setUnreadCount(c => c + 1)
  })
}, [])
```

**连接生命周期**:
```
layout.tsx mounted → WSProvider
  → if (token) connect ws://host/ws/chat?token=xxx
  → onmessage → 解析 { type, payload } → 分发给所有订阅者
  → onclose → 指数退避重连 (delays: [1,2,4,8,16,30]s, max 30s)
  → token 变化 → 断开重连
  → 用户登出 → 关闭连接，清空订阅
```

**退避算法**:
```typescript
const BACKOFF = [1000, 2000, 4000, 8000, 16000, 30000]
let attempt = 0
const delay = BACKOFF[Math.min(attempt++, BACKOFF.length - 1)]
```

---

#### [INFRA-B02] WS 消息路由策略

**优先级**: P1
**说明**: 不是单独文件，是在各业务 Hook 中通过 `subscribe` 实现的约定。

```
ws message type → 处理动作
─────────────────────────────────────────────────────
'notification'     → 通知中心: unreadCount+1, queryClient.invalidateQueries(['notifications'])
'chat'             → 会话列表: 更新 last_message, queryClient.invalidateQueries(['messages', convId])
'post_moderation'  → 帖子详情: 更新 post.moderation_status (通过 queryClient.setQueryData)
```

**WS 消息约定** (与后端 WSMessage 格式对齐):
```typescript
interface WSMessage {
  type: 'chat' | 'notification' | 'post_moderation' | 'ping' | 'pong'
  conversation_id?: string
  payload: unknown
}
```

---

### Phase 3: 安全防御 (Security)

---

#### [SEC-A01] OSS 上传客户端安全加固

**优先级**: P0
**实现要点**:
- Policy/Signature **禁止** `console.log`，只在内存中使用
- `XMLHttpRequest` timeout 设置 30s，避免挂起
- 上传前校验: `file.size <= 10 * 1024 * 1024`（与服务端 Policy 条件一致）
- 上传前校验: `['image/jpeg','image/png','image/gif','image/webp'].includes(file.type)`
- `AbortController` 支持：用户离开页面时取消进行中的上传

---

#### [SEC-B01] WS 重连安全

**优先级**: P1
**实现要点**:
- 重连前检查 localStorage token 是否仍有效（`exp` 字段）
- 如 Token 已过期 → 先调 `/auth/refresh` → 再重连
- 同一时刻最多 1 个 pending 重连 Timer（清除旧 Timer）
- 页面不可见时（`document.visibilityState === 'hidden'`）暂停重连，可见时恢复

---

#### [SEC-B02] 内容渲染 XSS 防御审计

**优先级**: P1
**实现要点**:
- 全局搜索 `dangerouslySetInnerHTML`，确认无用法（当前 PostCard 用 `whitespace-pre-wrap` ✓）
- `PostCard.content` 渲染保持 text-only，不支持 Markdown/HTML
- `author_username` 在渲染前不做 HTML 解码（Next.js JSX 默认转义 ✓）

---

### Phase 4: 测试先行 (TDD)

---

#### [TEST-B01] WSContext 单元测试

**优先级**: P1
**文件**: `apps/web/src/__tests__/ws-context.test.tsx`

**测试用例**:
```
✓ 挂载时使用正确 token 建立连接
✓ onclose 后按退避策略重连
✓ subscribe 注册的 handler 在收到对应 type 消息时被调用
✓ 同一 type 多个 subscribe handler 都被调用
✓ unsubscribe 后 handler 不再被调用
✓ 用户登出后 WS 关闭
```

---

#### [TEST-A01] OSS 上传 Hook 测试

**优先级**: P1
**文件**: `apps/web/src/__tests__/use-oss-upload.test.ts`

**测试用例**:
```
✓ 超过 10MB 文件被拒绝（不调用 API）
✓ 非白名单 MIME 类型被拒绝
✓ happy path: getOSSPolicy → XHR POST → 返回 URL
✓ API 失败时 error 状态被正确设置
✓ progress 从 0 → 100 正确触发回调
```

---

### Phase 5: 业务落地 (Business)

---

#### [BIZ-A01] createPost — AI 生成标签 + OSS 直传迁移

**优先级**: P0
**文件**: `apps/web/src/app/posts/create/page.tsx`

**变更点**:
1. 引入 `useOSSUpload` hook，替换 `apiClient.uploadFile('/upload/image', file)` 调用
2. 每张图片显示独立上传进度条 (0-100%)
3. 新增 AI 标签勾选项：
```tsx
<label className="flex items-center gap-2 text-sm text-gray-600">
  <input
    type="checkbox"
    checked={isAIGenerated}
    onChange={e => setIsAIGenerated(e.target.checked)}
    className="rounded"
  />
  <span>此内容包含 AI 生成内容（请如实标注）</span>
</label>
```
4. `createPost` 调用增加 `is_ai_generated: isAIGenerated`
5. 草稿自动保存（见 BIZ-B05）

---

#### [BIZ-A02] Feed/Explore — 审核状态展示

**优先级**: P0
**文件**: `apps/web/src/components/post/post-card.tsx`

**展示规则**:
```
moderation_status === 'pending'
  → 灰色半透明遮罩 (opacity-60)
  → 右上角 Badge: "审核中" (yellow)
  → 互动按钮禁用（点赞、评论）

moderation_status === 'blocked'
  → 强遮罩 + "内容不符合社区规范"
  → 仅作者自己可见（后端已过滤，此处是防御性展示）
```

**TypeScript 实现**:
```tsx
{post.moderation_status === 'pending' && (
  <div className="absolute inset-0 bg-gray-100/60 rounded-xl flex items-center justify-center z-10">
    <span className="bg-yellow-100 text-yellow-800 text-xs px-2 py-1 rounded-full">
      ⏳ 审核中
    </span>
  </div>
)}
```

---

#### [BIZ-A03] Explore — AI 标签过滤 Toggle

**优先级**: P1
**文件**: `apps/web/src/app/explore/page.tsx`

**变更点**:
- 在 tag 列表旁新增 Toggle 按钮组:
  ```
  [ 全部 ] [ 人工创作 ] [ AI 生成 ]
  ```
- 前端过滤（`content_labels?.is_ai_generated === true/false`）
- 默认"全部"

---

#### [BIZ-A04] 赞助页完善

**优先级**: P1
**文件**: `apps/web/src/app/sponsor/page.tsx`

**增强内容**:
1. 进度条 mount 动画 (CSS `transition-all duration-1000 ease-out`)
2. "本月剩余" 金额展示: `¥{(goal - raised).toFixed(2)}`
3. 鸣谢名录 section（静态 Placeholder，5个匿名ID格式）
4. 一键复制功能（点击收款码旁按钮复制"请转账至XXX"提示文字）
5. Header 导航增加"赞助"入口（在 Profile 旁边加 ❤️ 图标）

---

#### [BIZ-B01] 乐观点赞更新

**优先级**: P1
**文件**: `apps/web/src/components/post/post-card.tsx`

**模式**: React Query `useMutation` with `onMutate` + rollback

```typescript
const likeMutation = useMutation({
  mutationFn: () => apiClient.likePost(post.id),
  onMutate: () => {
    // 立即更新本地状态
    setLiked(true)
    setLikeCount(c => c + 1)
    return { prevLiked: liked, prevCount: likeCount }
  },
  onError: (_, __, context) => {
    // 回滚
    setLiked(context!.prevLiked)
    setLikeCount(context!.prevCount)
  }
})
```

心形动画: `animate-bounce` 触发一次（使用 `key` prop 重置动画）

---

#### [BIZ-B02] 实时审核状态通知

**优先级**: P1
**文件**: `apps/web/src/app/posts/create/page.tsx` 或全局 Toast

**流程**:
```
WS message: { type: 'notification', payload: { type: 'system', ... } }
  → 如果 payload 包含 post_id 且 status === 'approved'
  → 显示 Toast: "✓ 您的帖子已通过审核"
  → queryClient.invalidateQueries(['post', postId])
```

**Toast 实现**: 简单的 CSS 动画 div（不引入新库），4s 后自动消失。

---

#### [BIZ-B03] Header 实时未读角标（替换 60s 轮询）

**优先级**: P1
**文件**: `apps/web/src/components/layout/header.tsx`

**变更点**:
```typescript
// 删除
setInterval(fetchUnread, 60000)

// 新增
const { subscribe } = useWS()
useEffect(() => {
  return subscribe('notification', () => {
    setUnreadCount(c => c + 1)
  })
}, [subscribe])
```

初始加载仍调用一次 `getUnreadCount()`（WS 只推增量，不提供总量）。

---

#### [BIZ-B04] 消息列表实时更新

**优先级**: P1
**文件**: `apps/web/src/app/messages/page.tsx`

**变更点**:
```typescript
useEffect(() => {
  return subscribe('chat', (payload: any) => {
    // 更新会话列表的 last_message
    setConversations(convs => convs.map(c =>
      c.id === payload.conversation_id
        ? { ...c, last_message: payload, unread_count: (c.unread_count ?? 0) + 1 }
        : c
    ))
  })
}, [subscribe])
```

---

#### [BIZ-B05] 发帖草稿自动保存

**优先级**: P2
**文件**: `apps/web/src/app/posts/create/page.tsx`

**实现**:
```typescript
const DRAFT_KEY = 'post_draft'

// 防抖保存 (2s)
useEffect(() => {
  const timer = setTimeout(() => {
    if (content || title) {
      localStorage.setItem(DRAFT_KEY, JSON.stringify({ title, content, tags }))
    }
  }, 2000)
  return () => clearTimeout(timer)
}, [title, content, tags])

// 挂载时恢复
useEffect(() => {
  const draft = localStorage.getItem(DRAFT_KEY)
  if (draft) {
    const { title, content, tags } = JSON.parse(draft)
    // 提示用户: "检测到未保存草稿，是否恢复？"
  }
}, [])

// 提交后清除
localStorage.removeItem(DRAFT_KEY)
```

草稿恢复弹窗: 简单 `confirm()` 或内联 Banner。

---

## 📊 执行优先级汇总

```
Week 1 (P0 同步对齐):
  INFRA-A01  TypeScript 接口全量同步
  INFRA-A02  OSS 直传 Hook
  BIZ-A01    createPost AI 标签 + OSS 迁移
  BIZ-A02    Feed/Explore 审核状态展示
  SEC-A01    OSS 上传客户端安全加固

Week 2 (P0/P1 实时底座):
  INFRA-B01  全局 WSContext (单连接 + 退避重连)
  INFRA-B02  WS 消息路由策略
  SEC-B01    WS 重连安全
  BIZ-B03    Header 实时未读（替换轮询）
  BIZ-B04    消息列表实时更新

Week 3 (P1 体验升级):
  TEST-B01   WSContext 单元测试
  TEST-A01   OSS 上传 Hook 测试
  SEC-B02    XSS 防御审计
  BIZ-A03    Explore AI 标签过滤
  BIZ-A04    赞助页完善
  BIZ-B01    乐观点赞更新
  BIZ-B02    实时审核通知 Toast
  BIZ-B05    草稿自动保存
```

---

## 🚨 关键风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| OSS 跨域 CORS 未配置 | 高 | 高 | 阿里云控制台提前配置 Bucket CORS 规则（允许 POST） |
| WSContext 循环重渲染 | 中 | 中 | `subscribe/unsubscribe` 用 `useCallback` + `useRef` 存 handlers，避免引用变化 |
| 乐观更新 rollback 失败 | 低 | 低 | `onError` 始终有 context 参数，TypeScript 强类型保证 |
| 草稿恢复循环弹窗 | 低 | 低 | 恢复后立即清除 DRAFT_KEY（无论选择恢复还是放弃） |
| WS Token 刷新竞态 | 中 | 中 | Token Refresh 操作加锁（单例 Promise），多个 WS/fetch 共用同一刷新流程 |

---

## ✅ 完成标志（全部满足才算完成）

- [ ] `pnpm build` 无 TypeScript 错误
- [ ] Post 接口包含 `moderation_status` 和 `content_labels`
- [ ] 图片上传经由 OSS Policy（`/upload/image` 不再被调用）
- [ ] 全局 WSContext 挂载在 layout.tsx
- [ ] Header 不再使用 60s setInterval 轮询
- [ ] `pending` 帖子在 PostCard 显示"审核中"遮罩
- [ ] createPost 有 AI 标签勾选框
- [ ] 代码已 `git commit`
- [ ] 无任何 `// TODO`、`// placeholder`、`...` 残留

---

*Frontend Blueprint v1 — A+B Mixed | 2026-03-10*
