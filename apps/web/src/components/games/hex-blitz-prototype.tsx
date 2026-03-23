'use client';

import { useEffect, useRef, useState } from 'react';
import {
  Flame,
  Play,
  RotateCcw,
  Timer,
  Trophy,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { HexBlitzBoardState, HexBlitzTile } from '@/lib/api-client';
import { HexBlitzBoard } from '@/components/games/hex-blitz-board';

type Tile = HexBlitzTile;

const BOARD_RADIUS = 2;
const GAME_DURATION_MS = 75_000;
const COMBO_WINDOW_MS = 2_500;
const DIRECTIONS: Array<[number, number]> = [
  [1, 0],
  [1, -1],
  [0, -1],
  [-1, 0],
  [-1, 1],
  [0, 1],
];

const SHOWCASE_PATTERN: Array<[Tile['color'], Tile['special']]> = [
  ['ember', 'none'],
  ['ember', 'spark'],
  ['lagoon', 'none'],
  ['mint', 'none'],
  ['lagoon', 'burst'],
  ['sun', 'none'],
  ['violet', 'none'],
  ['sun', 'spark'],
  ['lagoon', 'none'],
  ['mint', 'none'],
  ['ember', 'burst'],
  ['violet', 'none'],
  ['sun', 'none'],
  ['mint', 'none'],
  ['violet', 'spark'],
  ['lagoon', 'none'],
  ['ember', 'none'],
  ['sun', 'none'],
  ['mint', 'none'],
];

function coordKey(q: number, r: number) {
  return `${q}:${r}`;
}

function buildCoords() {
  const coords: Array<{ q: number; r: number }> = [];

  for (let q = -BOARD_RADIUS; q <= BOARD_RADIUS; q += 1) {
    const rMin = Math.max(-BOARD_RADIUS, -q - BOARD_RADIUS);
    const rMax = Math.min(BOARD_RADIUS, -q + BOARD_RADIUS);
    for (let r = rMin; r <= rMax; r += 1) {
      coords.push({ q, r });
    }
  }

  return coords.sort((a, b) => (a.r === b.r ? a.q - b.q : a.r - b.r));
}

const BOARD_COORDS = buildCoords();

function createShowcaseBoard() {
  return BOARD_COORDS.map((coord, index) => {
    const [color, special] = SHOWCASE_PATTERN[index % SHOWCASE_PATTERN.length];

    return {
      id: coordKey(coord.q, coord.r),
      q: coord.q,
      r: coord.r,
      color,
      special,
    } satisfies Tile;
  });
}

function randomColor(): Tile['color'] {
  const colors: Tile['color'][] = ['ember', 'lagoon', 'mint', 'sun', 'violet'];
  return colors[Math.floor(Math.random() * colors.length)];
}

function randomSpecial(): Tile['special'] {
  const roll = Math.random();
  if (roll < 0.08) {
    return 'burst';
  }
  if (roll < 0.2) {
    return 'spark';
  }
  return 'none';
}

function createRandomBoard() {
  return ensurePlayableBoard(
    BOARD_COORDS.map((coord) => ({
      id: coordKey(coord.q, coord.r),
      q: coord.q,
      r: coord.r,
      color: randomColor(),
      special: randomSpecial(),
    }))
  );
}

function tileMapFrom(tiles: Tile[]) {
  const map = new Map<string, Tile>();
  for (const tile of tiles) {
    map.set(tile.id, tile);
  }
  return map;
}

function collectGroup(tiles: Tile[], startId: string | null) {
  if (!startId) {
    return [];
  }

  const tileMap = tileMapFrom(tiles);
  const start = tileMap.get(startId);

  if (!start) {
    return [];
  }

  const queue = [start];
  const visited = new Set<string>([start.id]);
  const group: Tile[] = [];

  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) {
      continue;
    }

    group.push(current);

    for (const [dq, dr] of DIRECTIONS) {
      const neighbor = tileMap.get(coordKey(current.q + dq, current.r + dr));
      if (!neighbor || visited.has(neighbor.id)) {
        continue;
      }
      if (neighbor.color !== start.color) {
        continue;
      }
      visited.add(neighbor.id);
      queue.push(neighbor);
    }
  }

  return group;
}

function hasPlayableMove(tiles: Tile[]) {
  return tiles.some((tile) => collectGroup(tiles, tile.id).length >= 2);
}

