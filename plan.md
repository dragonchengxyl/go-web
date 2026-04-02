# AI Agent Workspace 实施计划

> 面向当前仓库，不是独立 AI demo。
> 更新时间：2026-04-02
> 目标：在现有 `Assistant + Copilot + RAG + Multimodal` 基础上，落地一个真正可执行、可审批、可观测的 `Agent Workspace`，V1 聚焦“发帖 Agent”。

---

## 1. 核心目标

我们现在要做的不是：

- 直接把现有站内浮窗助手硬改成万能 Agent
- 新开一个独立 `agent-web/` 或 `agent-service/` 子项目
- 让模型直接拼 SQL、直接调 repo、直接越过 usecase 写业务数据
- 一上来做“全站万能自动化”，覆盖发帖、活动、圈子、私信、后台全部场景
- 先做复杂多 Agent 编排，再回头补审批、审计和运行态

我们现在要做的是：

- 保留现有浮窗助手，继续承担轻量问答、站内导览和页面 Copilot
- 新开一个全页 `Agent Workspace`，承载真正的执行型任务
- 先把 Agent 范围收窄到一个清晰闭环：`发帖 Agent`
- 明确区分只读工具和写入工具，所有写入动作必须可审批
- 把运行态、步骤、工具调用、审批记录、产物都沉淀下来，支持回看和审计

一句话定义：

> 在现有网站里新增一个“Agent 工作台”，让用户可以围绕发帖目标发起任务，由服务端 Agent 在受控工具集合内做规划、检索、分析、生成草稿，并在用户批准后执行有限写操作。

---

## 2. 当前仓库基线

当前项目已经具备一条很好的 AI 接入基础，但它更偏“助手”，还不是严格意义上的 Agent：

- 站内浮窗助手：`apps/web/src/components/assistant/furry-assistant.tsx`
- 页面级上下文注入：`apps/web/src/contexts/assistant-page-context.tsx`
- 助手核心服务：`internal/usecase/assistant_service.go`
- 页面 Copilot：`internal/usecase/assistant_copilot.go`
- 知识索引与反馈：`internal/usecase/assistant_rag.go`
- 多模态分析：`internal/usecase/multimodal_service.go`
- 后台 AI 工具：`internal/usecase/admin_ai_tools.go`
- SSE 接口：`internal/transport/http/handler/assistant_handler.go`
- 多模态接口：`internal/transport/http/handler/multimodal_handler.go`
- 后台工作台：`apps/web/src/app/admin/assistant/page.tsx`
- 主服务接线：`cmd/server/main.go`

这意味着我们不需要从零造轮子，但也不应该继续把 `AssistantService` 往“万能执行器”方向堆。

---

## 3. 产品范围定义

## 3.1 V1 必须做到的范围

- 保留现有浮窗助手，不替换
- 新增全页 `Agent Workspace`
- 新增 `发帖 Agent` 单一任务域
- 支持从 `发布动态` 页面带着上下文进入 Agent
- Agent 运行过程可流式展示
- 展示规划步骤、工具调用、观察结果、最终建议
- 支持结构化产物输出：
  - 标题备选
  - 正文润色稿
  - 标签建议
  - 可见性建议
  - 图片理解摘要
- 支持审批型动作：
  - 保存草稿
  - 回填发帖页草稿
- 运行记录持久化
- 支持取消运行、查看历史运行结果
- 写入动作全量审计

## 3.2 V1 推荐一并做

- `Agent Runs` 历史页
- 每次运行的 `step timeline`
- 失败重试与错误摘要
- 运行状态统计与后台观测
- 将现有多模态缓存和页面上下文更稳定地复用给 Agent

## 3.3 V1 明确不做

- 全站万能 Agent
- 多 Agent 并行协作
- 长期人格记忆
- 自动发帖且无审批
- 自动加入圈子、自动报名活动、自动发私信
- 任意代码执行、任意 HTTP 调用、任意数据库写入
- 复杂 workflow builder

---

## 4. 关键架构决策

## 4.1 双轨并存，不替换现有助手

现有浮窗适合：

- 快速提问
- 页面导览
- 轻量 Copilot 建议

新的 Agent Workspace 适合：

- 明确目标
- 多步骤执行
- 产物整理
- 审批写入

也就是说：

- `Assistant` 继续做“聊天”
- `Agent` 开始做“任务执行”

## 4.2 Agent 必须是服务端循环，不是前端拼 prompt

不能让前端自己组织“步骤 -> 工具 -> 结果”的循环。

必须坚持：

