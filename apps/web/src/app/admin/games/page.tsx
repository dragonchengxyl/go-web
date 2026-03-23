'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  Bot,
  Gamepad2,
  Gauge,
  Loader2,
  RadioTower,
  ShieldCheck,
  Swords,
  Trophy,
  Users,
} from 'lucide-react'
import {
  AdminGameOverview,
  apiClient,
  DoudizhuLeaderboardEntry,
  DoudizhuMatchSummary,
  DoudizhuRoom,
  HexBlitzMatchSummary,
  HexBlitzRoom,
} from '@/lib/api-client'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

function formatHexRoomStatus(status: string) {
  switch (status) {
    case 'countdown':
      return '倒计时'
    case 'running':
      return '进行中'
    case 'finished':
      return '已结算'
    default:
      return '待准备'
  }
}

function formatDoudizhuRoomStatus(status: string) {
  switch (status) {
    case 'bidding':
      return '叫分中'
    case 'playing':
      return '对局中'
    case 'settlement':
      return '已结算'
    case 'redeal':
      return '流局重发'
    default:
      return '待准备'
  }
}

function formatSeat(seat?: number) {
  switch (seat) {
    case 0:
      return '一号位'
    case 1:
      return '二号位'
    case 2:
      return '三号位'
    default:
      return '--'
  }
}

function HexMatchCard({ match }: { match: HexBlitzMatchSummary }) {
  return (
    <div className="rounded-[22px] border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-lg font-semibold text-slate-950">{match.room_title}</span>
            <Badge variant="outline" className="border-slate-200 text-slate-600">
              {match.room_code}
            </Badge>
          </div>
          <p className="mt-2 text-sm text-slate-500">
            {new Date(match.finished_at).toLocaleString('zh-CN')} · {match.duration_sec}s ·{' '}
            {match.player_count} 人
          </p>
        </div>

        <div className="text-right">
          <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Winner</p>
          <p className="mt-1 text-lg font-semibold text-slate-950">{match.winner_name}</p>
          <p className="mt-1 text-sm text-sky-600">{match.winner_score} 分</p>
        </div>
      </div>

      <div className="mt-4">
        <Link
          href={`/games/hex-blitz/matches/${match.match_id}`}
          className="text-sm font-medium text-sky-600 hover:text-sky-700"
          target="_blank"
        >
          查看战报详情
        </Link>
      </div>
    </div>
  )
}

function HexRoomCard({ room }: { room: HexBlitzRoom }) {
  return (
    <div className="rounded-[22px] border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-slate-950">{room.title}</span>
            <Badge variant="outline" className="border-slate-200 text-slate-600">
              {room.code}
            </Badge>
          </div>
          <p className="mt-2 text-sm text-slate-500">
            {room.player_count} / 4 人 · 已准备 {room.ready_count} 人
          </p>
        </div>

        <Badge variant="outline" className="border-slate-200 bg-slate-50 text-slate-700">
          {formatHexRoomStatus(room.status)}
        </Badge>
      </div>
    </div>
  )
}

function DoudizhuRoomCard({ room }: { room: DoudizhuRoom }) {
  return (
    <div className="rounded-[22px] border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-slate-950">{room.title}</span>
            <Badge variant="outline" className="border-slate-200 text-slate-600">
              {room.code}
            </Badge>
            <Badge
              variant="outline"
              className={
                room.match_mode === 'demo_ai'
                  ? 'border-amber-200 bg-amber-50 text-amber-700'
                  : 'border-sky-200 bg-sky-50 text-sky-700'
              }
            >
              {room.match_mode === 'demo_ai' ? 'AI 演示' : '真人房'}
            </Badge>
          </div>
          <p className="mt-2 text-sm text-slate-500">
            {room.player_count} / 3 人 · 已准备 {room.ready_count} 人
          </p>
        </div>

        <Badge variant="outline" className="border-slate-200 bg-slate-50 text-slate-700">
          {formatDoudizhuRoomStatus(room.status)}
        </Badge>
      </div>
    </div>
  )
}