function ensurePlayableBoard(tiles: Tile[]) {
  let next = tiles;
  let attempts = 0;

  while (!hasPlayableMove(next) && attempts < 8) {
    next = next.map((tile) => ({
      ...tile,
      color: randomColor(),
      special: randomSpecial(),
    }));
    attempts += 1;
  }

  return next;
}

function expandWithBurstTiles(group: Tile[], tiles: Tile[]) {
  const tileMap = tileMapFrom(tiles);
  const cleared = new Set(group.map((tile) => tile.id));

  for (const tile of group) {
    if (tile.special !== 'burst') {
      continue;
    }
    for (const [dq, dr] of DIRECTIONS) {
      const neighbor = tileMap.get(coordKey(tile.q + dq, tile.r + dr));
      if (neighbor) {
        cleared.add(neighbor.id);
      }
    }
  }

  return cleared;
}

function replaceClearedTiles(tiles: Tile[], cleared: Set<string>) {
  const next = tiles.map((tile) => {
    if (!cleared.has(tile.id)) {
      return tile;
    }
    return {
      ...tile,
      color: randomColor(),
      special: randomSpecial(),
    };
  });

  return ensurePlayableBoard(next);
}

function formatTime(ms: number) {
  return (ms / 1000).toFixed(1);
}

function getTier(score: number) {
  if (score >= 4200) {
    return '钻石冲刺';
  }
  if (score >= 3400) {
    return '黄金节奏';
  }
  if (score >= 2600) {
    return '白银连段';
  }
  if (score >= 1800) {
    return '青铜起步';
  }
  return '热身中';
}

interface HexBlitzRoomMode {
  enabled: boolean;
  phase: 'waiting' | 'countdown' | 'running' | 'finished';
  matchKey?: string;
  endsAt?: string;
  infoText?: string;
}

interface HexBlitzPrototypeProps {
  roomMode?: HexBlitzRoomMode;
  boardState?: HexBlitzBoardState | null;
  onBoardMove?: (tileId: string) => void;
}