- 规划逻辑在服务端
- 工具调用在服务端
- 权限校验在服务端
- 审批拦截在服务端
- 前端只负责呈现运行态和提交批准

## 4.3 工具必须是 typed tools，不允许模型直接碰 usecase 之外的底层

Agent 不应该直接访问：

- SQL
- repo
- Redis 原始 key
- 任意内部服务对象

正确方式是：

- Agent 只能看到一个受控工具注册表
- 每个工具都有明确输入输出
- 每个工具内部再调用现有 usecase

## 4.4 写操作默认需要审批

V1 必须把风险控住：

- `read_only` 工具可直接执行
- `write_requires_approval` 工具必须等用户批准
- `admin_only` 工具暂不进入用户 Agent

## 4.5 先做单域闭环，再扩场景

第一阶段只做 `发帖 Agent`，原因很现实：

- 当前发帖页已有页面上下文
- 已有图片分析结果
- 产物天然是结构化的
- 风险相对可控
- 审批边界清晰

---

## 5. 目标接入形态

## 5.1 面向用户的页面

- `/agent`
  - Agent Workspace 主页面
  - 创建任务、查看运行态、审批动作、查看产物
- `/agent/runs/[id]`
  - 单次运行详情页
  - 支持回放执行时间线

## 5.2 面向页面场景的入口

- `/posts/create`
  - 新增“交给 Agent 处理”入口
  - 自动带入草稿、标签、可见性、图片分析和目标圈子信息
- 第二阶段可选接入：
  - `/groups/[id]`
  - `/events/[id]`

## 5.3 面向后台的页面

- 先复用 `/admin/assistant`
- 后续可扩：
  - `Agent 运行量`
  - `审批通过率`
  - `步骤失败率`
  - `各工具调用量`

---

## 6. 后端设计

## 6.1 新增领域对象

建议新增 `internal/domain/agent`：

- `Run`
- `Step`
- `ToolCall`
- `Approval`
- `Artifact`

建议的核心状态：

- `queued`
- `running`
- `waiting_approval`
- `completed`
- `failed`
- `cancelled`

## 6.2 建议新增数据表

- `agent_runs`
- `agent_steps`
- `agent_tool_calls`
- `agent_approvals`
- `agent_artifacts`

其中：

- `agent_runs` 记录目标、状态、所属用户、场景、最终摘要
- `agent_steps` 记录规划步骤和执行结果
- `agent_tool_calls` 记录工具输入输出、耗时、错误
- `agent_approvals` 记录待批准动作、批准人、批准时间
- `agent_artifacts` 记录草稿产物和结构化结果

## 6.3 建议新增用例层

建议新增：

- `internal/usecase/agent_service.go`
- `internal/usecase/agent_tools.go`
- `internal/usecase/agent_runner.go`

职责拆分：

- `AgentService`
  - 创建运行
  - 获取运行
  - 流式执行
  - 审批继续
  - 取消运行
- `ToolRegistry`
  - 注册工具
  - 校验工具权限
  - 统一执行工具
- `Runner`
  - 执行“规划 -> 工具 -> 观察 -> 再规划”的循环

## 6.4 不建议直接改造现有 AssistantService 为 AgentService

推荐关系：

- `AssistantService`
  - 继续负责聊天和 RAG
- `AgentService`
  - 复用 `AssistantService` 的部分能力
  - 但单独维护运行态和工具循环

也就是说：

- Agent 可以调用现有助手能力
- 但 Agent 不等于助手

---

## 7. 工具设计

## 7.1 V1 只读工具

建议优先实现：

- `get_post_create_context`
- `analyze_attached_media`
- `get_group_context`
- `search_similar_posts`
- `suggest_tags`
- `summarize_draft_risks`
- `generate_post_variants`

这些工具的特点：

- 不修改用户数据
- 输出结构稳定
- 易于测试

## 7.2 V1 审批型写工具

建议只开放：

- `save_post_draft`
- `apply_agent_artifact_to_draft`

V1 不建议直接开放：

- `publish_post`
- `join_group`
- `attend_event`

## 7.3 工具协议要求

每个工具至少应有：

- `name`
- `description`
- `access_level`
- `input_schema`
- `output_schema`
- `timeout`
- `idempotent`

---

## 8. API 与流式协议

## 8.1 HTTP 接口建议

- `POST /api/v1/agent/runs`
- `GET /api/v1/agent/runs`
- `GET /api/v1/agent/runs/:id`
- `POST /api/v1/agent/runs/:id/stream`
- `POST /api/v1/agent/runs/:id/approve`
- `POST /api/v1/agent/runs/:id/cancel`

## 8.2 SSE 事件建议

