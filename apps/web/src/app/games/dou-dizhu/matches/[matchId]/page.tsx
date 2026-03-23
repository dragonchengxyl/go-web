'use client';

import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowLeft,
  Clock3,
  Crown,
  Layers3,
  ListOrdered,
  Trophy,
  Users2,
} from 'lucide-react';
import { useParams } from 'next/navigation';
import { apiClient, DoudizhuReplay } from '@/lib/api-client';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';

function seatLabel(seat: number) {
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

function comboLabel(type?: string) {
  if (!type) {
    return '普通操作';
  }
  const map: Record<string, string> = {
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
  return map[type] ?? type;
}

function cardLabel(card: { suit: string; rank: number }) {
  const suitMap: Record<string, string> = {
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
  return `${suitMap[card.suit] ?? ''}${rankMap[card.rank] ?? String(card.rank)}`;
}

export default function DoudizhuReplayPage() {
  const params = useParams<{ matchId: string }>();
  const matchId = params.matchId;

  const { data, isLoading, error } = useQuery<DoudizhuReplay>({
    queryKey: ['doudizhu-replay', matchId],
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
          返回斗地主实验页
        </Link>

        <div className="mt-6 mb-8 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.32em] text-slate-500">Match Replay</p>
            <h1 className="mt-2 text-4xl font-black tracking-tight text-white">
              {data?.match.room_title ?? '斗地主战报'}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
              当前版本先提供完整操作时间线和结果榜，便于验证服务端裁决、战报落库和最近对局链路。
            </p>
          </div>
          {data?.match && (
            <Badge className="border-white/15 bg-white/8 text-white">{data.match.room_code}</Badge>
          )}
        </div>

        {isLoading && (
          <Card className="border-white/10 bg-white/[0.04] text-white">
            <CardContent className="p-6 text-sm text-slate-300">正在加载战报...</CardContent>
          </Card>
        )}

        {error && (
          <Card className="border-red-400/20 bg-red-400/10 text-red-50">
            <CardContent className="p-6 text-sm">
              {error instanceof Error ? error.message : '加载战报失败'}
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
                  <div className="text-2xl font-black tracking-tight">{data.results.length}</div>
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
                          1000
                      )
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
                  <div className="text-2xl font-black tracking-tight">x{data.match.multiplier}</div>
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
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
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div>
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="font-semibold text-white">{result.display_name}</span>
                              <Badge className="border-white/15 bg-white/8 text-white">
                                {seatLabel(result.seat)}
                              </Badge>
                              <Badge
                                className={
                                  result.role === 'landlord'
                                    ? 'border-red-400/20 bg-red-400/10 text-red-100'
                                    : 'border-sky-300/20 bg-sky-300/10 text-sky-100'
                                }
                              >
                                {result.role === 'landlord' ? '地主' : '农民'}
                              </Badge>
                              {result.is_bot && (
                                <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                                  机器人
                                </Badge>
                              )}
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              叫分 {result.bid_score} · 剩余手牌 {result.cards_left} 张
                            </div>
                          </div>

                          <div className="text-right">
                            <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
                              Delta
                            </div>
                            <div className="mt-1 text-2xl font-black tracking-tight text-white">
                              {result.score_delta > 0 ? `+${result.score_delta}` : result.score_delta}
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
                    <h2 className="text-2xl font-black tracking-tight">操作时间线</h2>
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
                              <span className="font-semibold text-white">{event.display_name}</span>
                              <Badge className="border-white/15 bg-white/8 text-white">
                                {seatLabel(event.seat)}
                              </Badge>
                              <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                                {event.action_type}
                              </Badge>
                            </div>
                            <div className="mt-2 text-sm text-slate-400">
                              第 {event.turn_no || 0} 轮 · {new Date(event.occurred_at).toLocaleString('zh-CN')}
                            </div>
                            {event.combo && (
                              <div className="mt-2 text-sm text-slate-300">
                                {comboLabel(event.combo.type)} · 主值 {event.combo.main_rank}
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
