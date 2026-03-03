# 独立游戏工作室全矩阵中台与社区架构蓝图 (Master Plan)

**Version:** 2.0.0
**Architecture:** Cloud-Native Monorepo / Domain-Driven Design (DDD)
**Core Stack:** Go, PostgreSQL, Redis, OSS/CDN, Docker/K8s
**Frontend Stack:** React / Next.js, TypeScript, Tailwind CSS
**Last Updated:** 2026-03-03

---

## 目录 (Table of Contents)

1. [项目愿景与核心价值观](#epic-0)
2. [Epic 1 - 基础设施与云原生基座](#epic-1)
3. [Epic 2 - 核心通行证与用户中心](#epic-2)
4. [Epic 3 - 游戏分发与版本控制中台](#epic-3)
5. [Epic 4 - 沉浸式流媒体与 OST 平台](#epic-4)
6. [Epic 5 - 高频互动与创作者社区](#epic-5)
7. [Epic 6 - 开发者与运维自动化流水线](#epic-6)
8. [Epic 7 - 电商与支付系统](#epic-7)
9. [Epic 8 - 前端全站架构](#epic-8)
10. [Epic 9 - 管理后台系统](#epic-9)
11. [Epic 10 - 成就与游戏化系统](#epic-10)
12. [Epic 11 - 数据分析与商业智能](#epic-11)
13. [Epic 12 - 本地化与国际化](#epic-12)
14. [Epic 13 - 邮件营销与推送通知](#epic-13)
15. [Epic 14 - 搜索引擎与推荐系统](#epic-14)
16. [Epic 15 - 客户支持与工单系统](#epic-15)
17. [Epic 16 - 开发者 SDK 与开放 API](#epic-16)
18. [Epic 17 - 移动端适配与 PWA](#epic-17)
19. [Epic 18 - 数据安全、合规与隐私](#epic-18)
20. [Epic 19 - 容灾、备份与高可用](#epic-19)
21. [Epic 20 - 性能工程与压力测试](#epic-20)
22. [数据库 ER 总览](#db-overview)
23. [接口规范与 API 契约](#api-contract)
24. [项目里程碑与优先级矩阵](#milestones)
25. [技术债务管理](#tech-debt)

---

## 摘要与 AI 交互守则

本文件为工作室核心后端的宏观开发计划。AI 在读取本计划时，需遵循以下原则：

1. **模块化解耦**：严格遵循定义好的接口边界，不允许跨域直接读写数据库。每个 Epic 均为独立的 Go 模块或微服务，通过内部 gRPC / 事件总线通信。
2. **渐进式实现**：优先实现 [Epic 1] 和 [Epic 2] 的核心链路，将复杂特性（如推荐算法、增量更新）推迟至后期迭代。
3. **面向容灾**：所有涉及文件（游戏本体、音乐）的操作，必须采用云端对象存储配合预签名 URL (Pre-signed URL)，严禁本地物理读写。
4. **安全优先**：每个接口必须经过认证鉴权检查，SQL 操作必须使用参数化查询，前端输入必须经过验证与转义。
5. **可观测性**：每个关键操作必须有日志记录和指标上报，确保生产问题可追溯。
6. **向后兼容**：API 版本变更需通过版本号（`/api/v1/`、`/api/v2/`）区分，旧版本需保留至少 6 个月。

---

## [Epic 0] 项目愿景与核心价值观 {#epic-0}

### 目标
打造一个面向全球玩家的专业独立游戏工作室官方平台，集游戏分发、音乐欣赏、玩家社区、创作者交流为一体的大型综合型网站。

### 核心功能定位

```
┌─────────────────────────────────────────────────────────┐
│                     游戏工作室官网                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │
│  │  游戏商城  │  │  社区论坛  │  │  OST平台  │  │ 创作者  │  │
│  │ 分发平台  │  │  玩家互动  │  │  音乐欣赏  │  │  博客  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │
│  │  成就系统  │  │  活动中心  │  │  支持中心  │  │ 开发者  │  │
│  │  游戏化   │  │  赛事报名  │  │  工单系统  │  │  SDK  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 非功能性需求 (NFR)

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 可用性 | 99.9% | 允许每月约 43 分钟停机 |
| 响应时间 (P99) | < 200ms | 核心 API 接口 |
| 并发用户 | 10,000+ | 峰值同时在线 |
| 文件下载峰值带宽 | 10 Gbps | 游戏新版本发布期间 |
| 数据备份 RPO | < 1 小时 | 最大数据丢失容忍 |
| 故障恢复 RTO | < 15 分钟 | 最大服务恢复时间 |
| 安全扫描 | 每次 CI | 自动漏洞扫描 |

### 技术选型理由

**Go (Golang) 作为后端核心语言**：
- 天然的高并发支持（Goroutine）
- 编译型语言，部署产物为单一二进制
- 标准库强大，生态成熟（Gin/Echo/Fiber 框架）
- 静态类型，减少运行时错误

**PostgreSQL 作为主数据库**：
- JSONB 支持灵活的半结构化数据
- 强事务一致性，适合电商支付场景
- 丰富的全文搜索能力（tsvector）
- 活跃的开源社区与云原生支持

**Redis 作为缓存与消息队列**：
- 亚毫秒级读取延迟
- 丰富的数据结构（Stream、Sorted Set、HyperLogLog）
- 支持发布/订阅模式实现简单消息队列

---

## [Epic 1] 基础设施与云原生基座 (Infrastructure Base) {#epic-1}

**目标**：构建高可用、易扩展的底层运行环境，接管所有网络流量与监控。

### [Feature 1.1] 流量接入与网关层 (API Gateway)

* **[Task 1.1.1]** 配置 Nginx / Traefik 作为全局反向代理，处理 SSL 证书自动化续期 (Let's Encrypt)。
  - 配置 HTTP/2 与 QUIC (HTTP/3) 协议支持
  - 开启 Brotli 压缩（比 gzip 压缩率高约 20%）
  - 配置 HSTS、CSP、X-Frame-Options 等安全响应头

* **[Task 1.1.2]** 设计全局统一的 HTTP 响应结构：
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {},
    "request_id": "uuid-v4",
    "timestamp": 1700000000
  }
  ```
  错误码规范：
  - `0` 成功
  - `4xxxx` 客户端错误（40001 参数错误，40101 未认证，40301 无权限，40401 资源不存在）
  - `5xxxx` 服务端错误（50001 内部错误，50002 依赖服务不可用）

* **[Task 1.1.3]** 实现基于 IP 和 User-ID 的全局限流策略 (Redis Token Bucket)：
  - 未认证用户：60 次/分钟（按 IP）
  - 认证普通用户：200 次/分钟（按 User-ID）
  - 管理员：1000 次/分钟
  - 下载接口：5 次/小时（按 User-ID，独立计数）
  - 超限响应 `429 Too Many Requests`，Header 返回 `Retry-After` 秒数

* **[Task 1.1.4]** 请求链路追踪：
  - 网关注入 `X-Request-ID` 和 `X-Trace-ID`
  - 所有下游服务透传并记录到日志
  - 集成 OpenTelemetry，支持 Jaeger / Zipkin 可视化追踪

* **[Task 1.1.5]** WebSocket 连接管理：
  - 通过网关代理 WebSocket 升级请求
  - 配置心跳检测（30 秒 ping/pong）
  - 连接数限制：单用户最多 3 个并发 WebSocket 连接

### [Feature 1.2] 存储与缓存设施

* **[Task 1.2.1]** 部署 PostgreSQL 主从架构：
  - 主节点：处理所有写操作
  - 只读副本（≥1 个）：处理读操作（读写分离）
  - 规划数据库表分区策略（按时间分区：评论表、日志表、事件表）
  - 连接池：使用 `pgx` + `pgxpool`，最大连接数 = CPU 核心数 × 4
  - 慢查询日志：记录超过 100ms 的 SQL

* **[Task 1.2.2]** 部署 Redis 哨兵集群（或 Redis Cluster）：
  - Key 命名规范：`{domain}:{entity}:{id}:{field}`
    ```
    studio:user:session:123          # 用户会话
    studio:game:version:456:meta     # 游戏版本元数据缓存
    studio:ost:track:789:playcount   # 音轨播放计数
    studio:ratelimit:ip:1.2.3.4      # 限流计数器
    studio:lock:payment:order:abc    # 分布式锁
    ```
  - TTL 规范：会话 7 天、游戏元数据 1 小时、排行榜 5 分钟

* **[Task 1.2.3]** 封装云服务商的对象存储 (OSS) SDK：
  ```go
  type StorageService interface {
      Upload(ctx context.Context, key string, reader io.Reader, opts UploadOptions) error
      GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
      Delete(ctx context.Context, key string) error
      BatchDelete(ctx context.Context, keys []string) error
      GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error)
  }
  ```
  - 支持阿里云 OSS、AWS S3、腾讯云 COS 三种实现，通过配置切换
  - 上传前校验文件类型（Magic Number，而非仅后缀）
  - 文件大小限制：头像 5MB，游戏包 10GB，音频文件 500MB

* **[Task 1.2.4]** 消息队列基础设施：
  - 使用 Redis Stream 作为轻量消息队列
  - 消费者组模式，确保消息至少消费一次
  - 死信队列：消费失败 3 次的消息进入 DLQ，人工介入
  - 主要消息主题：
    - `studio:mq:email`（邮件发送任务）
    - `studio:mq:notify`（站内通知）
    - `studio:mq:download:log`（下载记录异步写入）
    - `studio:mq:analytics`（用户行为事件）

### [Feature 1.3] 遥测、监控与安全 (Observability & Security)

* **[Task 1.3.1]** 引入结构化日志 (Zap)：
  - 日志级别：DEBUG（仅开发）、INFO、WARN、ERROR、FATAL
  - 必须记录的字段：`timestamp`, `level`, `service`, `request_id`, `user_id`, `action`, `duration_ms`, `error`
  - 接入 ELK / Fluent-bit 收集全链路日志
  - 生产环境日志保留 30 天，归档到冷存储

* **[Task 1.3.2]** 实现 Prometheus 埋点：
  - HTTP 接口：QPS、P50/P95/P99 延迟、错误率（按状态码）
  - 游戏下载：下载峰值带宽、下载成功/失败率、防盗链拦截次数
  - 业务指标：DAU、新注册用户数、付款成功率、OST 播放量
  - Grafana Dashboard 可视化，告警规则（PagerDuty / 钉钉机器人通知）

* **[Task 1.3.3]** 网络安全防御：
  - **接口签名**：客户端请求携带 `X-Timestamp`（允许偏差 ±5 分钟）和 `X-Signature`（HMAC-SHA256 签名）
  - **SQL 注入**：全部使用 ORM 参数化查询，禁止字符串拼接 SQL
  - **XSS 防御**：服务端对所有用户输入进行 HTML 转义，前端使用 CSP
  - **CSRF 防御**：SameSite Cookie + CSRF Token 双重保护
  - **文件上传安全**：服务端二次校验文件类型，病毒扫描（集成 ClamAV）
  - **DDoS 防御**：接入云厂商 DDoS 高防 IP，Nginx 层限制单 IP 连接数

* **[Task 1.3.4]** 健康检查与熔断器：
  - 每个服务暴露 `/health`（存活探针）和 `/ready`（就绪探针）
  - 集成 Hystrix-go 或自实现熔断器，防止级联故障
  - 数据库连接池监控，超时自动断开重连

---

## [Epic 2] 核心通行证与用户中心 (Identity & Access Management) {#epic-2}

**目标**：构建一处注册、全站通用的用户系统（类似于单点登录 SSO）。

### [Feature 2.1] 认证与授权模块

* **[Task 2.1.1]** 实现 JWT 完整生命周期：
  - Access Token：有效期 15 分钟，Payload 含 `user_id`, `role`, `permissions`
  - Refresh Token：有效期 30 天，存入 Redis（`studio:user:refresh:{user_id}:{device_id}`）
  - 主动注销：将 Access Token JTI 加入 Redis 黑名单（TTL = Token 剩余有效期）
  - 设备管理：支持查看和踢出指定设备的登录态

* **[Task 2.1.2]** 实现 RBAC 权限控制：
  ```
  角色层次：
  SuperAdmin > Admin > Moderator > Creator > Premium Player > Player > Guest

  权限示例：
  - game:release:create   (仅 Admin+)
  - comment:delete:any    (仅 Moderator+)
  - comment:delete:own    (Player+)
  - ost:download:hifi     (Premium Player+)
  - dashboard:view        (Admin+)
  ```
  - 权限数据缓存在 Redis，TTL 5 分钟，角色变更时主动失效

* **[Task 2.1.3]** 第三方 OAuth2 登录：
  - GitHub、Google（国内用户优先 GitHub）
  - Steam（游戏玩家首选）
  - 微信扫码登录（国内用户）
  - QQ 登录（国内用户）
  - 账号绑定：一个账号可绑定多个第三方登录源

* **[Task 2.1.4]** 多因素认证 (MFA)：
  - 支持 TOTP（Google Authenticator、Authy）
  - 短信验证码（高敏感操作：修改密码、绑定支付方式）
  - 邮箱验证码（注册、找回密码）
  - 备用码：生成 8 个一次性备用码，安全存储（哈希后入库）

* **[Task 2.1.5]** 登录安全增强：
  - 登录失败 5 次，账号冻结 15 分钟
  - 异地登录检测：对比上次登录 IP 的 ASN / 国家，触发邮件告警
  - 可疑登录：要求 CAPTCHA 验证（集成 hCaptcha 或 Cloudflare Turnstile）

### [Feature 2.2] 用户画像与资产

* **[Task 2.2.1]** 用户基本信息表设计：
  ```sql
  CREATE TABLE users (
      id          BIGSERIAL PRIMARY KEY,
      uuid        UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
      username    VARCHAR(32) NOT NULL UNIQUE,
      email       VARCHAR(255) NOT NULL UNIQUE,
      password    VARCHAR(255),                  -- Argon2id 哈希
      avatar_key  VARCHAR(512),                  -- OSS 对象 Key
      bio         TEXT,
      website     VARCHAR(255),
      location    VARCHAR(100),
      role        VARCHAR(20) NOT NULL DEFAULT 'player',
      status      VARCHAR(20) NOT NULL DEFAULT 'active',
      email_verified_at TIMESTAMPTZ,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      last_login_at TIMESTAMPTZ,
      last_login_ip INET
  );
  ```

* **[Task 2.2.2]** 用户游戏资产库：
  ```sql
  CREATE TABLE user_game_assets (
      id          BIGSERIAL PRIMARY KEY,
      user_id     BIGINT NOT NULL REFERENCES users(id),
      game_id     BIGINT NOT NULL REFERENCES games(id),
      asset_type  VARCHAR(50) NOT NULL,  -- 'base_game', 'dlc', 'ost', 'artbook'
      asset_id    BIGINT NOT NULL,
      obtained_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      source      VARCHAR(50) NOT NULL,  -- 'purchase', 'gift', 'free_claim', 'redeem_code'
      UNIQUE(user_id, asset_type, asset_id)
  );
  ```

* **[Task 2.2.3]** 用户偏好设置：
  - 语言偏好（影响邮件语言和内容展示）
  - 通知偏好（邮件通知、站内通知的粒度开关）
  - 隐私设置（游戏库是否公开、评论是否实名）
  - 主题偏好（暗色/亮色模式偏好存云端，跨设备同步）

* **[Task 2.2.4]** 用户关注/粉丝系统：
  - 关注其他玩家（关注后可在动态流看到其游戏进展、评论）
  - 关注游戏（游戏更新时收到通知）
  - 粉丝数、关注数展示
  - 防刷：每日最多关注 200 个新账号

---

## [Epic 3] 游戏分发与版本控制中台 (CMS & Distribution) {#epic-3}

**目标**：专为 GVN-Engine 设计的版本迭代管理系统，支持多分支、DLC 与热更新。

### [Feature 3.1] 游戏元数据管理

* **[Task 3.1.1]** 核心数据库设计：
  ```sql
  -- 游戏基本信息
  CREATE TABLE games (
      id              BIGSERIAL PRIMARY KEY,
      slug            VARCHAR(100) NOT NULL UNIQUE,    -- URL 友好标识
      title           VARCHAR(255) NOT NULL,
      subtitle        VARCHAR(255),
      description     TEXT,
      cover_key       VARCHAR(512),                    -- 封面图 OSS Key
      banner_key      VARCHAR(512),                    -- 横幅图 OSS Key
      trailer_url     VARCHAR(512),                    -- 宣传片 URL（YouTube/B站）
      genre           VARCHAR(50)[],                   -- 游戏类型标签数组
      tags            VARCHAR(50)[],
      engine          VARCHAR(50) NOT NULL DEFAULT 'gvn',
      status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active/archived/coming_soon
      release_date    DATE,
      developer_id    BIGINT REFERENCES users(id),
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- 游戏截图
  CREATE TABLE game_screenshots (
      id          BIGSERIAL PRIMARY KEY,
      game_id     BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
      oss_key     VARCHAR(512) NOT NULL,
      caption     VARCHAR(255),
      sort_order  INT NOT NULL DEFAULT 0
  );

  -- 发行分支（main / beta / demo）
  CREATE TABLE game_branches (
      id          BIGSERIAL PRIMARY KEY,
      game_id     BIGINT NOT NULL REFERENCES games(id),
      name        VARCHAR(50) NOT NULL,       -- 'main', 'beta', 'demo'
      description TEXT,
      is_default  BOOLEAN NOT NULL DEFAULT FALSE,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      UNIQUE(game_id, name)
  );

  -- 版本实体
  CREATE TABLE game_releases (
      id              BIGSERIAL PRIMARY KEY,
      branch_id       BIGINT NOT NULL REFERENCES game_branches(id),
      version         VARCHAR(20) NOT NULL,    -- semver: v1.2.3
      title           VARCHAR(255),            -- 发布标题，如"雷 - 音乐教师 日常更新"
      changelog       TEXT,                    -- Markdown 格式更新日志
      oss_key         VARCHAR(512),            -- 游戏包 OSS Key
      manifest_key    VARCHAR(512),            -- manifest.json OSS Key（热更新用）
      file_size       BIGINT,                  -- 字节数
      checksum_sha256 VARCHAR(64),             -- 完整性校验
      min_os_version  VARCHAR(20),
      is_published    BOOLEAN NOT NULL DEFAULT FALSE,
      published_at    TIMESTAMPTZ,
      created_by      BIGINT REFERENCES users(id),
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      UNIQUE(branch_id, version)
  );
  ```

* **[Task 3.1.2]** DLC 与附加内容管理：
  ```sql
  CREATE TABLE dlc (
      id              BIGSERIAL PRIMARY KEY,
      game_id         BIGINT NOT NULL REFERENCES games(id),
      slug            VARCHAR(100) NOT NULL,
      title           VARCHAR(255) NOT NULL,
      description     TEXT,
      cover_key       VARCHAR(512),
      dlc_type        VARCHAR(50),    -- 'story', 'costume', 'artbook', 'ost', 'bundle'
      price_cents     INT,            -- 单位：分（0 表示免费）
      currency        CHAR(3) NOT NULL DEFAULT 'CNY',
      is_free         BOOLEAN NOT NULL DEFAULT FALSE,
      release_date    DATE,
      oss_key         VARCHAR(512),
      file_size       BIGINT,
      checksum_sha256 VARCHAR(64),
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```

* **[Task 3.1.3]** 版本更新日志管理：
  - 支持 Markdown 格式富文本（后端存 raw Markdown，前端渲染）
  - 更新类型标签：`新功能`、`剧情更新`、`BUG修复`、`性能优化`、`音乐更新`
  - 多语言更新日志（中文/英文分开维护）
  - 更新日志的 RSS/Atom Feed 订阅支持

### [Feature 3.2] 高并发防盗链下载分发

* **[Task 3.2.1]** 安全下载接口开发：
  ```
  GET /api/v1/games/{game_id}/releases/{release_id}/download

  处理流程：
  1. 验证用户认证（JWT）
  2. 检查用户是否拥有该游戏/版本权限
  3. 检查下载频率限制（Redis）
  4. 记录下载意图（异步写入下载日志）
  5. 调用 OSS SDK 生成 15 分钟有效期 Pre-signed URL
  6. 返回加密下载直链（用户浏览器直接从 CDN 下载）
  ```

* **[Task 3.2.2]** 防刷机制：
  - 单账号每日最多下载 3 次同一文件
  - 同一 IP 每小时最多触发 10 次下载请求
  - 检测到异常批量下载行为（如 Bot），自动封禁 IP 并告警
  - 下载日志记录：`user_id`, `release_id`, `client_ip`, `user_agent`, `downloaded_at`

* **[Task 3.2.3]** CDN 加速配置：
  - 静态资源（图片、音频试听片段）托管到 CDN 边缘节点
  - 游戏大包直接通过 CDN 回源 OSS 下发，不经过应用服务器
  - 多 CDN 冗余（国内 + 海外节点）

* **[Task 3.2.4]** 断点续传支持：
  - OSS Pre-signed URL 支持 HTTP Range 请求
  - 前端下载器集成断点续传逻辑
  - 下载进度持久化，浏览器刷新后可继续

### [Feature 3.3] 游戏客户端热更新支持

* **[Task 3.3.1]** Manifest 差异化补丁分发：
  ```json
  // manifest.json 结构示例
  {
    "version": "v1.2.3",
    "files": [
      {
        "path": "scripts/chapter1.ks",
        "size": 102400,
        "sha256": "abc123...",
        "url_key": "games/thunder/v1.2.3/scripts/chapter1.ks"
      }
    ],
    "total_size": 204800,
    "generated_at": "2026-03-03T12:00:00Z"
  }
  ```
  - GVN-Engine 客户端对比本地 manifest 与云端 manifest，计算差异文件列表
  - 仅下载变更的剧本文件或素材，而非全量包（节省带宽 ≥ 90%）

* **[Task 3.3.2]** 热更新版本管理 API：
  ```
  GET /api/v1/games/{game_id}/update-check?local_version=v1.2.0&branch=main

  响应：
  {
    "has_update": true,
    "latest_version": "v1.2.3",
    "patch_files": [...],  // 仅需下载的差异文件列表
    "full_download_url": "..."  // 若差异过大，建议全量下载
  }
  ```

* **[Task 3.3.3]** 强制更新机制：
  - 管理员可将某个版本标记为 `force_update`
  - 客户端检测到强制更新时，阻止进入游戏，弹出更新提示
  - 版本兼容性矩阵：记录各版本存档的前后兼容关系

---

## [Epic 4] 沉浸式流媒体与 OST 平台 (Media Streaming) {#epic-4}

**目标**：提供游戏原声带的在线试听、歌词同步与无损格式下载。

### [Feature 4.1] 音乐元数据与专辑管理

* **[Task 4.1.1]** 数据库设计：
  ```sql
  CREATE TABLE albums (
      id              BIGSERIAL PRIMARY KEY,
      game_id         BIGINT REFERENCES games(id),     -- 可为空（独立专辑）
      slug            VARCHAR(100) NOT NULL UNIQUE,
      title           VARCHAR(255) NOT NULL,
      subtitle        VARCHAR(255),
      description     TEXT,
      cover_key       VARCHAR(512),
      artist          VARCHAR(255),
      composer        VARCHAR(255),
      arranger        VARCHAR(255),
      lyricist        VARCHAR(255),
      total_tracks    INT,
      duration_sec    INT,
      release_date    DATE,
      album_type      VARCHAR(50) DEFAULT 'ost',  -- 'ost', 'drama', 'vocal', 'bgm'
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE tracks (
      id              BIGSERIAL PRIMARY KEY,
      album_id        BIGINT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
      track_number    INT NOT NULL,
      disc_number     INT NOT NULL DEFAULT 1,
      title           VARCHAR(255) NOT NULL,
      artist          VARCHAR(255),
      duration_sec    INT,
      -- 流媒体文件（试听用，MP3 128kbps）
      stream_key      VARCHAR(512),
      stream_size     BIGINT,
      -- 高保真文件（购买后下载，FLAC/WAV）
      hifi_key        VARCHAR(512),
      hifi_format     VARCHAR(10),              -- 'flac', 'wav', 'aiff'
      hifi_bitdepth   SMALLINT,                -- 16/24/32 bit
      hifi_samplerate INT,                     -- 44100/48000/96000 Hz
      hifi_size       BIGINT,
      -- 歌词
      lrc_key         VARCHAR(512),            -- LRC 格式歌词 OSS Key
      -- 统计
      play_count      BIGINT NOT NULL DEFAULT 0,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      UNIQUE(album_id, disc_number, track_number)
  );
  ```

* **[Task 4.1.2]** 音频上传流程：
  - 管理员上传 FLAC 源文件 → 后端异步转码为 MP3 128kbps（FFmpeg Worker）
  - 转码完成后，两份文件均存入 OSS
  - 音频波形数据生成（用于前端波形显示）

### [Feature 4.2] 音频流媒体分发引擎

* **[Task 4.2.1]** 试听接口开发：
  ```
  GET /api/v1/albums/{album_id}/tracks/{track_id}/stream

  响应：重定向到带签名的 CDN 流媒体 URL
  支持 HTTP Range 请求（断点续传 + 拖拽进度）

  签名有效期：30 分钟
  防盗链：签名 URL 绑定请求 User-Agent 前缀
  ```

* **[Task 4.2.2]** 动态播放列表生成：
  ```json
  GET /api/v1/albums/{album_id}/playlist
  {
    "album": {...},
    "tracks": [
      {
        "id": 1,
        "title": "雷的日常",
        "artist": "工作室原创",
        "duration": 185,
        "stream_url": "https://cdn.studio.com/tracks/1/stream?sig=xxx&expires=xxx",
        "waveform_url": "https://cdn.studio.com/tracks/1/waveform.json"
      }
    ],
    "expires_at": "2026-03-03T13:00:00Z"
  }
  ```

* **[Task 4.2.3]** 播放统计与热度计算：
  - 播放开始时：Redis INCR `studio:ost:track:{id}:playcount`
  - 每 5 分钟定时任务：将 Redis 计数批量刷入 PostgreSQL
  - 热度排行榜：Redis Sorted Set，Score = 播放量，定时刷新
  - 按游戏聚合：支持查询某游戏 OST 的总播放量

### [Feature 4.3] 交互式媒体特性

* **[Task 4.3.1]** LRC 歌词同步：
  - 后端提供 LRC 文件 URL，前端实现歌词滚动高亮
  - 支持逐字歌词（增强版 LRC / ASS 格式，用于 Karaoke 效果）
  - 歌词翻译：支持中日英三语对照显示

* **[Task 4.3.2]** 社交分享功能：
  - 生成专辑/曲目的分享卡片（OG Image，动态生成带封面和曲名的图片）
  - 支持分享到微博、微信、Twitter
  - 生成深度链接（打开官网并自动定位到对应曲目）

* **[Task 4.3.3]** 无损下载管理：
  - 用户购买 OST 后解锁 HIFI 下载权限
  - 下载接口生成 10 分钟有效的 Pre-signed URL
  - 打包下载：后台异步生成整张专辑的 ZIP 压缩包，完成后发送邮件通知

---

## [Epic 5] 高频互动与创作者社区 (Community Forums) {#epic-5}

**目标**：建立高粘性的玩家讨论区，处理高并发的读写请求。

### [Feature 5.1] 模块化评论树引擎

* **[Task 5.1.1]** 数据库模型（路径枚举 + 邻接表混合方案）：
  ```sql
  CREATE TABLE comments (
      id              BIGSERIAL PRIMARY KEY,
      -- 多态关联
      target_type     VARCHAR(50) NOT NULL,  -- 'game', 'release', 'track', 'post', 'dlc'
      target_id       BIGINT NOT NULL,
      -- 树形结构
      parent_id       BIGINT REFERENCES comments(id),
      root_id         BIGINT REFERENCES comments(id),   -- 根评论 ID，方便分组
      path            TEXT,                              -- 路径，如 '1/23/456'
      depth           INT NOT NULL DEFAULT 0,
      -- 内容
      user_id         BIGINT NOT NULL REFERENCES users(id),
      content         TEXT NOT NULL,
      content_type    VARCHAR(20) DEFAULT 'markdown',   -- 'plain', 'markdown'
      -- 状态
      status          VARCHAR(20) DEFAULT 'visible',    -- 'visible', 'hidden', 'deleted'
      is_pinned       BOOLEAN DEFAULT FALSE,
      -- 互动统计
      like_count      INT NOT NULL DEFAULT 0,
      dislike_count   INT NOT NULL DEFAULT 0,
      reply_count     INT NOT NULL DEFAULT 0,
      -- 时间
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at      TIMESTAMPTZ                       -- 软删除
  );

  CREATE INDEX idx_comments_target ON comments(target_type, target_id, created_at DESC);
  CREATE INDEX idx_comments_root ON comments(root_id, created_at ASC);
  CREATE INDEX idx_comments_user ON comments(user_id);
  ```

* **[Task 5.1.2]** 高性能评论读取策略：
  - 一级评论分页（`cursor-based` 游标分页，避免 OFFSET 性能问题）
  - 每条一级评论预加载前 3 条热门子评论（减少前端额外请求）
  - Redis 缓存首页热门评论（TTL 5 分钟，新评论发布时主动失效）
  - 评论数使用 Redis 计数器，避免 COUNT(*) 查询

* **[Task 5.1.3]** 评论审核流程：
  - 新用户（注册 < 7 天 或 信誉分 < 50）发帖进入审核队列
  - 关键词命中敏感词库，自动标记为"待审核"
  - 管理员审核界面：批量审核、一键通过/拒绝
  - 误判申诉：用户可对被拒绝的评论提交申诉

### [Feature 5.2] 社区帖子系统 (论坛)

* **[Task 5.2.1]** 帖子数据模型：
  ```sql
  CREATE TABLE posts (
      id              BIGSERIAL PRIMARY KEY,
      user_id         BIGINT NOT NULL REFERENCES users(id),
      category_id     INT NOT NULL,
      title           VARCHAR(255) NOT NULL,
      content         TEXT NOT NULL,          -- Markdown
      cover_key       VARCHAR(512),           -- 可选封面图
      tags            VARCHAR(50)[],
      status          VARCHAR(20) DEFAULT 'published',
      is_pinned       BOOLEAN DEFAULT FALSE,
      is_highlighted  BOOLEAN DEFAULT FALSE,  -- 编辑推荐
      view_count      BIGINT DEFAULT 0,
      like_count      INT DEFAULT 0,
      comment_count   INT DEFAULT 0,
      last_comment_at TIMESTAMPTZ,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE post_categories (
      id          SERIAL PRIMARY KEY,
      name        VARCHAR(100) NOT NULL,
      slug        VARCHAR(100) NOT NULL UNIQUE,
      description TEXT,
      icon        VARCHAR(50),              -- 图标名称
      sort_order  INT DEFAULT 0,
      color       VARCHAR(7)               -- Hex 颜色
  );
  ```
  预设分类：
  - `general`（综合讨论）
  - `bug-report`（BUG 反馈）
  - `feature-request`（功能建议）
  - `fan-art`（同人创作）
  - `guides`（攻略分享）
  - `off-topic`（灌水区）

* **[Task 5.2.2]** 富文本编辑器支持：
  - 前端集成 TipTap 或 ByteMD 编辑器
  - 支持图片上传（用户配图，限制 10MB，自动压缩到 1080p 以内）
  - 支持代码高亮、表格、引用、提及（@用户名）
  - 草稿自动保存（LocalStorage + 云端草稿，防止意外丢失）

### [Feature 5.3] 用户互动与通知反馈

* **[Task 5.3.1]** 点赞/踩系统：
  ```go
  // 点赞防重复：Redis SET
  key := fmt.Sprintf("studio:like:comment:%d:users", commentID)
  isNew := redis.SAdd(ctx, key, userID).Val() == 1
  if isNew {
      // 异步刷入数据库，更新 like_count
      mq.Publish("studio:mq:like", LikeEvent{...})
  }
  ```
  - 使用 Redis `SADD` 防止重复点赞（O(1) 复杂度）
  - 通过 Go Goroutine 定时（每 30 秒）批量将点赞数同步到 PostgreSQL
  - 用户的点赞记录表（用于"我的点赞"功能）

* **[Task 5.3.2]** 实时通知系统（WebSocket）：
  ```
  通知类型：
  - COMMENT_REPLY    有人回复了你的评论
  - POST_COMMENT     有人评论了你的帖子
  - USER_FOLLOW      有人关注了你
  - GAME_UPDATE      你关注的游戏发布了新版本
  - ACHIEVEMENT      你解锁了新成就
  - SYSTEM           系统公告
  ```
  - WebSocket Hub 管理在线用户连接
  - 离线通知积压：用户断线期间的通知存入 `notifications` 表
  - 重连后自动拉取未读通知
  - 通知红点计数：Redis `studio:user:{id}:unread_notify_count`

* **[Task 5.3.3]** 消息系统（站内信）：
  ```sql
  CREATE TABLE messages (
      id          BIGSERIAL PRIMARY KEY,
      from_user_id BIGINT REFERENCES users(id),  -- NULL 表示系统消息
      to_user_id  BIGINT NOT NULL REFERENCES users(id),
      subject     VARCHAR(255),
      content     TEXT NOT NULL,
      is_read     BOOLEAN DEFAULT FALSE,
      read_at     TIMESTAMPTZ,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```
  - 防止垃圾消息：仅互相关注的用户可以发送私信
  - 用户可屏蔽指定用户的消息

### [Feature 5.4] 社区内容治理 (Anti-Spam)

* **[Task 5.4.1]** 敏感词过滤架构：
  - DFA（确定性有穷自动机）算法实现，在内存中构建前缀树
  - 支持通配符词汇（如 `f*ck`）和变体识别（全角/半角、谐音字）
  - 敏感词库热更新（修改后无需重启服务）
  - 违规内容策略：警告（隐藏待审核）→ 删除 → 封禁账号

* **[Task 5.4.2]** 用户信誉分系统：
  ```
  信誉分变化规则：
  + 5   发帖被推荐
  + 2   评论被点赞（每个）
  + 10  连续登录 7 天
  - 10  评论被举报并审核通过
  - 50  账号被封禁一次

  信誉分区间权益：
  0-30    (新手) 每日发帖 3 条上限，评论需审核
  31-100  (普通) 每日发帖 10 条上限
  101+    (信任) 无上限，评论免审核
  ```

* **[Task 5.4.3]** 举报系统：
  ```sql
  CREATE TABLE reports (
      id              BIGSERIAL PRIMARY KEY,
      reporter_id     BIGINT NOT NULL REFERENCES users(id),
      target_type     VARCHAR(50) NOT NULL,   -- 'comment', 'post', 'user'
      target_id       BIGINT NOT NULL,
      reason          VARCHAR(50) NOT NULL,    -- 'spam', 'harassment', 'inappropriate', 'other'
      description     TEXT,
      status          VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'resolved', 'dismissed'
      handled_by      BIGINT REFERENCES users(id),
      handled_at      TIMESTAMPTZ,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```
  - 同一内容被举报 3 次，自动隐藏等待人工审核
  - 管理员处理界面：展示被举报内容上下文、举报者信誉分

---

## [Epic 6] 开发者与运维自动化流水线 (CI/CD Pipeline) {#epic-6}

**目标**：实现"不停上传新版本"的极致顺畅体验，全自动化构建部署。

### [Feature 6.1] 后端代码持续集成 (CI)

* **[Task 6.1.1]** GitHub Actions 工作流（`.github/workflows/ci.yml`）：
  ```yaml
  name: CI
  on: [push, pull_request]
  jobs:
    lint:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.23' }
        - name: golangci-lint
          uses: golangci/golangci-lint-action@v6
          with: { args: '--timeout=5m' }
    test:
      runs-on: ubuntu-latest
      services:
        postgres:
          image: postgres:16
          env: { POSTGRES_PASSWORD: test }
        redis:
          image: redis:7-alpine
      steps:
        - run: go test ./... -race -coverprofile=coverage.out
        - name: Upload coverage
          uses: codecov/codecov-action@v4
    build:
      needs: [lint, test]
      steps:
        - run: go build -ldflags="-w -s" -o server ./cmd/server
  ```

* **[Task 6.1.2]** Docker 镜像自动构建：
  ```dockerfile
  # 多阶段构建，最终镜像 < 20MB
  FROM golang:1.23-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o server ./cmd/server

  FROM gcr.io/distroless/static:nonroot
  COPY --from=builder /app/server /
  EXPOSE 8080
  ENTRYPOINT ["/server"]
  ```
  - 推送到私有镜像仓库（Harbor / 阿里云 ACR）
  - 镜像 Tag 策略：`{branch}-{commit-sha}-{date}`（如 `main-abc1234-20260303`）

### [Feature 6.2] 持续部署与同步 (CD)

* **[Task 6.2.1]** 生产环境滚动更新：
  - Kubernetes Deployment 滚动更新策略（`maxUnavailable: 0`, `maxSurge: 1`）
  - 健康检查通过后才切流量（Readiness Probe）
  - 异常自动回滚（监控到 Error Rate 上升时触发 Argo Rollouts 回滚）

* **[Task 6.2.2]** 游戏包发布 CLI 工具：
  ```bash
  # 工作室开发者使用
  $ studio-cli publish \
    --game thunder \
    --branch main \
    --version v1.2.3 \
    --title "雷 - 音乐教师 日常更新" \
    --changelog ./CHANGELOG.md \
    --package ./dist/thunder-v1.2.3.zip \
    --auto-publish

  # CLI 执行流程：
  # 1. 校验包完整性（SHA256）
  # 2. 分片上传到 OSS（支持断点续传）
  # 3. 生成并上传 manifest.json
  # 4. 调用 API 创建 Release 记录
  # 5. 若指定 --auto-publish，自动发布并通知订阅用户
  ```

* **[Task 6.2.3]** 数据库迁移自动化：
  - 使用 `golang-migrate` 管理 SQL 迁移文件
  - CD 流程中，服务启动前自动执行待执行的迁移
  - 迁移必须包含回滚脚本（`up.sql` + `down.sql`）
  - 禁止在生产迁移中执行破坏性操作（`DROP TABLE`、`DROP COLUMN` 需分多次迁移完成）

---

## [Epic 7] 电商与支付系统 (E-Commerce & Payments) {#epic-7}

**目标**：为游戏、DLC、OST 等数字商品提供安全、便捷的购买体验。

### [Feature 7.1] 商品目录与定价

* **[Task 7.1.1]** 商品数据模型：
  ```sql
  CREATE TABLE products (
      id              BIGSERIAL PRIMARY KEY,
      sku             VARCHAR(100) NOT NULL UNIQUE,
      name            VARCHAR(255) NOT NULL,
      description     TEXT,
      product_type    VARCHAR(50) NOT NULL,   -- 'game', 'dlc', 'ost', 'bundle', 'membership'
      entity_id       BIGINT,                 -- 关联的 game_id / dlc_id / album_id
      price_cents     INT NOT NULL,           -- 定价（分）
      currency        CHAR(3) NOT NULL DEFAULT 'CNY',
      original_price_cents INT,              -- 原价（用于划线价展示）
      is_active       BOOLEAN DEFAULT TRUE,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE discount_rules (
      id              BIGSERIAL PRIMARY KEY,
      product_id      BIGINT REFERENCES products(id),
      discount_type   VARCHAR(20),  -- 'percentage', 'fixed', 'buy_x_get_y'
      discount_value  NUMERIC(5,2),
      start_at        TIMESTAMPTZ NOT NULL,
      end_at          TIMESTAMPTZ NOT NULL,
      max_uses        INT,
      used_count      INT DEFAULT 0
  );
  ```

* **[Task 7.1.2]** 捆绑包 (Bundle) 逻辑：
  - 定义包含多个商品的捆绑包
  - 捆绑包价格低于单独购买总价
  - 用户已购部分商品时，显示"升级价格"（仅补差价）

### [Feature 7.2] 购物车与订单系统

* **[Task 7.2.1]** 购物车（Redis 实现）：
  ```
  Key: studio:cart:{user_id}
  Type: Hash
  Field: {product_id}
  Value: {quantity, added_at}
  TTL: 30 天
  ```
  - 登录时自动合并游客购物车（LocalStorage → Redis）
  - 实时价格计算（折扣、优惠码应用）

* **[Task 7.2.2]** 订单核心流程：
  ```
  下单状态机：
  PENDING_PAYMENT → PAID → FULFILLED → REFUNDED
                  → CANCELLED（超时未支付）
                  → FAILED（支付失败）

  幂等性：使用 idempotency_key 防止重复下单
  库存：数字商品无限量，不需要库存管理
  ```

  ```sql
  CREATE TABLE orders (
      id              BIGSERIAL PRIMARY KEY,
      order_no        VARCHAR(32) NOT NULL UNIQUE,  -- 业务订单号
      user_id         BIGINT NOT NULL REFERENCES users(id),
      status          VARCHAR(30) NOT NULL DEFAULT 'pending_payment',
      total_cents     INT NOT NULL,
      currency        CHAR(3) NOT NULL DEFAULT 'CNY',
      discount_cents  INT DEFAULT 0,
      coupon_code     VARCHAR(50),
      payment_method  VARCHAR(50),    -- 'alipay', 'wechat', 'stripe', 'paypal'
      paid_at         TIMESTAMPTZ,
      idempotency_key VARCHAR(64) UNIQUE,
      metadata        JSONB,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      expires_at      TIMESTAMPTZ,    -- 未支付订单过期时间（30 分钟）
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE order_items (
      id          BIGSERIAL PRIMARY KEY,
      order_id    BIGINT NOT NULL REFERENCES orders(id),
      product_id  BIGINT NOT NULL REFERENCES products(id),
      price_cents INT NOT NULL,
      quantity    INT NOT NULL DEFAULT 1
  );
  ```

### [Feature 7.3] 支付网关集成

* **[Task 7.3.1]** 国内支付接入：
  - **支付宝**：PC 端扫码支付、手机端 H5 支付、APP 支付
  - **微信支付**：Native 扫码、JSAPI、H5 支付
  - 回调验签：对支付网关的异步回调进行签名验证，防伪造

* **[Task 7.3.2]** 海外支付接入：
  - **Stripe**：信用卡、Apple Pay、Google Pay
  - **PayPal**：支持全球主要货币
  - 货币自动换算（汇率每日更新）

* **[Task 7.3.3]** 支付安全：
  - 所有支付回调必须在服务端验签（不信任客户端传递的支付结果）
  - 订单金额二次校验（数据库价格 vs 支付网关回调金额）
  - 敏感操作日志审计：支付、退款操作必须留完整日志

### [Feature 7.4] 兑换码与优惠券系统

* **[Task 7.4.1]** 兑换码生成：
  ```go
  // 生成安全的兑换码（Base32 编码，去除易混淆字符）
  // 格式：XXXX-XXXX-XXXX（共 12 位）
  func GenerateRedeemCode() string { ... }
  ```
  - 批量生成（用于媒体宣传、活动赠送）
  - 兑换码与商品的关联关系（单个商品、捆绑包、会员时长）
  - 兑换码使用记录

* **[Task 7.4.2]** 优惠券系统：
  - 优惠类型：折扣码（8折）、满减（满50减10）、定额优惠（减5元）
  - 适用范围：全场、指定游戏、指定分类
  - 使用限制：每人限用1次、新用户专属、有效期

---

## [Epic 8] 前端全站架构 (Frontend Architecture) {#epic-8}

**目标**：构建高性能、SEO 友好、体验一流的全站前端。

### [Feature 8.1] 技术选型与项目结构

* **[Task 8.1.1]** 核心技术栈：
  ```
  框架：     Next.js 15 (App Router)
  语言：     TypeScript 5.x
  样式：     Tailwind CSS v4 + CSS Modules（局部复杂样式）
  状态管理：  Zustand（客户端全局状态）+ React Query（服务端状态）
  动画：     Framer Motion（页面过渡）+ GSAP（复杂动画）
  UI组件库：  自研组件库（基于 Radix UI 无障碍原语）
  图标：     Lucide React
  表单：     React Hook Form + Zod
  ```

* **[Task 8.1.2]** Monorepo 项目结构：
  ```
  apps/
  ├── web/          # 主站 (Next.js)
  ├── admin/        # 管理后台 (Next.js)
  └── docs/         # 开发者文档 (Nextra)
  packages/
  ├── ui/           # 共享 UI 组件库
  ├── api-client/   # 自动生成的 API 客户端 (openapi-typescript-codegen)
  ├── config/       # 共享配置 (ESLint, Tailwind, tsconfig)
  └── utils/        # 工具函数
  ```

### [Feature 8.2] 核心页面设计规范

* **[Task 8.2.1]** 首页 (Landing Page)：
  ```
  布局：
  ┌─────────────────────────────────────┐
  │  导航栏（固定顶部，透明→模糊渐变）       │
  ├─────────────────────────────────────┤
  │  Hero 区（全屏视差背景，游戏 Trailer）   │
  ├─────────────────────────────────────┤
  │  最新游戏展示（卡片轮播）               │
  ├─────────────────────────────────────┤
  │  OST 精选播放（迷你播放器）             │
  ├─────────────────────────────────────┤
  │  社区动态（最新帖子、评论）             │
  ├─────────────────────────────────────┤
  │  工作室介绍 + 团队成员                  │
  ├─────────────────────────────────────┤
  │  邮件订阅 Banner                      │
  ├─────────────────────────────────────┤
  │  页脚（友情链接、社交媒体）              │
  └─────────────────────────────────────┘
  ```

* **[Task 8.2.2]** 游戏详情页：
  - 游戏封面大图 + 截图画廊（Lightbox 查看器）
  - 游戏介绍（支持 Markdown 渲染，图文混排）
  - 版本历史时间轴（可展开查看每个版本的更新日志）
  - DLC 列表（购买状态、价格、简介）
  - 用户评分（五星评分 + 文字评价）
  - 评论区（嵌套评论树）
  - 相关 OST 专辑入口
  - 下载/购买按钮（根据用户购买状态显示不同文案）

* **[Task 8.2.3]** OST 播放器设计（全局悬浮播放器）：
  ```
  桌面端：底部固定播放条
  - 封面缩略图 | 曲名/专辑名 | 播放控制 | 进度条 | 音量 | 播放列表

  移动端：底部 Mini 播放器 + 全屏展开模式
  - 全屏展开：封面大图、歌词滚动、进度条、播放控制
  ```
  - 支持键盘快捷键（空格暂停、左右键快进/快退）
  - 媒体会话 API（锁屏界面控制）
  - 播放队列管理（自定义播放顺序）

### [Feature 8.3] SEO 与性能优化

* **[Task 8.3.1]** SEO 策略：
  - Next.js App Router SSR / SSG / ISR 混合渲染策略
  - 动态 meta 标签（title, description, OG Image）
  - 结构化数据 (Schema.org)：Game、MusicAlbum、Article
  - Sitemap 自动生成（每日更新）
  - robots.txt 配置

* **[Task 8.3.2]** Core Web Vitals 优化目标：
  - LCP (Largest Contentful Paint) < 2.5 秒
  - FID (First Input Delay) < 100ms
  - CLS (Cumulative Layout Shift) < 0.1
  - 图片优化：Next.js Image 组件，WebP/AVIF 自动转换
  - 字体优化：`next/font` 子集化加载，避免 FOUT
  - 代码分割：动态 import，路由级代码块

* **[Task 8.3.3]** 国际化 (i18n)：
  - `next-intl` 实现多语言路由（`/zh/games`, `/en/games`）
  - 翻译文件：JSON 格式，按页面分割
  - 语言自动检测（Accept-Language Header）+ 手动切换

---

## [Epic 9] 管理后台系统 (Admin Dashboard) {#epic-9}

**目标**：为工作室内部团队提供高效的内容管理和数据运营工具。

### [Feature 9.1] 内容管理模块

* **[Task 9.1.1]** 游戏/版本管理：
  - 游戏信息 CRUD（所见即所得预览）
  - 版本发布流程（草稿 → 预发布 → 正式发布）
  - 分支管理界面（main / beta / demo 分支切换）
  - 游戏包上传（拖拽上传，分片断点续传，实时进度条）
  - 一键回滚到历史版本

* **[Task 9.1.2]** OST 管理：
  - 专辑和音轨的增删改查
  - 批量上传音频文件（自动解析 ID3 元数据）
  - 音频转码进度监控
  - 歌词文件上传与关联

* **[Task 9.1.3]** 社区内容审核：
  - 评论/帖子审核队列（待审核 / 已通过 / 已拒绝 筛选）
  - 批量操作（批量通过、批量删除）
  - 举报处理界面（展示被举报内容与举报原因）
  - 用户管理（搜索、封禁、解封、角色变更）

### [Feature 9.2] 数据运营看板

* **[Task 9.2.1]** 实时运营面板（每分钟刷新）：
  ```
  ┌──────────┬──────────┬──────────┬──────────┐
  │ 在线用户  │  今日新增  │  今日下载  │  今日收入  │
  │  1,234   │   567    │   890    │  ¥12,345 │
  └──────────┴──────────┴──────────┴──────────┘
  ┌─────────────────────┬───────────────────────┐
  │  下载量趋势（折线图）  │  收入来源分布（饼图）  │
  └─────────────────────┴───────────────────────┘
  ```

* **[Task 9.2.2]** 用户行为分析：
  - 新用户注册趋势（日/周/月）
  - 用户留存率分析（Day 1 / Day 7 / Day 30）
  - 最受欢迎的游戏和 OST 曲目
  - 用户地理分布热力图

* **[Task 9.2.3]** 销售与财务报表：
  - 每日/月收入报表（支持导出 Excel/CSV）
  - 商品销售排行（数量 + 金额）
  - 退款率监控
  - 优惠券使用效果分析

---

## [Epic 10] 成就与游戏化系统 (Gamification) {#epic-10}

**目标**：通过成就、勋章、积分等游戏化机制提升用户粘性。

### [Feature 10.1] 成就系统设计

* **[Task 10.1.1]** 成就数据模型：
  ```sql
  CREATE TABLE achievements (
      id          SERIAL PRIMARY KEY,
      slug        VARCHAR(100) NOT NULL UNIQUE,
      name        VARCHAR(100) NOT NULL,
      description TEXT,
      icon_key    VARCHAR(512),
      rarity      VARCHAR(20),    -- 'common', 'rare', 'epic', 'legendary'
      points      INT DEFAULT 0,
      trigger     VARCHAR(50),    -- 'manual', 'auto'
      condition_type VARCHAR(50), -- 'game_download', 'comment_count', 'login_streak'
      condition_value JSONB,      -- 条件参数，如 {"count": 10}
      is_secret   BOOLEAN DEFAULT FALSE,  -- 隐藏成就
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE user_achievements (
      id              BIGSERIAL PRIMARY KEY,
      user_id         BIGINT NOT NULL REFERENCES users(id),
      achievement_id  INT NOT NULL REFERENCES achievements(id),
      obtained_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      UNIQUE(user_id, achievement_id)
  );
  ```

* **[Task 10.1.2]** 预设成就列表：
  ```
  游戏相关：
  - "初次探索"：首次下载任意游戏
  - "全家桶"：购买工作室全部游戏
  - "早鸟"：在游戏正式发布前下载 Demo

  社区相关：
  - "话痨"：发布 100 条评论
  - "社区明星"：评论累计获得 1000 个赞
  - "创作者"：发布 10 篇攻略帖

  忠实粉丝：
  - "常客"：连续登录 30 天
  - "老玩家"：注册满 1 周年
  - "首批支持者"：注册于工作室成立后 30 天内

  音乐相关：
  - "音乐鉴赏家"：累计试听 OST 超过 100 首
  - "无损主义"：购买并下载无损音频
  ```

* **[Task 10.1.3]** 成就触发引擎：
  - 事件驱动：用户行为发布到 Redis Stream，Consumer 监听并计算成就条件
  - 实时解锁：条件满足时立即发放，触发 WebSocket 通知弹窗
  - 成就进度追踪（如"100 条评论：当前 47/100"）

### [Feature 10.2] 积分与排行榜

* **[Task 10.2.1]** 积分系统：
  ```
  积分来源：
  + 10  首次注册
  + 5   每日签到（连续 7 天额外 +20）
  + 1   发布评论
  + 3   评论获赞
  + 50  购买游戏
  + 20  解锁成就

  积分用途：
  - 兑换工作室周边（实体商品）
  - 兑换游戏优惠券
  - 账号等级展示
  ```

* **[Task 10.2.2]** 排行榜（Redis Sorted Set）：
  - 积分周榜/月榜/总榜
  - OST 播放量排行榜
  - 社区活跃度排行榜（按评论数、获赞数）
  - 排行榜数据每 5 分钟更新一次

---

## [Epic 11] 数据分析与商业智能 (Analytics & BI) {#epic-11}

**目标**：构建数据驱动的运营体系，支持精细化运营决策。

### [Feature 11.1] 用户行为事件追踪

* **[Task 11.1.1]** 事件采集 SDK（前端）：
  ```typescript
  // 前端埋点示例
  analytics.track('game_page_view', {
    game_id: 123,
    game_title: '雷 - 音乐教师',
    referrer: 'homepage_banner',
    session_id: 'xxx'
  });

  analytics.track('download_initiated', {
    game_id: 123,
    release_id: 456,
    version: 'v1.2.3'
  });
  ```
  - 事件通过 Beacon API 异步上报，不阻塞页面交互
  - 批量上报：累积 10 条事件或每 10 秒批量发送
  - 离线缓存：网络断开时存 LocalStorage，恢复后补报

* **[Task 11.1.2]** 事件处理管道：
  ```
  前端 → API Gateway → Redis Stream → Consumer → PostgreSQL (事件表)
                                               → ClickHouse（分析数据仓库，可选）
  ```

* **[Task 11.1.3]** 核心分析指标定义：
  - **DAU/MAU**：日活/月活用户数
  - **转化漏斗**：首页浏览 → 游戏详情 → 注册 → 购买（各步骤转化率）
  - **LTV（用户生命周期价值）**：平均每用户累计付费金额
  - **NPS（净推荐值）**：定期问卷调查用户满意度

### [Feature 11.2] A/B 测试框架

* **[Task 11.2.1]** 实验框架设计：
  ```go
  type Experiment struct {
      ID          string
      Name        string
      Variants    []Variant   // e.g., control, treatment_a, treatment_b
      TrafficPct  float64     // 参与实验的流量比例
      StartAt     time.Time
      EndAt       time.Time
  }

  // 用户分配（基于 user_id 的一致性哈希，同一用户始终进入同一分组）
  func AssignVariant(userID int64, experiment Experiment) string { ... }
  ```
  - 实验结果分析：统计显著性检验（P 值 < 0.05）
  - 典型实验：定价策略测试、首页布局测试、邮件主题行测试

---

## [Epic 12] 本地化与国际化 (L10n & i18n) {#epic-12}

**目标**：支持全球玩家访问，提供本地化的内容和体验。

### [Feature 12.1] 多语言内容管理

* **[Task 12.1.1]** 支持的语言：
  | 语言 | 代码 | 优先级 |
  |------|------|--------|
  | 简体中文 | zh-CN | P0（主要） |
  | 繁体中文 | zh-TW | P1 |
  | 英语 | en | P1 |
  | 日语 | ja | P2 |
  | 韩语 | ko | P2 |

* **[Task 12.1.2]** 内容本地化策略：
  - 游戏标题、描述、更新日志支持多语言字段
  - 数据库使用 JSONB 存储多语言内容：
    ```json
    {
      "zh-CN": "雷 - 音乐教师",
      "en": "Thunder - Music Teacher",
      "ja": "レイ - 音楽の先生"
    }
    ```
  - 查询时按用户语言偏好返回对应语言内容，降级到默认语言（zh-CN）

* **[Task 12.1.3]** 货币与时区本地化：
  - 自动检测用户地区，显示对应货币价格（CNY / USD / JPY / KRW）
  - 汇率每日从汇率 API 更新
  - 时区：所有时间存储 UTC，前端按用户时区显示

---

## [Epic 13] 邮件营销与推送通知 (Email & Push) {#epic-13}

**目标**：通过精准的邮件和推送触达用户，提升活跃度与转化率。

### [Feature 13.1] 事务型邮件

* **[Task 13.1.1]** 邮件模板系统：
  - 使用 Go HTML 模板引擎渲染邮件
  - 邮件类型：
    - 欢迎邮件（注册成功）
    - 邮箱验证
    - 密码重置
    - 支付成功通知
    - 游戏新版本发布通知
    - 成就解锁通知
    - 安全告警（异地登录）
  - 所有邮件模板支持中英文双语

* **[Task 13.1.2]** 邮件发送基础设施：
  - SMTP 集成（阿里云邮件推送 / SendGrid / Resend）
  - 异步发送（Redis Stream 消息队列）
  - 发送失败自动重试（指数退避，最多 3 次）
  - SPF / DKIM / DMARC 配置，确保邮件进入收件箱而非垃圾箱

### [Feature 13.2] 营销邮件与订阅

* **[Task 13.2.1]** 邮件订阅系统：
  - 首页邮件订阅入口
  - 确认订阅邮件（Double Opt-in）
  - 订阅分类：新游戏发布、重大更新、活动促销
  - 一键退订链接（GDPR 合规）

* **[Task 13.2.2]** Web Push 通知（PWA）：
  - 基于 Web Push Protocol（使用 VAPID 密钥对）
  - 推送触发场景：游戏更新、成就解锁、社区回复
  - 推送统计：推送量、点击率、关闭率

---

## [Epic 14] 搜索引擎与推荐系统 (Search & Recommendation) {#epic-14}

**目标**：帮助用户快速找到感兴趣的游戏、音乐和社区内容。

### [Feature 14.1] 全站搜索

* **[Task 14.1.1]** 搜索后端实现（PostgreSQL 全文搜索）：
  ```sql
  -- 为游戏表添加全文搜索索引
  ALTER TABLE games ADD COLUMN search_vector TSVECTOR;
  CREATE INDEX idx_games_search ON games USING GIN(search_vector);

  -- 触发器自动更新 search_vector
  CREATE TRIGGER update_games_search
  BEFORE INSERT OR UPDATE ON games
  FOR EACH ROW EXECUTE FUNCTION
  tsvector_update_trigger(search_vector, 'pg_catalog.chinese', title, description);
  ```
  - 支持模糊匹配（trigram 扩展）
  - 搜索结果按相关度 + 热度加权排序

* **[Task 14.1.2]** 搜索范围：
  - 游戏（标题、标签、描述）
  - OST 专辑/曲目（曲名、艺术家）
  - 社区帖子（标题）
  - 用户（用户名）
  - 统一搜索结果页，按类型分 Tab 展示

* **[Task 14.1.3]** 搜索增强功能：
  - 实时搜索建议（输入时即时展示候选词）
  - 搜索历史记录（最近 10 次搜索，存 LocalStorage）
  - 热门搜索词展示（Redis Sorted Set 统计搜索频率）
  - 搜索无结果时推荐相关内容

### [Feature 14.2] 内容推荐

* **[Task 14.2.1]** 基于规则的推荐（初期）：
  - "你可能也喜欢"：同类型游戏、同专辑的其他曲目
  - 基于购买记录推荐相关 DLC
  - "热门本周"：按下载量、播放量排序
  - 新用户冷启动：展示工作室编辑推荐榜单

* **[Task 14.2.2]** 协同过滤推荐（进阶）：
  - 用户行为数据：游戏浏览、下载、OST 播放记录
  - Item-based 协同过滤（相似游戏推荐）
  - 定期（每日）离线计算，结果存入 Redis

---

## [Epic 15] 客户支持与工单系统 (Support & Tickets) {#epic-15}

**目标**：为用户提供高效的技术支持和问题解决渠道。

### [Feature 15.1] 帮助中心

* **[Task 15.1.1]** FAQ 知识库：
  - 文章分类：账号相关、游戏下载、购买支付、技术问题
  - 搜索功能：全文搜索 FAQ 文章
  - 文章评价：有帮助/没帮助（收集用户反馈）
  - 热门文章：展示解决率最高的 FAQ

* **[Task 15.1.2]** 智能机器人（规则引擎）：
  - 接入前置问题引导（选择问题类型）
  - 自动匹配 FAQ 文章
  - 无法解决时，引导创建人工工单

### [Feature 15.2] 工单系统

* **[Task 15.2.1]** 工单数据模型：
  ```sql
  CREATE TABLE tickets (
      id          BIGSERIAL PRIMARY KEY,
      ticket_no   VARCHAR(20) NOT NULL UNIQUE,
      user_id     BIGINT NOT NULL REFERENCES users(id),
      category    VARCHAR(50) NOT NULL,
      subject     VARCHAR(255) NOT NULL,
      status      VARCHAR(20) DEFAULT 'open',  -- 'open', 'in_progress', 'resolved', 'closed'
      priority    VARCHAR(20) DEFAULT 'normal',  -- 'low', 'normal', 'high', 'urgent'
      assigned_to BIGINT REFERENCES users(id),
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      resolved_at TIMESTAMPTZ,
      satisfaction INT  -- 1-5 满意度评分
  );

  CREATE TABLE ticket_messages (
      id          BIGSERIAL PRIMARY KEY,
      ticket_id   BIGINT NOT NULL REFERENCES tickets(id),
      sender_id   BIGINT NOT NULL REFERENCES users(id),
      content     TEXT NOT NULL,
      attachments JSONB,  -- 附件列表（图片 OSS Key）
      is_staff    BOOLEAN DEFAULT FALSE,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```

* **[Task 15.2.2]** 工单工作流：
  - 用户创建工单 → 自动分配优先级 → 发送确认邮件
  - 客服回复 → 邮件/站内信通知用户
  - 72 小时无响应 → 自动升级优先级
  - 解决后 → 发送满意度调查

---

## [Epic 16] 开发者 SDK 与开放 API (Developer Platform) {#epic-16}

**目标**：对外开放部分 API 能力，支持第三方开发者集成。

### [Feature 16.1] 开放 API 规范

* **[Task 16.1.1]** RESTful API 设计原则：
  - 版本控制：`/api/v1/`, `/api/v2/`
  - 资源命名：复数名词（`/games`, `/albums`）
  - HTTP 方法语义：GET 读取、POST 创建、PUT 全量更新、PATCH 部分更新、DELETE 删除
  - 分页：统一使用 cursor-based 分页
  - 过滤：`?filter[status]=published&sort=-created_at&limit=20`

* **[Task 16.1.2]** API 文档：
  - OpenAPI 3.0 规范（`openapi.yaml` 自动从代码注释生成）
  - Swagger UI / Redoc 在线文档
  - API 变更日志（Breaking Changes 需提前 3 个月通知）
  - 代码示例（Go、Python、JavaScript、curl）

* **[Task 16.1.3]** API Key 管理：
  ```sql
  CREATE TABLE api_keys (
      id          BIGSERIAL PRIMARY KEY,
      user_id     BIGINT NOT NULL REFERENCES users(id),
      name        VARCHAR(100) NOT NULL,
      key_hash    VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 哈希存储，不存明文
      prefix      VARCHAR(10) NOT NULL,         -- 前缀（展示用），如 'sk_live_'
      scopes      VARCHAR(50)[],               -- 权限范围
      last_used_at TIMESTAMPTZ,
      expires_at  TIMESTAMPTZ,
      is_active   BOOLEAN DEFAULT TRUE,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```

### [Feature 16.2] GVN-Engine 专用 SDK

* **[Task 16.2.1]** Go SDK（后端调用）：
  ```go
  client := studio.NewClient(apiKey, studio.WithBaseURL("https://api.studio.com"))

  // 检查游戏更新
  update, err := client.Games.CheckUpdate(ctx, &studio.UpdateCheckRequest{
      GameID:       123,
      LocalVersion: "v1.2.0",
      Branch:       "main",
  })

  // 获取下载链接
  url, err := client.Games.GetDownloadURL(ctx, &studio.DownloadRequest{
      ReleaseID: 456,
  })
  ```

* **[Task 16.2.2]** JavaScript SDK（前端调用，通过 npm 分发）：
  ```bash
  npm install @studio/sdk
  ```
  ```javascript
  import { StudioClient } from '@studio/sdk';

  const client = new StudioClient({ apiKey: 'your-key' });
  const games = await client.games.list({ status: 'published' });
  ```

---

## [Epic 17] 移动端适配与 PWA {#epic-17}

**目标**：提供接近原生 App 的移动端体验，无需用户安装 App。

### [Feature 17.1] 响应式设计规范

* **[Task 17.1.1]** 断点规范：
  ```
  sm:  640px   手机横屏
  md:  768px   平板竖屏
  lg:  1024px  平板横屏 / 小桌面
  xl:  1280px  标准桌面
  2xl: 1536px  宽屏桌面
  ```

* **[Task 17.1.2]** 移动端特殊优化：
  - 触摸目标区域 ≥ 44×44px（WCAG 标准）
  - 滑动手势支持（左滑关闭抽屉、上滑展开播放器）
  - 防误触：重要操作（删除、支付）需二次确认
  - 底部导航栏（移动端主导航）

### [Feature 17.2] PWA 配置

* **[Task 17.2.1]** Service Worker 策略：
  - **Cache-First**：静态资源（图片、字体、JS/CSS）
  - **Network-First**：API 请求（数据实时性优先）
  - **Stale-While-Revalidate**：次要页面（社区帖子列表）
  - 离线页面：网络断开时展示缓存内容 + 友好提示

* **[Task 17.2.2]** 安装体验：
  - Web App Manifest 配置（图标、主题色、启动画面）
  - 安装提示时机优化（用户访问 3 次后提示）
  - 添加到桌面后支持全屏启动（无浏览器 UI）

---

## [Epic 18] 数据安全、合规与隐私 (Security & Compliance) {#epic-18}

**目标**：保护用户数据安全，满足各地区法规要求。

### [Feature 18.1] 数据隐私合规

* **[Task 18.1.1]** GDPR 合规（欧洲用户）：
  - 隐私政策与服务条款（每次重大变更需用户重新确认）
  - Cookie 同意横幅（非必要 Cookie 需用户主动授权）
  - 数据访问权：用户可申请导出个人数据（7 天内响应）
  - 数据删除权：用户注销账号后，30 天内删除或匿名化个人数据
  - 数据处理记录（Data Processing Records）

* **[Task 18.1.2]** 中国《个人信息保护法》合规：
  - 收集必要性原则（只收集业务所需的最少个人信息）
  - 用户知情同意（隐私政策弹窗，注册时明确说明）
  - 数据跨境传输（用户数据不出境，国内节点存储）
  - 未成年人保护（年龄核实，14 岁以下需家长同意）

### [Feature 18.2] 安全加固

* **[Task 18.2.1]** 密码安全：
  - 密码哈希：Argon2id（time=3, memory=64MB, threads=4）
  - 密码强度要求：≥8 字符，含大小写+数字
  - Pwned Passwords API 检测是否为已泄露密码

* **[Task 18.2.2]** 安全审计：
  - 所有管理员操作记录到 `audit_logs` 表（谁、何时、做了什么、操作前后数据）
  - 定期（每季度）渗透测试
  - 漏洞披露政策（安全研究员可提交漏洞）

* **[Task 18.2.3]** 依赖安全：
  - 定期运行 `govulncheck` 检查 Go 依赖漏洞
  - 前端：`npm audit` + Dependabot 自动 PR
  - Docker 基础镜像定期更新，使用 Trivy 扫描镜像漏洞

---

## [Epic 19] 容灾、备份与高可用 (Disaster Recovery) {#epic-19}

**目标**：确保在硬件故障、机房灾难等极端情况下，服务能快速恢复。

### [Feature 19.1] 数据备份策略

* **[Task 19.1.1]** PostgreSQL 备份：
  - 连续归档（WAL-G + OSS）：RPO = 0（理论上零数据丢失）
  - 每日全量备份快照，保留 30 天
  - 每周备份迁移到冷存储（成本更低），保留 1 年
  - 备份恢复演练：每月执行一次完整的恢复演练

* **[Task 19.1.2]** Redis 备份：
  - AOF 持久化（每秒 fsync）
  - 每小时 RDB 快照，存入 OSS
  - 故障切换：哨兵自动选主，RTO < 30 秒

* **[Task 19.1.3]** OSS 数据冗余：
  - 跨地域复制（主站 + 容灾站异地复制）
  - 版本控制开启（防止误删除）
  - 生命周期策略：临时文件 7 天自动清理

### [Feature 19.2] 多区域高可用

* **[Task 19.2.1]** 多可用区部署：
  - 应用服务：跨 3 个可用区部署（负载均衡器分发）
  - 数据库：主从跨可用区（主库故障时自动切换到从库）
  - 缓存：Redis Sentinel 跨可用区

* **[Task 19.2.2]** 故障切换演练（混沌工程）：
  - 定期（每季度）随机关闭一台服务器，验证自动故障切换
  - 模拟数据库主库故障，测量故障切换时间
  - 演练结果归档，优化应急预案

---

## [Epic 20] 性能工程与压力测试 (Performance Engineering) {#epic-20}

**目标**：在新版本发布、活动高峰等极端场景下确保系统稳定。

### [Feature 20.1] 性能基准测试

* **[Task 20.1.1]** API 性能基准（`wrk` / `k6` 工具）：
  ```bash
  # 压测示例
  k6 run --vus 1000 --duration 30s scripts/load_test.js

  # 目标指标
  - /api/v1/games: P99 < 50ms，错误率 < 0.1%
  - /api/v1/users/login: P99 < 200ms，错误率 < 0.01%
  - WebSocket 连接: 支持 5000 并发连接
  ```

* **[Task 20.1.2]** 数据库查询优化：
  - 慢查询分析（EXPLAIN ANALYZE）
  - 索引覆盖率检查（确保热路径查询都走索引）
  - 避免 N+1 查询（使用 JOIN 或批量查询）
  - 定期执行 VACUUM ANALYZE 防止表膨胀

* **[Task 20.1.3]** 缓存命中率优化：
  - 目标：核心接口缓存命中率 > 90%
  - 热点数据识别与预热（应用启动时预加载）
  - 缓存穿透防护（布隆过滤器）
  - 缓存雪崩防护（TTL 加随机抖动）
  - 缓存击穿防护（singleflight 合并并发请求）

### [Feature 20.2] 弹性扩容

* **[Task 20.2.1]** Kubernetes HPA（水平自动扩缩容）：
  ```yaml
  apiVersion: autoscaling/v2
  kind: HorizontalPodAutoscaler
  spec:
    minReplicas: 2
    maxReplicas: 20
    metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 70
  ```
  - CPU 使用率超过 70% 自动扩容，低于 30% 缩容
  - 提前扩容：根据历史流量预测，在游戏发布前 1 小时预扩容

---

## 数据库 ER 总览 {#db-overview}

```
users ──────────────────── user_achievements ── achievements
  │
  ├── user_game_assets ─── games ──── game_branches ── game_releases
  │         │                │
  │         │                ├── game_screenshots
  │         │                └── dlc
  │
  ├── orders ──────────────── order_items ──── products
  │
  ├── comments ──────────────(polymorphic: game/track/post)
  │
  ├── posts ────────────────── post_categories
  │
  ├── messages
  │
  ├── tickets ──────────────── ticket_messages
  │
  ├── api_keys
  │
  ├── notifications
  │
  └── audit_logs

albums ────────────────────── tracks
  └── (game_id → games)
```

---

## 接口规范与 API 契约 {#api-contract}

### 全局约定

**请求头：**
```
Authorization: Bearer {access_token}
Content-Type: application/json
Accept-Language: zh-CN
X-Client-Version: 1.0.0
X-Request-ID: {uuid}（由客户端生成，用于日志追踪）
```

**分页参数（cursor-based）：**
```
GET /api/v1/games?cursor=eyJpZCI6MTAwfQ&limit=20&sort=-created_at

响应：
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6ODB9",
    "has_more": true,
    "total": 150
  }
}
```

**错误响应格式：**
```json
{
  "code": 40400,
  "message": "游戏不存在",
  "data": null,
  "request_id": "uuid",
  "timestamp": 1700000000,
  "details": {
    "field": "game_id",
    "reason": "resource_not_found"
  }
}
```

### 核心接口列表

| 模块 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 认证 | POST | `/api/v1/auth/register` | 注册 |
| 认证 | POST | `/api/v1/auth/login` | 登录 |
| 认证 | POST | `/api/v1/auth/refresh` | 刷新 Token |
| 认证 | POST | `/api/v1/auth/logout` | 注销 |
| 认证 | POST | `/api/v1/auth/oauth/{provider}` | OAuth 登录 |
| 用户 | GET | `/api/v1/users/me` | 当前用户信息 |
| 用户 | PATCH | `/api/v1/users/me` | 更新用户信息 |
| 用户 | GET | `/api/v1/users/{id}` | 用户公开信息 |
| 游戏 | GET | `/api/v1/games` | 游戏列表 |
| 游戏 | GET | `/api/v1/games/{slug}` | 游戏详情 |
| 游戏 | GET | `/api/v1/games/{id}/releases` | 版本列表 |
| 下载 | GET | `/api/v1/releases/{id}/download` | 获取下载链接 |
| 专辑 | GET | `/api/v1/albums` | 专辑列表 |
| 专辑 | GET | `/api/v1/albums/{id}` | 专辑详情 |
| 音轨 | GET | `/api/v1/tracks/{id}/stream` | 音频流地址 |
| 评论 | GET | `/api/v1/comments` | 评论列表（多态） |
| 评论 | POST | `/api/v1/comments` | 发布评论 |
| 评论 | DELETE | `/api/v1/comments/{id}` | 删除评论 |
| 帖子 | GET | `/api/v1/posts` | 帖子列表 |
| 帖子 | POST | `/api/v1/posts` | 创建帖子 |
| 帖子 | GET | `/api/v1/posts/{id}` | 帖子详情 |
| 搜索 | GET | `/api/v1/search` | 全站搜索 |
| 商品 | GET | `/api/v1/products` | 商品列表 |
| 订单 | POST | `/api/v1/orders` | 创建订单 |
| 订单 | GET | `/api/v1/orders/{id}` | 订单详情 |
| 支付 | POST | `/api/v1/payments/create` | 发起支付 |
| 通知 | GET | `/api/v1/notifications` | 通知列表 |
| 通知 | PATCH | `/api/v1/notifications/read-all` | 标记全部已读 |
| 成就 | GET | `/api/v1/achievements` | 成就列表 |
| 工单 | POST | `/api/v1/tickets` | 创建工单 |
| 上传 | POST | `/api/v1/uploads/presign` | 获取上传预签名 URL |

---

## 项目里程碑与优先级矩阵 {#milestones}

### 里程碑规划

```
Phase 0 (M0) - 基础骨架 [预计 2 周]
  ✓ 项目初始化（Monorepo 结构、Go 模块、数据库迁移框架）
  ✓ 基础设施本地开发环境（Docker Compose）
  ✓ 全局 HTTP 中间件（日志、限流、CORS、认证）

Phase 1 (M1) - 核心用户系统 [预计 3 周]
  □ 用户注册/登录（邮箱+密码、JWT）
  □ 基本 RBAC 权限控制
  □ 用户信息管理

Phase 2 (M2) - 游戏分发核心 [预计 4 周]
  □ 游戏元数据 CRUD（管理后台）
  □ 版本发布流程
  □ 安全下载接口（Pre-signed URL + 防盗链）

Phase 3 (M3) - OST 平台 [预计 3 周]
  □ 专辑/音轨管理
  □ 音频流媒体接口
  □ 前端播放器组件

Phase 4 (M4) - 社区基础 [预计 4 周]
  □ 评论系统（支持嵌套）
  □ 帖子系统（论坛）
  □ 点赞/通知基础功能

Phase 5 (M5) - 商业化 [预计 5 周]
  □ 商品目录与定价
  □ 订单系统
  □ 支付宝/微信支付接入

Phase 6 (M6) - 增强功能 [持续迭代]
  □ 成就系统
  □ 搜索与推荐
  □ 数据分析看板
  □ 邮件营销
```

### 优先级矩阵

| 功能 | 业务价值 | 实现难度 | 优先级 |
|------|---------|---------|--------|
| 用户注册/登录 | 极高 | 中 | P0 |
| 游戏下载分发 | 极高 | 中 | P0 |
| OST 试听 | 高 | 低 | P0 |
| 评论系统 | 高 | 中 | P1 |
| 支付系统 | 极高 | 高 | P1 |
| 成就系统 | 中 | 中 | P2 |
| 推荐系统 | 高 | 高 | P2 |
| 热更新 | 中 | 高 | P3 |
| A/B 测试 | 低 | 高 | P3 |

---

## 技术债务管理 {#tech-debt}

### 技术债务登记

每个已知的技术债务需在此处登记，并在合适时机偿还：

| ID | 描述 | 影响范围 | 偿还时机 |
|----|------|---------|---------|
| TD-001 | 初期使用 `OFFSET` 分页，高偏移量性能差 | 评论、帖子列表 | M4 完成后迁移到 cursor-based |
| TD-002 | 音频转码暂用同步处理，阻塞上传接口 | 音频上传 | M3 后期引入异步 Worker |
| TD-003 | 初期统计数据实时查 PostgreSQL，高并发下有压力 | 下载量、播放量 | M5 后引入 Redis 计数 + 批量刷入 |
| TD-004 | 敏感词库硬编码在配置文件 | 内容治理 | M4 后期实现动态热更新 |

### 重构原则

1. 不允许为了重构而重构——只在有性能或维护性问题时进行
2. 重构前必须有测试覆盖（测试覆盖率 > 80%）
3. 重构以小步骤进行，每个 PR 只做一件事
4. 重构必须在 Staging 环境验证后才能上线

---

*文档版本：2.0.0 | 最后更新：2026-03-03 | 维护者：工作室技术团队*
