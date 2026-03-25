"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
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
import {
  DoudizhuLeaderboardCard,
  DoudizhuMatchLinkCard,
} from "@/components/games/doudizhu/doudizhu-history-panel";
import { DoudizhuLobbyPanel } from "@/components/games/doudizhu/doudizhu-lobby-panel";
import {
  DoudizhuDisplayCard,
  DoudizhuPlayCard,
} from "@/components/games/doudizhu/doudizhu-play-card";
import { DoudizhuSeat } from "@/components/games/doudizhu/doudizhu-seat";
import { cn } from "@/lib/utils";
import { cardKey, cardLabel } from "@/lib/games/doudizhu/cards";
import {
  comboLabel,
  evaluateSelectedCombo,
} from "@/lib/games/doudizhu/combo";
import {
  actionTypeLabel,
  buildRoomNotice,
  connectionStatusLabel,
  DOUDIZHU_LOBBY_NAME,
  DOUDIZHU_PRODUCT_NAME,
  formatRemaining,
  roleLabel,
  roomPagePath,
  roomModeLabel,
  roomStatusLabel,
  seatLabel,
} from "@/lib/games/doudizhu/presenter";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
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
  if (!bottom) {
    return {
      bottom: null as DoudizhuRoomPlayer | null,
      left: null as DoudizhuRoomPlayer | null,
      right: null as DoudizhuRoomPlayer | null,
    };
  }

  const seatMap = new Map<number, DoudizhuRoomPlayer>(
    players.map((player) => [player.seat, player]),
  );

  return {
    bottom,
    left: seatMap.get((bottom.seat + 1) % 3) ?? null,
    right: seatMap.get((bottom.seat + 2) % 3) ?? null,
  };
}

interface DouDizhuPlayStageProps {
  immersive?: boolean;
  fixedRoomId?: string;
}

