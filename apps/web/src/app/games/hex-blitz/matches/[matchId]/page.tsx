'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  ArrowLeft,
  Clock3,
  Layers3,
  PlaySquare,
  Sparkles,
  Trophy,
  Users2,
} from 'lucide-react'
import { useParams } from 'next/navigation'
import { apiClient, HexBlitzReplay } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function HexBlitzReplayPage() {
  const params = useParams<{ matchId: string }>()
  const matchId = params.matchId

  const { data, isLoading, error } = useQuery<HexBlitzReplay>({
    queryKey: ['hex-blitz-replay', matchId],
    queryFn: () => apiClient.getHexBlitzReplay(matchId),
    enabled: !!matchId,
  })

  const replay = data

  return (
    <main className="min-h-screen bg-[#07131b] pb-20 pt-24 text-white">
      <section className="container mx-auto px-4">
        <Link
          href="/games/hex-blitz/play"
          className="inline-flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          返回 Hex Blitz 实验室
        </Link>

        <div className="mt-6 mb-8 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.32em] text-slate-500">
              Match Replay Foundation
            </p>
            <h1 className="mt-2 text-4xl font-black tracking-tight text-white">
              {replay?.match.room_title ?? '对局战报'}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
              当前版本先提供“回放基础”：一局对局的元信息、最终结果和完整操作时间线。下一步再把它升级成逐步演示棋盘状态的可视化回放。
            </p>
          </div>
          {replay?.match && (
            <Badge className="border-white/15 bg-white/8 text-white">
              {replay.match.room_code}
            </Badge>
          )}
        </div>

        {isLoading && (
          <Card className="border-white/10 bg-white/[0.04] text-white">
            <CardContent className="p-6 text-sm text-slate-300">
              正在加载对局战报...
            </CardContent>
          </Card>
        )}

        {error && (
          <Card className="border-red-400/20 bg-red-400/10 text-red-50">
            <CardContent className="p-6 text-sm">
              {error instanceof Error ? error.message : '加载回放失败'}
            </CardContent>
          </Card>
        )}

        {replay && (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Users2 className="h-4 w-4" />
                    玩家数
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {replay.results.length}
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Clock3 className="h-4 w-4" />
                    对局时长
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {replay.match.duration_sec}s
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Layers3 className="h-4 w-4" />
                    Seed
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {replay.match.seed}
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <PlaySquare className="h-4 w-4" />
                    操作数
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {replay.events.length}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-6 xl:grid-cols-[0.82fr_1.18fr]">
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-6">
                  <div className="mb-5 flex items-center gap-3">
                    <Trophy className="h-5 w-5 text-amber-300" />
                    <h2 className="text-2xl font-black tracking-tight">结果榜</h2>
                  </div>
                  <div className="space-y-3">
                    {replay.results.map((result) => (
                      <div
                        key={result.id}
                        className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                      >
                        <div className="flex items-center justify-between gap-4">
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="text-lg font-semibold text-white">
                                #{result.rank} {result.display_name}
                              </span>
                              {result.rank === 1 && (
                                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                                  冠军
                                </Badge>
                              )}
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              session: {result.user_id ?? result.player_name}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                              Score
                            </div>
                            <div className="mt-1 text-2xl font-black tracking-tight text-white">
                              {result.score}
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
                  <div className="mb-5 flex items-center gap-3">
                    <Sparkles className="h-5 w-5 text-sky-300" />
                    <h2 className="text-2xl font-black tracking-tight">操作时间线</h2>
                  </div>
                  <div className="space-y-3">
                    {replay.events.length === 0 && (
                      <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                        当前这局没有操作事件。
                      </div>
                    )}

                    {replay.events.map((event) => (
                      <div
                        key={event.id}
                        className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                      >
                        <div className="flex flex-wrap items-center justify-between gap-4">
                          <div>
                            <div className="font-semibold text-white">
                              Step {event.move_index} · {event.display_name}
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              tile {event.tile_id} · combo {event.combo_after}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-sm text-sky-200">+{event.gained_score} 分</div>
                            <div className="mt-1 text-sm text-slate-400">
                              清除 {event.cleared_count} 格，累计 {event.score_after} 分
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        )}
      </section>
    </main>
  )
}