function DoudizhuMatchCard({ match }: { match: DoudizhuMatchSummary }) {
  return (
    <div className="rounded-[22px] border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-lg font-semibold text-slate-950">{match.room_title}</span>
            <Badge variant="outline" className="border-slate-200 text-slate-600">
              {match.room_code}
            </Badge>
            <Badge
              variant="outline"
              className={
                match.match_mode === 'demo_ai'
                  ? 'border-amber-200 bg-amber-50 text-amber-700'
                  : 'border-sky-200 bg-sky-50 text-sky-700'
              }
            >
              {match.match_mode === 'demo_ai' ? 'AI 演示' : '真人房'}
            </Badge>
          </div>
          <p className="mt-2 text-sm text-slate-500">
            {new Date(match.finished_at).toLocaleString('zh-CN')} · 地主 {formatSeat(match.landlord_seat)} ·{' '}
            {match.player_count} 人
          </p>
        </div>

        <div className="text-right">
          <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Winner</p>
          <p className="mt-1 text-lg font-semibold text-slate-950">
            {match.winner_side === 'landlord' ? '地主胜' : '农民胜'}
          </p>
          <p className="mt-1 text-sm text-sky-600">倍率 x{match.multiplier}</p>
        </div>
      </div>

      <div className="mt-4">
        <Link
          href={`/games/dou-dizhu/matches/${match.match_id}`}
          className="text-sm font-medium text-sky-600 hover:text-sky-700"
          target="_blank"
        >
          查看战报详情
        </Link>
      </div>
    </div>
  )
}

function DoudizhuLeaderboardCard({ entry }: { entry: DoudizhuLeaderboardEntry }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-[22px] border border-slate-200 bg-white px-4 py-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div>
        <div className="font-semibold text-slate-950">
          #{entry.rank} {entry.display_name}
        </div>
        <div className="mt-1 text-sm text-slate-500">
          {entry.matches} 局 · 胜 {entry.wins} 局 · 最近于{' '}
          {new Date(entry.last_played).toLocaleString('zh-CN')}
        </div>
      </div>
      <div className="text-right">
        <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Total</p>
        <p className="mt-1 text-xl font-semibold text-slate-950">{entry.total_score}</p>
      </div>
    </div>
  )
}