export function DouDizhuPlayStage({
  immersive = false,
  fixedRoomId,
}: DouDizhuPlayStageProps = {}) {
  const { user } = useAuth();
  const router = useRouter();
  const [sessionId, setSessionId] = useState("");
  const [playerName, setPlayerName] = useState("");
  const [roomTitle, setRoomTitle] = useState("涂油牌局");
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
  const [notice, setNotice] = useState("直接开始一局人机热身，或者拉起一桌三人牌局。");
  const [errorMessage, setErrorMessage] = useState("");
  const [timeTick, setTimeTick] = useState(Date.now());
  const [infoDialogOpen, setInfoDialogOpen] = useState(false);
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
  const reconnectAttemptsRef = useRef(0);
  const audioContextRef = useRef<AudioContext | null>(null);
  const lastSettledRoomRef = useRef<string | null>(null);
  const connectToRoomRef = useRef<
    (roomId: string, isReconnect?: boolean) => void
  >(() => {});
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
      setPlayerName("涂油牌手");
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
    if (immersive) {
      return;
    }
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
  }, [activeRoomId, immersive]);

  useEffect(() => {
    if (immersive) {
      return;
    }
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
  }, [user, immersive]);

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
    if (!immersive || !fixedRoomId || !sessionId || !playerName.trim()) {
      return;
    }
    if (activeRoomIdRef.current === fixedRoomId && wsRef.current) {
      return;
    }
    connectToRoomRef.current(fixedRoomId);
  }, [fixedRoomId, immersive, playerName, sessionId]);

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

  function navigateToRoomPage(roomId: string) {
    router.push(roomPagePath(roomId));
  }

  function navigateToLobby() {
    router.push("/games/dou-dizhu/play");
  }

  function updateActiveRoom(room: DoudizhuRoom) {
    const nextPlayers = [...room.players].sort(playerSort);
    const nextRoom = { ...room, players: nextPlayers };
    setActiveRoom(nextRoom);
    setActiveRoomId(nextRoom.id);
    activeRoomIdRef.current = nextRoom.id;
    setNotice(buildRoomNotice(nextRoom));
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

  function connectToRoom(roomId: string, isReconnect = false) {
    if (!sessionId) {
      setErrorMessage("本地 session 还未准备好，请稍后重试。");
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage("请先填写玩家名称。");
      return;
    }

    closeSocket(false);
    if (!isReconnect) {
      reconnectAttemptsRef.current = 0;
    }
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
      reconnectAttemptsRef.current = 0;
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
            if (immersive) {
              navigateToLobby();
            }
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
        if (immersive) {
          navigateToLobby();
        }
        return;
      }

      reconnectAttemptsRef.current += 1;
      if (reconnectAttemptsRef.current > 6) {
        reconnectEnabledRef.current = false;
        setWsStatus("idle");
        setErrorMessage(`牌桌连接已中断，请返回${DOUDIZHU_LOBBY_NAME}重新进入。`);
        setNotice("当前牌桌连接已断开，回到大厅后重新入桌更稳妥。");
        if (immersive) {
          navigateToLobby();
        }
        return;
      }

      const reconnectDelay = Math.min(
        1200 * 2 ** (reconnectAttemptsRef.current - 1),
        8000,
      );
      window.setTimeout(() => {
        if (activeRoomIdRef.current) {
          connectToRoom(activeRoomIdRef.current, true);
        }
      }, reconnectDelay);
    };
  }

  connectToRoomRef.current = connectToRoom;

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
        title: roomTitle.trim() || "涂油牌局",
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      if (immersive) {
        connectToRoom(room.id);
      } else {
        navigateToRoomPage(room.id);
      }
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
        title: roomTitle.trim() || "人机热身房",
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      if (immersive) {
        connectToRoom(room.id);
      } else {
        navigateToRoomPage(room.id);
      }
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "创建热身房失败",
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
  const useStackedHand = currentHand.length >= 11;
  const handOverlapClass =
    currentHand.length >= 20
      ? "-ml-12"
      : currentHand.length >= 17
        ? "-ml-10"
        : currentHand.length >= 14
          ? "-ml-8"
          : currentHand.length >= 11
            ? "-ml-6"
            : "";
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
    <div
      className={cn("space-y-6", immersive ? "min-h-[calc(100vh-48px)]" : "")}
    >
      <section
        className={cn(
          "relative overflow-hidden border border-white/10 bg-[#081013] text-white shadow-[0_30px_90px_-50px_rgba(0,0,0,0.85)]",
          immersive
            ? "min-h-[calc(100vh-48px)] rounded-[28px]"
            : "rounded-[36px]",
        )}
      >
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,207,120,0.15),transparent_28%),radial-gradient(circle_at_20%_20%,rgba(255,255,255,0.08),transparent_18%),linear-gradient(180deg,#0d382c_0%,#082019_100%)]" />
        <div className="relative p-5 md:p-8">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              {immersive ? (
                <>
                  <h2 className="text-3xl font-black tracking-tight text-white md:text-4xl">
                    {activeRoom?.title ?? "正在进入牌桌"}
                  </h2>
                </>
              ) : (
                <>
                  <p className="text-xs uppercase tracking-[0.32em] text-emerald-100/50">
                    Oil Table
                  </p>
                  <h2 className="mt-3 text-3xl font-black tracking-tight text-white md:text-5xl">
                    {DOUDIZHU_PRODUCT_NAME}
                  </h2>
                </>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <Badge className="border-white/15 bg-white/8 text-white">
                连线: {connectionStatusLabel(wsStatus)}
              </Badge>
              {activeRoom && (
                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                  {roomStatusLabel(activeRoom.status)}
                </Badge>
              )}
              {activeRoom && (
                <Badge className="border-white/15 bg-black/25 text-white">
                  {roomModeLabel(activeRoom.match_mode)}
                </Badge>
              )}
            </div>
          </div>

          {!activeRoom && !immersive && (
            <DoudizhuLobbyPanel
              playerName={playerName}
              roomTitle={roomTitle}
              errorMessage={errorMessage}
              onPlayerNameChange={setPlayerName}
              onRoomTitleChange={setRoomTitle}
              onCreateDemoRoom={handleCreateDemoRoom}
              onCreateRoom={handleCreateRoom}
            />
          )}

          {!activeRoom && immersive && (
            <div className="mt-8 rounded-[32px] border border-white/10 bg-black/20 px-6 py-12 text-center">
              <div className="mx-auto max-w-xl">
                <div className="text-xs uppercase tracking-[0.28em] text-emerald-100/45">
                  Room Loading
                </div>
                <div className="mt-3 text-3xl font-black tracking-tight text-white">
                  正在连接牌桌
                </div>
                <div className="mt-4 text-sm leading-7 text-emerald-50/75">
                  {fixedRoomId
                    ? `房间 ${fixedRoomId.slice(0, 8)} 已进入独立大屏模式，正在建立实时连接。`
                    : "正在准备独立牌桌页面。"}
                </div>
                {errorMessage && (
                  <div className="mt-6 rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
                    {errorMessage}
                  </div>
                )}
              </div>
            </div>
          )}

          {activeRoom && (
            <div className="mt-8 space-y-5">
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
                      {notice ? (
                        <div className="mt-2 text-sm leading-7 text-amber-50/80">
                          {notice}
                        </div>
                      ) : null}
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

                <div className="flex flex-wrap items-center justify-between gap-3 rounded-[26px] border border-white/10 bg-black/20 px-4 py-4">
                  <div className="flex flex-wrap gap-2">
                    {[
                      { label: "最高叫分", value: String(activeRoom.highest_bid) },
                      { label: "当前倍率", value: `x${activeRoom.multiplier}` },
                      { label: "我的角色", value: roleLabel(privateState?.role) },
                      { label: "剩余手牌", value: String(currentHand.length) },
                      {
                        label: "当前连线",
                        value: connectionStatusLabel(wsStatus),
                      },
                    ].map((item) => (
                      <div
                        key={item.label}
                        className="rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 text-sm text-white"
                      >
                        <span className="mr-2 text-white/45">{item.label}</span>
                        <span className="font-semibold">{item.value}</span>
                      </div>
                    ))}
                  </div>

                  <Dialog open={infoDialogOpen} onOpenChange={setInfoDialogOpen}>
                    <DialogTrigger asChild>
                      <Button
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                      >
                        查看牌局信息
                      </Button>
                    </DialogTrigger>
                    <DialogContent className="max-w-3xl border-white/10 bg-[#081013] text-white">
                      <DialogHeader>
                        <DialogTitle className="text-2xl font-black tracking-tight text-white">
                          牌局信息
                        </DialogTitle>
                        <DialogDescription className="text-sm leading-6 text-slate-400">
                          倍率细节和最近操作默认收起，避免长期占用桌面空间。
                        </DialogDescription>
                      </DialogHeader>

                      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                        <div className="rounded-[22px] border border-white/10 bg-black/25 p-4">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            最高叫分
                          </div>
                          <div className="mt-2 text-3xl font-black text-white">
                            {activeRoom.highest_bid}
                          </div>
                        </div>
                        <div className="rounded-[22px] border border-white/10 bg-black/25 p-4">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            当前倍率
                          </div>
                          <div className="mt-2 text-3xl font-black text-white">
                            x{activeRoom.multiplier}
                          </div>
                        </div>
                        <div className="rounded-[22px] border border-white/10 bg-black/25 p-4">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            我的角色
                          </div>
                          <div className="mt-2 text-3xl font-black text-white">
                            {roleLabel(privateState?.role)}
                          </div>
                        </div>
                        <div className="rounded-[22px] border border-white/10 bg-black/25 p-4">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            剩余手牌
                          </div>
                          <div className="mt-2 text-3xl font-black text-white">
                            {currentHand.length}
                          </div>
                        </div>
                        <div className="rounded-[22px] border border-white/10 bg-black/25 p-4">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            当前连线
                          </div>
                          <div className="mt-2 text-3xl font-black text-white">
                            {connectionStatusLabel(wsStatus)}
                          </div>
                        </div>
                      </div>

                      <div className="rounded-[24px] border border-white/10 bg-black/20 p-5">
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

                      <div className="rounded-[24px] border border-white/10 bg-black/20 p-5">
                        <div className="mb-3 text-lg font-semibold text-white">
                          最近操作
                        </div>
                        <div className="max-h-[320px] space-y-2 overflow-y-auto pr-1">
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
                    </DialogContent>
                  </Dialog>
                </div>

                <div className="relative overflow-hidden rounded-[38px] border border-white/10 bg-[linear-gradient(180deg,#0a3528_0%,#082119_100%)] shadow-[inset_0_1px_0_rgba(255,255,255,0.06),0_30px_90px_-40px_rgba(0,0,0,0.85)]">
                  <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,214,120,0.18),transparent_28%),radial-gradient(circle_at_center,rgba(255,255,255,0.08),transparent_24%),radial-gradient(circle_at_bottom,rgba(0,0,0,0.26),transparent_38%)]" />
                  <div className="relative p-4 md:p-6">
                    <div className="grid gap-4 lg:grid-cols-2 xl:hidden">
                      <DoudizhuSeat
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
                      <DoudizhuSeat
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
                          <DoudizhuSeat
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
                          <DoudizhuSeat
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
                          {notice ? (
                            <div className="mt-3 text-sm leading-7 text-emerald-50/75">
                              {notice}
                            </div>
                          ) : null}

                          {(privateState?.bottom_cards?.length ||
                            activeRoom.bottom_cards?.length) && (
                            <div className="mt-4">
                              <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                                地主底牌
                              </div>
                              <div className="mt-3 flex flex-wrap justify-center gap-3">
                                {(
                                  privateState?.bottom_cards ??
                                  activeRoom.bottom_cards ??
                                  []
                                ).map((card, index) => (
                                  <DoudizhuDisplayCard
                                    key={`${cardKey(card)}-${index}`}
                                    card={card}
                                  />
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
                              <div className="mt-4 flex flex-wrap justify-center gap-3">
                                {activeRoom.last_play_cards.map(
                                  (card, index) => (
                                    <div
                                      key={`${cardKey(card)}-${index}`}
                                      className={cn(
                                        "transition-all",
                                        tableEffect === "play" ||
                                          tableEffect === "bomb"
                                          ? "translate-y-0 scale-100"
                                          : "",
                                      )}
                                    >
                                      <DoudizhuDisplayCard card={card} />
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
                        <DoudizhuSeat
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
                          ? `角色：${roleLabel(privateState.role)} · 共 ${currentHand.length} 张`
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
                        <div
                          className={cn(
                            "flex min-w-max items-end pr-6",
                            useStackedHand ? "gap-0" : "gap-3",
                          )}
                        >
                          {currentHand.map((card, index) => {
                            const selected = selectedCards.includes(
                              cardKey(card),
                            );
                            return (
                              <div
                                key={cardKey(card)}
                                className={cn(
                                  "shrink-0 transition-all duration-200",
                                  useStackedHand && currentHand.length > 1
                                    ? index === 0
                                      ? ""
                                      : handOverlapClass
                                    : "",
                                  selected
                                    ? "z-30"
                                    : "z-10 hover:z-20",
                                  useStackedHand && currentHand.length >= 17
                                    ? "scale-[0.96] origin-bottom"
                                    : "",
                                )}
                              >
                                <DoudizhuPlayCard
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
                              </div>
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
          )}
        </div>
      </section>

      {!immersive && (
        <Tabs defaultValue="start" className="space-y-4">
          <TabsList className="border border-white/10 bg-black/20">
            <TabsTrigger
              value="start"
              className="data-[state=active]:bg-white data-[state=active]:text-slate-950"
            >
              {DOUDIZHU_LOBBY_NAME}
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
                        立刻上桌
                      </h3>
                      <p className="mt-1 text-sm text-slate-400">
                        房间能力还在，但现在退到辅助入口，不再压过真正的牌桌体验。
                      </p>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label
                        htmlFor="ddz-room-title-panel"
                        className="text-white/85"
                      >
                        牌局标题
                      </Label>
                      <Input
                        id="ddz-room-title-panel"
                        value={roomTitle}
                        onChange={(event) => setRoomTitle(event.target.value)}
                        className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                        placeholder="例如：夜场涂油局"
                      />
                    </div>

                    <div className="flex flex-wrap gap-3">
                      <Button
                        onClick={handleCreateDemoRoom}
                        className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                      >
                        <Bot className="mr-2 h-4 w-4" />
                        开始人机热身
                      </Button>
                      <Button
                        onClick={handleCreateRoom}
                        variant="outline"
                        className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                      >
                        <Users2 className="mr-2 h-4 w-4" />
                        创建三人牌局
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
                              牌桌列表
                            </h3>
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
                        当前还没有可加入的牌桌。最简单的试玩方式是先开一把人机热身。
                      </div>
                    )}

                    {rooms.map((room) => {
                      const active = activeRoomId === room.id;
                      return (
                        <button
                          key={room.id}
                          type="button"
                          onClick={() =>
                            immersive
                              ? connectToRoom(room.id)
                              : navigateToRoomPage(room.id)
                          }
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
                                  {roomModeLabel(room.match_mode)}
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
                        打完就能直接回看，不会再出现“这一把打完就蒸发”的情况。
                      </p>
                    </div>
                  </div>
                  <div className="space-y-3">
                    {user && myMatches.length === 0 && (
                      <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                        你还没有可展示的涂油牌局。
                      </div>
                    )}
                    {myMatches.map((match) => (
                      <DoudizhuMatchLinkCard
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
                          人机热身和三人联机都会留下战报。
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
                        还没有新的涂油战报。
                      </div>
                    )}
                    {recentMatches.map((match) => (
                      <DoudizhuMatchLinkCard
                        key={match.match_id}
                        match={match}
                      />
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
                      当前只统计三人联机，人机热身不会进入榜单。
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
                    <DoudizhuLeaderboardCard
                      key={`${entry.user_id ?? entry.player_name}-${entry.rank}`}
                      entry={entry}
                    />
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      )}
    </div>
  );
}
