export type GameStatus = 'playable' | 'prototype' | 'coming_soon';

export interface GameCatalogEntry {
  slug: string;
  title: string;
  subtitle: string;
  status: GameStatus;
  playPageEnabled?: boolean;
  genre: string;
  playerMode: string;
  roundTime: string;
  onlineNow: number;
  shortDescription: string;
  heroDescription: string;
  accentFrom: string;
  accentTo: string;
  atmosphere: string;
  loops: string[];
  highlights: string[];
  roadmap: string[];
  fitForJD: string[];
}

export const gamesCatalog: GameCatalogEntry[] = [
  {
    slug: 'hex-blitz',
    title: 'Hex Blitz',
    subtitle: '六角冲分赛',
    status: 'playable',
    playPageEnabled: true,
    genre: '休闲竞技 / 连消冲分',
    playerMode: '当前单机，规划 2-4 人实时对局',
    roundTime: '75 秒一局',
    onlineNow: 18,
    shortDescription:
      '在六角棋盘上快速清除相邻同色块，叠高连击和爆裂特效，把分数一路推上去。',
    heroDescription:
      '这是第一款真正落地的网页休闲游戏原型。当前阶段先验证手感、分数循环和页面体验，下一阶段接入房间匹配、实时同步和排行榜。',
    accentFrom: '#ff8a3d',
    accentTo: '#34d2ff',
    atmosphere: '高饱和街机厅 / 琥珀 + 青蓝',
    loops: [
      '观察棋盘，优先找 2 格以上相邻同色块。',
      '连续清除会触发连击倍率，窗口时间只有 2.5 秒。',
      '星辉块提供额外分数，爆裂块会顺带清掉周围六角格。',
      '75 秒结束后结算本局分数，后续可接周榜和房间排位。',
    ],
    highlights: [
      '单局节奏短，适合休闲产品面试演示。',
      '规则清楚，后续扩展房间对战和服务端结算非常自然。',
      '美术资源需求轻，适合先用免费素材快速打样。',
    ],
    roadmap: [
      'Phase 1：单机原型、分数循环、页面动效、游戏中心接入。',
      'Phase 2：房间创建、准备态、Go 服务端 Tick、WebSocket 同步。',
      'Phase 3：排行榜、战报分享、社区帖子联动。',
      'Phase 4：压测、日志、指标、断线重连、后台游戏数据。',
    ],
    fitForJD: [
      '能讲清楚游戏服务器如何设计房间、Tick 和状态同步。',
      '可以展示你如何和策划一起压缩规则，和美术一起定义反馈。',
      '后续很适合补性能优化、稳定性和压测结果。',
    ],
  },
  {
    slug: 'dou-dizhu',
    title: '涂油斗地主',
    subtitle: '三人牌桌对战',
    status: 'prototype',
    playPageEnabled: true,
    genre: '棋牌对战 / 三人牌桌',
    playerMode: '支持 3 人联机，也支持 1 人 + 2 陪练热身',
    roundTime: '4-8 分钟一局',
    onlineNow: 0,
    shortDescription:
      '围绕深绿牌桌和金橙油彩重做的站内斗地主体验，已经能开人机热身、建三人牌局并查看战报。',
    heroDescription:
      '这不是旁边临时挂着的实验页，而是一条已经能直接上桌的棋牌产品线。当前版本支持人机热身、三人联机、实时牌桌同步和战报回看，接下来继续打磨选牌反馈、桌面表现和正式产品质感。',
    accentFrom: '#f4b63f',
    accentTo: '#db5a3f',
    atmosphere: '涂油牌桌 / 深绿桌布 + 金橙油彩',
    loops: [
      '三位牌手凑齐后直接开局，发牌、叫分、地主归属和胜负都由服务端统一裁决。',
      '一个人也能直接开一把人机热身，快速把完整流程跑通。',
      '每局结束后会生成最近战报，方便继续回看倍率变化和关键出牌。',
      '当前版本重点在产品化重构，继续把牌桌、手牌反馈和结算体验做得更专业。',
    ],
    highlights: [
      '更贴近真实棋牌服务端场景，能完整展示状态机、裁决和断线重连。',
      '单人热身有陪练兜底，演示和联调时不必临时凑齐三个人。',
      '天然适合接战报、排行榜、后台观测和社区分享链路。',
    ],
    roadmap: [
      'Phase 1：统一产品命名、入口和页面叙事，确立“涂油斗地主”品牌感。',
      'Phase 2：拆分大组件，抽离牌型、牌面和文案映射逻辑。',
      'Phase 3：重做牌桌视觉、手牌交互和结算表现。',
      'Phase 4：继续打磨战报、大厅和后台观测能力。',
    ],
    fitForJD: [
      '能展示服务端权威裁决、房间状态机和实时同步，而不是只做单机原型。',
      '陪练模式让演示和录屏场景更稳定，不会被真人人数卡住。',
      '后续很适合继续扩成更完整的棋牌游戏中台，而不会脱离现有网站。',
    ],
  },
  {
    slug: 'lucky-current',
    title: 'Lucky Current',
    subtitle: '好运潮汐',
    status: 'prototype',
    genre: '轻捕鱼 / 路线选择',
    playerMode: '1-2 人合作原型',
    roundTime: '3 分钟',
    onlineNow: 0,
    shortDescription:
      '偏收集和倍率路线的轻捕鱼企划，重点不是炮台复杂度，而是短回合收益选择。',
    heroDescription:
      '这条线更接近“休闲捕鱼”的表达，但目前只保留在产品企划阶段，等 Hex Blitz 的实时底座稳定后再推进。',
    accentFrom: '#00d68f',
    accentTo: '#2fb7ff',
    atmosphere: '清亮海湾 / 海绿 + 浅蓝',
    loops: [
      '选择路线卡，改变当前潮流和目标鱼群。',
      '通过命中连段叠倍率，而不是追求重数值炮台。',
      '两位玩家共享增益，争取在短时间内拉高收益。 ',
    ],
    highlights: [
      '更靠近途游公开产品方向，但资源和反馈系统更重。',
      '适合作为第二款游戏，不适合在第一阶段同时推进。',
      '未来可作为房间服和事件系统的第二验证场景。',
    ],
    roadmap: [
      '先保留玩法草案。',
      '等房间与结算链路稳定后再开始立项。',
    ],
    fitForJD: [
      '能证明你不是只会做单一玩法，而是能围绕公司赛道迭代选题。',
    ],
  },
  {
    slug: 'tile-tempo',
    title: 'Tile Tempo',
    subtitle: '节拍拼拼',
    status: 'coming_soon',
    genre: '节奏拼图 / 三消变体',
    playerMode: '1 人挑战',
    roundTime: '90 秒',
    onlineNow: 0,
    shortDescription:
      '把消除和节奏点结合起来的快反玩法，适合做低成本高反馈的第二梯队项目。',
    heroDescription:
      '这是为后续内容池准备的项目，目的是让游戏中心不只挂一款作品，而是逐步形成休闲合集。',
    accentFrom: '#f95dd5',
    accentTo: '#ffa84a',
    atmosphere: '霓虹糖果 / 洋红 + 橙黄',
    loops: [
      '卡点输入，触发更高倍率。',
      '节奏错误会重置连段。',
      '适合接排行榜，不强依赖房间同步。',
    ],
    highlights: [
      '可复用大量现有 UI 资产。',
      '适合做“低资源也能出效果”的休闲游戏样板。',
    ],
    roadmap: ['待 Hex Blitz 完成多人化后评估。'],
    fitForJD: ['能体现你有产品矩阵意识，而不是只做单点 Demo。'],
  },
  {
    slug: 'paw-chess-dash',
    title: 'Paw Chess Dash',
    subtitle: '掌爪快棋',
    status: 'coming_soon',
    genre: '轻棋类 / 超短局',
    playerMode: '2 人实时',
    roundTime: '2 分钟',
    onlineNow: 0,
    shortDescription:
      '更偏棋类和策略短局的设想，保留给后续验证规则驱动型房间服务。',
    heroDescription:
      '如果后续要更贴近棋类产品线，这会是比直接上麻将/斗地主更轻的一条切入路径。',
    accentFrom: '#b785ff',
    accentTo: '#6be0a8',
    atmosphere: '策略沙盘 / 薰衣紫 + 青绿',
    loops: [
      '极简规则，快速对局。',
      '更强调房间同步和回合制状态机。',
    ],
    highlights: [
      '适合在实时消除项目稳定后，验证另一种同步模型。',
    ],
    roadmap: ['仅保留方向，不进入当前开发批次。'],
    fitForJD: ['能体现你理解不同游戏品类对服务端模型的差异。'],
  },
];

export const playableGameSlug = 'hex-blitz';

export function getGameBySlug(slug: string) {
  return gamesCatalog.find((game) => game.slug === slug);
}

export function getStatusMeta(status: GameStatus) {
  switch (status) {
    case 'playable':
      return {
        label: '可试玩',
        className:
          'border-emerald-400/30 bg-emerald-400/15 text-emerald-100',
      };
    case 'prototype':
      return {
        label: '原型阶段',
        className: 'border-sky-400/30 bg-sky-400/15 text-sky-100',
      };
    default:
      return {
        label: '筹备中',
        className: 'border-white/15 bg-white/8 text-white/80',
      };
  }
}
