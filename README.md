# 独立游戏工作室全矩阵中台平台

一个功能完整的独立游戏工作室全矩阵中台与社区平台，包含游戏分发、音乐流媒体、电商系统、社区互动等功能。

## 项目架构

### 后端 (Go)
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + Redis
- **架构**: DDD (领域驱动设计)
- **认证**: JWT + RBAC

### 前端 (Next.js)
- **框架**: Next.js 14 (App Router)
- **语言**: TypeScript 5.x
- **样式**: Tailwind CSS
- **状态管理**: Zustand + React Query
- **动画**: Framer Motion

## 已实现的 Epic

### ✅ Epic 1 - 基础设施与云原生基座
- PostgreSQL 数据库连接池
- Redis 缓存
- 全局中间件（日志、限流、CORS、认证）
- 配置管理

### ✅ Epic 2 - 核心通行证与用户中心
- 用户注册/登录（JWT）
- RBAC 权限控制（7个角色级别）
- 用户信息管理
- Token 黑名单机制

### ✅ Epic 3 - 游戏分发与版本控制中台
- 游戏元数据管理
- 分支管理（main/beta/demo）
- 版本发布系统
- 用户资产管理
- 安全下载接口

### ✅ Epic 4 - 沉浸式流媒体与 OST 平台
- 专辑/音轨管理
- HiFi 音频元数据
- 音频流媒体接口

### ✅ Epic 5 - 高频互动与创作者社区
- 评论系统（支持嵌套）
- 点赞功能
- 软删除机制
- 多态关联（游戏/专辑/音轨）

### ✅ Epic 7 - 电商与支付系统
- 商品管理（游戏/DLC/OST/捆绑包/会员）
- 订单系统（状态机、幂等性）
- 折扣规则
- 优惠券系统
- 兑换码生成与管理
- 支付网关框架（支付宝/微信/Stripe/PayPal）

### ✅ Epic 8 - 前端全站架构
- Next.js 14 项目搭建
- Monorepo 架构（Turborepo）
- 响应式首页
- UI 组件库
- API 客户端封装
- 主题切换（亮色/暗色）

## 快速开始

### 后端

1. 安装依赖
```bash
go mod download
```

2. 启动服务
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

### 前端

1. 安装依赖
```bash
pnpm install
```

2. 启动开发服务器
```bash
cd apps/web
pnpm dev
```

前端将在 `http://localhost:3000` 启动

## 技术特性

### 后端
- ✅ DDD 架构设计
- ✅ Repository 模式
- ✅ JWT 认证 + RBAC 权限
- ✅ Redis Token 黑名单
- ✅ 请求限流
- ✅ 统一错误处理
- ✅ 结构化日志
- ✅ 幂等性保证
- ✅ 软删除支持

### 前端
- ✅ TypeScript 类型安全
- ✅ 响应式设计
- ✅ 暗色模式
- ✅ 动画效果
- ✅ SEO 优化
- ✅ 性能优化
- ✅ Monorepo 架构

## 许可证

MIT License
