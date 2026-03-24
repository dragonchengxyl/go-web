"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Bot,
  Crown,
  DoorOpen,
  Loader2,
  Radio,
  RefreshCcw,
  Sparkles,
  Swords,
  Trophy,
  Users2,
} from "lucide-react";
import { useAuth } from "@/contexts/auth-context";
import {
  apiClient,
  DoudizhuActionResult,
  DoudizhuCard,
  DoudizhuCombo,
  DoudizhuLeaderboardEntry,
  DoudizhuMatchSummary,
  DoudizhuPrivateState,
  DoudizhuRoom,
  DoudizhuRoomPlayer,
} from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

type RoomWSStatus = "idle" | "connecting" | "connected" | "closed";

interface RoomServerMessage {
  type: string;
  payload: any;
}

interface RoomHintResult {
  action_type: "bid" | "play_cards" | "pass_turn";
  bid_score?: number;
  cards?: DoudizhuCard[];
  combo?: DoudizhuCombo;
  message?: string;
}

function getGameWsBase() {
  if (typeof window === "undefined") {
    return "ws://localhost:8080";
  }

  const configured = process.env.NEXT_PUBLIC_WS_URL;
  if (configured) {
    return configured;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}`;
}

function createLocalSessionID() {
  if (typeof window === "undefined") {
    return "server-session";
  }
  if (window.crypto?.randomUUID) {
    return window.crypto.randomUUID();
  }
  return `session-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`;
}

function roomStatusLabel(status?: string) {
  switch (status) {
    case "bidding":
      return "叫分中";
    case "playing":
      return "对局中";
    case "settlement":
      return "已结算";
    case "redeal":
      return "流局重发";
    default:
      return "待准备";
  }
}

function roomModeLabel(mode?: string) {
  return mode === "demo_ai" ? "AI 演示" : "真人房";
}

function seatLabel(seat?: number) {
  switch (seat) {
    case 0:
      return "一号位";
    case 1:
      return "二号位";
    case 2:
      return "三号位";
    default:
      return "--";
  }
}

function formatRemaining(target?: string, nowMs: number = Date.now()) {
  if (!target) {
    return "--";
  }
  const targetMs = new Date(target).getTime();
  const remaining = Math.max(0, targetMs - nowMs);
  return `${Math.ceil(remaining / 1000)}s`;
}

function cardKey(card: DoudizhuCard) {
  return `${card.suit}-${card.rank}`;
}

function cardSuit(card: DoudizhuCard) {
  switch (card.suit) {
    case "spade":
      return "♠";
    case "heart":
      return "♥";
    case "club":
      return "♣";
    case "diamond":
      return "♦";
    default:
      return "J";
  }
}

function cardRank(card: DoudizhuCard) {
  const rankMap: Record<number, string> = {
    11: "J",
    12: "Q",
    13: "K",
    14: "A",
    15: "2",
    16: "SJ",
    17: "BJ",
  };
  return rankMap[card.rank] ?? String(card.rank);
}

function cardLabel(card: DoudizhuCard) {
  return `${cardSuit(card)}${cardRank(card)}`;
}

function comboLabel(combo?: DoudizhuCombo | null) {
  if (!combo) {
    return "等待首家出牌";
  }

  const typeLabelMap: Record<DoudizhuCombo["type"], string> = {
    single: "单张",
    pair: "对子",
    triple: "三张",
    triple_with_single: "三带一",
    triple_with_pair: "三带二",
    straight: "顺子",
    straight_pairs: "连对",
    airplane: "飞机",
    airplane_with_single: "飞机带单",
    airplane_with_pair: "飞机带对",
    four_with_two_single: "四带二",
    four_with_two_pair: "四带两对",
    bomb: "炸弹",
    rocket: "王炸",
  };
  return `${typeLabelMap[combo.type]} · 主值 ${combo.main_rank}`;
}

function actionTypeLabel(actionType?: string) {
  switch (actionType) {
    case "bid":
      return "叫分";
    case "auto_bid":
      return "托管叫分";
    case "play_cards":
      return "出牌";
    case "auto_play_cards":
      return "托管出牌";
    case "pass_turn":
      return "过牌";
    case "auto_pass_turn":
      return "托管过牌";
    case "timeout_auto_play":
      return "超时托管";
    case "landlord_assigned":
      return "地主确定";
    case "settlement":
      return "本局结算";
    default:
      return actionType ?? "操作";
  }
}

function sortedSelection(cards: DoudizhuCard[]) {
  return [...cards].sort((a, b) => {
    if (a.rank === b.rank) {
      return a.suit.localeCompare(b.suit);
    }
    return a.rank - b.rank;
  });
}

function rankCounts(cards: DoudizhuCard[]) {
  const counts = new Map<number, number>();
  for (const card of cards) {
    counts.set(card.rank, (counts.get(card.rank) ?? 0) + 1);
  }
  return counts;
}

function sortedRanks(counts: Map<number, number>) {
  return [...counts.keys()].sort((a, b) => a - b);
}

function rankWithCount(counts: Map<number, number>, expected: number) {
  for (const [rank, count] of counts.entries()) {
    if (count === expected) {
      return rank;
    }
  }
  return null;
}

function hasCount(counts: Map<number, number>, expected: number) {
  return [...counts.values()].some((count) => count === expected);
}

function countPattern(counts: Map<number, number>, pattern: number[]) {
  const found = [...counts.values()].sort((a, b) => a - b);
  const expected = [...pattern].sort((a, b) => a - b);
  return (
    found.length === expected.length &&
    found.every((value, index) => value === expected[index])
  );
}

function areConsecutive(ranks: number[]) {
  for (let index = 1; index < ranks.length; index += 1) {
    if (ranks[index] !== ranks[index - 1] + 1) {
      return false;
    }
  }
  return true;
}

function isStraight(counts: Map<number, number>) {
  if (counts.size < 5) {
    return false;
  }
  const ranks = sortedRanks(counts);
  return ranks.every((rank, index) => {
    if (rank > 14 || counts.get(rank) !== 1) {
      return false;
    }
    return index === 0 || rank === ranks[index - 1] + 1;
  });
}

function isStraightPairs(counts: Map<number, number>) {
  if (counts.size < 3) {
    return false;
  }
  const ranks = sortedRanks(counts);
  return ranks.every((rank, index) => {
    if (rank > 14 || counts.get(rank) !== 2) {
      return false;
    }
    return index === 0 || rank === ranks[index - 1] + 1;
  });
}

function remainingPattern(
  counts: Map<number, number>,
  expectedCount: number,
  expectedKinds: number,
) {
  let kinds = 0;
  for (const count of counts.values()) {
    if (count === 0) {
      continue;
    }
    if (count !== expectedCount) {
      return false;
    }
    kinds += 1;
  }
  return kinds === expectedKinds;
}

function airplaneCombo(
  counts: Map<number, number>,
  segment: number[],
  total: number,
): DoudizhuCombo | null {
  const remaining = new Map(counts);
  for (const rank of segment) {
    remaining.set(rank, (remaining.get(rank) ?? 0) - 3);
  }
  const runLength = segment.length;
  const remainingCards = total - runLength * 3;

  let type: DoudizhuCombo["type"] | null = null;
  if (remainingCards === 0) {
    type = "airplane";
  } else if (
    remainingCards === runLength &&
    remainingPattern(remaining, 1, runLength)
  ) {
    type = "airplane_with_single";
  } else if (
    remainingCards === runLength * 2 &&
    remainingPattern(remaining, 2, runLength)
  ) {
    type = "airplane_with_pair";
  }

  if (!type) {
    return null;
  }
  return {
    type,
    main_rank: segment[segment.length - 1],
    sequence_length: runLength,
    total_cards: total,
  };
}

function findAirplane(counts: Map<number, number>, total: number) {
  const triples = [...counts.entries()]
    .filter(([rank, count]) => count >= 3 && rank <= 14)
    .map(([rank]) => rank)
    .sort((a, b) => a - b);

  for (let runLength = triples.length; runLength >= 2; runLength -= 1) {
    for (let start = 0; start + runLength <= triples.length; start += 1) {
      const segment = triples.slice(start, start + runLength);
      if (!areConsecutive(segment)) {
        continue;
      }
      const combo = airplaneCombo(counts, segment, total);
      if (combo) {
        return combo;
      }
    }
  }
  return null;
}

function evaluateSelectedCombo(cards: DoudizhuCard[]): {
  combo: DoudizhuCombo | null;
  error: string;
} {
  if (cards.length === 0) {
    return { combo: null as DoudizhuCombo | null, error: "" };
  }

  const sorted = sortedSelection(cards);
  const counts = rankCounts(sorted);
  const ranks = sortedRanks(counts);
  const total = sorted.length;

  if (total === 1) {
    return {
      combo: {
        type: "single",
        main_rank: sorted[0].rank,
        sequence_length: 1,
        total_cards: total,
      },
      error: "",
    };
  }
  if (total === 2) {
    if (sorted[0].rank === 16 && sorted[1].rank === 17) {
      return {
        combo: {
          type: "rocket",
          main_rank: 17,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
    if (ranks.length === 1) {
      return {
        combo: {
          type: "pair",
          main_rank: ranks[0],
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 3 && ranks.length === 1) {
    return {
      combo: {
        type: "triple",
        main_rank: ranks[0],
        sequence_length: 1,
        total_cards: total,
      },
      error: "",
    };
  }
  if (total === 4) {
    if (ranks.length === 1) {
      return {
        combo: {
          type: "bomb",
          main_rank: ranks[0],
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
    const tripleRank = rankWithCount(counts, 3);
    if (tripleRank) {
      return {
        combo: {
          type: "triple_with_single",
          main_rank: tripleRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 5) {
    const tripleRank = rankWithCount(counts, 3);
    if (tripleRank && hasCount(counts, 2)) {
      return {
        combo: {
          type: "triple_with_pair",
          main_rank: tripleRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (isStraight(counts)) {
    return {
      combo: {
        type: "straight",
        main_rank: ranks[ranks.length - 1],
        sequence_length: ranks.length,
        total_cards: total,
      },
      error: "",
    };
  }
  if (isStraightPairs(counts)) {
    return {
      combo: {
        type: "straight_pairs",
        main_rank: ranks[ranks.length - 1],
        sequence_length: ranks.length,
        total_cards: total,
      },
      error: "",
    };
  }
  const airplane = findAirplane(counts, total);
  if (airplane) {
    return { combo: airplane, error: "" };
  }
  if (total === 6) {
    const fourRank = rankWithCount(counts, 4);
    if (fourRank && countPattern(counts, [4, 1, 1])) {
      return {
        combo: {
          type: "four_with_two_single",
          main_rank: fourRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 8) {
    const fourRank = rankWithCount(counts, 4);
    if (fourRank && countPattern(counts, [4, 2, 2])) {
      return {
        combo: {
          type: "four_with_two_pair",
          main_rank: fourRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  return {
    combo: null as DoudizhuCombo | null,
    error: "当前选择不是可出的合法牌型",
  };
}

function playerSort(a: DoudizhuRoomPlayer, b: DoudizhuRoomPlayer) {
  return a.seat - b.seat;
}

function buildBoardSeats(
  activeRoom: DoudizhuRoom | null,
  me: DoudizhuRoomPlayer | null,
) {
  if (!activeRoom) {
    return {
      bottom: null as DoudizhuRoomPlayer | null,
      left: null as DoudizhuRoomPlayer | null,
      right: null as DoudizhuRoomPlayer | null,
    };
  }

  const players = [...activeRoom.players].sort(playerSort);
  const bottom = me ?? players[0] ?? null;
  const others = players.filter((player) =>
    bottom ? player.seat !== bottom.seat : true,
  );

  return {
    bottom,
    left: others[0] ?? null,
    right: others[1] ?? null,
  };
}

function TablePlayerSeat({
  player,
  position,
  isCurrentTurn,
  isCurrentBidder,
  isLandlord,
  isMe,
}: {
  player: DoudizhuRoomPlayer | null;
  position: "left" | "right" | "bottom";
  isCurrentTurn: boolean;
  isCurrentBidder: boolean;
  isLandlord: boolean;
  isMe: boolean;
}) {
  if (!player) {
    return null;
  }

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-[30px] border px-4 py-4 backdrop-blur-md",
        isMe
          ? "border-amber-300/35 bg-[linear-gradient(180deg,rgba(255,203,120,0.16),rgba(255,203,120,0.06))]"
          : "border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.09),rgba(255,255,255,0.03))]",
        position === "bottom"
          ? "shadow-[0_28px_90px_-34px_rgba(0,0,0,0.65)]"
          : "",
      )}
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.08),transparent_35%)]" />
      <div className="relative flex items-start gap-3">
        <div
          className={cn(
            "flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border text-lg font-black",
            isLandlord
              ? "border-red-300/30 bg-red-400/15 text-red-50"
              : "border-white/10 bg-black/25 text-white",
          )}
        >
          {isLandlord ? <Crown className="h-5 w-5" /> : player.name.slice(0, 1)}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="truncate text-lg font-semibold text-white">
              {player.name}
            </span>
            <Badge className="border-white/15 bg-black/25 text-white">
              {seatLabel(player.seat)}
            </Badge>
            {player.is_host && (
              <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                房主
              </Badge>
            )}
            {player.is_bot && (
              <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                机器人
              </Badge>
            )}
            {isLandlord && (
              <Badge className="border-red-400/20 bg-red-400/10 text-red-100">
                地主
              </Badge>
            )}
            {(isCurrentTurn || isCurrentBidder) && (
              <Badge className="border-emerald-300/20 bg-emerald-300/10 text-emerald-100">
                <span className="mr-1 inline-block h-2 w-2 rounded-full bg-emerald-300" />
                {isCurrentTurn ? "行动中" : "叫分中"}
              </Badge>
            )}
          </div>

          <div className="mt-2 flex flex-wrap items-center justify-between gap-3 text-xs text-slate-300">
            <div>手牌 {player.card_count} 张</div>
            <div>{player.connected ? "在线" : "离线"}</div>
            <div>{player.auto_play ? "托管中" : "手动操作"}</div>
          </div>
        </div>
      </div>

      <div className="relative mt-4 flex flex-wrap gap-2">
        {Array.from({
          length: Math.min(player.card_count, position === "bottom" ? 10 : 8),
        }).map((_, index) => (
          <div
            key={`${player.session_id}-${index}`}
            className={cn(
              "rounded-xl border px-2 py-1 text-xs font-medium shadow-[0_14px_28px_-20px_rgba(0,0,0,0.8)]",
              position === "bottom"
                ? "border-white/10 bg-black/35 text-slate-300"
                : "border-white/10 bg-white/8 text-slate-300",
            )}
          >
            牌
          </div>
        ))}
      </div>
    </div>
  );
}

function PlayCard({
  card,
  selected,
  invalid,
  pulse,
  onClick,
}: {
  card: DoudizhuCard;
  selected: boolean;
  invalid?: boolean;
  pulse?: boolean;
  onClick: () => void;
}) {
  const red =
    card.suit === "heart" || card.suit === "diamond" || card.suit === "joker";

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "relative h-28 w-20 shrink-0 rounded-[22px] border bg-[linear-gradient(180deg,#fffaf1_0%,#ffffff_55%,#f0ece4_100%)] px-3 py-2 text-left shadow-[0_24px_42px_-20px_rgba(0,0,0,0.65)] transition-all",
        pulse ? "scale-[1.03]" : "",
        selected
          ? "-translate-y-5 border-amber-400 ring-2 ring-amber-300/40"
          : "border-slate-300/90 hover:-translate-y-2",
        invalid ? "border-red-400 ring-2 ring-red-300/40" : "",
      )}
    >
      <div className="absolute inset-x-2 top-1 h-5 rounded-full bg-white/80 blur-sm" />
      <div
        className={cn(
          "text-xs font-semibold",
          red ? "text-red-500" : "text-slate-900",
        )}
      >
        {cardSuit(card)}
      </div>
      <div
        className={cn(
          "mt-2 text-2xl font-black",
          red ? "text-red-500" : "text-slate-900",
        )}
      >
        {cardRank(card)}
      </div>
      <div
        className={cn(
          "absolute bottom-2 right-2 text-lg font-semibold",
          red ? "text-red-400" : "text-slate-400",
        )}
      >
        {cardSuit(card)}
      </div>
    </button>
  );
}

function MatchLinkCard({
  match,
  mine = false,
}: {
  match: DoudizhuMatchSummary;
  mine?: boolean;
}) {
  return (
    <Link
      href={`/games/dou-dizhu/matches/${match.match_id}`}
      className={cn(
        "block rounded-[24px] border px-4 py-4 transition-colors",
        mine
          ? "border-sky-300/15 bg-sky-300/10 hover:border-sky-300/30 hover:bg-sky-300/14"
          : "border-white/10 bg-black/20 hover:border-white/20 hover:bg-white/[0.05]",
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-white">{match.room_title}</span>
            <Badge className="border-white/15 bg-white/8 text-white">
              {match.match_mode === "demo_ai" ? "AI 演示" : "真人房"}
            </Badge>
          </div>
          <div className="mt-1 text-sm text-slate-400">
            {new Date(match.finished_at).toLocaleString("zh-CN")} · 地主{" "}
            {seatLabel(match.landlord_seat)}
          </div>
        </div>

        <div className="text-right text-sm text-slate-300">
          <div>{match.winner_side === "landlord" ? "地主胜" : "农民胜"}</div>
          <div className="mt-1">倍率 x{match.multiplier}</div>
        </div>
      </div>
    </Link>
  );
}

function LeaderboardCard({ entry }: { entry: DoudizhuLeaderboardEntry }) {
  return (
    <div className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="font-semibold text-white">
            #{entry.rank} {entry.display_name}
          </div>
          <div className="mt-1 text-sm text-slate-400">
            {entry.matches} 局 · 胜 {entry.wins} 局
          </div>
        </div>
        <div className="text-right">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
            Total
          </div>
          <div className="mt-1 text-2xl font-black tracking-tight text-white">
            {entry.total_score}
          </div>
        </div>
      </div>
    </div>
  );
}

export function DouDizhuPlayStage() {
  const { user } = useAuth();
  const [sessionId, setSessionId] = useState("");
  const [playerName, setPlayerName] = useState("");
  const [roomTitle, setRoomTitle] = useState("周末牌局");
  const [rooms, setRooms] = useState<DoudizhuRoom[]>([]);
  const [roomsLoading, setRoomsLoading] = useState(false);
  const [metaLoading, setMetaLoading] = useState(false);
  const [activeRoom, setActiveRoom] = useState<DoudizhuRoom | null>(null);
  const [activeRoomId, setActiveRoomId] = useState<string | null>(null);
  const [privateState, setPrivateState] = useState<DoudizhuPrivateState | null>(
    null,
  );
  const [latestAction, setLatestAction] = useState<DoudizhuActionResult | null>(
    null,
  );
  const [selectedCards, setSelectedCards] = useState<string[]>([]);
  const [selectionError, setSelectionError] = useState("");
  const [tableEffect, setTableEffect] = useState<
    "idle" | "play" | "bomb" | "settlement" | "error"
  >("idle");
  const [settlementVisible, setSettlementVisible] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [wsStatus, setWsStatus] = useState<RoomWSStatus>("idle");
  const [notice, setNotice] = useState(
    "直接开始 AI 演示，或者创建一个真人房。",
  );
  const [errorMessage, setErrorMessage] = useState("");
  const [timeTick, setTimeTick] = useState(Date.now());
  const [leaderboard, setLeaderboard] = useState<DoudizhuLeaderboardEntry[]>(
    [],
  );
  const [recentMatches, setRecentMatches] = useState<DoudizhuMatchSummary[]>(
    [],
  );
  const [myMatches, setMyMatches] = useState<DoudizhuMatchSummary[]>([]);

  const wsRef = useRef<WebSocket | null>(null);
  const activeRoomIdRef = useRef<string | null>(null);
  const reconnectEnabledRef = useRef(false);
  const audioContextRef = useRef<AudioContext | null>(null);
  const lastSettledRoomRef = useRef<string | null>(null);
  const playFeedbackToneRef = useRef<
    (effect: "hint" | "play" | "bomb" | "settlement" | "error") => void
  >(() => {});

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    let storedSessionID = localStorage.getItem("doudizhu_session_id");
    if (!storedSessionID) {
      storedSessionID = createLocalSessionID();
      localStorage.setItem("doudizhu_session_id", storedSessionID);
    }
    setSessionId(storedSessionID);

    const storedPlayerName = localStorage.getItem("doudizhu_player_name");
    if (storedPlayerName) {
      setPlayerName(storedPlayerName);
    } else if (user?.username) {
      setPlayerName(user.username);
    } else {
      setPlayerName("牌桌玩家");
    }
  }, [user?.username]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    if (playerName.trim()) {
      localStorage.setItem("doudizhu_player_name", playerName.trim());
    }
  }, [playerName]);

  useEffect(() => {
    let cancelled = false;

    async function loadRooms() {
      setRoomsLoading(true);
      try {
        const data = await apiClient.getDoudizhuRooms();
        if (!cancelled) {
          setRooms(data.rooms);
          if (activeRoomIdRef.current) {
            const current = data.rooms.find(
              (item) => item.id === activeRoomIdRef.current,
            );
            if (current) {
              setActiveRoom(current);
            }
          }
        }
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(
            error instanceof Error ? error.message : "加载房间失败",
          );
        }
      } finally {
        if (!cancelled) {
          setRoomsLoading(false);
        }
      }
    }

    void loadRooms();
    const timer = window.setInterval(
      () => {
        void loadRooms();
      },
      activeRoomId ? 8000 : 5000,
    );

    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [activeRoomId]);

  useEffect(() => {
    let cancelled = false;

    async function loadMeta() {
      setMetaLoading(true);
      try {
        const [leaderboardData, matchesData, myMatchesData] = await Promise.all(
          [
            apiClient.getDoudizhuLeaderboard(10),
            apiClient.getDoudizhuRecentMatches(6),
            user
              ? apiClient.getMyDoudizhuRecentMatches(4)
              : Promise.resolve({ matches: [] }),
          ],
        );
        if (!cancelled) {
          setLeaderboard(leaderboardData.entries);
          setRecentMatches(matchesData.matches);
          setMyMatches(myMatchesData.matches);
        }
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(
            error instanceof Error ? error.message : "加载战报失败",
          );
        }
      } finally {
        if (!cancelled) {
          setMetaLoading(false);
        }
      }
    }

    void loadMeta();
    const timer = window.setInterval(() => {
      void loadMeta();
    }, 10000);

    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [user]);

  useEffect(() => {
    if (!activeRoom) {
      return;
    }

    const timer = window.setInterval(() => {
      setTimeTick(Date.now());
    }, 250);
    return () => window.clearInterval(timer);
  }, [activeRoom]);

  useEffect(() => {
    return () => {
      reconnectEnabledRef.current = false;
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!privateState) {
      setSelectedCards([]);
      return;
    }
    const hand = Array.isArray(privateState.hand) ? privateState.hand : [];
    setSelectedCards((current) =>
      current.filter((item) => hand.some((card) => cardKey(card) === item)),
    );
  }, [privateState]);

  function triggerTableEffect(
    effect: "play" | "bomb" | "settlement" | "error",
    duration = 850,
  ) {
    setTableEffect(effect);
    window.setTimeout(() => {
      setTableEffect((current) => (current === effect ? "idle" : current));
    }, duration);
  }

  function playFeedbackTone(
    effect: "hint" | "play" | "bomb" | "settlement" | "error",
  ) {
    if (!soundEnabled || typeof window === "undefined") {
      return;
    }

    const AudioContextClass = window.AudioContext;
    if (!AudioContextClass) {
      return;
    }
    const context = audioContextRef.current ?? new AudioContextClass();
    audioContextRef.current = context;

    const patterns: Record<typeof effect, number[]> = {
      hint: [620],
      play: [480],
      bomb: [220, 160],
      settlement: [520, 660, 820],
      error: [180, 150],
    };

    let offset = 0;
    for (const frequency of patterns[effect]) {
      const oscillator = context.createOscillator();
      const gain = context.createGain();
      oscillator.type = effect === "bomb" ? "square" : "triangle";
      oscillator.frequency.value = frequency;
      gain.gain.value = effect === "error" ? 0.015 : 0.025;
      oscillator.connect(gain);
      gain.connect(context.destination);
      oscillator.start(context.currentTime + offset);
      oscillator.stop(context.currentTime + offset + 0.08);
      offset += 0.1;
    }
  }

  playFeedbackToneRef.current = playFeedbackTone;

  function updateActiveRoom(room: DoudizhuRoom) {
    const nextPlayers = [...room.players].sort(playerSort);
    const nextRoom = { ...room, players: nextPlayers };
    setActiveRoom(nextRoom);
    setActiveRoomId(nextRoom.id);
    activeRoomIdRef.current = nextRoom.id;

    switch (nextRoom.status) {
      case "bidding":
        setNotice("叫分阶段已开始。地主归属由服务端统一裁定。");
        break;
      case "playing":
        setNotice(
          nextRoom.match_mode === "demo_ai"
            ? "AI 演示局进行中。你现在在真正的牌桌里和两名机器人对局。"
            : "真人对局进行中。当前轮转、压牌和胜负都由服务端处理。",
        );
        break;
      case "settlement":
        setNotice("本局已结算。你可以继续留在房间再开一局，或直接查看战报。");
        break;
      case "redeal":
        setNotice("这轮没有确定地主，服务端会重新发起一轮。");
        break;
      default:
        setNotice(
          nextRoom.match_mode === "demo_ai"
            ? "这是 AI 演示房。点准备后，房主可以直接开始。"
            : "这是真人房。凑齐 3 人并全部准备后即可开始。",
        );
    }
  }

  useEffect(() => {
    if (!activeRoom) {
      lastSettledRoomRef.current = null;
      setSettlementVisible(false);
      return;
    }

    let effectTimer: number | undefined;
    let hideTimer: number | undefined;
    if (
      activeRoom.status === "settlement" &&
      lastSettledRoomRef.current !== activeRoom.id
    ) {
      lastSettledRoomRef.current = activeRoom.id;
      setSettlementVisible(true);
      setTableEffect("settlement");
      effectTimer = window.setTimeout(() => {
        setTableEffect((current) =>
          current === "settlement" ? "idle" : current,
        );
      }, 1400);
      playFeedbackToneRef.current("settlement");
      hideTimer = window.setTimeout(() => {
        setSettlementVisible(false);
      }, 3200);
      return () => {
        if (effectTimer) {
          window.clearTimeout(effectTimer);
        }
        if (hideTimer) {
          window.clearTimeout(hideTimer);
        }
      };
    }
    if (activeRoom.status !== "settlement") {
      setSettlementVisible(false);
      lastSettledRoomRef.current = null;
    }
    return () => {
      if (effectTimer) {
        window.clearTimeout(effectTimer);
      }
      if (hideTimer) {
        window.clearTimeout(hideTimer);
      }
    };
  }, [activeRoom, soundEnabled]);

  function closeSocket(allowReconnect: boolean) {
    reconnectEnabledRef.current = allowReconnect;
    if (wsRef.current) {
      wsRef.current.onclose = null;
      wsRef.current.close();
      wsRef.current = null;
    }
  }

  function connectToRoom(roomId: string) {
    if (!sessionId) {
      setErrorMessage("本地 session 还未准备好，请稍后重试。");
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage("请先填写玩家名称。");
      return;
    }

    closeSocket(false);
    setErrorMessage("");
    setWsStatus("connecting");
    setNotice("正在连接牌桌...");
    activeRoomIdRef.current = roomId;
    setActiveRoomId(roomId);
    reconnectEnabledRef.current = true;
    setPrivateState(null);
    setLatestAction(null);
    setSelectionError("");
    setSelectedCards([]);

    const query = new URLSearchParams({
      room_id: roomId,
      session_id: sessionId,
      player_name: playerName.trim(),
    });
    const token = apiClient.getToken();
    if (token) {
      query.set("token", token);
    }

    const ws = new WebSocket(
      `${getGameWsBase()}/ws/game/dou-dizhu?${query.toString()}`,
    );
    wsRef.current = ws;

    ws.onopen = () => {
      setWsStatus("connected");
    };

    ws.onmessage = (event) => {
      try {
        const message: RoomServerMessage = JSON.parse(event.data);
        switch (message.type) {
          case "joined":
            if (message.payload?.room) {
              updateActiveRoom(message.payload.room as DoudizhuRoom);
            }
            if (message.payload?.private_state) {
              setPrivateState(
                message.payload.private_state as DoudizhuPrivateState,
              );
            }
            if (
              message.payload?.session_id &&
              typeof message.payload.session_id === "string"
            ) {
              setSessionId(message.payload.session_id);
              if (typeof window !== "undefined") {
                localStorage.setItem(
                  "doudizhu_session_id",
                  message.payload.session_id,
                );
              }
            }
            break;
          case "room_state":
            updateActiveRoom(message.payload as DoudizhuRoom);
            break;
          case "private_state":
            setPrivateState(message.payload as DoudizhuPrivateState);
            break;
          case "action_result":
            setLatestAction(message.payload as DoudizhuActionResult);
            setSelectionError("");
            if (typeof message.payload?.message === "string") {
              setNotice(message.payload.message);
            }
            if (
              message.payload?.action_type === "play_cards" ||
              message.payload?.action_type === "pass_turn"
            ) {
              setSelectedCards([]);
            }
            if (message.payload?.action_type === "play_cards") {
              if (
                message.payload?.combo?.type === "bomb" ||
                message.payload?.combo?.type === "rocket"
              ) {
                triggerTableEffect("bomb", 1100);
                playFeedbackTone("bomb");
              } else {
                triggerTableEffect("play");
                playFeedbackTone("play");
              }
            }
            break;
          case "hint_result": {
            const hint = message.payload as RoomHintResult;
            setSelectionError("");
            if (
              hint.action_type === "play_cards" &&
              Array.isArray(hint.cards)
            ) {
              setSelectedCards(hint.cards.map((card) => cardKey(card)));
            }
            if (hint.action_type === "pass_turn") {
              setSelectedCards([]);
            }
            playFeedbackTone("hint");
            if (typeof hint.message === "string") {
              if (
                hint.action_type === "bid" &&
                typeof hint.bid_score === "number"
              ) {
                setNotice(
                  `${hint.message} ${hint.bid_score > 0 ? `你可以直接叫 ${hint.bid_score} 分。` : ""}`,
                );
              } else {
                setNotice(hint.message);
              }
            }
            break;
          }
          case "room_closed":
            setNotice("房间已关闭。");
            setActiveRoom(null);
            setActiveRoomId(null);
            setPrivateState(null);
            setLatestAction(null);
            setSelectedCards([]);
            activeRoomIdRef.current = null;
            reconnectEnabledRef.current = false;
            setWsStatus("idle");
            setSelectionError("");
            break;
          case "error":
            if (
              typeof message.payload?.message === "string" &&
              /出牌|牌型|压过|不匹配|过牌/.test(message.payload.message)
            ) {
              setSelectionError(message.payload.message);
              triggerTableEffect("error", 700);
              playFeedbackTone("error");
            }
            setErrorMessage(
              typeof message.payload?.message === "string"
                ? message.payload.message
                : "房间操作失败",
            );
            break;
          default:
            break;
        }
      } catch {
        setErrorMessage("房间消息解析失败");
      }
    };

    ws.onerror = () => {
      setErrorMessage("牌桌连接异常，请稍后重试。");
    };

    ws.onclose = () => {
      setWsStatus("closed");
      wsRef.current = null;

      if (!reconnectEnabledRef.current || !activeRoomIdRef.current) {
        setActiveRoom(null);
        setActiveRoomId(null);
        setPrivateState(null);
        setLatestAction(null);
        setSelectedCards([]);
        setSelectionError("");
        setNotice("你已离开房间。");
        return;
      }

      window.setTimeout(() => {
        if (activeRoomIdRef.current) {
          connectToRoom(activeRoomIdRef.current);
        }
      }, 1200);
    };
  }

  function sendRoomMessage(type: string, payload?: Record<string, unknown>) {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setErrorMessage("房间连接尚未建立");
      return;
    }
    if (audioContextRef.current?.state === "suspended") {
      void audioContextRef.current.resume();
    }
    wsRef.current.send(JSON.stringify({ type, payload }));
  }

  async function handleCreateRoom() {
    if (!sessionId) {
      setErrorMessage("本地 session 还未准备好");
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage("请先填写玩家名称");
      return;
    }

    setErrorMessage("");
    try {
      const room = await apiClient.createDoudizhuRoom({
        title: roomTitle.trim() || "周末牌局",
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      connectToRoom(room.id);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "创建房间失败");
    }
  }

  async function handleCreateDemoRoom() {
    if (!sessionId) {
      setErrorMessage("本地 session 还未准备好");
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage("请先填写玩家名称");
      return;
    }

    setErrorMessage("");
    try {
      const room = await apiClient.createDoudizhuDemoRoom({
        title: roomTitle.trim() || "单人演示房",
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      connectToRoom(room.id);
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "创建演示房失败",
      );
    }
  }

  const me =
    activeRoom?.players.find((player) => player.session_id === sessionId) ??
    null;
  const boardSeats = useMemo(
    () => buildBoardSeats(activeRoom, me),
    [activeRoom, me],
  );
  const isHost = !!me?.is_host;
  const canStart =
    !!activeRoom &&
    activeRoom.status !== "playing" &&
    activeRoom.status !== "bidding" &&
    activeRoom.players
      .filter((player) => !player.is_bot)
      .every((player) => player.ready && player.connected);

  const isMyBidTurn =
    !!activeRoom &&
    !!me &&
    activeRoom.status === "bidding" &&
    activeRoom.current_bidder === me.seat;
  const isMyPlayTurn =
    !!activeRoom &&
    !!me &&
    activeRoom.status === "playing" &&
    activeRoom.current_turn === me.seat;

  const selectedHandCards = useMemo(() => {
    if (!privateState || !Array.isArray(privateState.hand)) {
      return [];
    }
    return privateState.hand.filter((card) =>
      selectedCards.includes(cardKey(card)),
    );
  }, [privateState, selectedCards]);

  const currentHand = Array.isArray(privateState?.hand)
    ? privateState.hand
    : [];
  const selectedComboPreview = useMemo(
    () => evaluateSelectedCombo(selectedHandCards),
    [selectedHandCards],
  );

  useEffect(() => {
    if (!activeRoom || activeRoom.status !== "settlement") {
      return;
    }

    const timer = window.setTimeout(() => {
      apiClient
        .getDoudizhuLeaderboard(10)
        .then((data) => setLeaderboard(data.entries))
        .catch(() => {});
      apiClient
        .getDoudizhuRecentMatches(6)
        .then((data) => setRecentMatches(data.matches))
        .catch(() => {});
      if (user) {
        apiClient
          .getMyDoudizhuRecentMatches(4)
          .then((data) => setMyMatches(data.matches))
          .catch(() => {});
      }
    }, 500);

    return () => window.clearTimeout(timer);
  }, [activeRoom, user]);

  return (
    <div className="space-y-6">
      <section className="relative overflow-hidden rounded-[36px] border border-white/10 bg-[#081013] text-white shadow-[0_30px_90px_-50px_rgba(0,0,0,0.85)]">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,207,120,0.15),transparent_28%),radial-gradient(circle_at_20%_20%,rgba(255,255,255,0.08),transparent_18%),linear-gradient(180deg,#0d382c_0%,#082019_100%)]" />
        <div className="relative p-5 md:p-8">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-[0.32em] text-emerald-100/50">
                Game Table
              </p>
              <h2 className="mt-3 text-3xl font-black tracking-tight text-white md:text-5xl">
                经典斗地主
              </h2>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-emerald-50/75 md:text-base">
                现在页面的中心是牌桌本身，而不是房间工具。你可以直接开始 AI
                演示，或者创建真人房，把叫分、 出牌和结算完整走一遍。
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <Badge className="border-white/15 bg-white/8 text-white">
                WS: {wsStatus}
              </Badge>
              {activeRoom && (
                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                  {roomStatusLabel(activeRoom.status)}
                </Badge>
              )}
              {activeRoom && (
                <Badge className="border-white/15 bg-black/25 text-white">
                  {activeRoom.match_mode === "demo_ai" ? "AI 演示" : "真人房"}
                </Badge>
              )}
            </div>
          </div>

          {!activeRoom && (
            <div className="mt-8 grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
              <div className="rounded-[30px] border border-white/10 bg-black/20 p-6">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2 sm:col-span-2">
                    <Label htmlFor="ddz-player-name" className="text-white/85">
                      玩家名称
                    </Label>
                    <Input
                      id="ddz-player-name"
                      value={playerName}
                      onChange={(event) => setPlayerName(event.target.value)}
                      className="border-white/10 bg-black/30 text-white placeholder:text-emerald-50/35"
                      placeholder="输入一个牌桌显示名称"
                    />
                  </div>

                  <div className="space-y-2 sm:col-span-2">
                    <Label htmlFor="ddz-room-title" className="text-white/85">
                      房间标题
                    </Label>
                    <Input
                      id="ddz-room-title"
                      value={roomTitle}
                      onChange={(event) => setRoomTitle(event.target.value)}
                      className="border-white/10 bg-black/30 text-white placeholder:text-emerald-50/35"
                      placeholder="例如：周五牌局"
                    />
                  </div>
                </div>

                <div className="mt-5 flex flex-wrap gap-3">
                  <Button
                    onClick={handleCreateDemoRoom}
                    className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                  >
                    <Bot className="mr-2 h-4 w-4" />
                    立即开始 AI 对局
                  </Button>
                  <Button
                    onClick={handleCreateRoom}
                    variant="outline"
                    className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                  >
                    <Users2 className="mr-2 h-4 w-4" />
                    创建真人房
                  </Button>
                </div>

                <div className="mt-5 rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-4 text-sm leading-7 text-emerald-50/70">
                  这页优先保证单人也能直接进入一局真实对战，不需要先理解房间系统。AI
                  演示房会自动补齐两名机器人， 你点准备后由房主直接开始。
                </div>

                {errorMessage && (
                  <div className="mt-4 rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
                    {errorMessage}
                  </div>
                )}
              </div>

              <div className="rounded-[30px] border border-white/10 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.08),transparent_35%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.03))] p-6">
                <div className="grid gap-4 md:grid-cols-2">
                  {[
                    {
                      title: "真正的牌桌布局",
                      text: "中央展示当前阶段、上一手和玩家座位，底部手牌和操作直接围绕对局展开。",
                    },
                    {
                      title: "AI 演示兜底",
                      text: "单人也能完整跑一局，适合演示、录屏和服务端联调。",
                    },
                    {
                      title: "真人房继续保留",
                      text: "需要 3 个真实玩家时，仍然可以从当前页面创建并进入真人房。",
                    },
                    {
                      title: "战报已接入",
                      text: "打完一局即可在最近对局里跳转到战报页，不再是打一局就蒸发。",
                    },
                  ].map((item) => (
                    <div
                      key={item.title}
                      className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4"
                    >
                      <div className="font-semibold text-white">
                        {item.title}
                      </div>
                      <div className="mt-2 text-sm leading-6 text-emerald-50/70">
                        {item.text}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {activeRoom && (
            <div className="mt-8 grid gap-6 xl:grid-cols-[1.14fr_0.86fr]">
              <div className="space-y-5">
                <div className="rounded-[30px] border border-amber-300/15 bg-[linear-gradient(135deg,rgba(244,182,63,0.16),rgba(219,90,63,0.08))] p-4 shadow-[0_18px_60px_-40px_rgba(0,0,0,0.7)]">
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge className="border-white/15 bg-black/25 text-white">
                          {activeRoom.code}
                        </Badge>
                        <Badge className="border-white/15 bg-black/25 text-white">
                          {roomModeLabel(activeRoom.match_mode)}
                        </Badge>
                        <Badge className="border-emerald-300/20 bg-emerald-300/10 text-emerald-100">
                          {roomStatusLabel(activeRoom.status)}
                        </Badge>
                        {activeRoom.landlord !== undefined &&
                          activeRoom.landlord !== null && (
                            <Badge className="border-red-400/20 bg-red-400/10 text-red-100">
                              地主 {seatLabel(activeRoom.landlord)}
                            </Badge>
                          )}
                      </div>
                      <div className="mt-3 text-2xl font-black tracking-tight text-white">
                        {activeRoom.title}
                      </div>
                      <div className="mt-2 text-sm leading-7 text-amber-50/80">
                        {notice}
                      </div>
                    </div>

                    <div className="flex flex-wrap gap-2">
                      <Button
                        onClick={() =>
                          sendRoomMessage("set_ready", {
                            ready: me ? !me.ready : true,
                          })
                        }
                        disabled={
                          !me ||
                          wsStatus !== "connected" ||
                          activeRoom.status === "playing" ||
                          activeRoom.status === "bidding"
                        }
                        className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110 disabled:opacity-50"
                      >
                        {me?.ready ? "取消准备" : "准备就绪"}
                      </Button>
                      <Button
                        onClick={() => sendRoomMessage("start_round")}
                        disabled={
                          !isHost || !canStart || wsStatus !== "connected"
                        }
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                      >
                        <Radio className="mr-2 h-4 w-4" />
                        房主开始
                      </Button>
                      <Button
                        onClick={() =>
                          sendRoomMessage("toggle_auto_play", {
                            enabled: !(me?.auto_play ?? false),
                          })
                        }
                        disabled={!me || wsStatus !== "connected"}
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                      >
                        {me?.auto_play ? "关闭托管" : "开启托管"}
                      </Button>
                      <Button
                        onClick={() => {
                          reconnectEnabledRef.current = false;
                          activeRoomIdRef.current = null;
                          sendRoomMessage("leave_room");
                        }}
                        disabled={
                          activeRoom.status === "playing" ||
                          activeRoom.status === "bidding"
                        }
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                      >
                        <DoorOpen className="mr-2 h-4 w-4" />
                        离开房间
                      </Button>
                    </div>
                  </div>
                </div>

                <div className="relative overflow-hidden rounded-[38px] border border-white/10 bg-[linear-gradient(180deg,#0a3528_0%,#082119_100%)] shadow-[inset_0_1px_0_rgba(255,255,255,0.06),0_30px_90px_-40px_rgba(0,0,0,0.85)]">
                  <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,214,120,0.18),transparent_28%),radial-gradient(circle_at_center,rgba(255,255,255,0.08),transparent_24%),radial-gradient(circle_at_bottom,rgba(0,0,0,0.26),transparent_38%)]" />
                  <div className="relative p-4 md:p-6">
                    <div className="grid gap-4 lg:grid-cols-2 xl:hidden">
                      <TablePlayerSeat
                        player={boardSeats.left}
                        position="left"
                        isCurrentTurn={
                          !!boardSeats.left &&
                          activeRoom.current_turn === boardSeats.left.seat
                        }
                        isCurrentBidder={
                          !!boardSeats.left &&
                          activeRoom.current_bidder === boardSeats.left.seat
                        }
                        isLandlord={
                          !!boardSeats.left &&
                          activeRoom.landlord === boardSeats.left.seat
                        }
                        isMe={
                          !!boardSeats.left && me?.seat === boardSeats.left.seat
                        }
                      />
                      <TablePlayerSeat
                        player={boardSeats.right}
                        position="right"
                        isCurrentTurn={
                          !!boardSeats.right &&
                          activeRoom.current_turn === boardSeats.right.seat
                        }
                        isCurrentBidder={
                          !!boardSeats.right &&
                          activeRoom.current_bidder === boardSeats.right.seat
                        }
                        isLandlord={
                          !!boardSeats.right &&
                          activeRoom.landlord === boardSeats.right.seat
                        }
                        isMe={
                          !!boardSeats.right &&
                          me?.seat === boardSeats.right.seat
                        }
                      />
                    </div>

                    <div className="relative min-h-[460px] xl:min-h-[590px]">
                      <div className="hidden xl:block">
                        <div className="absolute left-0 top-4 w-[285px]">
                          <TablePlayerSeat
                            player={boardSeats.left}
                            position="left"
                            isCurrentTurn={
                              !!boardSeats.left &&
                              activeRoom.current_turn === boardSeats.left.seat
                            }
                            isCurrentBidder={
                              !!boardSeats.left &&
                              activeRoom.current_bidder === boardSeats.left.seat
                            }
                            isLandlord={
                              !!boardSeats.left &&
                              activeRoom.landlord === boardSeats.left.seat
                            }
                            isMe={
                              !!boardSeats.left &&
                              me?.seat === boardSeats.left.seat
                            }
                          />
                        </div>
                        <div className="absolute right-0 top-4 w-[285px]">
                          <TablePlayerSeat
                            player={boardSeats.right}
                            position="right"
                            isCurrentTurn={
                              !!boardSeats.right &&
                              activeRoom.current_turn === boardSeats.right.seat
                            }
                            isCurrentBidder={
                              !!boardSeats.right &&
                              activeRoom.current_bidder ===
                                boardSeats.right.seat
                            }
                            isLandlord={
                              !!boardSeats.right &&
                              activeRoom.landlord === boardSeats.right.seat
                            }
                            isMe={
                              !!boardSeats.right &&
                              me?.seat === boardSeats.right.seat
                            }
                          />
                        </div>
                      </div>

                      <div
                        className={cn(
                          "mx-auto flex max-w-[500px] flex-col items-center gap-5 rounded-[34px] border border-white/10 bg-[linear-gradient(180deg,rgba(0,0,0,0.26),rgba(255,255,255,0.05))] px-5 py-6 text-center shadow-[0_24px_70px_-40px_rgba(0,0,0,0.9)] transition-all duration-300",
                          tableEffect === "play"
                            ? "scale-[1.01] border-emerald-300/30 shadow-[0_24px_90px_-34px_rgba(16,185,129,0.45)]"
                            : "",
                          tableEffect === "bomb"
                            ? "scale-[1.02] border-amber-300/35 shadow-[0_24px_100px_-34px_rgba(245,158,11,0.55)]"
                            : "",
                          tableEffect === "error"
                            ? "border-red-300/30 shadow-[0_24px_80px_-34px_rgba(239,68,68,0.45)]"
                            : "",
                          tableEffect === "settlement"
                            ? "border-sky-300/35 shadow-[0_24px_100px_-34px_rgba(59,130,246,0.48)]"
                            : "",
                        )}
                      >
                        <div className="grid w-full gap-3 sm:grid-cols-3">
                          <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3">
                            <div className="text-xs uppercase tracking-[0.24em] text-emerald-100/45">
                              当前阶段
                            </div>
                            <div className="mt-2 text-xl font-black tracking-tight text-white">
                              {roomStatusLabel(activeRoom.status)}
                            </div>
                          </div>
                          <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3">
                            <div className="text-xs uppercase tracking-[0.24em] text-emerald-100/45">
                              当前目标
                            </div>
                            <div className="mt-2 text-xl font-black tracking-tight text-white">
                              {activeRoom.status === "bidding"
                                ? seatLabel(activeRoom.current_bidder)
                                : seatLabel(activeRoom.current_turn)}
                            </div>
                          </div>
                          <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3">
                            <div className="text-xs uppercase tracking-[0.24em] text-emerald-100/45">
                              当前计时
                            </div>
                            <div className="mt-2 text-xl font-black tracking-tight text-white">
                              {formatRemaining(
                                privateState?.turn_expires_at ??
                                  activeRoom.turn_expires_at,
                                timeTick,
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="w-full rounded-[28px] border border-white/10 bg-black/20 px-4 py-5">
                          <div className="text-xs uppercase tracking-[0.28em] text-emerald-100/45">
                            中心牌桌
                          </div>
                          <div className="mt-3 text-sm leading-7 text-emerald-50/75">
                            {notice}
                          </div>

                          {(privateState?.bottom_cards?.length ||
                            activeRoom.bottom_cards?.length) && (
                            <div className="mt-4">
                              <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                                地主底牌
                              </div>
                              <div className="mt-3 flex flex-wrap justify-center gap-2">
                                {(
                                  privateState?.bottom_cards ??
                                  activeRoom.bottom_cards ??
                                  []
                                ).map((card, index) => (
                                  <div
                                    key={`${cardKey(card)}-${index}`}
                                    className="rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-sm font-medium text-white"
                                  >
                                    {cardLabel(card)}
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          <div className="mt-5 rounded-[24px] border border-white/10 bg-white/[0.04] px-4 py-4">
                            <div className="text-xs uppercase tracking-[0.24em] text-emerald-100/45">
                              上一手
                            </div>
                            <div className="mt-2 text-lg font-semibold text-white">
                              {comboLabel(
                                privateState?.last_play ?? activeRoom.last_play,
                              )}
                            </div>
                            <div className="mt-2 text-sm text-emerald-50/65">
                              {privateState?.last_play_seat !== undefined ||
                              activeRoom.last_play_seat !== undefined
                                ? `上一手来自 ${seatLabel(
                                    privateState?.last_play_seat ??
                                      activeRoom.last_play_seat,
                                  )}`
                                : "等待首家出牌"}
                            </div>
                            {activeRoom.last_play_cards?.length ? (
                              <div className="mt-4 flex flex-wrap justify-center gap-2">
                                {activeRoom.last_play_cards.map(
                                  (card, index) => (
                                    <div
                                      key={`${cardKey(card)}-${index}`}
                                      className={cn(
                                        "rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-sm font-medium text-white transition-all",
                                        tableEffect === "play" ||
                                          tableEffect === "bomb"
                                          ? "translate-y-0 scale-100"
                                          : "",
                                      )}
                                    >
                                      {cardLabel(card)}
                                    </div>
                                  ),
                                )}
                              </div>
                            ) : null}
                          </div>
                        </div>

                        {latestAction && (
                          <div className="w-full rounded-2xl border border-sky-300/15 bg-sky-300/10 px-4 py-3 text-sm text-sky-50">
                            {latestAction.message ??
                              `${latestAction.actor_name} 完成了一次操作。`}
                          </div>
                        )}

                        {settlementVisible && activeRoom.winning_side && (
                          <div className="w-full rounded-[26px] border border-amber-300/20 bg-[linear-gradient(135deg,rgba(244,182,63,0.24),rgba(59,130,246,0.14))] px-5 py-5 text-white">
                            <div className="text-xs uppercase tracking-[0.28em] text-white/60">
                              Settlement
                            </div>
                            <div className="mt-2 text-3xl font-black tracking-tight">
                              {activeRoom.winning_side === "landlord"
                                ? "地主胜利"
                                : "农民胜利"}
                            </div>
                            <div className="mt-3 flex flex-wrap justify-center gap-2 text-sm">
                              <Badge className="border-white/15 bg-black/20 text-white">
                                倍率 x{activeRoom.multiplier}
                              </Badge>
                              <Badge className="border-white/15 bg-black/20 text-white">
                                炸弹 {activeRoom.bomb_count}
                              </Badge>
                              {activeRoom.spring && (
                                <Badge className="border-white/15 bg-black/20 text-white">
                                  春天
                                </Badge>
                              )}
                              {activeRoom.anti_spring && (
                                <Badge className="border-white/15 bg-black/20 text-white">
                                  反春
                                </Badge>
                              )}
                            </div>
                          </div>
                        )}
                      </div>

                      <div className="mt-6 xl:absolute xl:bottom-0 xl:left-1/2 xl:w-[82%] xl:-translate-x-1/2">
                        <TablePlayerSeat
                          player={boardSeats.bottom}
                          position="bottom"
                          isCurrentTurn={
                            !!boardSeats.bottom &&
                            activeRoom.current_turn === boardSeats.bottom.seat
                          }
                          isCurrentBidder={
                            !!boardSeats.bottom &&
                            activeRoom.current_bidder === boardSeats.bottom.seat
                          }
                          isLandlord={
                            !!boardSeats.bottom &&
                            activeRoom.landlord === boardSeats.bottom.seat
                          }
                          isMe={
                            !!boardSeats.bottom &&
                            me?.seat === boardSeats.bottom.seat
                          }
                        />
                      </div>
                    </div>
                  </div>
                </div>

                <div className="rounded-[32px] border border-white/10 bg-black/20 p-5 shadow-[0_24px_70px_-50px_rgba(0,0,0,0.8)]">
                  <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <div className="text-lg font-semibold text-white">
                        我的手牌
                      </div>
                      <div className="mt-1 text-sm text-slate-400">
                        {privateState
                          ? `角色：${privateState.role === "landlord" ? "地主" : "农民"} · 共 ${currentHand.length} 张`
                          : "进入房间后会显示你的实际手牌"}
                      </div>
                    </div>
                    <div className="rounded-full border border-white/10 bg-white/[0.04] px-3 py-1 text-xs text-slate-300">
                      已选 {selectedCards.length} 张
                    </div>
                  </div>

                  <div className="mb-4 grid gap-3 md:grid-cols-[1fr_auto]">
                    <div className="rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-sm">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        牌型提示
                      </div>
                      <div className="mt-2 font-medium text-white">
                        {selectedHandCards.length === 0
                          ? "未选中手牌"
                          : selectedComboPreview.combo
                            ? comboLabel(selectedComboPreview.combo)
                            : selectedComboPreview.error}
                      </div>
                    </div>
                    <Button
                      onClick={() => setSoundEnabled((current) => !current)}
                      variant="outline"
                      className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                    >
                      {soundEnabled ? "音效开" : "音效关"}
                    </Button>
                  </div>

                  {!privateState && (
                    <div className="rounded-2xl border border-dashed border-white/15 bg-white/[0.03] px-4 py-8 text-center text-sm text-slate-400">
                      先加入一个房间，再开始叫分和出牌。
                    </div>
                  )}

                  {privateState && (
                    <>
                      <div className="overflow-x-auto pb-4">
                        <div className="flex min-w-max items-end gap-3 pr-2">
                          {currentHand.map((card) => {
                            const selected = selectedCards.includes(
                              cardKey(card),
                            );
                            return (
                              <PlayCard
                                key={cardKey(card)}
                                card={card}
                                selected={selected}
                                invalid={selected && !!selectionError}
                                pulse={selected && tableEffect === "error"}
                                onClick={() => {
                                  setSelectionError("");
                                  setSelectedCards((current) =>
                                    current.includes(cardKey(card))
                                      ? current.filter(
                                          (item) => item !== cardKey(card),
                                        )
                                      : [...current, cardKey(card)],
                                  );
                                }}
                              />
                            );
                          })}
                        </div>
                      </div>

                      {selectionError && (
                        <div className="rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
                          {selectionError}
                        </div>
                      )}

                      <div className="mt-5 flex flex-wrap gap-3">
                        {isMyBidTurn && (
                          <>
                            <Button
                              onClick={() => sendRoomMessage("request_hint")}
                              disabled={wsStatus !== "connected"}
                              variant="outline"
                              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                            >
                              <Sparkles className="mr-2 h-4 w-4" />
                              提示
                            </Button>
                            {[1, 2, 3].map((score) => (
                              <Button
                                key={score}
                                onClick={() =>
                                  sendRoomMessage("bid", { score })
                                }
                                disabled={score <= activeRoom.highest_bid}
                                className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                              >
                                叫 {score} 分
                              </Button>
                            ))}
                            <Button
                              onClick={() => sendRoomMessage("pass_bid")}
                              variant="outline"
                              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                            >
                              不叫
                            </Button>
                          </>
                        )}

                        {!isMyBidTurn && (
                          <>
                            <Button
                              onClick={() => sendRoomMessage("request_hint")}
                              disabled={
                                !isMyPlayTurn || wsStatus !== "connected"
                              }
                              variant="outline"
                              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                            >
                              <Sparkles className="mr-2 h-4 w-4" />
                              提示
                            </Button>
                            <Button
                              onClick={() =>
                                sendRoomMessage("play_cards", {
                                  cards: selectedHandCards,
                                })
                              }
                              disabled={
                                !isMyPlayTurn ||
                                selectedHandCards.length === 0 ||
                                !selectedComboPreview.combo
                              }
                              className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110 disabled:opacity-50"
                            >
                              出牌
                            </Button>
                            <Button
                              onClick={() => sendRoomMessage("pass_turn")}
                              disabled={!isMyPlayTurn || !privateState.can_pass}
                              variant="outline"
                              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                            >
                              过牌
                            </Button>
                            <Button
                              onClick={() => {
                                setSelectionError("");
                                setSelectedCards([]);
                              }}
                              variant="outline"
                              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                            >
                              <RefreshCcw className="mr-2 h-4 w-4" />
                              重选
                            </Button>
                          </>
                        )}
                      </div>
                    </>
                  )}
                </div>
              </div>

              <div className="space-y-4">
                <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-1">
                  <div className="rounded-[28px] border border-white/10 bg-black/20 p-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      最高叫分
                    </div>
                    <div className="mt-2 text-3xl font-black tracking-tight text-white">
                      {activeRoom.highest_bid}
                    </div>
                  </div>
                  <div className="rounded-[28px] border border-white/10 bg-black/20 p-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      当前倍率
                    </div>
                    <div className="mt-2 text-3xl font-black tracking-tight text-white">
                      x{activeRoom.multiplier}
                    </div>
                  </div>
                  <div className="rounded-[28px] border border-white/10 bg-black/20 p-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      我的角色
                    </div>
                    <div className="mt-2 text-3xl font-black tracking-tight text-white">
                      {privateState?.role === "landlord" ? "地主" : "农民"}
                    </div>
                  </div>
                  <div className="rounded-[28px] border border-white/10 bg-black/20 p-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      剩余手牌
                    </div>
                    <div className="mt-2 text-3xl font-black tracking-tight text-white">
                      {currentHand.length}
                    </div>
                  </div>
                  <div className="rounded-[28px] border border-white/10 bg-black/20 p-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      当前连线
                    </div>
                    <div className="mt-2 text-3xl font-black tracking-tight text-white">
                      {wsStatus === "connected" ? "稳定" : "重连中"}
                    </div>
                  </div>
                </div>

                <div className="rounded-[30px] border border-white/10 bg-black/20 p-5">
                  <div className="mb-3 text-lg font-semibold text-white">
                    倍率细节
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Badge className="border-white/15 bg-white/8 text-white">
                      炸弹 {activeRoom.bomb_count}
                    </Badge>
                    {activeRoom.spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        春天
                      </Badge>
                    )}
                    {activeRoom.anti_spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        反春
                      </Badge>
                    )}
                    {!activeRoom.spring && !activeRoom.anti_spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        常规倍率
                      </Badge>
                    )}
                  </div>
                </div>

                <div className="rounded-[30px] border border-white/10 bg-black/20 p-5">
                  <div className="mb-3 text-lg font-semibold text-white">
                    最近操作
                  </div>
                  <div className="space-y-2">
                    {(activeRoom.recent_actions ?? []).length === 0 && (
                      <div className="rounded-2xl border border-dashed border-white/15 bg-white/[0.03] px-4 py-5 text-sm text-slate-400">
                        还没有操作记录。
                      </div>
                    )}
                    {(activeRoom.recent_actions ?? [])
                      .slice()
                      .reverse()
                      .map((action, index) => (
                        <div
                          key={`${action.at}-${index}`}
                          className="rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-sm text-slate-300"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-3">
                            <div className="font-medium text-white">
                              {action.actor_name ?? seatLabel(action.seat)}
                            </div>
                            <div className="text-xs text-slate-500">
                              {new Date(action.at).toLocaleTimeString("zh-CN")}
                            </div>
                          </div>
                          <div className="mt-1">
                            {action.message ??
                              actionTypeLabel(action.action_type)}
                          </div>
                          {action.cards?.length ? (
                            <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-400">
                              {action.cards.map((card, index) => (
                                <span key={`${cardKey(card)}-${index}`}>
                                  {cardLabel(card)}
                                </span>
                              ))}
                            </div>
                          ) : null}
                        </div>
                      ))}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </section>

      <Tabs defaultValue="start" className="space-y-4">
        <TabsList className="border border-white/10 bg-black/20">
          <TabsTrigger
            value="start"
            className="data-[state=active]:bg-white data-[state=active]:text-slate-950"
          >
            房间大厅
          </TabsTrigger>
          <TabsTrigger
            value="matches"
            className="data-[state=active]:bg-white data-[state=active]:text-slate-950"
          >
            最近对局
          </TabsTrigger>
          <TabsTrigger
            value="leaderboard"
            className="data-[state=active]:bg-white data-[state=active]:text-slate-950"
          >
            排行榜
          </TabsTrigger>
        </TabsList>

        <TabsContent value="start">
          <div className="grid gap-6 xl:grid-cols-[0.96fr_1.04fr]">
            <Card className="border-white/10 bg-white/[0.04] text-white">
              <CardContent className="p-6">
                <div className="mb-5 flex items-center gap-3">
                  <Sparkles className="h-5 w-5 text-amber-300" />
                  <div>
                    <h3 className="text-2xl font-black tracking-tight text-white">
                      快速开局
                    </h3>
                    <p className="mt-1 text-sm text-slate-400">
                      继续保留房间能力，但它现在退到辅助入口，不再压过牌桌本身。
                    </p>
                  </div>
                </div>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label
                      htmlFor="ddz-room-title-panel"
                      className="text-white/85"
                    >
                      房间标题
                    </Label>
                    <Input
                      id="ddz-room-title-panel"
                      value={roomTitle}
                      onChange={(event) => setRoomTitle(event.target.value)}
                      className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                      placeholder="例如：夜场训练局"
                    />
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <Button
                      onClick={handleCreateDemoRoom}
                      className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                    >
                      <Bot className="mr-2 h-4 w-4" />
                      开始 AI 演示
                    </Button>
                    <Button
                      onClick={handleCreateRoom}
                      variant="outline"
                      className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                    >
                      <Users2 className="mr-2 h-4 w-4" />
                      创建真人房
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-white/10 bg-white/[0.04] text-white">
              <CardContent className="p-6">
                <div className="mb-5 flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <Users2 className="h-5 w-5 text-sky-300" />
                    <div>
                      <h3 className="text-2xl font-black tracking-tight text-white">
                        房间列表
                      </h3>
                      <p className="mt-1 text-sm text-slate-400">
                        如果你要和真人联机，这里仍然可以直接加入当前开放中的房间。
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 rounded-full border border-white/10 bg-black/20 px-3 py-2 text-sm text-slate-300">
                    {roomsLoading ? (
                      <Loader2 className="h-4 w-4 animate-spin text-sky-300" />
                    ) : (
                      <Radio className="h-4 w-4 text-emerald-300" />
                    )}
                    {rooms.length} 个房间
                  </div>
                </div>

                <div className="space-y-3">
                  {rooms.length === 0 && (
                    <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                      当前还没有房间。最简单的试玩方式依然是直接点 AI 演示。
                    </div>
                  )}

                  {rooms.map((room) => {
                    const active = activeRoomId === room.id;
                    return (
                      <button
                        key={room.id}
                        type="button"
                        onClick={() => connectToRoom(room.id)}
                        className={cn(
                          "w-full rounded-[22px] border px-4 py-4 text-left transition-all",
                          active
                            ? "border-amber-300/35 bg-amber-300/10"
                            : "border-white/10 bg-black/20 hover:border-white/20 hover:bg-white/[0.04]",
                        )}
                      >
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-white">
                                {room.title}
                              </span>
                              <Badge className="border-white/15 bg-white/8 text-white">
                                {room.code}
                              </Badge>
                              <Badge
                                className={
                                  room.match_mode === "demo_ai"
                                    ? "border-amber-300/20 bg-amber-300/10 text-amber-100"
                                    : "border-sky-300/20 bg-sky-300/10 text-sky-100"
                                }
                              >
                                {room.match_mode === "demo_ai"
                                  ? "AI 演示"
                                  : "真人房"}
                              </Badge>
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              {room.player_count} / 3 人，状态{" "}
                              {roomStatusLabel(room.status)}
                            </div>
                          </div>
                          <div className="text-sm text-slate-300">
                            {active ? "已连接" : "进入房间"}
                          </div>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="matches">
          <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
            <Card className="border-white/10 bg-white/[0.04] text-white">
              <CardContent className="p-6">
                <div className="mb-5 flex items-center gap-3">
                  <Crown className="h-5 w-5 text-sky-300" />
                  <div>
                    <h3 className="text-2xl font-black tracking-tight text-white">
                      我的最近对局
                    </h3>
                    <p className="mt-1 text-sm text-slate-400">
                      打完就能直接回看，不再是“玩了但没有留下任何结果”。
                    </p>
                  </div>
                </div>
                <div className="space-y-3">
                  {user && myMatches.length === 0 && (
                    <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                      你还没有可展示的斗地主对局。
                    </div>
                  )}
                  {myMatches.map((match) => (
                    <MatchLinkCard
                      key={`me-${match.match_id}`}
                      match={match}
                      mine
                    />
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card className="border-white/10 bg-white/[0.04] text-white">
              <CardContent className="p-6">
                <div className="mb-5 flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <Swords className="h-5 w-5 text-amber-300" />
                    <div>
                      <h3 className="text-2xl font-black tracking-tight text-white">
                        全站最近对局
                      </h3>
                      <p className="mt-1 text-sm text-slate-400">
                        AI 演示和真人房都能形成战报。
                      </p>
                    </div>
                  </div>
                  {metaLoading && (
                    <Loader2 className="h-4 w-4 animate-spin text-sky-300" />
                  )}
                </div>

                <div className="space-y-3">
                  {recentMatches.length === 0 && (
                    <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                      还没有斗地主战报。
                    </div>
                  )}
                  {recentMatches.map((match) => (
                    <MatchLinkCard key={match.match_id} match={match} />
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="leaderboard">
          <Card className="border-white/10 bg-white/[0.04] text-white">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Trophy className="h-5 w-5 text-amber-300" />
                <div>
                  <h3 className="text-2xl font-black tracking-tight text-white">
                    排行榜
                  </h3>
                  <p className="mt-1 text-sm text-slate-400">
                    当前只统计真人对局，AI 演示房不会进入榜单。
                  </p>
                </div>
              </div>

              <div className="grid gap-3 lg:grid-cols-2">
                {leaderboard.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                    还没有可展示的真人对局结果。
                  </div>
                )}
                {leaderboard.map((entry) => (
                  <LeaderboardCard
                    key={`${entry.user_id ?? entry.player_name}-${entry.rank}`}
                    entry={entry}
                  />
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