现有助手 SSE 只有聊天粒度，不够 Agent 用。Agent 需要扩成：

- `meta`
- `plan`
- `step`
- `tool_call`
- `tool_result`
- `approval_required`
- `artifact`
- `final`
- `error`

## 8.3 前后端协议原则

- 前端不推断 Agent 状态
- 所有状态跃迁由服务端给出
- 所有审批动作必须带 `run_id` 和 `approval_id`

---

## 9. 前端工作台设计

## 9.1 页面结构

`/agent` 推荐采用三栏或双栏信息架构：

- 左侧：
  - 任务创建
  - 场景选择
  - 历史运行列表
- 中间：
  - 执行时间线
  - 当前步骤
  - 工具结果
- 右侧：
  - 最终产物
  - 审批卡片
  - 可应用到草稿的操作

## 9.2 交互原则

- 聊天不是主角，任务进度才是主角
- 用户始终能知道 Agent 在做什么
- 用户始终能知道哪些动作还没被执行，只是在等待批准

## 9.3 与现有页面上下文的关系

Agent 需要复用已有 `AssistantPageContext`，但不能只停留在提示词层面。

应该做到：

- 从页面上下文构造初始任务输入
- 把页面上下文同时存到 `agent_runs.context_snapshot`
- 后续每个工具都可读取这个快照，而不是依赖前端重复提交

---

## 10. V1 发帖 Agent 目标链路

期望形成这样一条真实链路：

`/posts/create` -> 点击“交给 Agent 处理” -> 跳转 `/agent` 并带入当前草稿上下文 -> Agent 先拆解任务 -> 调用图片分析 / 圈子上下文 / 相似内容检索 / 标签建议工具 -> 生成结构化草稿结果 -> 用户批准“应用到草稿” -> 回填发布页 -> 用户自行确认并发布

这里必须注意：

- V1 的最终发布动作仍由用户自己点
- Agent 只负责把草稿做得更完整、更可发

---

## 11. 权限、安全与治理

## 11.1 用户边界

- Agent 只能操作当前用户自己的草稿和上下文
- 不能读取他人私有数据
- 不能调用后台工具

## 11.2 审计要求

以下行为必须记审计：

- 创建运行
- 触发审批
- 批准写入
- 取消运行
- 工具执行失败

## 11.3 运行保护

V1 需要硬性限制：

- 单次运行最大步骤数
- 单次运行最大工具调用数
- 单工具超时
- 总运行超时
- 同用户并发运行数

---

## 12. 观测与指标

建议新增或扩展指标：

- `agent_runs_total`
- `agent_run_duration_seconds`
- `agent_steps_total`
- `agent_tool_calls_total`
- `agent_tool_call_duration_seconds`
- `agent_approval_wait_seconds`
- `agent_failures_total`

后台至少要能看到：

- 最近运行数
- 完成率
- 失败率
- 审批通过率
- 最常调用工具

---

## 13. 文件落点建议

后端建议新增：

- `internal/domain/agent/entity.go`
- `internal/domain/agent/repository.go`
- `internal/infra/postgres/agent_repo.go`
- `internal/usecase/agent_service.go`
- `internal/usecase/agent_tools.go`
- `internal/transport/http/handler/agent_handler.go`
- `migrations/*_create_agent_tables.up.sql`

前端建议新增：

- `apps/web/src/app/agent/page.tsx`
- `apps/web/src/app/agent/runs/[id]/page.tsx`
- `apps/web/src/components/agent/*`
- `apps/web/src/lib/agent-client.ts`

现有文件建议增量接入：

- `apps/web/src/app/posts/create/page.tsx`
- `apps/web/src/lib/api-client.ts`
- `internal/transport/http/router.go`
- `cmd/server/main.go`

---

## 14. 开发阶段建议

## 14.1 Phase 0：骨架搭建

- 建表
- 建 domain / repo / usecase 骨架
- 建 `/agent` 页面空壳
- 建运行状态和时间线 UI 基础组件

## 14.2 Phase 1：只读 Agent 跑通

- 支持创建运行
- 支持只读工具调用
- 支持时间线流式展示
- 支持生成结构化草稿产物

## 14.3 Phase 2：审批型写工具

- 增加审批卡片
- 增加“应用到草稿”
- 跑通从 Agent 到发帖页的回填链路

## 14.4 Phase 3：观测与扩展

- 后台运行指标
- 历史运行页
- 更多任务域：
  - 圈子 Agent
  - 活动 Agent

---

## 15. 验收标准

V1 完成后，项目里应具备以下真实能力：

