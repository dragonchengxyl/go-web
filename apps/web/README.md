# 独立游戏工作室平台 - 前端

这是一个基于 Next.js 14 的现代化独立游戏工作室平台前端项目。

## 技术栈

- **框架**: Next.js 14 (App Router)
- **语言**: TypeScript 5.x
- **样式**: Tailwind CSS v3
- **状态管理**: Zustand + React Query
- **动画**: Framer Motion
- **UI组件**: 基于 Radix UI 的自研组件库
- **图标**: Lucide React
- **表单**: React Hook Form + Zod

## 项目结构

```
apps/
├── web/          # 主站 (Next.js)
├── admin/        # 管理后台 (Next.js)
└── docs/         # 开发者文档 (Nextra)

packages/
├── ui/           # 共享 UI 组件库
├── api-client/   # API 客户端
├── config/       # 共享配置
└── utils/        # 工具函数
```

## 开发指南

### 安装依赖

```bash
pnpm install
```

### 启动开发服务器

```bash
# 启动所有应用
pnpm dev

# 仅启动主站
cd apps/web && pnpm dev
```

### 构建生产版本

```bash
pnpm build
```

## 环境变量

在 `apps/web` 目录下创建 `.env.local` 文件：

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## 核心功能

### 已实现

- ✅ 响应式导航栏（固定顶部，滚动渐变）
- ✅ Hero 区域（视差效果，动画）
- ✅ 游戏卡片组件
- ✅ 基础 UI 组件库（Button 等）
- ✅ API 客户端封装
- ✅ 主题切换（亮色/暗色模式）
- ✅ 国际化准备

### 开发中

- 🚧 游戏详情页
- 🚧 OST 播放器
- 🚧 用户认证流程
- 🚧 购物车功能
- 🚧 订单管理
- 🚧 社区评论系统

## 性能优化

- 使用 Next.js Image 组件优化图片加载
- 代码分割和懒加载
- 字体优化（next/font）
- 服务端渲染（SSR）和静态生成（SSG）

## SEO 优化

- 动态 meta 标签
- 结构化数据（Schema.org）
- Sitemap 自动生成
- robots.txt 配置

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License
