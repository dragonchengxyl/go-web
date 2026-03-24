"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Clock3,
  Crown,
  Layers3,
  ListOrdered,
  Trophy,
  Users2,
} from "lucide-react";
import { useParams } from "next/navigation";
import { apiClient, DoudizhuReplay } from "@/lib/api-client";
import { cardLabel } from "@/lib/games/doudizhu/cards";
import { comboTypeLabel } from "@/lib/games/doudizhu/combo";
import {
  actionTypeLabel,
  DOUDIZHU_LOBBY_NAME,
  DOUDIZHU_REPLAY_NAME,
  roomModeLabel,
  seatLabel,
} from "@/lib/games/doudizhu/presenter";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

export default function DoudizhuReplayPage() {
  const params = useParams<{ matchId: string }>();
  const matchId = params.matchId;

  const { data, isLoading, error } = useQuery<DoudizhuReplay>({
    queryKey: ["doudizhu-replay", matchId],
    queryFn: () => apiClient.getDoudizhuReplay(matchId),
    enabled: !!matchId,
  });

  return (
    <main className="min-h-screen bg-[#07131b] pb-20 pt-24 text-white">
      <section className="container mx-auto px-4">
        <Link
          href="/games/dou-dizhu/play"
          className="inline-flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          返回{DOUDIZHU_LOBBY_NAME}
        </Link>

        <div className="mt-6 mb-8 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.32em] text-slate-500">
              对局战报
            </p>
            <h1 className="mt-2 text-4xl font-black tracking-tight text-white">
              {data?.match.room_title ?? DOUDIZHU_REPLAY_NAME}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
              这里保留了整局时间线、倍率变化和最终结算，方便回看这把牌是怎么打成的。
            </p>
          </div>
          {data?.match && (
            <div className="flex flex-wrap gap-2">
              <Badge className="border-white/15 bg-white/8 text-white">
                {data.match.room_code}
              </Badge>
              <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                {roomModeLabel(data.match.match_mode)}
              </Badge>
            </div>
          )}
        </div>

        {isLoading && (
          <Card className="border-white/10 bg-white/[0.04] text-white">
            <CardContent className="p-6 text-sm text-slate-300">
              正在加载战报...
            </CardContent>
          </Card>
        )}

        {error && (
          <Card className="border-red-400/20 bg-red-400/10 text-red-50">
            <CardContent className="p-6 text-sm">
              {error instanceof Error ? error.message : "加载战报失败"}
            </CardContent>
          </Card>
        )}

        {data && (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
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
                    {Math.max(
                      0,
                      Math.round(
                        (new Date(data.match.finished_at).getTime() -
                          new Date(data.match.started_at).getTime()) /
                          1000,
                      ),
                    )}
                    s
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Crown className="h-4 w-4" />
                    地主位
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    {seatLabel(data.match.landlord_seat)}
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Layers3 className="h-4 w-4" />
                    倍率
                  </div>
                  <div className="text-2xl font-black tracking-tight">
                    x{data.match.multiplier}
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <ListOrdered className="h-4 w-4" />
                    特殊倍率
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Badge className="border-white/15 bg-white/8 text-white">
                      炸弹 {data.match.bomb_count}
                    </Badge>
                    {data.match.spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        春天
                      </Badge>
                    )}
                    {data.match.anti_spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        反春
                      </Badge>
                    )}
                    {!data.match.spring && !data.match.anti_spring && (
                      <Badge className="border-white/15 bg-white/8 text-white">
                        无额外加倍
                      </Badge>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
              <Card className="border-white/10 bg-white/[0.04] text-white">
                <CardContent className="p-6">
                  <div className="mb-5 flex items-center gap-3">
                    <Trophy className="h-5 w-5 text-amber-300" />
                    <h2 className="text-2xl font-black tracking-tight">
                      本局结算
                    </h2>
                  </div>
                  <div className="space-y-3">
                    {data.results.map((result) => (
                      <div
                        key={result.id}
                        className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                      >
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div>
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="font-semibold text-white">
                                {result.display_name}
                              </span>
                              <Badge className="border-white/15 bg-white/8 text-white">
                                {seatLabel(result.seat)}
                              </Badge>
                              <Badge
                                className={
                                  result.role === "landlord"
                                    ? "border-red-400/20 bg-red-400/10 text-red-100"
                                    : "border-sky-300/20 bg-sky-300/10 text-sky-100"
                                }
                              >
                                {result.role === "landlord" ? "地主" : "农民"}
                              </Badge>
                              {result.is_bot && (
                                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                                  陪练
                                </Badge>
                              )}
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              叫分 {result.bid_score} · 剩余手牌{" "}
                              {result.cards_left} 张
                            </div>
                          </div>

                          <div className="text-right">
                            <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                              Delta
                            </div>
                            <div className="mt-1 text-2xl font-black tracking-tight text-white">
                              {result.score_delta > 0
                                ? `+${result.score_delta}`
                                : result.score_delta}
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
                    <ListOrdered className="h-5 w-5 text-sky-300" />
                    <h2 className="text-2xl font-black tracking-tight">
                      出牌过程
                    </h2>
                  </div>
                  <div className="space-y-3">
                    {data.events.length === 0 && (
                      <div className="rounded-2xl border border-dashed border-white/15 bg-black/20 px-4 py-5 text-sm text-slate-400">
                        当前没有记录到操作日志。
                      </div>
                    )}
                    {data.events.map((event) => (
                      <div
                        key={event.id}
                        className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4"
                      >
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div>
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="font-semibold text-white">
                                {event.display_name}
                              </span>
                              <Badge className="border-white/15 bg-white/8 text-white">
                                {seatLabel(event.seat)}
                              </Badge>
                              <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                                {actionTypeLabel(event.action_type)}
                              </Badge>
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              第 {event.turn_no || 0} 轮 ·{" "}
                              {new Date(event.occurred_at).toLocaleString(
                                "zh-CN",
                              )}
                            </div>
                            {event.combo && (
                              <div className="mt-2 text-sm text-slate-300">
                                {comboTypeLabel(event.combo.type)} · 主值{" "}
                                {event.combo.main_rank}
                              </div>
                            )}
                            {event.cards?.length ? (
                              <div className="mt-3 flex flex-wrap gap-2">
                                {event.cards.map((card, index) => (
                                  <div
                                    key={`${event.id}-${card.suit}-${card.rank}-${index}`}
                                    className="rounded-lg border border-white/10 bg-white/8 px-2 py-1 text-xs text-white"
                                  >
                                    {cardLabel(card)}
                                  </div>
                                ))}
                              </div>
                            ) : null}
                          </div>

                          <div className="text-right text-sm text-slate-300">
                            <div>倍率 x{event.multiplier_after}</div>
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
  );
}