- 用户可以进入 `/agent`
- 用户可以从发帖页带着草稿进入 Agent
- Agent 能展示规划、步骤、工具调用和结果
- Agent 能生成可读、可复用的发帖产物
- 写动作不会直接执行，必须经过批准
- 用户批准后可以把结果回填到草稿
- 运行过程和结果可以回看
- 后端可以审计和观测这条链路

---

## 16. AI / 开发执行顺序

如果让 AI 或开发者按顺序推进，建议严格按下面执行：

1. 先建 `agent` 的数据结构和运行状态，不要先写大段 prompt
2. 再建工具注册表和只读工具，不要先开放写工具
3. 再做 `/agent` 页面和 SSE 时间线，不要先把功能塞进浮窗
4. 再接入发帖页场景，让闭环真正跑通
5. 最后再做审批写入、历史运行和后台观测

优先级永远是：

1. 可控
2. 可解释
3. 可执行
4. 可观测
5. 可扩展

---

## 17. 下一步最小实现范围

下一步不应该直接做“全套 Agent 平台”，而应该只做最小可跑版本：

- 新增 `/agent`
- 新增 `AgentService` 骨架
- 新增 `agent_runs`、`agent_steps`、`agent_tool_calls`
- 接一个 `发帖 Agent`
- 先支持只读工具和结构化产物
- 暂时把写入动作收敛到“批准后回填草稿”

---

## 附录 A. 历史规划保留

以下保留原有斗地主接入计划，避免覆盖已有方案；后续可以再单独拆到 `docs/`。

---

# 斗地主接入现有网站实施计划

> 面向当前仓库，不是独立项目模板。
> 更新时间：2026-03-23
> 目标：在现有 `Next.js + Gin + PostgreSQL + Redis + WebSocket + Games Hub` 基础上，落地一个可以真正接入网站、可演示、可继续迭代的斗地主版本。

---

## 1. 核心目标

我们现在要做的不是：

- 新开一个 `doudizhu/` 子项目
- 再写一个独立 Vue 验证前端
- 回到 MySQL + GORM + 单仓库 demo 的旧方案

我们现在要做的是：

- 把斗地主作为站内第二款可运行游戏，接入现有 `/games` 体系
- 复用当前用户系统、鉴权、WebSocket、管理后台、排行榜和运维能力
- 让它能自然接到社区、战报、后台观测，而不是成为孤立 demo

一句话定义：

> 在现有网站中做一个“经典三人斗地主房间版”，支持登录用户创建房间、加入房间、实时对局、断线重连、结算落库、战报查询和后台观测，并提供 1 人 + 2 机器人的演示模式。

---

## 2. 当前仓库基线

当前项目已经具备一条完整的游戏接入样板，虽然只服务于 `Hex Blitz`，但它已经证明这些能力可用：

- 前端游戏中心：`apps/web/src/app/games`
- 游戏详情页 / 试玩页：`apps/web/src/app/games/[slug]/page.tsx`、`apps/web/src/app/games/[slug]/play/page.tsx`
- 游戏目录配置：`apps/web/src/lib/games.ts`
- 游戏 API SDK：`apps/web/src/lib/api-client.ts`
- 实时房间链路：`internal/transport/http/handler/game_handler.go`
- 房间服务与内存态：`internal/usecase/hex_blitz_room_service.go`
- 游戏结果落库：`internal/infra/postgres/hex_blitz_repo.go`
- 后台观测页：`apps/web/src/app/admin/games/page.tsx`
- 主服务接线：`cmd/server/main.go`
- 路由注册：`internal/transport/http/router.go`

这意味着斗地主不应该另起炉灶，而应该沿着这条链路接入。

---

## 3. 产品范围定义

## 3.1 V1 必须做到的范围

- 经典三人斗地主
- 登录用户创建房间、加入房间
- 人机演示模式，支持 1 名真人玩家 + 2 个服务端机器人完整开局
- 房主开局
- 发牌、叫分 / 抢地主、确定地主
- 轮流出牌、过牌、服务端校验牌型与大小
- 回合倒计时与超时托管
- 对局结算
- 对局结果落库
- 最近对局列表
- 基础排行榜
- WebSocket 实时同步
- 同进程断线重连
- 站内详情页与可玩页接入
- 后台基础观测

## 3.2 V1 推荐一并做

- Redis 房间快照，降低 API 进程重启带来的房间丢失风险
- 对局操作日志落库，用于复盘 / 回放
- “我的斗地主最近战绩”
- 基础战报页
- 机器人基础策略可配置，便于后续调 demo 体验

## 3.3 V1 明确不做

