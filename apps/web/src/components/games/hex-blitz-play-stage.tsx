'use client';

import { useEffect, useRef, useState } from 'react';
import {
  AlertCircle,
  ArrowRight,
  Crown,
  Loader2,
  Medal,
  Radio,
  RefreshCcw,
  Rocket,
  Signal,
  Trophy,
  Users2,
} from 'lucide-react';
import { useAuth } from '@/contexts/auth-context';
import {
  apiClient,
  HexBlitzLeaderboardEntry,
  HexBlitzMatchSummary,
  HexBlitzRoom,
  HexBlitzRoomPlayer,
} from '@/lib/api-client';
import { cn } from '@/lib/utils';
import { HexBlitzPrototype } from '@/components/games/hex-blitz-prototype';
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

function formatRemaining(targetMs?: number, nowMs: number = Date.now()) {
  if (!targetMs) {
    return '--';
  }
  const remaining = Math.max(0, targetMs - nowMs);
  return `${(remaining / 1000).toFixed(1)}s`;
}

function roomStatusLabel(room: HexBlitzRoom | null) {
  switch (room?.status) {
    case 'countdown':
      return '倒计时';
    case 'running':
      return '进行中';
    case 'finished':
      return '已结算';
    default:
      return '待准备';
  }
}

export function HexBlitzPlayStage() {
  const { user } = useAuth();
  const [sessionId, setSessionId] = useState('');
  const [playerName, setPlayerName] = useState('');
  const [roomTitle, setRoomTitle] = useState('好友训练房');
  const [rooms, setRooms] = useState<HexBlitzRoom[]>([]);
  const [roomsLoading, setRoomsLoading] = useState(false);
  const [metaLoading, setMetaLoading] = useState(false);
  const [activeRoom, setActiveRoom] = useState<HexBlitzRoom | null>(null);
  const [activeRoomId, setActiveRoomId] = useState<string | null>(null);
  const [wsStatus, setWsStatus] = useState<RoomWSStatus>('idle');
  const [notice, setNotice] = useState('可以先单机试玩，也可以创建房间进入多人实验室。');
  const [errorMessage, setErrorMessage] = useState('');
  const [timeTick, setTimeTick] = useState(Date.now());
  const [leaderboard, setLeaderboard] = useState<HexBlitzLeaderboardEntry[]>([]);
  const [recentMatches, setRecentMatches] = useState<HexBlitzMatchSummary[]>([]);
  const [myMatches, setMyMatches] = useState<HexBlitzMatchSummary[]>([]);

  const wsRef = useRef<WebSocket | null>(null);
  const activeRoomIdRef = useRef<string | null>(null);
  const reconnectEnabledRef = useRef(false);
  const lastReportedScoreRef = useRef(-1);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    let storedSessionID = localStorage.getItem('hex_blitz_session_id');
    if (!storedSessionID) {
      storedSessionID = createLocalSessionID();
      localStorage.setItem('hex_blitz_session_id', storedSessionID);
    }
    setSessionId(storedSessionID);

    const storedPlayerName = localStorage.getItem('hex_blitz_player_name');
    if (storedPlayerName) {
      setPlayerName(storedPlayerName);
    } else if (user?.username) {
      setPlayerName(user.username);
    } else {
      setPlayerName('游客玩家');
    }
  }, [user?.username]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    if (playerName.trim()) {
      localStorage.setItem('hex_blitz_player_name', playerName.trim());
    }
  }, [playerName]);

  useEffect(() => {
    let cancelled = false;

    async function loadRooms() {
      setRoomsLoading(true);
      try {
        const data = await apiClient.getHexBlitzRooms();
        if (!cancelled) {
          setRooms(data.rooms);
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
          apiClient.getHexBlitzLeaderboard(10),
          apiClient.getHexBlitzRecentMatches(6),
          user ? apiClient.getMyHexBlitzRecentMatches(4) : Promise.resolve({ matches: [] }),
        ]);
        if (!cancelled) {
          setLeaderboard(leaderboardData.entries);
          setRecentMatches(matchesData.matches);
          setMyMatches(myMatchesData.matches);
        }
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(error instanceof Error ? error.message : '加载榜单失败');
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
    }, 200);

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

  function updateActiveRoom(room: HexBlitzRoom) {
    setActiveRoom(room);
    setActiveRoomId(room.id);
    activeRoomIdRef.current = room.id;
    if (room.status === 'running') {
      setNotice('多人对局进行中，你的本地分数会实时同步到房间记分板。');
    } else if (room.status === 'countdown') {
      setNotice('房主已经开始倒计时，盘面会在开局时自动激活。');
    } else if (room.status === 'finished') {
      setNotice('多人对局已结算，可以再次准备开始下一局。');
    } else {
      setNotice('房间已连接。所有在线玩家准备后，由房主开始对局。');
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
      setErrorMessage('先填一个玩家名称，再进入房间。');
      return;
    }

    closeSocket(false);
    setErrorMessage('');
    setWsStatus('connecting');
    setNotice('正在连接房间...');
    activeRoomIdRef.current = roomId;
    setActiveRoomId(roomId);
    reconnectEnabledRef.current = true;
    lastReportedScoreRef.current = -1;

    const query = new URLSearchParams({
      room_id: roomId,
      session_id: sessionId,
      player_name: playerName.trim(),
    });
    const token = apiClient.getToken();
    if (token) {
      query.set('token', token);
    }

    const ws = new WebSocket(`${getGameWsBase()}/ws/game/hex-blitz?${query.toString()}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setWsStatus('connected');
    };

    ws.onmessage = (event) => {
      try {
        const message: RoomServerMessage = JSON.parse(event.data);
        if (message.type === 'joined' && message.payload?.room) {
          updateActiveRoom(message.payload.room as HexBlitzRoom);
          if (message.payload?.session_id && typeof message.payload.session_id === 'string') {
            setSessionId(message.payload.session_id);
            if (typeof window !== 'undefined') {
              localStorage.setItem('hex_blitz_session_id', message.payload.session_id);
            }
          }
          return;
        }
        if (message.type === 'room_state') {
          updateActiveRoom(message.payload as HexBlitzRoom);
          return;
        }
        if (message.type === 'room_closed') {
          setNotice('房间已关闭。');
          setActiveRoom(null);
          setActiveRoomId(null);
          activeRoomIdRef.current = null;
          setWsStatus('idle');
          reconnectEnabledRef.current = false;
          return;
        }
        if (message.type === 'error') {
          const nextError =
            typeof message.payload?.message === 'string'
              ? message.payload.message
              : '房间操作失败';
          setErrorMessage(nextError);
          return;
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
      const room = await apiClient.createHexBlitzRoom({
        title: roomTitle.trim() || '好友训练房',
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

  const me =
    activeRoom?.players.find((player) => player.session_id === sessionId) ?? null;
  const isHost = !!me?.is_host;
  const canStart =
    !!activeRoom &&
    activeRoom.status === 'waiting' &&
    activeRoom.players.filter((player) => player.connected).length > 0 &&
    activeRoom.players
      .filter((player) => player.connected)
      .every((player) => player.ready);

  const roomMode = activeRoom
    ? {
        enabled: true,
        phase: activeRoom.status,
        matchKey:
          activeRoom.started_at ??
          activeRoom.countdown_started_at ??
          activeRoom.updated_at,
        endsAt: activeRoom.ends_at,
        infoText: notice,
      }
    : undefined;

  useEffect(() => {
    if (activeRoom?.status !== 'finished') {
      return;
    }

    const timer = window.setTimeout(() => {
      apiClient
        .getHexBlitzLeaderboard(10)
        .then((data) => setLeaderboard(data.entries))
        .catch(() => {});
      apiClient
        .getHexBlitzRecentMatches(6)
        .then((data) => setRecentMatches(data.matches))
        .catch(() => {});
      if (user) {
        apiClient
          .getMyHexBlitzRecentMatches(4)
          .then((data) => setMyMatches(data.matches))
          .catch(() => {});
      }
    }, 500);

    return () => window.clearTimeout(timer);
  }, [activeRoom?.id, activeRoom?.status, user]);

  return (
    <div className="space-y-6">
      <HexBlitzPrototype
        roomMode={roomMode}
        onScoreChange={(score) => {
          if (!activeRoom || activeRoom.status !== 'running') {
            return;
          }
          if (lastReportedScoreRef.current === score) {
            return;
          }
          lastReportedScoreRef.current = score;
          sendRoomMessage('score_update', { score });
        }}
      />

      <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center gap-3">
              <Rocket className="h-5 w-5 text-amber-300" />
              <div>
                <h2 className="text-2xl font-black tracking-tight">Room Lab</h2>
                <p className="mt-1 text-sm text-slate-400">
                  Phase 2 的目标是把“休闲原型”接成“多人实验室”。
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="player-name" className="text-white/85">
                  玩家名称
                </Label>
                <Input
                  id="player-name"
                  value={playerName}
                  onChange={(event) => setPlayerName(event.target.value)}
                  className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                  placeholder="输入一个房间内显示的名字"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="room-title" className="text-white/85">
                  房间标题
                </Label>
                <Input
                  id="room-title"
                  value={roomTitle}
                  onChange={(event) => setRoomTitle(event.target.value)}
                  className="border-white/10 bg-black/20 text-white placeholder:text-slate-500"
                  placeholder="例如：周末练习赛"
                />
              </div>

              <div className="flex flex-wrap gap-3">
                <Button
                  onClick={handleCreateRoom}
                  className="border-0 bg-[linear-gradient(135deg,#ff8a3d_0%,#34d2ff_100%)] text-slate-950 hover:brightness-110"
                >
                  创建房间
                </Button>
                <Button
                  onClick={() => {
                    setRoomsLoading(true);
                    apiClient
                      .getHexBlitzRooms()
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
                这轮多人版先做成房间实验室：房间创建、准备、倒计时、开局和实时记分板都是真实的
                Go + WebSocket 链路，但棋盘解题仍然由前端本地执行。
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
                    当前可加入的实验室房间。房主开始后，新玩家就不能中途加入。
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2 rounded-full border border-white/10 bg-black/20 px-3 py-2 text-sm text-slate-300">
                {roomsLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-sky-300" />
                ) : (
                  <Signal className="h-4 w-4 text-emerald-300" />
                )}
                {rooms.length} 个房间
              </div>
            </div>

            <div className="space-y-3">
              {rooms.length === 0 && (
                <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                  还没有房间。你可以直接创建一个，把这页发给另一位同学一起进来测试。
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
                        ? 'border-sky-300/40 bg-sky-300/10'
                        : 'border-white/10 bg-black/20 hover:border-white/20 hover:bg-white/[0.04]'
                    )}
                  >
                    <div className="flex flex-wrap items-start justify-between gap-4">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-lg font-semibold text-white">
                            {room.title}
                          </span>
                          <Badge className="border-white/15 bg-white/8 text-white">
                            {room.code}
                          </Badge>
                          <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                            {roomStatusLabel(room)}
                          </Badge>
                        </div>
                        <div className="mt-2 text-sm text-slate-400">
                          {room.player_count} / 4 人，已准备 {room.ready_count} 人
                        </div>
                      </div>

                      <div className="flex items-center gap-2 text-sm text-slate-300">
                        {active ? '已连接' : '加入'}
                        <ArrowRight className="h-4 w-4" />
                      </div>
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
              <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
                Live Room State
              </p>
              <h2 className="mt-2 text-2xl font-black tracking-tight">
                {activeRoom ? activeRoom.title : '还未加入房间'}
              </h2>
              <p className="mt-2 text-sm text-slate-400">{notice}</p>
            </div>

            <div className="flex flex-wrap gap-2">
              <Badge className="border-white/15 bg-white/8 text-white">
                WS: {wsStatus}
              </Badge>
              {activeRoom && (
                <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                  状态：{roomStatusLabel(activeRoom)}
                </Badge>
              )}
            </div>
          </div>

          {!activeRoom && (
            <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
              选一个房间进入后，这里会显示实时记分板、准备态和开局控制。
            </div>
          )}

          {activeRoom && (
            <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
              <div className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-3">
                  <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      房间码
                    </div>
                    <div className="mt-2 text-xl font-black tracking-tight">
                      {activeRoom.code}
                    </div>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      轮次计时
                    </div>
                    <div className="mt-2 text-xl font-black tracking-tight">
                      {activeRoom.status === 'running'
                        ? formatRemaining(
                            activeRoom.ends_at
                              ? new Date(activeRoom.ends_at).getTime()
                              : undefined,
                            timeTick
                          )
                        : `${activeRoom.round_duration_sec}s`}
                    </div>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      倒计时
                    </div>
                    <div className="mt-2 text-xl font-black tracking-tight">
                      {activeRoom.status === 'countdown'
                        ? formatRemaining(
                            activeRoom.countdown_started_at
                              ? new Date(activeRoom.countdown_started_at).getTime() +
                                  activeRoom.countdown_sec * 1000
                              : undefined,
                            timeTick
                          )
                        : `${activeRoom.countdown_sec}s`}
                    </div>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3">
                  <Button
                    onClick={() => {
                      sendRoomMessage('set_ready', {
                        ready: me ? !me.ready : true,
                      });
                    }}
                    disabled={!me || wsStatus !== 'connected' || activeRoom.status !== 'waiting'}
                    className="border-0 bg-[linear-gradient(135deg,#34d2ff_0%,#7af6b5_100%)] text-slate-950 hover:brightness-110 disabled:opacity-50"
                  >
                    {me?.ready ? '取消准备' : '准备就绪'}
                  </Button>
                  <Button
                    onClick={() => sendRoomMessage('start_match')}
                    disabled={!isHost || !canStart || wsStatus !== 'connected'}
                    variant="outline"
                    className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white disabled:opacity-50"
                  >
                    <Radio className="mr-2 h-4 w-4" />
                    房主开始
                  </Button>
                  <Button
                    onClick={() => {
                      reconnectEnabledRef.current = false;
                      activeRoomIdRef.current = null;
                      sendRoomMessage('leave_room');
                      setActiveRoom(null);
                      setActiveRoomId(null);
                      setNotice('你已离开房间。');
                      setWsStatus('idle');
                    }}
                    variant="outline"
                    className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                  >
                    离开房间
                  </Button>
                </div>

                <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4 text-sm leading-7 text-slate-300">
                  现在这条链路已经具备真实的多人房间状态机：创建房间、进入房间、在线状态、
                  准备、倒计时、开始、局内记分同步、结束。
                </div>
              </div>

              <div className="space-y-3">
                {activeRoom.players.map((player: HexBlitzRoomPlayer, index) => (
                  <div
                    key={player.session_id}
                    className={cn(
                      'rounded-[22px] border px-4 py-4',
                      player.session_id === sessionId
                        ? 'border-sky-300/35 bg-sky-300/10'
                        : 'border-white/10 bg-black/20'
                    )}
                  >
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-lg font-semibold text-white">
                            {player.name}
                          </span>
                          {player.is_host && (
                            <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                              <Crown className="mr-1 h-3 w-3" />
                              房主
                            </Badge>
                          )}
                          {player.session_id === sessionId && (
                            <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                              你
                            </Badge>
                          )}
                        </div>
                        <div className="mt-2 text-sm text-slate-400">
                          排名 #{index + 1} · {player.connected ? '在线' : '离线'}
                          {player.ready ? ' · 已准备' : ''}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                          Score
                        </div>
                        <div className="mt-1 text-3xl font-black tracking-tight text-white">
                          {player.score}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="mt-5 flex items-start gap-3 rounded-2xl border border-white/10 bg-black/20 px-4 py-4 text-sm leading-7 text-slate-300">
            <AlertCircle className="mt-1 h-4 w-4 flex-shrink-0 text-amber-300" />
            <div>
              当前多人版仍是“客户端记分、服务端同步房间状态”的实验室实现，目的是先把房间、WS
              协议和多人演示链路立起来。下一阶段再把计分和结算逻辑逐步收回到服务端。
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <Trophy className="h-5 w-5 text-amber-300" />
                <div>
                  <h2 className="text-2xl font-black tracking-tight">实时榜单</h2>
                  <p className="mt-1 text-sm text-slate-400">
                    当前基于已落库的 Hex Blitz 对局结果生成。
                  </p>
                </div>
              </div>
              {metaLoading && <Loader2 className="h-4 w-4 animate-spin text-sky-300" />}
            </div>

            <div className="space-y-3">
              {leaderboard.length === 0 && (
                <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                  还没有已结算的正式对局。先拉一局多人房间，把结果写进榜单。
                </div>
              )}

              {leaderboard.map((entry) => (
                <div
                  key={`${entry.user_id ?? entry.player_name}-${entry.rank}`}
                  className="flex items-center justify-between gap-4 rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={cn(
                        'flex h-10 w-10 items-center justify-center rounded-2xl text-sm font-black',
                        entry.rank <= 3
                          ? 'bg-amber-300/15 text-amber-100'
                          : 'bg-white/8 text-white/80'
                      )}
                    >
                      #{entry.rank}
                    </div>
                    <div>
                      <div className="font-semibold text-white">{entry.display_name}</div>
                      <div className="mt-1 text-sm text-slate-400">
                        {entry.matches} 场 · 最近于{' '}
                        {new Date(entry.last_played).toLocaleString('zh-CN', {
                          month: 'numeric',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                      BEST
                    </div>
                    <div className="mt-1 text-2xl font-black tracking-tight text-white">
                      {entry.best_score}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center gap-3">
              <Medal className="h-5 w-5 text-sky-300" />
              <div>
                <h2 className="text-2xl font-black tracking-tight">近期战报</h2>
                <p className="mt-1 text-sm text-slate-400">
                  最近落库的对局摘要，方便展示“从房间到结果”的完整链路。
                </p>
              </div>
            </div>

            <div className="space-y-4">
              {recentMatches.length === 0 && (
                <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                  还没有战报。完成一局多人对局后，这里会出现近期结果。
                </div>
              )}

              {recentMatches.map((match) => (
                <div
                  key={match.match_id}
                  className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4"
                >
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-lg font-semibold text-white">
                          {match.room_title}
                        </span>
                        <Badge className="border-white/15 bg-white/8 text-white">
                          {match.room_code}
                        </Badge>
                      </div>
                      <div className="mt-2 text-sm text-slate-400">
                        {new Date(match.finished_at).toLocaleString('zh-CN', {
                          month: 'numeric',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}{' '}
                        · {match.duration_sec}s · {match.player_count} 人
                      </div>
                    </div>

                    <div className="text-right">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        WINNER
                      </div>
                      <div className="mt-1 text-lg font-semibold text-white">
                        {match.winner_name}
                      </div>
                      <div className="mt-1 text-sm text-sky-200">
                        {match.winner_score} 分
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 grid gap-3 sm:grid-cols-3">
                    {match.top_results.map((result) => (
                      <div
                        key={result.id}
                        className="rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3"
                      >
                        <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                          #{result.rank}
                        </div>
                        <div className="mt-2 font-semibold text-white">
                          {result.display_name}
                        </div>
                        <div className="mt-1 text-sm text-slate-400">{result.score} 分</div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {user && (
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center gap-3">
              <Users2 className="h-5 w-5 text-emerald-300" />
              <div>
                <h2 className="text-2xl font-black tracking-tight">我的近期对局</h2>
                <p className="mt-1 text-sm text-slate-400">
                  这里展示当前登录用户最近落库的 Hex Blitz 对局。
                </p>
              </div>
            </div>

            <div className="space-y-3">
              {myMatches.length === 0 && (
                <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                  你还没有已落库的战报。完成一局多人对局后，这里就会出现。
                </div>
              )}

              {myMatches.map((match) => {
                const myResult = match.top_results.find(
                  (result) => result.user_id && result.user_id === user.id
                );
                return (
                  <div
                    key={match.match_id}
                    className="flex flex-wrap items-center justify-between gap-4 rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-white">{match.room_title}</span>
                        <Badge className="border-white/15 bg-white/8 text-white">
                          {match.room_code}
                        </Badge>
                      </div>
                      <div className="mt-2 text-sm text-slate-400">
                        {new Date(match.finished_at).toLocaleString('zh-CN', {
                          month: 'numeric',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </div>

                    <div className="text-right">
                      <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                        我的成绩
                      </div>
                      <div className="mt-1 text-lg font-semibold text-white">
                        {myResult ? `#${myResult.rank} · ${myResult.score} 分` : '未进入前 3'}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
