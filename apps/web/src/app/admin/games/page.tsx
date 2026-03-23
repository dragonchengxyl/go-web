'use client'

import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
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
  HexBlitzMatchSummary,
  HexBlitzRoom,
} from '@/lib/api-client'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

function formatRoomStatus(status: string) {
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

function MatchCard({ match }: { match: HexBlitzMatchSummary }) {
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

      <div className="mt-4 grid gap-3 sm:grid-cols-3">
        {match.top_results.map((result) => (
          <div
            key={result.id}
            className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3"
          >
            <p className="text-xs uppercase tracking-[0.24em] text-slate-400">#{result.rank}</p>
            <p className="mt-2 font-medium text-slate-950">{result.display_name}</p>
            <p className="mt-1 text-sm text-slate-500">{result.score} 分</p>
          </div>
        ))}
      </div>
    </div>
  )
}

function ActiveRoomCard({ room }: { room: HexBlitzRoom }) {
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

        <Badge
          variant="outline"
          className="border-slate-200 bg-slate-50 text-slate-700"
        >
          {formatRoomStatus(room.status)}
        </Badge>
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

  const metrics = data?.metrics
  const rooms = data?.rooms ?? []
  const leaderboard = data?.leaderboard ?? []
  const recentMatches = data?.recent_matches ?? []

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Game Runtime"
        title="游戏观测"
        description="把 Hex Blitz 的房间运行态、异常上报、已完成对局和榜单集中展示，便于演示与排查。"
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminMetricCard
          label="活跃房间"
          value={(metrics?.active_rooms ?? 0).toLocaleString()}
          hint="当前内存房间数"
          icon={Gamepad2}
          tone="brand"
        />
        <AdminMetricCard
          label="活跃玩家"
          value={(metrics?.active_players ?? 0).toLocaleString()}
          hint="当前房间内玩家总数"
          icon={Users}
          tone="success"
        />
        <AdminMetricCard
          label="已完成对局"
          value={(metrics?.matches_finished_total ?? 0).toLocaleString()}
          hint="累计已落库结果"
          icon={Swords}
          tone="default"
        />
        <AdminMetricCard
          label="拒绝上报"
          value={(metrics?.rejected_score_reports ?? 0).toLocaleString()}
          hint="基础反作弊拒绝次数"
          icon={ShieldCheck}
          tone={(metrics?.rejected_score_reports ?? 0) > 0 ? 'warning' : 'success'}
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
                  这里能直接看出房间有没有卡在等待、倒计时或进行中。
                </p>
              </div>
            </div>

            <div className="space-y-3">
              {rooms.length === 0 && (
                <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                  当前没有活动房间。
                </div>
              )}
              {rooms.map((room) => (
                <ActiveRoomCard key={room.id} room={room} />
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
                  分数上报已经不再裸收，异常会进入 Prometheus 指标。
                </p>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                <p className="text-xs uppercase tracking-[0.24em] text-slate-400">
                  Score Reports
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(metrics?.score_reports_total ?? 0).toLocaleString()}
                </p>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                <p className="text-xs uppercase tracking-[0.24em] text-slate-400">
                  Active WS
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(metrics?.active_connections ?? 0).toLocaleString()}
                </p>
              </div>
            </div>

            <div className="mt-4 space-y-3">
              {Object.entries(metrics?.score_report_reasons ?? {}).length === 0 && (
                <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                  当前没有异常上报被拒绝。
                </div>
              )}

              {Object.entries(metrics?.score_report_reasons ?? {}).map(([reason, count]) => (
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
                <p className="mt-1 text-sm text-slate-500">
                  当前基于已落库战报生成。
                </p>
              </div>
            </div>

            <div className="space-y-3">
              {leaderboard.length === 0 && (
                <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                  暂无排行榜数据。
                </div>
              )}

              {leaderboard.map((entry) => (
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
              {recentMatches.length === 0 && (
                <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                  暂无对局结果。
                </div>
              )}
              {recentMatches.map((match) => (
                <MatchCard key={match.match_id} match={match} />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

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