- 高复杂度 AI、博弈搜索或多难度机器人系统
- 匹配池 / 段位排位
- 金币经济、道具、商城
- 观战模式
- 语音聊天
- 好友邀请通知
- 复杂规则变体：癞子、残局、天地癞子、比赛场
- 跨进程房间迁移

---

## 4. 关键架构决策

## 4.1 不先做“通用游戏引擎大重构”

当前代码里的游戏能力高度偏向 `Hex Blitz`。如果一上来为了斗地主去重构出一套“万能游戏框架”，项目会显著变慢。

本计划采取更稳妥的路线：

- 保留 `Hex Blitz` 现状，不为接第二款游戏而大改现有逻辑
- 斗地主先走“平行落地”方案
- 只抽离真正会复用的层：游戏目录、路由接线、后台观测入口、公共 WS 鉴权方式

也就是说：

- `Hex Blitz` 继续使用现有 `HexBlitzRoomService`
- 斗地主新增 `DouDizhuService` / `DouDizhuHandler`
- 当两款游戏都稳定后，再抽象共用的 `GameRegistry` 或管理接口

## 4.2 斗地主必须是服务端权威

和 `Hex Blitz` 不同，斗地主不能接受任何“客户端算好结果再上报”的模型。

必须坚持：

- 发牌由服务端完成
- 叫分、出牌、过牌由客户端只发送“意图”
- 牌型识别、合法性校验、胜负判断全部在服务端
- 客户端只负责展示，不负责裁决

## 4.3 房间态优先，匹配延后

V1 不做自动匹配，先做站内最容易接入、最容易演示、最符合现有社区网站气质的模式：

- 创建房间
- 房间列表
- 房间码加入
- 3 人凑齐后开始

这样可以直接复用当前 `Games Hub -> 游戏详情 -> 进入房间实验室` 的信息架构。

## 4.4 先接网站，再谈长期治理

斗地主第一版不是为了做一个棋牌平台，而是为了让你现在的网站里真的多出一款多人实时游戏。

所以第一阶段的优先级是：

1. 能进入网站
2. 能登录玩
3. 能稳定打一局
4. 能落库和查看战绩
5. 能在后台观测

## 4.5 机器人只做“演示兜底”，不做独立 AI 项目

加入人机模式的原因很现实：

- 演示时不一定临时凑得到 3 个真人
- 面试或录屏需要稳定跑完整局
- 网站中第二款游戏需要一个“单人即可体验”的入口

但机器人在 V1 的定位必须收住：

- 机器人运行在服务端，和真人一样走统一规则引擎
- 不做机器学习，不做复杂博弈搜索
- 先实现“合法、稳定、节奏自然”的基础策略
- 优先保证 demo 可跑通，而不是追求高胜率或高拟人度

---

## 5. 目标接入形态

## 5.1 面向用户的页面

- `/games`
  - 游戏中心新增斗地主卡片
- `/games/dou-dizhu`
  - 斗地主详情页
  - 展示规则摘要、玩法定位、开发状态、路线图
- `/games/dou-dizhu/play`
  - 斗地主房间大厅 + 对局界面
  - 支持创建真人房或快速开始人机演示房
- `/games/dou-dizhu/matches/[matchId]`
  - 斗地主战报页

## 5.2 面向后台的页面

- `/admin/games`
  - 从“仅 Hex Blitz 观测页”升级为“多游戏概览页”
- 可选新增：
  - `/admin/games/dou-dizhu`
    - 斗地主专属运行态、活跃房间、最近对局、异常计数

## 5.3 面向后端的接口

HTTP：

- `GET /api/v1/games/dou-dizhu/rooms`
- `GET /api/v1/games/dou-dizhu/rooms/:id`
- `POST /api/v1/games/dou-dizhu/rooms`
- `POST /api/v1/games/dou-dizhu/rooms/demo`
- `GET /api/v1/games/dou-dizhu/leaderboard`
- `GET /api/v1/games/dou-dizhu/matches`
- `GET /api/v1/games/dou-dizhu/matches/:match_id/replay`
- `GET /api/v1/games/dou-dizhu/matches/me`

WebSocket：

- `GET /ws/game/dou-dizhu`

鉴权方式沿用现有游戏 WS 方案：

- 登录用户通过 query `token` 传入 access token
- 仍保留 `session_id` 作为房间内会话标识

---

## 6. 数据与状态设计

## 6.1 房间运行时状态

房间运行时先放在内存中，由 `DouDizhuService` 管理：

- 房间基础信息
- 3 个座位
- 玩家在线状态
- 机器人座位与机器人配置
- 准备状态
- 当前局状态机
- 当前手牌
- 地主牌
- 当前轮到谁
- 上一手牌型
- 倍数 / 炸弹 / 春天状态
- 托管状态

