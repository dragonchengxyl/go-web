'use client'

import Link from 'next/link'
import { useEffect, useMemo, useState } from 'react'
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
import { HexBlitzBoard } from '@/components/games/hex-blitz-board'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function HexBlitzReplayPage() {
  const params = useParams<{ matchId: string }>()
  const matchId = params.matchId
  const [selectedSession, setSelectedSession] = useState('')
  const [selectedStep, setSelectedStep] = useState(0)

  const { data, isLoading, error } = useQuery<HexBlitzReplay>({
    queryKey: ['hex-blitz-replay', matchId],
    queryFn: () => apiClient.getHexBlitzReplay(matchId),
    enabled: !!matchId,
  })

  useEffect(() => {
    if (!data?.players?.length) {
      return
    }
    setSelectedSession((current) => current || data.players[0].session_id)
  }, [data])

  const selectedPlayer = useMemo(
    () => data?.players.find((player) => player.session_id === selectedSession) ?? data?.players[0],
    [data, selectedSession]
  )

  useEffect(() => {
    setSelectedStep(0)
  }, [selectedPlayer?.session_id])

  const selectedFrame = selectedPlayer
    ? selectedPlayer.frames[Math.min(selectedStep, selectedPlayer.frames.length - 1)]
    : undefined

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
              {data?.match.room_title ?? '对局战报'}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
              当前版本已经支持按玩家查看棋盘快照，并通过 step slider 逐步回放本局操作。下一步可以继续演进成自动播放和逐帧动画。
            </p>
          </div>
          {data?.match && (
            <Badge className="border-white/15 bg-white/8 text-white">
              {data.match.room_code}
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

        {data && (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Users2 className="h-4 w-4" />
                    玩家数
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {data.results.length}
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
                    {data.match.duration_sec}s
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
                    {data.match.seed}
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
                    {data.events.length}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-6 xl:grid-cols-[1.08fr_0.92fr]">
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-6">
                  <div className="mb-5 flex items-center gap-3">
                    <Sparkles className="h-5 w-5 text-sky-300" />
                    <div>
                      <h2 className="text-2xl font-black tracking-tight">逐步回放</h2>
                      <p className="mt-1 text-sm text-slate-400">
                        先选择玩家，再拖动 step slider 查看该玩家棋盘如何演变。
                      </p>
                    </div>
                  </div>

                  {!selectedPlayer ? (
                    <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                      当前没有可回放的玩家棋盘。
                    </div>
                  ) : (
                    <div className="space-y-5">
                      <div className="flex flex-wrap gap-2">
                        {data.players.map((player) => (
                          <button
                            key={player.session_id}
                            type="button"
                            onClick={() => setSelectedSession(player.session_id)}
                            className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
                              selectedPlayer.session_id === player.session_id
                                ? 'border-sky-300/30 bg-sky-300/10 text-sky-100'
                                : 'border-white/10 bg-black/20 text-slate-300 hover:border-white/20 hover:text-white'
                            }`}
                          >
                            {player.display_name}
                          </button>
                        ))}
                      </div>

                      <HexBlitzBoard tiles={selectedFrame?.board.tiles ?? []} />

                      <div className="rounded-[24px] border border-white/10 bg-black/20 px-4 py-4">
                        <div className="flex flex-wrap items-center justify-between gap-4">
                          <div>
                            <div className="text-sm text-slate-400">
                              {selectedStep === 0
                                ? 'Step 0 · 初始棋盘'
                                : `Step ${selectedStep} · move_index ${selectedFrame?.move_index ?? 0}`}
                            </div>
                            <div className="mt-2 text-lg font-semibold text-white">
                              {selectedFrame?.event
                                ? `点击 ${selectedFrame.event.tile_id}，清除 ${selectedFrame.event.cleared_count} 格`
                                : '服务端按 seed 生成的初始棋盘'}
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              {selectedFrame?.board.message}
                            </div>
                          </div>

                          <div className="text-right">
                            <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                              Score
                            </div>
                            <div className="mt-1 text-2xl font-black tracking-tight text-white">
                              {selectedFrame?.board.score ?? 0}
                            </div>
                            <div className="mt-1 text-sm text-sky-200">
                              combo x{selectedFrame?.board.combo ?? 0}
                            </div>
                          </div>
                        </div>

                        <div className="mt-5">
                          <input
                            type="range"
                            min={0}
                            max={Math.max(0, (selectedPlayer?.frames.length ?? 1) - 1)}
                            value={selectedStep}
                            onChange={(event) => setSelectedStep(Number(event.target.value))}
                            className="h-2 w-full cursor-pointer accent-sky-400"
                          />
                        </div>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-6">
                  <div className="mb-5 flex items-center gap-3">
                    <Trophy className="h-5 w-5 text-amber-300" />
                    <h2 className="text-2xl font-black tracking-tight">结果榜</h2>
                  </div>
                  <div className="space-y-3">
                    {data.results.map((result) => (
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
                              session: {result.session_id}
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
            </div>
          </div>
        )}
      </section>
    </main>
  )
}
