# 🗺️ Project Vision & Implementation Blueprint

## Furry 同好社区平台 — MVP 减法重构与合规演进路线图

**生成日期**: 2026-03-10
**战略方向**: 纯免费内容驱动 + 站长服务器透明赞助模式 (国内合规版)
**核心目标**: 极低运营成本、规避国内支付与涉黄合规风险、专注 DAU 与内容生态沉淀。

---

## 🔭 战略决议复盘

### 砍掉的复杂模块（做减法）

* ❌ **交易与分润体系**：剥离 Stripe 依赖、委托交易状态机（Escrow/Dispute）、创作者订阅制与数字商品商店。避免国内"资金二清"红线，砍掉 60% 的开发工作量。

### 选定路线：纯粹社区 + Buy Me a Coffee

当前 MVP 的发力点必须聚焦于两处：

1. **高性能与高可用**：解决 WebSocket 单点瓶颈，引入轻量级相似度推荐（pgvector）。
2. **国内合规防线**：用阿里云/腾讯云全面替代海外服务（R2/Sightengine），建立极严的内容机审防火墙与防刷机制。同时上线"站长赞助看板"，通过情绪价值覆盖服务器开销。

---

## ⚖️ 核心技术选型推荐 (国内平替版)

### C-1: 分布式 WebSocket Hub

**决策：Redis Pub/Sub**。利用已有的 `go-redis/v9`，零新增基础设施。

* 频道命名：`ws:user:{userID}`

### C-2: Feed 算法引擎

**推荐方案**：参与度加权评分 + pgvector 相似推荐（两阶段）。

* 依赖 PostgreSQL 的 `pgvector` 扩展。

### C-3: AI 内容审核 (国内合规生死线)

**决策：阿里云内容安全 (Aliyun Green)**

* **优势**：国内合规标准制定者，自带庞大的涉政、暴恐、涉黄图库与违禁词库。
* **机制**：异步流水线。上传完成 → `status: pending` → 后台调用阿里云 API → 命中 `block` 直接物理隔离并限制账号，`pass` 则进入信息流。

### C-4: 站长赞助系统与资源存储

**决策：阿里云 OSS + 纯前端展示**

* **存储**：阿里云 OSS（开启防盗链，绑定国内已备案域名）。
* **打赏**：不接入复杂支付 API。纯静态展示微信赞赏码与支付宝收款码，配合"本月服务器开销进度条"产生情绪共鸣。

---

## 🛡️ 威胁建模与防线 (国内合规特化)

| 威胁 | 攻击向量 | 影响 | 防线 |
| --- | --- | --- | --- |
| **涉黄/涉政内容爆破** | 凌晨批量上传高危 NSFW 图片 | 域名被墙，喝茶 | 阿里云内容安全 API 强行拦截，发帖必须经过 `pending` 状态机。 |
| **短信接口盗刷** | 恶意脚本疯狂请求注册短信 | 短信费用破产 | 短信下发前强制接入图形滑块验证码；单 IP 限频。 |
| **WebSocket 连接耗尽** | 攻击者建立大量 WS 连接 | 服务 OOM | 每用户最大 5 个并发连接；IP 级限速。 |
| **SSRF via MediaURL** | 上传含恶意 URL 的帖子 | 内网探测 | 媒体上传强制走 OSS 直传获取 STS Token，禁止外部 URL 直存。 |

---

## 📝 Issue 级别执行清单

> 严格按 Phase 2 → Phase 5 排序。每个 Issue 独立可测试、可合并。

---

### Phase 2: 基础设施 (Infrastructure)

#### [INFRA-01] 分布式 WebSocket Hub（Redis Pub/Sub）

**优先级**: P0 — 阻塞水平扩展
**实现要点**: 提取 `HubInterface`，实现基于 `PUBLISH ws:user:{userID}` 和 `SUBSCRIBE` 的跨节点转发。

#### [INFRA-02] WebSocket 每连接速率限制

**优先级**: P0 — 安全防线
**实现要点**: 在 `readPump` 中集成 Token Bucket（`golang.org/x/time/rate`），超限关闭连接并记录日志。

#### [INFRA-03] 阿里云 OSS 前端直传集成

**优先级**: P0 — 替换 R2，降本增效
**文件变更**: 新增 `internal/usecase/oss_service.go`
**实现要点**: 后端仅提供生成 STS 临时凭证的接口，前端使用凭证直传图片至 OSS，减轻服务器带宽压力。

#### [INFRA-04] pgvector 扩展与 Feed Bug 修复

**优先级**: P1 — 算法前置依赖
**实现要点**: 修复 `scanPosts` 的总数统计 Bug；引入 PostgreSQL 的 vector 扩展，并在 `posts` 表添加 `embedding` 列。