推荐增加 Redis 快照：

- Key：`game:doudizhu:room:{room_id}`
- TTL：2 小时
- 每次关键状态迁移时写入
- 用于断线重连与未来的重启恢复

## 6.2 数据库表设计

建议新增 3 组表，不直接复用 `hex_blitz_*` 表：

### `doudizhu_matches`

保存一局的主记录：

- `id`
- `room_id`
- `room_code`
- `room_title`
- `started_at`
- `finished_at`
- `match_mode`，值为 `pvp` / `demo_ai`
- `landlord_seat`
- `winner_side`，值为 `landlord` / `farmers`
- `multiplier`
- `bomb_count`
- `spring` / `anti_spring`
- `created_at`

### `doudizhu_match_players`

保存每个玩家在该局中的结果：

- `id`
- `match_id`
- `session_id`
- `user_id`
- `is_bot`
- `bot_level`
- `seat`
- `player_name`
- `display_name`
- `role`，值为 `landlord` / `farmer`
- `cards_left`
- `bid_score`
- `is_winner`
- `score_delta`
- `created_at`

### `doudizhu_action_events`

保存操作日志，供回放 / 审计 / 排查：

- `id`
- `match_id`
- `turn_no`
- `action_index`
- `session_id`
- `user_id`
- `seat`
- `action_type`
- `cards_json`
- `combo_type`
- `combo_rank`
- `multiplier_after`
- `occurred_at`

---

## 7. 服务端模块落点

## 7.1 建议新增的后端文件

### 领域层

- `internal/domain/doudizhu/entity.go`
- `internal/domain/doudizhu/rule.go`
- `internal/domain/doudizhu/repository.go`

### 用例层

- `internal/usecase/doudizhu_engine.go`
- `internal/usecase/doudizhu_bot.go`
- `internal/usecase/doudizhu_room_service.go`
- `internal/usecase/doudizhu_room_service_test.go`
- `internal/usecase/doudizhu_rule_engine_test.go`
- `internal/usecase/doudizhu_bot_test.go`

### 基础设施层

- `internal/infra/postgres/doudizhu_repo.go`
- 可选：`internal/infra/redis/doudizhu_room_store.go`

### 传输层

- `internal/transport/http/handler/doudizhu_handler.go`

### 迁移

- `migrations/*_create_doudizhu_matches.up.sql`
- `migrations/*_create_doudizhu_matches.down.sql`
- `migrations/*_create_doudizhu_action_events.up.sql`
- `migrations/*_create_doudizhu_action_events.down.sql`

## 7.2 需要修改的后端文件

- `cmd/server/main.go`
  - 初始化斗地主 repo / service
  - 注入 router
- `internal/transport/http/router.go`
  - 注册斗地主 HTTP 路由
  - 注册 `/ws/game/dou-dizhu`
- `internal/transport/http/handler/admin_handler.go`
  - 扩展为多游戏概览，不再只盯 `Hex Blitz`

---

## 8. 斗地主状态机

建议使用明确的阶段状态，而不是散落在多个 bool 上：

### 房间阶段

- `waiting`
- `ready`
- `dealing`
- `bidding`
- `playing`
- `settlement`

### 局内流程

1. 房主建房
2. 玩家加入，最多 3 人；人机演示房则自动补齐 2 个机器人
3. 所有人准备
4. 发牌
5. 叫分 / 抢地主
6. 确定地主并发 3 张底牌
7. 正式出牌
8. 一方手牌出完
9. 结算
10. 回到房间，等待下一局

### 超时策略

- 叫分超时：自动“不叫”或按最低优先级处理
- 出牌超时：
  - 有可过时默认“过”
  - 首出不能过时，自动按最小合法牌型代出
- 托管可手动开启 / 超时自动开启

---

## 9. WebSocket 协议设计

沿用现有房间类消息模式，但斗地主必须做“公共态广播 + 私有手牌单发”的区分。

## 9.1 客户端 -> 服务端

- `ready`
- `cancel_ready`
- `start_game`
- `bid`
- `pass_bid`
- `play_cards`
- `pass_turn`
- `toggle_auto_play`
- `leave_room`
- `ping`

## 9.2 服务端 -> 客户端

- `joined`
- `room_state`
- `seat_state`
- `game_state`
- `hand_state`
- `bid_state`
- `turn_state`
- `action_result`
- `settlement`
- `error`
- `pong`

## 9.3 重要约束

