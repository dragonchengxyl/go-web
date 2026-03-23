'use client';

import Link from 'next/link';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  Bot,
  Crown,
  Loader2,
  Radio,
  RefreshCcw,
  Swords,
  Trophy,
  Users2,
} from 'lucide-react';
import { useAuth } from '@/contexts/auth-context';
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
} from '@/lib/api-client';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

type RoomWSStatus = 'idle' | 'connecting' | 'connected' | 'closed';

interface RoomServerMessage {
  type: string;
  payload: any;
}

function getGameWsBase() {
  if (typeof window === 'undefined') {
    return 'ws://localhost:8080';
  }

  const configured = process.env.NEXT_PUBLIC_WS_URL;
  if (configured) {
    return configured;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}`;
}

function createLocalSessionID() {
  if (typeof window === 'undefined') {
    return 'server-session';
  }
  if (window.crypto?.randomUUID) {
    return window.crypto.randomUUID();
  }
  return `session-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`;
}

function roomStatusLabel(status?: string) {
  switch (status) {
    case 'bidding':
      return '叫分中';
    case 'playing':
      return '对局中';
    case 'settlement':
      return '已结算';
    case 'redeal':
      return '流局重发';
    default:
      return '待准备';
  }
}

function seatLabel(seat?: number) {
  switch (seat) {
    case 0:
      return '一号位';
    case 1:
      return '二号位';
    case 2:
      return '三号位';
    default:
      return '--';
  }
}

function formatRemaining(target?: string, nowMs: number = Date.now()) {
  if (!target) {
    return '--';
  }
  const targetMs = new Date(target).getTime();
  const remaining = Math.max(0, targetMs - nowMs);
  return `${Math.ceil(remaining / 1000)}s`;
}

function cardKey(card: DoudizhuCard) {
  return `${card.suit}-${card.rank}`;
}

function cardLabel(card: DoudizhuCard) {
  const suitMap: Record<DoudizhuCard['suit'], string> = {
    spade: '♠',
    heart: '♥',
    club: '♣',
    diamond: '♦',
    joker: 'J',
  };
  const rankMap: Record<number, string> = {
    11: 'J',
    12: 'Q',
    13: 'K',
    14: 'A',
    15: '2',
    16: '小王',
    17: '大王',
  };

  return `${suitMap[card.suit]}${rankMap[card.rank] ?? String(card.rank)}`;
}

function comboLabel(combo?: DoudizhuCombo | null) {
  if (!combo) {
    return '暂无牌型';
  }
  const typeLabelMap: Record<DoudizhuCombo['type'], string> = {
    single: '单张',
    pair: '对子',
    triple: '三张',
    triple_with_single: '三带一',
    triple_with_pair: '三带二',
    straight: '顺子',
    straight_pairs: '连对',
    airplane: '飞机',
    airplane_with_single: '飞机带单',
    airplane_with_pair: '飞机带对',
    four_with_two_single: '四带二',
    four_with_two_pair: '四带两对',
    bomb: '炸弹',
    rocket: '王炸',
  };
  return `${typeLabelMap[combo.type]} · 主值 ${combo.main_rank}`;
}

function playerSort(a: DoudizhuRoomPlayer, b: DoudizhuRoomPlayer) {
  return a.seat - b.seat;
}

export function DouDizhuPlayStage() {
  const { user } = useAuth();
  const [sessionId, setSessionId] = useState('');
  const [playerName, setPlayerName] = useState('');
  const [roomTitle, setRoomTitle] = useState('周末牌局');
  const [rooms, setRooms] = useState<DoudizhuRoom[]>([]);
  const [roomsLoading, setRoomsLoading] = useState(false);
  const [metaLoading, setMetaLoading] = useState(false);
  const [activeRoom, setActiveRoom] = useState<DoudizhuRoom | null>(null);
  const [activeRoomId, setActiveRoomId] = useState<string | null>(null);
  const [privateState, setPrivateState] = useState<DoudizhuPrivateState | null>(null);
  const [latestAction, setLatestAction] = useState<DoudizhuActionResult | null>(null);
  const [selectedCards, setSelectedCards] = useState<string[]>([]);
  const [wsStatus, setWsStatus] = useState<RoomWSStatus>('idle');
  const [notice, setNotice] = useState('可以创建真人房，也可以直接开始 1 人 + 2 机器人演示。');
  const [errorMessage, setErrorMessage] = useState('');
  const [timeTick, setTimeTick] = useState(Date.now());
  const [leaderboard, setLeaderboard] = useState<DoudizhuLeaderboardEntry[]>([]);
  const [recentMatches, setRecentMatches] = useState<DoudizhuMatchSummary[]>([]);
  const [myMatches, setMyMatches] = useState<DoudizhuMatchSummary[]>([]);

  const wsRef = useRef<WebSocket | null>(null);
  const activeRoomIdRef = useRef<string | null>(null);
  const reconnectEnabledRef = useRef(false);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    let storedSessionID = localStorage.getItem('doudizhu_session_id');
    if (!storedSessionID) {
      storedSessionID = createLocalSessionID();
      localStorage.setItem('doudizhu_session_id', storedSessionID);
    }
    setSessionId(storedSessionID);

    const storedPlayerName = localStorage.getItem('doudizhu_player_name');
    if (storedPlayerName) {
      setPlayerName(storedPlayerName);
    } else if (user?.username) {
      setPlayerName(user.username);
    } else {
      setPlayerName('牌桌玩家');
    }
  }, [user?.username]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    if (playerName.trim()) {
      localStorage.setItem('doudizhu_player_name', playerName.trim());
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
            const current = data.rooms.find((item) => item.id === activeRoomIdRef.current);
            if (current) {
              setActiveRoom(current);
            }
          }
        }
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(error instanceof Error ? error.message : '加载房间失败');
        }
      } finally {
        if (!cancelled) {
          setRoomsLoading(false);
        }
      }
    }

    void loadRooms();
    const timer = window.setInterval(() => {
      void loadRooms();
    }, activeRoomId ? 8000 : 5000);

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
        const [leaderboardData, matchesData, myMatchesData] = await Promise.all([
          apiClient.getDoudizhuLeaderboard(10),
          apiClient.getDoudizhuRecentMatches(6),
          user ? apiClient.getMyDoudizhuRecentMatches(4) : Promise.resolve({ matches: [] }),
        ]);
        if (!cancelled) {
          setLeaderboard(leaderboardData.entries);
          setRecentMatches(matchesData.matches);
          setMyMatches(myMatchesData.matches);
        }
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(error instanceof Error ? error.message : '加载战报失败');
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
    setSelectedCards((current) =>
      current.filter((item) => privateState.hand.some((card) => cardKey(card) === item))
    );
  }, [privateState]);

  function updateActiveRoom(room: DoudizhuRoom) {
    const nextPlayers = [...room.players].sort(playerSort);
    const nextRoom = { ...room, players: nextPlayers };
    setActiveRoom(nextRoom);
    setActiveRoomId(nextRoom.id);
    activeRoomIdRef.current = nextRoom.id;

    switch (nextRoom.status) {
      case 'bidding':
        setNotice('正在叫分，所有叫分和地主裁决都由服务端处理。');
        break;
      case 'playing':
        setNotice(
          nextRoom.match_mode === 'demo_ai'
            ? '人机演示局进行中，机器人正在按基础策略出牌。'
            : '真人房进行中，当前由服务端统一裁决出牌合法性与轮转。'
        );
        break;
      case 'settlement':
        setNotice('本局已结算，可以继续留在房间准备下一局。');
        break;
      case 'redeal':
        setNotice('这一轮没有确定地主，服务端会重新开始新一轮。');
        break;
      default:
        setNotice(
          nextRoom.match_mode === 'demo_ai'
            ? '这是人机演示房。你准备后可以由房主直接开始，系统会补两名机器人。'
            : '房间已连接。凑齐 3 名真人并全部准备后，由房主开始一局。'
        );
    }
  }

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
      setErrorMessage('本地 session 还未准备好，请稍后重试。');
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage('请先填写玩家名称。');
      return;
    }

    closeSocket(false);
    setErrorMessage('');
    setWsStatus('connecting');
    setNotice('正在连接房间...');
    activeRoomIdRef.current = roomId;
    setActiveRoomId(roomId);
    reconnectEnabledRef.current = true;
    setPrivateState(null);
    setLatestAction(null);
    setSelectedCards([]);

    const query = new URLSearchParams({
      room_id: roomId,
      session_id: sessionId,
      player_name: playerName.trim(),
    });
    const token = apiClient.getToken();
    if (token) {
      query.set('token', token);
    }

    const ws = new WebSocket(`${getGameWsBase()}/ws/game/dou-dizhu?${query.toString()}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setWsStatus('connected');
    };

    ws.onmessage = (event) => {
      try {
        const message: RoomServerMessage = JSON.parse(event.data);
        switch (message.type) {
          case 'joined':
            if (message.payload?.room) {
              updateActiveRoom(message.payload.room as DoudizhuRoom);
            }
            if (message.payload?.private_state) {
              setPrivateState(message.payload.private_state as DoudizhuPrivateState);
            }
            if (message.payload?.session_id && typeof message.payload.session_id === 'string') {
              setSessionId(message.payload.session_id);
              if (typeof window !== 'undefined') {
                localStorage.setItem('doudizhu_session_id', message.payload.session_id);
              }
            }
            break;
          case 'room_state':
            updateActiveRoom(message.payload as DoudizhuRoom);
            break;
          case 'private_state':
            setPrivateState(message.payload as DoudizhuPrivateState);
            break;
          case 'action_result':
            setLatestAction(message.payload as DoudizhuActionResult);
            if (typeof message.payload?.message === 'string') {
              setNotice(message.payload.message);
            }
            break;
          case 'room_closed':
            setNotice('房间已关闭。');
            setActiveRoom(null);
            setActiveRoomId(null);
            setPrivateState(null);
            setLatestAction(null);
            setSelectedCards([]);
            activeRoomIdRef.current = null;
            reconnectEnabledRef.current = false;
            setWsStatus('idle');
            break;
          case 'error':
            setErrorMessage(
              typeof message.payload?.message === 'string'
                ? message.payload.message
                : '房间操作失败'
            );
            break;
          default:
            break;
        }
      } catch {
        setErrorMessage('房间消息解析失败');
      }
    };

    ws.onerror = () => {
      setErrorMessage('房间连接异常，请稍后重试。');
    };

    ws.onclose = () => {
      setWsStatus('closed');
      wsRef.current = null;

      if (!reconnectEnabledRef.current || !activeRoomIdRef.current) {
        setActiveRoom(null);
        setActiveRoomId(null);
        setPrivateState(null);
        setLatestAction(null);
        setSelectedCards([]);
        setNotice('你已离开房间。');
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
      setErrorMessage('房间连接尚未建立');
      return;
    }
    wsRef.current.send(JSON.stringify({ type, payload }));
  }

  async function handleCreateRoom() {
    if (!sessionId) {
      setErrorMessage('本地 session 还未准备好');
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage('请先填写玩家名称');
      return;
    }

    setErrorMessage('');
    try {
      const room = await apiClient.createDoudizhuRoom({
        title: roomTitle.trim() || '周末牌局',
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      connectToRoom(room.id);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '创建房间失败');
    }
  }

  async function handleCreateDemoRoom() {
    if (!sessionId) {
      setErrorMessage('本地 session 还未准备好');
      return;
    }
    if (!playerName.trim()) {
      setErrorMessage('请先填写玩家名称');
      return;
    }

    setErrorMessage('');
    try {
      const room = await apiClient.createDoudizhuDemoRoom({
        title: roomTitle.trim() || '单人演示房',
        player_name: playerName.trim(),
        session_id: sessionId,
      });
      setRooms((current) => {
        const next = current.filter((item) => item.id !== room.id);
        return [room, ...next];
      });
      connectToRoom(room.id);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '创建演示房失败');
    }
  }

  const me =
    activeRoom?.players.find((player) => player.session_id === sessionId) ?? null;
  const isHost = !!me?.is_host;
  const canStart =
    !!activeRoom &&
    activeRoom.status !== 'playing' &&
    activeRoom.status !== 'bidding' &&
    activeRoom.players
      .filter((player) => !player.is_bot)
      .every((player) => player.ready && player.connected);

  const isMyBidTurn =
    !!activeRoom &&
    !!me &&
    activeRoom.status === 'bidding' &&
    activeRoom.current_bidder === me.seat;
  const isMyPlayTurn =
    !!activeRoom &&
    !!me &&
    activeRoom.status === 'playing' &&
    activeRoom.current_turn === me.seat;

  const selectedHandCards = useMemo(() => {
    if (!privateState) {
      return [];
    }
    return privateState.hand.filter((card) => selectedCards.includes(cardKey(card)));
  }, [privateState, selectedCards]);

  useEffect(() => {
    if (!activeRoom || activeRoom.status !== 'settlement') {
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
      <div className="grid gap-6 xl:grid-cols-[0.96fr_1.04fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center gap-3">
              <Swords className="h-5 w-5 text-amber-300" />
              <div>
                <p className="text-xs uppercase tracking-[0.28em] text-slate-500">
                  Room Entry
                </p>
                <h2 className="mt-2 text-2xl font-black tracking-tight text-white">
                  大厅与演示入口
                </h2>
              </div>
            </div>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="ddz-player-name" className="text-white/85">
                  玩家名称
                </Label>
                <Input
                  id="ddz-player-name"
                  value={playerName}
                  onChange={(event) => setPlayerName(event.target.value)}
                  className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                  placeholder="输入一个牌桌显示名称"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="ddz-room-title" className="text-white/85">
                  房间标题
                </Label>
                <Input
                  id="ddz-room-title"
                  value={roomTitle}
                  onChange={(event) => setRoomTitle(event.target.value)}
                  className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                  placeholder="例如：周五加练场"
                />
              </div>

              <div className="flex flex-wrap gap-3">
                <Button
                  onClick={handleCreateRoom}
                  className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                >
                  创建真人房
                </Button>
                <Button
                  onClick={handleCreateDemoRoom}
                  variant="outline"
                  className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                >
                  <Bot className="mr-2 h-4 w-4" />
                  快速 AI 演示
                </Button>
                <Button
                  onClick={() => {
                    setRoomsLoading(true);
                    apiClient
                      .getDoudizhuRooms()
                      .then((data) => setRooms(data.rooms))
                      .catch((error) =>
                        setErrorMessage(
                          error instanceof Error ? error.message : '刷新房间失败'
                        )
                      )
                      .finally(() => setRoomsLoading(false));
                  }}
                  variant="outline"
                  className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                >
                  <RefreshCcw className="mr-2 h-4 w-4" />
                  刷新房间
                </Button>
              </div>

              <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-7 text-slate-300">
                真人房模式用于 3 名登录用户联机；快速 AI 演示会自动补齐两名机器人，
                方便录屏、面试和单人验证整条链路。
              </div>

              {errorMessage && (
                <div className="rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
                  {errorMessage}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <Users2 className="h-5 w-5 text-sky-300" />
                <div>
                  <h2 className="text-2xl font-black tracking-tight">房间列表</h2>
                  <p className="mt-1 text-sm text-slate-400">
                    加入一个正在等待中的房间，或者直接创建新的真人房 / 演示房。
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
                  当前还没有房间。你可以自己建一个真人房，也可以直接点“快速 AI 演示”。
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
                      'w-full rounded-[22px] border px-4 py-4 text-left transition-all',
                      active
                        ? 'border-amber-300/35 bg-amber-300/10'
                        : 'border-white/10 bg-black/20 hover:border-white/20 hover:bg-white/[0.04]'
                    )}
                  >
                    <div className="flex flex-wrap items-start justify-between gap-4">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-lg font-semibold text-white">{room.title}</span>
                          <Badge className="border-white/15 bg-white/8 text-white">
                            {room.code}
                          </Badge>
                          <Badge
                            className={cn(
                              'border-white/15 bg-white/8 text-white',
                              room.match_mode === 'demo_ai' &&
                                'border-amber-300/20 bg-amber-300/10 text-amber-100'
                            )}
                          >
                            {room.match_mode === 'demo_ai' ? 'AI 演示' : '真人房'}
                          </Badge>
                        </div>
                        <div className="mt-2 text-sm text-slate-400">
                          {room.player_count} / 3 人，已准备 {room.ready_count} 人，当前状态{' '}
                          {roomStatusLabel(room.status)}
                        </div>
                      </div>

                      <div className="text-sm text-slate-300">{active ? '已连接' : '进入房间'}</div>
                    </div>
                  </button>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="border-white/10 bg-white/[0.04] text-white">
        <CardContent className="p-6">
          <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-[0.28em] text-slate-500">
                Live Table
              </p>
              <h2 className="mt-2 text-2xl font-black tracking-tight">
                {activeRoom ? activeRoom.title : '尚未进入房间'}
              </h2>
              <p className="mt-2 text-sm text-slate-400">{notice}</p>
            </div>
            <div className="flex flex-wrap gap-2">
              <Badge className="border-white/15 bg-white/8 text-white">WS: {wsStatus}</Badge>
              {activeRoom && (
                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                  {roomStatusLabel(activeRoom.status)}
                </Badge>
              )}
            </div>
          </div>

          {!activeRoom && (
            <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
              先进入一个房间，这里会显示座位、叫分、出牌区、手牌和最近操作。
            </div>
          )}

          {activeRoom && (
            <div className="space-y-6">
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                  <div className="text-xs uppercase tracking-[0.24em] text-slate-500">房间码</div>
                  <div className="mt-2 text-xl font-black tracking-tight">{activeRoom.code}</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                  <div className="text-xs uppercase tracking-[0.24em] text-slate-500">模式</div>
                  <div className="mt-2 text-xl font-black tracking-tight">
                    {activeRoom.match_mode === 'demo_ai' ? 'AI 演示' : '真人房'}
                  </div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                  <div className="text-xs uppercase tracking-[0.24em] text-slate-500">最高叫分</div>
                  <div className="mt-2 text-xl font-black tracking-tight">{activeRoom.highest_bid}</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                  <div className="text-xs uppercase tracking-[0.24em] text-slate-500">当前计时</div>
                  <div className="mt-2 text-xl font-black tracking-tight">
                    {formatRemaining(
                      privateState?.turn_expires_at ?? activeRoom.turn_expires_at,
                      timeTick
                    )}
                  </div>
                </div>
              </div>

              <div className="flex flex-wrap gap-3">
                <Button
                  onClick={() => sendRoomMessage('set_ready', { ready: me ? !me.ready : true })}
                  disabled={!me || wsStatus !== 'connected' || activeRoom.status === 'playing' || activeRoom.status === 'bidding'}
                  className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110 disabled:opacity-50"
                >
                  {me?.ready ? '取消准备' : '准备就绪'}
                </Button>
                <Button
                  onClick={() => sendRoomMessage('start_round')}
                  disabled={!isHost || !canStart || wsStatus !== 'connected'}
                  variant="outline"
                  className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                >
                  房主开始
                </Button>
                <Button
                  onClick={() =>
                    sendRoomMessage('toggle_auto_play', { enabled: !(me?.auto_play ?? false) })
                  }
                  disabled={!me || wsStatus !== 'connected'}
                  variant="outline"
                  className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                >
                  {me?.auto_play ? '关闭托管' : '开启托管'}
                </Button>
                <Button
                  onClick={() => {
                    reconnectEnabledRef.current = false;
                    activeRoomIdRef.current = null;
                    sendRoomMessage('leave_room');
                  }}
                  disabled={activeRoom.status === 'playing' || activeRoom.status === 'bidding'}
                  variant="outline"
                  className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                >
                  离开房间
                </Button>
              </div>

              <div className="grid gap-6 xl:grid-cols-[0.96fr_1.04fr]">
                <div className="space-y-4">
                  <div className="grid gap-3">
                    {[...activeRoom.players].sort(playerSort).map((player) => {
                      const isMe = player.session_id === sessionId;
                      const isCurrentTurn = activeRoom.current_turn === player.seat;
                      const isCurrentBidder = activeRoom.current_bidder === player.seat;
                      const isLandlord = activeRoom.landlord === player.seat;

                      return (
                        <div
                          key={player.session_id}
                          className={cn(
                            'rounded-[22px] border px-4 py-4',
                            isMe
                              ? 'border-amber-300/30 bg-amber-300/10'
                              : 'border-white/10 bg-black/20'
                          )}
                        >
                          <div className="flex flex-wrap items-start justify-between gap-4">
                            <div>
                              <div className="flex flex-wrap items-center gap-2">
                                <span className="font-semibold text-white">{player.name}</span>
                                <Badge className="border-white/15 bg-white/8 text-white">
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
                                {isCurrentTurn && (
                                  <Badge className="border-emerald-300/20 bg-emerald-300/10 text-emerald-100">
                                    当前出牌
                                  </Badge>
                                )}
                                {isCurrentBidder && activeRoom.status === 'bidding' && (
                                  <Badge className="border-emerald-300/20 bg-emerald-300/10 text-emerald-100">
                                    当前叫分
                                  </Badge>
                                )}
                              </div>
                              <div className="mt-2 text-sm text-slate-400">
                                手牌 {player.card_count} 张 · {player.connected ? '在线' : '离线'} ·{' '}
                                {player.ready ? '已准备' : '未准备'}
                              </div>
                            </div>

                            <div className="text-right text-sm text-slate-300">
                              <div>{player.role === 'landlord' ? '地主阵营' : '农民阵营'}</div>
                              <div className="mt-1">{player.auto_play ? '托管中' : '手动操作'}</div>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  {latestAction && (
                    <div className="rounded-2xl border border-sky-300/15 bg-sky-300/10 px-4 py-4 text-sm leading-7 text-sky-50">
                      最近结果：{latestAction.message ?? `${latestAction.actor_name} 完成了一次操作`}。
                    </div>
                  )}
                </div>

                <div className="space-y-4">
                  <div className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                          Last Play
                        </div>
                        <div className="mt-2 text-lg font-semibold text-white">
                          {comboLabel(privateState?.last_play ?? activeRoom.last_play)}
                        </div>
                        <div className="mt-2 text-sm text-slate-400">
                          上一手座位：{seatLabel(privateState?.last_play_seat ?? activeRoom.last_play_seat)}
                        </div>
                      </div>
                      {privateState?.bottom_cards?.length ? (
                        <div className="text-right">
                          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                            Bottom
                          </div>
                          <div className="mt-2 flex flex-wrap justify-end gap-2">
                            {privateState.bottom_cards.map((card) => (
                              <div
                                key={cardKey(card)}
                                className="rounded-lg border border-white/10 bg-white/10 px-2 py-1 text-sm text-white"
                              >
                                {cardLabel(card)}
                              </div>
                            ))}
                          </div>
                        </div>
                      ) : null}
                    </div>
                  </div>

                  {isMyBidTurn && (
                    <div className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4">
                      <div className="mb-3 text-lg font-semibold text-white">轮到你叫分</div>
                      <div className="flex flex-wrap gap-3">
                        {[1, 2, 3].map((score) => (
                          <Button
                            key={score}
                            onClick={() => sendRoomMessage('bid', { score })}
                            disabled={score <= activeRoom.highest_bid}
                            className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
                          >
                            叫 {score} 分
                          </Button>
                        ))}
                        <Button
                          onClick={() => sendRoomMessage('pass_bid')}
                          variant="outline"
                          className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                        >
                          不叫
                        </Button>
                      </div>
                    </div>
                  )}

                  {privateState && (
                    <div className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4">
                      <div className="mb-3 flex items-center justify-between gap-3">
                        <div>
                          <div className="text-lg font-semibold text-white">我的手牌</div>
                          <div className="mt-1 text-sm text-slate-400">
                            角色：{privateState.role === 'landlord' ? '地主' : '农民'} · 共{' '}
                            {privateState.hand.length} 张
                          </div>
                        </div>
                        <div className="text-sm text-slate-400">
                          已选 {selectedCards.length} 张
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        {privateState.hand.map((card) => {
                          const selected = selectedCards.includes(cardKey(card));
                          return (
                            <button
                              key={cardKey(card)}
                              type="button"
                              onClick={() =>
                                setSelectedCards((current) =>
                                  current.includes(cardKey(card))
                                    ? current.filter((item) => item !== cardKey(card))
                                    : [...current, cardKey(card)]
                                )
                              }
                              className={cn(
                                'rounded-xl border px-3 py-2 text-sm font-medium transition-all',
                                selected
                                  ? 'border-amber-300/40 bg-amber-300/12 text-amber-50 -translate-y-1'
                                  : 'border-white/10 bg-white/5 text-white hover:border-white/20 hover:bg-white/10'
                              )}
                            >
                              {cardLabel(card)}
                            </button>
                          );
                        })}
                      </div>

                      <div className="mt-4 flex flex-wrap gap-3">
                        <Button
                          onClick={() => sendRoomMessage('play_cards', { cards: selectedHandCards })}
                          disabled={!isMyPlayTurn || selectedHandCards.length === 0}
                          className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110 disabled:opacity-50"
                        >
                          出牌
                        </Button>
                        <Button
                          onClick={() => sendRoomMessage('pass_turn')}
                          disabled={!isMyPlayTurn || !privateState.can_pass}
                          variant="outline"
                          className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                        >
                          过牌
                        </Button>
                        <Button
                          onClick={() => setSelectedCards([])}
                          variant="outline"
                          className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                        >
                          清空选择
                        </Button>
                      </div>
                    </div>
                  )}

                  <div className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4">
                    <div className="mb-3 text-lg font-semibold text-white">最近操作</div>
                    <div className="space-y-2">
                      {(activeRoom.recent_actions ?? []).length === 0 && (
                        <div className="text-sm text-slate-400">还没有操作记录。</div>
                      )}
                      {(activeRoom.recent_actions ?? []).slice().reverse().map((action, index) => (
                        <div
                          key={`${action.at}-${index}`}
                          className="rounded-2xl border border-white/10 bg-white/5 px-3 py-3 text-sm text-slate-300"
                        >
                          <div className="font-medium text-white">
                            {action.actor_name ?? seatLabel(action.seat)}
                          </div>
                          <div className="mt-1">{action.message ?? action.action_type}</div>
                          {action.cards?.length ? (
                            <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-400">
                              {action.cards.map((card) => (
                                <span key={cardKey(card)}>{cardLabel(card)}</span>
                              ))}
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center gap-3">
              <Trophy className="h-5 w-5 text-amber-300" />
              <div>
                <h2 className="text-2xl font-black tracking-tight">排行榜</h2>
                <p className="mt-1 text-sm text-slate-400">
                  当前只统计真人对局，人机演示房不会进入榜单。
                </p>
              </div>
            </div>

            <div className="space-y-3">
              {leaderboard.length === 0 && (
                <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                  还没有可展示的真人对局结果。
                </div>
              )}
              {leaderboard.map((entry) => (
                <div
                  key={`${entry.rank}-${entry.display_name}`}
                  className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                >
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <div className="text-sm text-slate-400">#{entry.rank}</div>
                      <div className="mt-1 text-lg font-semibold text-white">{entry.display_name}</div>
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
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <Crown className="h-5 w-5 text-sky-300" />
                <div>
                  <h2 className="text-2xl font-black tracking-tight">最近对局</h2>
                  <p className="mt-1 text-sm text-slate-400">
                    可以直接进入战报页查看完整操作记录。
                  </p>
                </div>
              </div>
              {metaLoading && <Loader2 className="h-4 w-4 animate-spin text-sky-300" />}
            </div>

            <div className="space-y-4">
              {user && myMatches.length > 0 && (
                <div className="space-y-3">
                  <div className="text-sm font-medium text-sky-200">我的最近对局</div>
                  {myMatches.map((match) => (
                    <Link
                      key={`me-${match.match_id}`}
                      href={`/games/dou-dizhu/matches/${match.match_id}`}
                      className="block rounded-[22px] border border-sky-300/15 bg-sky-300/10 px-4 py-4 transition-colors hover:border-sky-300/30 hover:bg-sky-300/14"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-4">
                        <div>
                          <div className="font-semibold text-white">{match.room_title}</div>
                          <div className="mt-1 text-sm text-slate-300">
                            {new Date(match.finished_at).toLocaleString('zh-CN')} · 倍率 x
                            {match.multiplier}
                          </div>
                        </div>
                        <Badge className="border-white/15 bg-white/8 text-white">
                          {match.room_code}
                        </Badge>
                      </div>
                    </Link>
                  ))}
                </div>
              )}

              <div className="space-y-3">
                <div className="text-sm font-medium text-white">全站最近对局</div>
                {recentMatches.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                    还没有斗地主战报。
                  </div>
                )}
                {recentMatches.map((match) => (
                  <Link
                    key={match.match_id}
                    href={`/games/dou-dizhu/matches/${match.match_id}`}
                    className="block rounded-[22px] border border-white/10 bg-black/20 px-4 py-4 transition-colors hover:border-white/20 hover:bg-white/[0.05]"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-4">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-white">{match.room_title}</span>
                          <Badge className="border-white/15 bg-white/8 text-white">
                            {match.match_mode === 'demo_ai' ? 'AI 演示' : '真人房'}
                          </Badge>
                        </div>
                        <div className="mt-1 text-sm text-slate-400">
                          {new Date(match.finished_at).toLocaleString('zh-CN')} · 地主位{' '}
                          {seatLabel(match.landlord_seat)} · 倍率 x{match.multiplier}
                        </div>
                      </div>

                      <div className="text-right text-sm text-slate-300">
                        <div>{match.winner_side === 'landlord' ? '地主胜' : '农民胜'}</div>
                        <div className="mt-1">{match.player_count} 名玩家</div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
