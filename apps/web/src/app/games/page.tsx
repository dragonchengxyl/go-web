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
              游戏中心
            </Badge>
            <h1 className="mt-6 max-w-3xl text-4xl font-black tracking-tight text-white md:text-6xl">
              游戏中心
              <span className="block bg-[linear-gradient(90deg,#ffd572_0%,#7ce6ff_100%)] bg-clip-text text-transparent">
                随时开一局
              </span>
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-7 text-slate-300 md:text-lg">
              这里放的是可以直接进入的休闲玩法。短局、轻竞技、牌桌和冲分都可以从这里开始。
            </p>

            <div className="mt-10 grid gap-4 md:grid-cols-3">
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">现在能玩</div>
                <div className="text-xl font-semibold text-white">
                  多款短局
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  直接进入大厅或开一局，少解释，先上手。
                </div>
              </div>
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">整体风格</div>
                <div className="text-xl font-semibold text-white">
                  休闲竞技
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  短回合、高反馈、开局快，适合随时玩。
                </div>
              </div>
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-sm">
                <div className="mb-2 text-sm text-slate-400">适合场景</div>
                <div className="text-xl font-semibold text-white">
                  随时开玩
                </div>
                <div className="mt-2 text-sm leading-6 text-slate-300">
                  一个人能玩，朋友来了也能一起开桌。
                </div>
              </div>
            </div>

            <div className="mt-10 flex flex-wrap gap-3">
              <Button
                asChild
                className="border-0 bg-[linear-gradient(135deg,#ff8a3d_0%,#34d2ff_100%)] text-slate-950 hover:brightness-110"
              >
                <Link href="/games/hex-blitz/play">
                  直接开玩
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button
                asChild
                variant="outline"
                className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
              >
                <Link href="/games/hex-blitz">查看详情</Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto px-4 py-12">
        <div className="mb-6 flex items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
              Featured
            </p>
            <h2 className="mt-2 text-3xl font-black tracking-tight text-white">
              本周主打
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
            全部游戏
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
                最近更新
              </h3>
            </div>
            <div className="space-y-3">
              {[
                '游戏中心已经可以直接进入多种玩法。',
                '涂油斗地主支持人机热身和三人牌局。',
                '最近对局和战报会继续补得更完整。',
                '后续还会陆续加入新的轻竞技内容。',
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
                为什么值得玩
              </h3>
            </div>
            <div className="space-y-3 text-sm leading-6 text-slate-300">
              <p>
                这里的游戏都偏短局和高反馈，打开就能玩，不需要先读一堆说明。
              </p>
              <p>
                有的适合一个人冲分，有的适合直接拉朋友开桌，节奏都尽量做得轻快。
              </p>
              <p>
                如果你喜欢短回合、即时反馈和一点点胜负感，这里会比纯展示页更有意思。
              </p>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