- `room_state` 只广播公共信息
- `hand_state` 只发给对应玩家本人
- 其他玩家只能看到：
  - 手牌数量
  - 最近操作
  - 是否托管
  - 是否在线
- 任何时候都不能把别人的完整手牌广播出去

---

## 10. HTTP 接口职责

斗地主不要把所有事情都塞进 WebSocket。

建议职责划分如下：

### HTTP 负责

- 房间列表
- 房间详情
- 创建房间
- 快速开始人机演示房
- 最近对局
- 我的最近对局
- 排行榜
- 战报查询

### WebSocket 负责

- 加入实时房间
- 准备 / 开局
- 叫分 / 出牌 / 过牌
- 局内状态推进
- 断线重连后的状态补发

---

## 11. 前端接入设计

## 11.1 需要新增的前端文件

- `apps/web/src/components/games/doudizhu-play-stage.tsx`
- `apps/web/src/components/games/doudizhu-table.tsx`
- `apps/web/src/components/games/doudizhu-hand.tsx`
- `apps/web/src/components/games/doudizhu-bid-panel.tsx`
- `apps/web/src/components/games/doudizhu-action-log.tsx`
- `apps/web/src/app/games/dou-dizhu/matches/[matchId]/page.tsx`

## 11.2 需要修改的前端文件

- `apps/web/src/lib/games.ts`
  - 新增斗地主目录项
  - 说明其状态为 `playable` 或 `prototype`
- `apps/web/src/app/games/[slug]/page.tsx`
  - 支持斗地主详情页内容
- `apps/web/src/app/games/[slug]/play/page.tsx`
  - 不再只对 `hex-blitz` 特判
  - 为斗地主挂载对应 `PlayStage`
- `apps/web/src/lib/api-client.ts`
  - 新增斗地主类型和 API 方法
- `apps/web/src/app/admin/games/page.tsx`
  - 从 Hex Blitz 单游戏观测扩为多游戏视图

## 11.3 前端页面结构

`/games/dou-dizhu/play` 建议分成 4 块：

1. 房间大厅
   - 创建房间
   - 快速开始人机演示
   - 房间列表
   - 加入中的反馈
2. 房间头部
   - 房间标题
   - 房间码
   - 在线状态
   - 当前阶段
3. 牌桌主体
   - 三个座位
   - 当前轮次
   - 地主标记
   - 最近出牌区
4. 我的操作区
   - 手牌
   - 叫分按钮
   - 出牌 / 过牌按钮
   - 托管开关

## 11.4 UI 原则

- 第一版以可读性和状态清晰为第一优先级
- 不追求重棋牌美术
- 但必须让面试或演示时一眼看出：
  - 谁是地主
  - 轮到谁
  - 上一手是什么
  - 我当前能做什么

---

## 12. 社区与站内整合点

斗地主是网站功能，不是孤岛。

因此应明确留出这些整合点：

## 12.1 账号体系

- 只允许登录用户参与正式对局
- 机器人账号不走独立登录体系，由服务端虚拟玩家表示
- 使用现有 JWT 与用户资料
- `display_name` 优先显示站内用户名

## 12.2 战报体系

V1 至少做到：

- 最近对局列表
- 我的最近对局
- 战报详情页

V1.1 可接：

- “分享本局到社区” 按钮
- 自动生成简短战报卡片，供发帖引用

## 12.3 管理后台

管理员应能看到：

- 当前活跃房间数
- 当前在线玩家数
- 当前机器人对局数
- 当前进行中的局数
- 托管次数
- 超时出牌次数
- 最近对局
- 基础榜单

---

## 13. 观测与风控

## 13.1 Prometheus 指标

建议新增：

- `doudizhu_active_rooms`
- `doudizhu_active_players`
- `doudizhu_matches_finished_total`
- `doudizhu_ws_connections`
- `doudizhu_timeout_actions_total`
- `doudizhu_auto_play_actions_total`
- `doudizhu_reconnect_success_total`
- `doudizhu_invalid_actions_total`
- `doudizhu_bot_turn_total`

## 13.2 日志

每次关键动作必须记录：

- `room_id`
- `match_id`
- `user_id`
- `session_id`
- `seat`
- `action`
- `phase`
- `turn_no`
- `request_id`

## 13.3 基础风控

V1 至少做这些：

- 非当前行动玩家不能出牌
- 非法牌型直接拒绝
- 小于当前牌型的出牌直接拒绝
- 首出不能过
- 重复消息幂等处理
- 房间满员后拒绝额外加入

---

## 14. 分阶段实施计划

## Phase 0：接入准备

目标：让斗地主有进入网站的“壳”。

任务：

