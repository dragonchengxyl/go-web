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
    playerMode: '现在可单人开玩，后续会开放多人对战',
    roundTime: '75 秒一局',
    onlineNow: 18,
    shortDescription:
      '在六角棋盘上快速清除相邻同色块，叠高连击和爆裂特效，把分数一路推上去。',
    heroDescription:
      '一局够短，反馈够亮，适合随手打开冲几把分。节奏干净利落，玩起来就是盯盘面、找机会、抢高分。',
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
      '短局高反馈，随时都能开一把。',
      '规则清楚，上手很快。',
      '视觉明亮，连击感强。',
    ],
    roadmap: [
      '现在可以直接单人冲分。',
      '接下来会开放多人同场竞技。',
      '之后会补上排行榜和战报分享。',
      '长期会继续扩充更多模式和活动。',
    ],
    fitForJD: [
      '适合喜欢快节奏高反馈的玩家。',
      '适合想随时开一局、不想背复杂规则的人。',
      '适合喜欢冲高分和连击爽感的人。',
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
      '这是更偏正式牌桌感的一版斗地主。一个人能直接热身，三个人也能开桌，打完还可以回看整局战报。',
    accentFrom: '#f4b63f',
    accentTo: '#db5a3f',
    atmosphere: '涂油牌桌 / 深绿桌布 + 金橙油彩',
    loops: [
      '三位牌手凑齐后直接开局，发牌、叫分、地主归属和胜负都由牌桌统一判定。',
      '一个人也能直接开一把人机热身，快速把完整流程跑通。',
      '每局结束后会生成最近战报，方便继续回看倍率变化和关键出牌。',
      '当前版本重点在产品化重构，继续把牌桌、手牌反馈和结算体验做得更专业。',
    ],
    highlights: [
      '一个人就能马上开局。',
      '三人牌桌节奏完整，叫分和出牌都更有临场感。',
      '每局结束后都能回看战报。',
    ],
    roadmap: [
      '现在可以先开人机热身或三人牌局。',
      '接下来会继续打磨牌桌视觉和手牌体验。',
      '之后会让战报和大厅更完整。',
      '长期会继续补更多牌桌反馈和玩法细节。',
    ],
    fitForJD: [
      '适合喜欢传统牌桌节奏的玩家。',
      '适合想先单人熟悉一局再拉朋友开桌的人。',
      '适合喜欢打完以后回看整局过程的人。',
    ],
  },
  {
    slug: 'lucky-current',
    title: 'Lucky Current',
    subtitle: '好运潮汐',
    status: 'prototype',
    genre: '轻捕鱼 / 路线选择',
    playerMode: '1-2 人轻合作',
    roundTime: '3 分钟',
    onlineNow: 0,
    shortDescription:
      '偏收集和倍率路线的轻捕鱼企划，重点不是炮台复杂度，而是短回合收益选择。',
    heroDescription:
      '这是一条更轻、更清亮的海湾路线，主打短回合里做选择、叠倍率、抓节奏。',
    accentFrom: '#00d68f',
    accentTo: '#2fb7ff',
    atmosphere: '清亮海湾 / 海绿 + 浅蓝',
    loops: [
      '选择路线卡，改变当前潮流和目标鱼群。',
      '通过命中连段叠倍率，而不是追求重数值炮台。',
      '两位玩家共享增益，争取在短时间内拉高收益。 ',
    ],
    highlights: [
      '更偏轻松收集和路线选择。',
      '节奏比重火力玩法更轻。',
      '适合喜欢海湾氛围和短回合收益感的人。',
    ],
    roadmap: [
      '会在后续内容更新中考虑加入。',
      '当前先保留为待开放项目。',
    ],
    fitForJD: [
      '适合喜欢清亮海湾风格和倍率路线的人。',
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
      '节奏和拼图结合，适合轻松上手。',
      '色彩明亮，反馈直接。',
    ],
    roadmap: ['后续会根据内容更新安排开放。'],
    fitForJD: ['适合喜欢卡点和拼图混合玩法的人。'],
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
      '更偏棋类和策略短局的设想，主打短时间里快速判断和落子。',
    heroDescription:
      '这条线会更安静、更讲究取舍，适合喜欢短局策略感的人。',
    accentFrom: '#b785ff',
    accentTo: '#6be0a8',
    atmosphere: '策略沙盘 / 薰衣紫 + 青绿',
    loops: [
      '极简规则，快速对局。',
      '每一步都更看重判断和时机。',
    ],
    highlights: [
      '更偏轻策略和短局博弈。',
    ],
    roadmap: ['后续会根据整体内容节奏安排。'],
    fitForJD: ['适合喜欢轻棋类和快节奏对局的人。'],
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
        label: '现已可玩',
        className:
          'border-emerald-400/30 bg-emerald-400/15 text-emerald-100',
      };
    case 'prototype':
      return {
        label: '抢先体验',
        className: 'border-sky-400/30 bg-sky-400/15 text-sky-100',
      };
    default:
      return {
        label: '筹备中',
        className: 'border-white/15 bg-white/8 text-white/80',
      };
  }
}
