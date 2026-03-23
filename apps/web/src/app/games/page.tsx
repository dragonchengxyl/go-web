import Link from 'next/link';
import { ArrowRight, Sparkles, TowerControl, Trophy } from 'lucide-react';
import { GameCard } from '@/components/games/game-card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  gamesCatalog,
  getGameBySlug,
  playableGameSlug,
} from '@/lib/games';

export default function GamesPage() {
  const featuredGame = getGameBySlug(playableGameSlug);

  return (
    <main className="min-h-screen bg-[#07131b] pb-20 pt-24 text-white">
      <section className="relative overflow-hidden border-b border-white/10">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(52,210,255,0.24),transparent_25%),radial-gradient(circle_at_80%_20%,rgba(255,138,61,0.28),transparent_24%),linear-gradient(180deg,#0b1821_0%,#07131b_100%)]" />
        <div className="absolute inset-x-0 bottom-0 h-32 bg-[linear-gradient(to_top,rgba(7,19,27,1),transparent)]" />
        <div className="relative container mx-auto px-4 py-16 md:py-24">
          <div className="max-w-4xl">
            <Badge className="border-white/15 bg-white/8 text-white">
              阶段一已启动
            </Badge>
            <h1 className="mt-6 max-w-3xl text-4xl font-black tracking-tight text-white md:text-6xl">
              Games Hub
              <span className="block bg-[linear-gradient(90deg,#ffd572_0%,#7ce6ff_100%)] bg-clip-text text-transparent">
                把社区扩展成一个休闲游戏实验场
              </span>
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-7 text-slate-300 md:text-lg">
              这不是把旧商城模型捡回来，而是基于现有的用户、社区、WebSocket
              和排行榜能力，先做一个能真正玩的网页休闲游戏中心，再逐步扩成多人实时产品。
            </p>

            <div className="mt-10 grid gap-4 md:grid-cols-3">
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">当前阶段</div>
                <div className="text-xl font-semibold text-white">
                  Phase 1
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  游戏中心、详情页、单机原型先落地，保证能演示、能试玩、能继续长。
                </div>
              </div>
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">核心方向</div>
                <div className="text-xl font-semibold text-white">
                  休闲竞技
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  先做短回合高反馈玩法，后续再接房间、同步、结算和周榜。
                </div>
              </div>
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">和 JD 的关系</div>
                <div className="text-xl font-semibold text-white">
                  可讲服务端
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  不只是页面原型，而是天然能延伸到游戏房间、性能优化和稳定性。
                </div>
              </div>
            </div>

            <div className="mt-10 flex flex-wrap gap-3">
              <Button
                asChild
                className="border-0 bg-[linear-gradient(135deg,#ff8a3d_0%,#34d2ff_100%)] text-slate-950 hover:brightness-110"
              >
                <Link href="/games/hex-blitz/play">
                  直接试玩第一款
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button
                asChild
                variant="outline"
                className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
              >
                <Link href="/games/hex-blitz">查看产品定义</Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto px-4 py-12">
        <div className="mb-6 flex items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
              Featured Build
            </p>
            <h2 className="mt-2 text-3xl font-black tracking-tight text-white">
              第一款真游戏
            </h2>
          </div>
          <Link
            href="/games/hex-blitz/play"
            className="text-sm font-medium text-sky-300 transition-colors hover:text-white"
          >
            进入试玩
          </Link>
        </div>

        {featuredGame && <GameCard game={featuredGame} featured />}
      </section>

      <section className="container mx-auto px-4 py-8">
        <div className="mb-6 flex items-center gap-3">
          <Sparkles className="h-5 w-5 text-sky-300" />
          <h2 className="text-2xl font-black tracking-tight text-white">
            游戏池规划
          </h2>
        </div>
        <div className="grid gap-6 lg:grid-cols-2">
          {gamesCatalog.map((game) => (
            <GameCard key={game.slug} game={game} />
          ))}
        </div>
      </section>

      <section className="container mx-auto px-4 py-12">
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-[28px] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.03))] p-7">
            <div className="mb-5 flex items-center gap-3">
              <TowerControl className="h-5 w-5 text-amber-300" />
              <h3 className="text-xl font-semibold text-white">
                接下来怎么分阶段做
              </h3>
            </div>
            <div className="space-y-3">
              {[
                'Phase 1：Games Hub、Hex Blitz 详情页、单机可玩原型。',
                'Phase 2：Go 房间服务、准备态、比赛开始、实时同步。',
                'Phase 3：结算落库、排行榜、分享战报到社区。',
                'Phase 4：压测、日志、指标、断线恢复和后台观测。',
              ].map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
                >
                  {item}
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-[28px] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.03))] p-7">
            <div className="mb-5 flex items-center gap-3">
              <Trophy className="h-5 w-5 text-sky-300" />
              <h3 className="text-xl font-semibold text-white">
                这套方案为什么适合面试
              </h3>
            </div>
            <div className="space-y-3 text-sm leading-6 text-slate-300">
              <p>
                它既有产品感，也能自然展开到游戏服务器设计，不会陷在纯页面 Demo
                或重型棋牌规则里。
              </p>
              <p>
                你后续可以很清楚地讲：为什么先选休闲竞技、为什么先做单机手感验证、为什么服务端用房间
                + Tick + WebSocket。
              </p>
              <p>
                对 HR 来说，方向更贴合休闲游戏平台；对面试官来说，技术链路也足够完整。
              </p>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