#### [INFRA-05] CI/CD 流水线与 K8s HPA

**优先级**: P2
**实现要点**: 配置 GitHub Actions 测试流（使用 testcontainers-go）及基于 CPU 70% 的弹性扩缩容。

---

### Phase 3: 安全防御 (Security)

#### [SEC-01] 图形验证码与防刷防线

**优先级**: P0 — 保护钱包
**文件变更**: 修改 `auth_service.go`
**实现要点**: 引入行为验证码（如极验或开源滑块组件），并在 Redis 配置 `sms:ip:{ip}` 与 `sms:phone:{phone}` 的双重频率锁。

#### [SEC-02] 阿里云内容安全异步流水线

**优先级**: P0 — 社区护城河
**文件变更**: 新增 `internal/infra/moderation/aliyun_green.go`
**流程**:

```
POST /posts → 入库 (status: pending) → 返回成功
                    ↓ goroutine
           调用 Aliyun Green API 进行图文双审
                    ↓
           UPDATE posts SET moderation_status = 'approved'|'blocked'
           若 blocked → 触发系统通知，违规计数 +1
```

#### [SEC-03] OSS 媒体 URL 白名单验证

**优先级**: P1
**实现要点**: 发帖时校验 `MediaURLs`，必须匹配平台配置的 OSS 域名，阻断注入风险。

---

### Phase 4: 测试先行 (TDD)

#### [TEST-01] PostService 与 Aliyun Mock 测试

**优先级**: P1
**实现要点**: 编写 `CreatePost` 的单元测试；Mock 阿里云 Green 的返回结果，确保 `blocked` 状态的内容绝对不会出现在 Feed 流中。

#### [TEST-02] 分布式 Hub 集成测试

**优先级**: P1
**实现要点**: 使用 testcontainers 模拟多节点环境，验证 WS 消息跨进程路由的准确性与速率限制。

---

### Phase 5: 业务落地 (Business)

#### [BIZ-01] 站长服务器赞助模块 (Sponsor Dashboard)

**优先级**: P0 — 平台回血核心
**文件变更**:

* 新增 `internal/transport/http/handler/sponsor_handler.go`
* 前端新增 `/sponsor` 页面

**实现要点**: 接口返回简单的配置信息（硬编码在配置文件中即可）：

```json
{
  "monthly_goal": 500.00,
  "current_raised": 150.00,
  "alipay_qr_url": "https://oss.xxx.com/alipay.jpg",
  "wechat_qr_url": "https://oss.xxx.com/wechat.jpg",
  "message": "各位兽圈同好，本月阿里云 OSS 流量费还差 350 元，求投喂~"
}
```

#### [BIZ-02] Explore Feed 算法升级

**优先级**: P1
**实现要点**: 基于 `(like_count + comment_count * 3) / 衰减时间` 排序，打破纯时序列表，提升优质内容曝光。

#### [BIZ-03] AI 艺术内容标签强制政策

**优先级**: P1
**实现要点**: 增设 `content_labels JSONB` 字段，发帖前端强制要求用户勾选"是否为 AI 生成"，便于用户按偏好过滤信息流。

---

## 📊 执行优先级汇总

```
Sprint 1 (P0 生死线与基建):
  INFRA-01  分布式 WS Hub
  INFRA-02  WS 速率限制
  INFRA-03  阿里云 OSS 直传
  SEC-01    验证码与防刷墙
  SEC-02    阿里云内容安全流水线
  BIZ-01    站长赞助打赏模块

Sprint 2 (P1 体验提升与算法):
  INFRA-04  pgvector 修复与扩展
  SEC-03    URL 白名单
  TEST-01   核心业务逻辑测试
  TEST-02   分布式 Hub 集成测试
  BIZ-02    Explore 算法升级
  BIZ-03    AI 标签过滤机制

Sprint 3 (P2 稳定性):
  INFRA-05  CI/CD 流水线与 K8s HPA
```

---

## 🚨 关键风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
| --- | --- | --- | --- |
| 机审 API 费用失控 | 中 | 高 | 将鉴黄与涉政接口降级调用（信任分高的老用户抽检，新用户全检）。 |
| 恶意刷量消耗 OSS 流量 | 中 | 中 | OSS 开启 Referer 防盗链，配合 CDN 边缘节点限制单 IP 带宽。 |
| 打赏转化率过低 | 高 | 低 | 优化 `/sponsor` 页面文案，提供已打赏用户的"鸣谢名录"，增加情绪价值。 |

---

*Blueprint v2 — Domestic Version | 2026-03-10*