- 在 `games.ts` 增加斗地主目录项
- 让 `/games/[slug]`、`/games/[slug]/play` 支持不止一款游戏
- 为斗地主准备详情页文案和空白试玩页
- 明确真人房与人机演示房的入口文案
- 明确 API 类型与 WS 路径命名

完成标准：

- 网站里能看到斗地主入口
- 能进入斗地主详情页与占位试玩页

## Phase 1：规则引擎与状态机

目标：先把斗地主本身做对。

任务：

- 定义牌、组合、牌型比较规则
- 支持经典斗地主核心牌型
- 实现发牌和地主流程
- 实现局内状态推进
- 预留机器人决策所需的状态读取接口
- 补齐核心单元测试

完成标准：

- 不依赖前端，也能通过测试跑完整局
- 主要牌型比较测试覆盖通过

## Phase 2：房间服务与 WebSocket

目标：能在线打一局。

任务：

- 实现 `DouDizhuRoomService`
- 实现房间创建 / 加入 / 准备 / 开局
- 实现人机演示房自动补齐 2 个机器人
- 实现机器人基础决策：叫分、出牌、过牌
- 接入 `/ws/game/dou-dizhu`
- 实现局内广播和私有手牌下发
- 实现超时与托管
- 实现同进程断线重连

完成标准：

- 3 个浏览器用户可以完成一局对战
- 1 个真人用户也可以通过人机模式完整打完一局
- 中途断开后可重新连回本局

## Phase 3：落库、榜单、战报

目标：让斗地主不只是“打一局就没了”。

任务：

- 新增 SQL migration
- 实现 `doudizhu_repo`
- 保存对局主记录、玩家结果、操作事件
- 区分真人局与人机演示局
- 提供最近对局、我的对局、基础排行榜接口
- 提供战报页查询接口

完成标准：

- 打完一局后数据能在数据库中看到
- 前台能看到最近对局和战报

## Phase 4：前端完整接入

目标：把功能真正做成网站产品。

任务：

- 实现斗地主 `PlayStage`
- 牌桌、座位、手牌、操作面板、日志区完整接线
- 对接 HTTP 和 WS
- 完成错误提示、重连提示、阶段提示
- 打磨移动端基础可用性

完成标准：

- 普通用户进入网站即可完成“创建房间 -> 开局 -> 打完 -> 看战报”全链路

## Phase 5：后台观测与多游戏治理

目标：不让后台还只认识 `Hex Blitz`。

任务：

- 把 `/admin/games` 升级成多游戏视图
- 新增斗地主运行态指标展示
- 增加最近异常动作和超时统计
- 评估是否抽象 `GameRegistry`

完成标准：

- 管理后台能同时看到 Hex Blitz 和斗地主的运行情况

---

## 15. 验收标准

满足以下条件，才算“斗地主已经接入网站”：

- `Games Hub` 中出现斗地主入口
- 登录用户可以创建和加入斗地主房间
- 三名玩家能完成一局完整对战
- 单个登录用户可以通过人机模式完整跑通一局
- 服务端完整裁定发牌、叫分、出牌、胜负
- 断线重连后可以继续本局
- 结算数据能落库
- 能查看最近对局和战报
- 后台能看到斗地主基本运行态
- `go test ./...` 通过

---

## 16. 明确不建议的实现方式

以下做法会让项目偏离当前仓库，应避免：

- 再开一个独立 `frontend/` Vue 工程
- 引入 GORM / MySQL，只为斗地主单独走另一套栈
- 把斗地主做成独立服务后再反向嵌回主站
- 为了“通用化”重写整个 `Hex Blitz` 链路
- 先去做复杂 AI 决策系统，再回头补基本对局链路
- 把游戏裁决逻辑放到前端
- 不落库，只保留演示态

---

## 17. AI / 开发执行顺序

如果让 AI 或开发者按顺序推进，建议严格按下面执行：

1. 先改 `games` 目录和页面入口，让斗地主能出现在站内
2. 再写规则引擎和测试，不要先写大段前端页面
3. 再写房间服务、机器人基础策略和 WS，同步出最小可跑链路
4. 再补数据库和战报
5. 最后做后台观测和多游戏治理

优先级永远是：

1. 正确性
2. 可接入
3. 可演示
4. 可观测
5. 可扩展

---

## 18. 最终产出定义

本计划完成后，项目里应存在这样一条真实链路：

`/games` -> `/games/dou-dizhu` -> `/games/dou-dizhu/play` -> 创建真人房或人机演示房 -> 完成一局对局 -> 结算落库 -> 战报查询 -> 后台观测

这才是“现在能接入到网站中的斗地主版本”。