export default function AdminGamesPage() {
  const { data, isLoading } = useQuery<AdminGameOverview>({
    queryKey: ['admin-games-overview'],
    queryFn: () => apiClient.getAdminGamesOverview(),
    refetchInterval: 5000,
  })

  const hex = data?.hex_blitz
  const doudizhu = data?.doudizhu
  const hexMetrics = hex?.metrics
  const hexRooms = hex?.rooms ?? []
  const hexLeaderboard = hex?.leaderboard ?? []
  const hexRecentMatches = hex?.recent_matches ?? []
  const doudizhuMetrics = doudizhu?.metrics
  const doudizhuRooms = doudizhu?.rooms ?? []
  const doudizhuLeaderboard = doudizhu?.leaderboard ?? []
  const doudizhuRecentMatches = doudizhu?.recent_matches ?? []

  return (
    <div className="space-y-8">
      <AdminPageHeader
        eyebrow="Game Runtime"
        title="游戏观测"
        description="把游戏模块从单一的 Hex Blitz 扩成多游戏总览，便于同时观察房间态、战报和榜单。"
      />

      <section className="space-y-6">
        <div className="flex items-center gap-3">
          <Gamepad2 className="h-5 w-5 text-sky-600" />
          <div>
            <h2 className="text-xl font-semibold text-slate-950">Hex Blitz</h2>
            <p className="mt-1 text-sm text-slate-500">
              保留现有运行态与基础反作弊指标，继续用来排查实时实验室状态。
            </p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <AdminMetricCard
            label="活跃房间"
            value={(hexMetrics?.active_rooms ?? 0).toLocaleString()}
            hint="当前内存房间数"
            icon={Gamepad2}
            tone="brand"
          />
          <AdminMetricCard
            label="活跃玩家"
            value={(hexMetrics?.active_players ?? 0).toLocaleString()}
            hint="当前房间内玩家总数"
            icon={Users}
            tone="success"
          />
          <AdminMetricCard
            label="已完成对局"
            value={(hexMetrics?.matches_finished_total ?? 0).toLocaleString()}
            hint="累计已落库结果"
            icon={Swords}
            tone="default"
          />
          <AdminMetricCard
            label="拒绝上报"
            value={(hexMetrics?.rejected_score_reports ?? 0).toLocaleString()}
            hint="基础反作弊拒绝次数"
            icon={ShieldCheck}
            tone={(hexMetrics?.rejected_score_reports ?? 0) > 0 ? 'warning' : 'success'}
          />
        </div>

        <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <RadioTower className="h-5 w-5 text-sky-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">当前房间态</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    看房间有没有卡在等待、倒计时或进行中。
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                {hexRooms.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    当前没有活动房间。
                  </div>
                )}
                {hexRooms.map((room) => (
                  <HexRoomCard key={room.id} room={room} />
                ))}
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Gauge className="h-5 w-5 text-amber-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">服务端约束</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    当前异常上报和活跃连接仍主要围绕 Hex Blitz 展示。
                  </p>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Score Reports</p>
                  <p className="mt-2 text-2xl font-semibold text-slate-950">
                    {(hexMetrics?.score_reports_total ?? 0).toLocaleString()}
                  </p>
                </div>
                <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Active WS</p>
                  <p className="mt-2 text-2xl font-semibold text-slate-950">
                    {(hexMetrics?.active_connections ?? 0).toLocaleString()}
                  </p>
                </div>
              </div>

              <div className="mt-4 space-y-3">
                {Object.entries(hexMetrics?.score_report_reasons ?? {}).length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    当前没有异常上报被拒绝。
                  </div>
                )}

                {Object.entries(hexMetrics?.score_report_reasons ?? {}).map(([reason, count]) => (
                  <div
                    key={reason}
                    className="flex items-center justify-between rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3"
                  >
                    <div className="flex items-center gap-2 text-sm text-slate-600">
                      <AlertTriangle className="h-4 w-4 text-amber-600" />
                      {reason}
                    </div>
                    <div className="font-semibold text-slate-950">{count}</div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Trophy className="h-5 w-5 text-amber-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">排行榜前十</h2>
                  <p className="mt-1 text-sm text-slate-500">当前基于已落库战报生成。</p>
                </div>
              </div>

              <div className="space-y-3">
                {hexLeaderboard.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    暂无排行榜数据。
                  </div>
                )}

                {hexLeaderboard.map((entry) => (
                  <div
                    key={`${entry.user_id ?? entry.player_name}-${entry.rank}`}
                    className="flex items-center justify-between gap-4 rounded-[22px] border border-slate-200 bg-white px-4 py-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]"
                  >
                    <div>
                      <div className="font-semibold text-slate-950">
                        #{entry.rank} {entry.display_name}
                      </div>
                      <div className="mt-1 text-sm text-slate-500">
                        {entry.matches} 场 · 最近于{' '}
                        {new Date(entry.last_played).toLocaleString('zh-CN')}
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-xs uppercase tracking-[0.24em] text-slate-400">Best</p>
                      <p className="mt-1 text-xl font-semibold text-slate-950">{entry.best_score}</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Swords className="h-5 w-5 text-sky-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">最近对局</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    用于快速确认从开局到落库是否打通。
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                {hexRecentMatches.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    暂无对局结果。
                  </div>
                )}
                {hexRecentMatches.map((match) => (
                  <HexMatchCard key={match.match_id} match={match} />
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="space-y-6">
        <div className="flex items-center gap-3">
          <Bot className="h-5 w-5 text-amber-600" />
          <div>
            <h2 className="text-xl font-semibold text-slate-950">斗地主</h2>
            <p className="mt-1 text-sm text-slate-500">
              当前展示真人房与 AI 演示房的基础运行态、最近战报与榜单。
            </p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <AdminMetricCard
            label="活跃房间"
            value={(doudizhuMetrics?.active_rooms ?? 0).toLocaleString()}
            hint="当前斗地主内存房间"
            icon={Gamepad2}
            tone="brand"
          />
          <AdminMetricCard
            label="活跃玩家"
            value={(doudizhuMetrics?.active_players ?? 0).toLocaleString()}
            hint="含真人与机器人"
            icon={Users}
            tone="success"
          />
          <AdminMetricCard
            label="AI 演示房"
            value={(doudizhuMetrics?.demo_rooms ?? 0).toLocaleString()}
            hint="单人演示兜底房间"
            icon={Bot}
            tone="warning"
          />
          <AdminMetricCard
            label="最近已落库"
            value={(doudizhuMetrics?.recent_matches_count ?? 0).toLocaleString()}
            hint="当前面板抓取到的最近战报数"
            icon={Swords}
            tone="default"
          />
        </div>

        <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <RadioTower className="h-5 w-5 text-sky-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">当前房间态</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    这里能区分真人房和 AI 演示房的当前状态。
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                {doudizhuRooms.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    当前没有活动中的斗地主房间。
                  </div>
                )}
                {doudizhuRooms.map((room) => (
                  <DoudizhuRoomCard key={room.id} room={room} />
                ))}
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Gauge className="h-5 w-5 text-amber-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">房间分布</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    当前先用房间和最近战报数据做多游戏观测的第一版。
                  </p>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.24em] text-slate-400">真人房</p>
                  <p className="mt-2 text-2xl font-semibold text-slate-950">
                    {(doudizhuMetrics?.pvp_rooms ?? 0).toLocaleString()}
                  </p>
                </div>
                <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.24em] text-slate-400">AI 演示房</p>
                  <p className="mt-2 text-2xl font-semibold text-slate-950">
                    {(doudizhuMetrics?.demo_rooms ?? 0).toLocaleString()}
                  </p>
                </div>
              </div>

              <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-5 text-sm leading-7 text-slate-600">
                这块先解决“后台能看到斗地主已经在跑”这个问题。更细的托管次数、超时动作和机器人回合指标，可以在下一轮接专属观测指标时补上。
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Trophy className="h-5 w-5 text-amber-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">排行榜前十</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    当前只统计真人对局，不含 AI 演示房。
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                {doudizhuLeaderboard.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    暂无斗地主排行榜数据。
                  </div>
                )}
                {doudizhuLeaderboard.map((entry) => (
                  <DoudizhuLeaderboardCard
                    key={`${entry.user_id ?? entry.player_name}-${entry.rank}`}
                    entry={entry}
                  />
                ))}
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardContent className="p-6">
              <div className="mb-5 flex items-center gap-3">
                <Swords className="h-5 w-5 text-sky-600" />
                <div>
                  <h2 className="text-xl font-semibold text-slate-950">最近对局</h2>
                  <p className="mt-1 text-sm text-slate-500">
                    用于确认真人房和 AI 演示房都能形成战报。
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                {doudizhuRecentMatches.length === 0 && (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                    暂无斗地主对局结果。
                  </div>
                )}
                {doudizhuRecentMatches.map((match) => (
                  <DoudizhuMatchCard key={match.match_id} match={match} />
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {isLoading ? (
        <div className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <div className="flex items-center gap-3 text-sm text-slate-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            正在加载游戏观测面板...
          </div>
        </div>
      ) : null}
    </div>
  )
}
