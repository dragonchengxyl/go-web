# Furry 同好社区平台 - 前端

`apps/web` 是项目的 Next.js 14 前端，承载社区首页、动态流、圈子、活动、私信、通知、创作者页和管理后台页面。

## 技术栈

- **框架**: Next.js 14 App Router
- **语言**: TypeScript 5.x
- **样式**: Tailwind CSS v3
- **状态管理**: Zustand + TanStack Query
- **动画**: Framer Motion
- **组件**: Radix UI + 自定义 UI 组件

## 常用命令

```bash
pnpm --filter web dev
pnpm --filter web lint
pnpm --filter web type-check
pnpm --filter web build
```

## 环境变量

在 `apps/web/.env.local` 中配置：

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 已落地页面

- `/` 首页与社区导览
- `/feed` 关注流
- `/explore` 发现页
- `/search` 全局搜索
- `/posts/create` 发帖页
- `/posts/[id]` 帖子详情
- `/groups` / `/groups/[id]` 圈子列表与详情
- `/events` / `/events/[id]` 活动列表与详情
- `/messages` / `/messages/[id]` 私信会话
- `/notifications` 通知中心
- `/creator` 创作者面板
- `/admin/*` 管理后台

## 说明

- 当前仓库只有一个前端应用 `apps/web`，管理页已合并在同一个 Next.js 应用内。
- 音乐搜索结果目前支持展示，不提供独立专辑详情页入口。
