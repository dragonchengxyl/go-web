import Link from 'next/link';
import { notFound } from 'next/navigation';
import {
  ArrowLeft,
  ArrowRight,
  Gamepad2,
  Sparkles,
  Timer,
  Users2,
  Waves,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { getGameBySlug, getStatusMeta } from '@/lib/games';

export default function GameDetailPage({
  params,
}: {
  params: { slug: string };
}) {
  const game = getGameBySlug(params.slug);

  if (!game) {
    notFound();
  }

  const status = getStatusMeta(game.status);
  const hasPlayPage = game.playPageEnabled === true;
  const primaryActionLabel =
    game.slug === 'dou-dizhu'
      ? '进入涂油大厅'
      : game.status === 'playable'
        ? '直接开玩'
        : '查看内容';

  return (
    <main className="min-h-screen bg-[#07131b] pb-20 pt-24 text-white">
      <section className="container mx-auto px-4">
        <Link
          href="/games"
          className="inline-flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          返回游戏中心
        </Link>

        <div
          className="mt-6 overflow-hidden rounded-[32px] border border-white/10"
          style={{
            background: `radial-gradient(circle at top left, ${game.accentFrom}40, transparent 28%), linear-gradient(135deg, rgba(255,255,255,0.06), rgba(255,255,255,0.02)), linear-gradient(135deg, ${game.accentFrom}20, ${game.accentTo}16)`,
          }}
        >
          <div className="grid gap-8 p-8 lg:grid-cols-[1.1fr_0.9fr] lg:p-10">
            <div>
              <Badge className={`border ${status.className}`}>{status.label}</Badge>
              <p className="mt-5 text-xs uppercase tracking-[0.32em] text-slate-400">
                {game.genre}
              </p>
              <h1 className="mt-4 text-4xl font-black tracking-tight text-white md:text-6xl">
                {game.title}
              </h1>
              <p className="mt-2 text-xl text-white/75">{game.subtitle}</p>
              <p className="mt-6 max-w-2xl text-base leading-8 text-slate-300">
                {game.heroDescription}
              </p>

              <div className="mt-8 flex flex-wrap gap-3">
                {hasPlayPage ? (
                  <>
                    <Button
                      asChild
                      className="border-0 text-slate-950 hover:brightness-110"
                      style={{
                        backgroundImage: `linear-gradient(135deg, ${game.accentFrom} 0%, ${game.accentTo} 100%)`,
                      }}
                    >
                      <Link href={`/games/${game.slug}/play`}>
                        {primaryActionLabel}
                        <ArrowRight className="ml-2 h-4 w-4" />
                      </Link>
                    </Button>
                    <Button
                      asChild
                      variant="outline"
                      className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                    >
                      <Link href="/games">查看其他游戏</Link>
                    </Button>
                  </>
                ) : (
                  <Button
                    asChild
                    variant="outline"
                    className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
                  >
                    <Link href="/games">返回游戏中心</Link>
                  </Button>
                )}
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Users2 className="h-4 w-4" />
                    玩家模式
                  </div>
                  <div className="text-lg font-semibold">{game.playerMode}</div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Timer className="h-4 w-4" />
                    单局时长
                  </div>
                  <div className="text-lg font-semibold">{game.roundTime}</div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Gamepad2 className="h-4 w-4" />
                    当前热度
                  </div>
                  <div className="text-lg font-semibold">
                    {game.onlineNow > 0 ? `${game.onlineNow} 人正在看` : '尚未开放'}
                  </div>
                </CardContent>
              </Card>
              <Card className="border-white/10 bg-black/20 text-white">
                <CardContent className="p-5">
                  <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                    <Waves className="h-4 w-4" />
                    视觉方向
                  </div>
                  <div className="text-lg font-semibold">{game.atmosphere}</div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>

      <section className="container mx-auto grid gap-6 px-4 py-10 lg:grid-cols-[1.1fr_0.9fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-7">
            <div className="mb-5 flex items-center gap-3">
              <Sparkles className="h-5 w-5 text-sky-300" />
              <h2 className="text-2xl font-black tracking-tight">核心循环</h2>
            </div>
            <div className="space-y-3">
              {game.loops.map((loop) => (
                <div
                  key={loop}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
                >
                  {loop}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-7">
            <h2 className="text-2xl font-black tracking-tight">适合谁玩</h2>
            <div className="mt-5 space-y-3">
              {game.fitForJD.map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
                >
                  {item}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-7">
            <h2 className="text-2xl font-black tracking-tight">亮点</h2>
            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              {game.highlights.map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4 text-sm leading-6 text-slate-300"
                >
                  {item}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-7">
            <h2 className="text-2xl font-black tracking-tight">更多看点</h2>
            <div className="mt-5 space-y-3">
              {game.roadmap.map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
                >
                  {item}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}