export function HexBlitzPrototype({
  roomMode,
  boardState,
  onBoardMove,
}: HexBlitzPrototypeProps = {}) {
  const [tiles, setTiles] = useState<Tile[]>(() => createShowcaseBoard());
  const [phase, setPhase] = useState<'idle' | 'running' | 'ended'>('idle');
  const [score, setScore] = useState(0);
  const [bestScore, setBestScore] = useState(0);
  const [combo, setCombo] = useState(0);
  const [bestCombo, setBestCombo] = useState(0);
  const [moves, setMoves] = useState(0);
  const [timeLeftMs, setTimeLeftMs] = useState(GAME_DURATION_MS);
  const [hoveredTileId, setHoveredTileId] = useState<string | null>(null);
  const [message, setMessage] = useState('点击相邻同色六角块，尽量在 75 秒内把连击叠起来。');
  const [chainExpiresAt, setChainExpiresAt] = useState<number | null>(null);
  const [runSeed, setRunSeed] = useState(0);
  const lastRoomMatchKeyRef = useRef<string | null>(null);

  const isRoomMode = !!roomMode?.enabled;
  const roomPhase = roomMode?.phase;
  const roomMatchKey = roomMode?.matchKey;
  const roomInfoText = roomMode?.infoText;
  const remoteEndAtMs = roomMode?.endsAt ? new Date(roomMode.endsAt).getTime() : null;

  const previewGroup = collectGroup(tiles, hoveredTileId);
  const previewSet = new Set(previewGroup.map((tile) => tile.id));

  function resetToShowcase(nextMessage: string) {
    setTiles(createShowcaseBoard());
    setPhase('idle');
    setScore(0);
    setCombo(0);
    setMoves(0);
    setTimeLeftMs(GAME_DURATION_MS);
    setHoveredTileId(null);
    setChainExpiresAt(null);
    setMessage(nextMessage);
  }

  useEffect(() => {
    if (phase !== 'running') {
      return;
    }

    const fallbackEndAt = Date.now() + GAME_DURATION_MS;
    const timer = window.setInterval(() => {
      const endAt = isRoomMode && remoteEndAtMs ? remoteEndAtMs : fallbackEndAt;
      const remaining = Math.max(0, endAt - Date.now());
      setTimeLeftMs(remaining);

      if (remaining === 0) {
        window.clearInterval(timer);
        setPhase('ended');
      }
    }, 100);

    return () => {
      window.clearInterval(timer);
    };
  }, [phase, runSeed, isRoomMode, remoteEndAtMs]);

  useEffect(() => {
    if (!isRoomMode || !roomPhase) {
      lastRoomMatchKeyRef.current = null;
      return;
    }

    if (roomPhase === 'running' && roomMatchKey) {
      if (lastRoomMatchKeyRef.current === roomMatchKey) {
        return;
      }
      lastRoomMatchKeyRef.current = roomMatchKey;
      setPhase('running');
      setScore(0);
      setBestCombo(0);
      setCombo(0);
      setMoves(0);
      setHoveredTileId(null);
      setChainExpiresAt(null);
      setTimeLeftMs(remoteEndAtMs ? Math.max(0, remoteEndAtMs - Date.now()) : GAME_DURATION_MS);
      setMessage(roomInfoText || '房间对局已开始，正在等待服务端棋盘状态。');
      setRunSeed((current) => current + 1);
      return;
    }

    if (roomPhase === 'countdown') {
      resetToShowcase(roomInfoText || '房主已经开始倒计时，棋盘将在开局时自动激活。');
      return;
    }

    if (roomPhase === 'waiting') {
      lastRoomMatchKeyRef.current = null;
      resetToShowcase(roomInfoText || '你已连接房间。等所有在线玩家准备完毕后，由房主开始。');
      return;
    }

    if (roomPhase === 'finished') {
      setPhase('ended');
      setTimeLeftMs(0);
      setChainExpiresAt(null);
      setHoveredTileId(null);
      setMessage(roomInfoText || `房间对局已结束，当前分数 ${score}。`);
    }
  }, [isRoomMode, remoteEndAtMs, roomInfoText, roomMatchKey, roomPhase, score]);

  useEffect(() => {
    if (!isRoomMode || !boardState) {
      return;
    }

    setTiles(
      boardState.tiles.map((tile) => ({
        id: tile.id,
        q: tile.q,
        r: tile.r,
        color: tile.color,
        special: tile.special,
      }))
    );
    setScore(boardState.score);
    setCombo(boardState.combo);
    setBestCombo(boardState.best_combo);
    setMoves(boardState.moves);
    setMessage(boardState.message || roomInfoText || '服务端棋盘已同步。');
    if (boardState.phase === 'running') {
      setPhase('running');
    } else if (boardState.phase === 'finished') {
      setPhase('ended');
    }
  }, [boardState, isRoomMode, roomInfoText]);

  useEffect(() => {
    if (!chainExpiresAt || phase !== 'running') {
      return;
    }

    const delay = chainExpiresAt - Date.now();

    if (delay <= 0) {
      setCombo(0);
      setChainExpiresAt(null);
      return;
    }

    const timeout = window.setTimeout(() => {
      setCombo(0);
      setChainExpiresAt(null);
    }, delay);

    return () => window.clearTimeout(timeout);
  }, [chainExpiresAt, phase]);

  useEffect(() => {
    if (phase !== 'ended') {
      return;
    }

    setBestScore((current) => Math.max(current, score));
    setHoveredTileId(null);
    setChainExpiresAt(null);
    setCombo(0);
    setMessage(() => {
      if (isRoomMode) {
        return roomInfoText || `房间对局已结束，当前分数 ${score}。`;
      }
      return `本局结束，最终得分 ${score}。下一阶段会把这个循环接成多人房间和排行榜。`;
    });
  }, [isRoomMode, phase, roomInfoText, score]);

  function startGame() {
    if (isRoomMode) {
      return;
    }
    setTiles(createRandomBoard());
    setPhase('running');
    setScore(0);
    setBestCombo(0);
    setCombo(0);
    setMoves(0);
    setTimeLeftMs(GAME_DURATION_MS);
    setHoveredTileId(null);
    setChainExpiresAt(null);
    setMessage('开局了。优先找大团块，爆裂块适合在中盘拉高收益。');
    setRunSeed((current) => current + 1);
  }

  function handleTileClick(tileId: string) {
    if (isRoomMode) {
      if (phase === 'running' && onBoardMove) {
        onBoardMove(tileId);
      }
      return;
    }
    if (phase !== 'running') {
      return;
    }

    const group = collectGroup(tiles, tileId);
    if (group.length < 2) {
      setMessage('至少连接 2 个同色块才能清除。尽量把爆裂块留在大团里。');
      return;
    }

    const now = Date.now();
    const nextCombo = chainExpiresAt && chainExpiresAt > now ? combo + 1 : 1;
    const clearedIds = expandWithBurstTiles(group, tiles);
    const clearedTiles = tiles.filter((tile) => clearedIds.has(tile.id));
    const sparkCount = clearedTiles.filter((tile) => tile.special === 'spark').length;
    const burstCount = group.filter((tile) => tile.special === 'burst').length;
    const base = clearedTiles.length * clearedTiles.length * 12;
    const comboBonus = Math.round(base * Math.max(0, nextCombo - 1) * 0.35);
    const specialBonus = sparkCount * 70 + burstCount * 40;
    const gainedScore = base + comboBonus + specialBonus;

    setTiles(replaceClearedTiles(tiles, clearedIds));
    setScore((current) => current + gainedScore);
    setCombo(nextCombo);
    setBestCombo((current) => Math.max(current, nextCombo));
    setMoves((current) => current + 1);
    setChainExpiresAt(now + COMBO_WINDOW_MS);
    setHoveredTileId(null);
    setMessage(
      `清掉 ${clearedTiles.length} 格，拿到 ${gainedScore} 分。连击 x${nextCombo}${sparkCount > 0 ? `，星辉 +${sparkCount}` : ''}${burstCount > 0 ? `，爆裂 ${burstCount} 次` : ''}。`
    );
  }

  const tier = getTier(score);
  const secondsLeft = formatTime(timeLeftMs);
  const progress = Math.max(0, Math.min(100, (timeLeftMs / GAME_DURATION_MS) * 100));

  return (
    <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
      <Card className="overflow-hidden border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.03))] text-white">
        <CardContent className="p-0">
          <div className="border-b border-white/10 px-6 py-5">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.28em] text-slate-500">
                  Single Player Prototype
                </p>
                <h2 className="mt-2 text-3xl font-black tracking-tight text-white">
                  Hex Blitz Board
                </h2>
              </div>
              <div className="flex flex-wrap gap-2">
                <Badge className="border-white/15 bg-white/8 text-white">
                  当前段位：{tier}
                </Badge>
                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                  {phase === 'running' ? '进行中' : phase === 'ended' ? '已结算' : '待开始'}
                </Badge>
              </div>
            </div>
          </div>

          <div className="grid gap-6 p-6 lg:grid-cols-[0.95fr_1.05fr]">
            <div className="rounded-[28px] border border-white/10 bg-[radial-gradient(circle_at_top,rgba(52,210,255,0.14),transparent_28%),linear-gradient(180deg,#101923_0%,#0a1018_100%)] p-5">
              <div className="mb-4 flex items-center justify-between">
                <div>
                  <div className="text-sm text-slate-400">剩余时间</div>
                  <div className="mt-1 flex items-end gap-2">
                    <span className="text-4xl font-black tracking-tight text-white">
                      {secondsLeft}
                    </span>
                    <span className="pb-1 text-sm text-slate-400">秒</span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm text-slate-400">当前分数</div>
                  <div className="mt-1 text-3xl font-black tracking-tight text-white">
                    {score}
                  </div>
                </div>
              </div>

              <div className="h-2 rounded-full bg-white/8">
                <div
                  className="h-full rounded-full bg-[linear-gradient(90deg,#ffb547_0%,#35d4ff_100%)] transition-[width] duration-150"
                  style={{ width: `${progress}%` }}
                />
              </div>

              <HexBlitzBoard
                className="mt-6"
                tiles={tiles}
                interactive={phase === 'running'}
                highlightedTileIds={previewGroup.length >= 2 ? previewSet : undefined}
                onHoverChange={setHoveredTileId}
                onTileClick={handleTileClick}
              />
            </div>

            <div className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <Card className="border-white/10 bg-black/20 text-white">
                  <CardContent className="p-5">
                    <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                      <Flame className="h-4 w-4" />
                      当前连击
                    </div>
                    <div className="text-3xl font-black tracking-tight">{combo}</div>
                    <div className="mt-2 text-sm text-slate-400">
                      2.5 秒内继续清除就能叠上去
                    </div>
                  </CardContent>
                </Card>
                <Card className="border-white/10 bg-black/20 text-white">
                  <CardContent className="p-5">
                    <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                      <Trophy className="h-4 w-4" />
                      最佳记录
                    </div>
                    <div className="text-3xl font-black tracking-tight">{bestScore}</div>
                    <div className="mt-2 text-sm text-slate-400">
                      历史最佳连击 x{bestCombo}
                    </div>
                  </CardContent>
                </Card>
              </div>

              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-3 flex items-center gap-2 text-sm text-slate-400">
                    <Timer className="h-4 w-4" />
                    当前局势
                  </div>
                  <p className="text-sm leading-7 text-slate-300">{message}</p>
                  <div className="mt-4 grid gap-3 sm:grid-cols-3">
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        Moves
                      </div>
                      <div className="mt-2 text-2xl font-black tracking-tight">{moves}</div>
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        Tier
                      </div>
                      <div className="mt-2 text-lg font-semibold text-white">{tier}</div>
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        Target
                      </div>
                      <div className="mt-2 text-lg font-semibold text-white">4200+</div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  {isRoomMode ? (
                    <div className="space-y-3">
                      <div className="rounded-2xl border border-sky-300/15 bg-sky-300/10 px-4 py-3 text-sm leading-7 text-sky-50">
                        当前正在房间模式下运行。棋盘会在房主开局后自动启动，分数也会实时推送到房间。
                      </div>
                      <div className="flex flex-wrap gap-3">
                        <Button
                          disabled
                          className="border-0 bg-[linear-gradient(135deg,#34d2ff_0%,#7af6b5_100%)] text-slate-950 opacity-70"
                        >
                          <Play className="mr-2 h-4 w-4" />
                          等待房间控制
                        </Button>
                        <Button
                          disabled
                          variant="outline"
                          className="border-white/15 bg-transparent text-white opacity-70"
                        >
                          <RotateCcw className="mr-2 h-4 w-4" />
                          房间状态驱动
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-3">
                      <Button
                        onClick={startGame}
                        className="border-0 bg-[linear-gradient(135deg,#ff8a3d_0%,#34d2ff_100%)] text-slate-950 hover:brightness-110"
                      >
                        <Play className="mr-2 h-4 w-4" />
                        {phase === 'running' ? '重新开局' : '开始一局'}
                      </Button>
                      <Button
                        onClick={() => {
                          resetToShowcase('已回到展示盘面。准备好后再开始一局。');
                        }}
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                      >
                        <RotateCcw className="mr-2 h-4 w-4" />
                        重置展示盘面
                      </Button>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-4 text-sm font-medium text-white/85">规则速览</div>
                  <div className="space-y-3 text-sm leading-6 text-slate-300">
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      点击相邻同色六角块，2 格起消。团块越大，基础得分越高。
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      <span className="font-medium text-white">星辉块</span>
                      ：清掉时额外加分，适合放进大团块里一起结算。
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      <span className="font-medium text-white">爆裂块</span>
                      ：如果它在被清除的团块里，会顺带炸掉周围六格。
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
                      连击窗口只有 2.5 秒，适合先扫局部大团，再补中型团维持节奏。
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-3 text-xs uppercase tracking-[0.28em] text-slate-500">
              Phase 2
            </div>
            <h3 className="text-xl font-semibold">房间和同步</h3>
            <p className="mt-3 text-sm leading-7 text-slate-300">
              下一阶段会把这个得分循环接成 2-4 人房间模式，服务端统一发题、统一计时、统一结算。
            </p>
          </CardContent>
        </Card>
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-3 text-xs uppercase tracking-[0.28em] text-slate-500">
              Phase 3
            </div>
            <h3 className="text-xl font-semibold">排行榜和战报</h3>
            <p className="mt-3 text-sm leading-7 text-slate-300">
              结算结果会落 PostgreSQL，生成日榜、周榜，并可以一键分享本局战绩到社区帖子。
            </p>
          </CardContent>
        </Card>
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-3 text-xs uppercase tracking-[0.28em] text-slate-500">
              Phase 4
            </div>
            <h3 className="text-xl font-semibold">面试可讲的工程化</h3>
            <p className="mt-3 text-sm leading-7 text-slate-300">
              补日志、指标、压测和断线恢复后，这条线就可以完整覆盖“设计、开发、优化、稳定性”。
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
