import Link from 'next/link';
import { notFound } from 'next/navigation';
import { ArrowLeft, Sparkles } from 'lucide-react';
import { DouDizhuPlayStage } from '@/components/games/doudizhu-play-stage';
import { HexBlitzPlayStage } from '@/components/games/hex-blitz-play-stage';
import { getGameBySlug } from '@/lib/games';

const playPageConfig = {
  'hex-blitz': {
    eyebrow: 'Playable Prototype',
    description:
      '当前版本先验证休闲手感和分数循环。下一阶段会把这块接成多人房间模式，补 WebSocket 同步、结算落库和排行榜。',
    badgeLabel: 'Stage 2',
    badgeText: '房间实验室已接入',
  },
  'dou-dizhu': {
    eyebrow: 'Professional Table',
    description: '',
    badgeLabel: 'Table Upgrade',
    badgeText: '涂油大厅已开放',
  },
} as const;

export default function GamePlayPage({
  params,
}: {
  params: { slug: string };
}) {
  const game = getGameBySlug(params.slug);

  if (!game || !game.playPageEnabled) {
    notFound();
  }

  const pageConfig = playPageConfig[game.slug as keyof typeof playPageConfig];
  if (!pageConfig) {
    notFound();
  }

  return (
    <main className="min-h-screen bg-[#07131b] pb-20 pt-24 text-white">
      <section className="container mx-auto px-4">
        <Link
          href={`/games/${game.slug}`}
          className="inline-flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          返回游戏详情
        </Link>

        <div className="mt-6 mb-8 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.32em] text-slate-500">
              {pageConfig.eyebrow}
            </p>
            <h1 className="mt-2 text-4xl font-black tracking-tight text-white">
              {game.title}
              <span className="ml-3 text-xl font-medium text-white/60">
                {game.subtitle}
              </span>
            </h1>
            {pageConfig.description ? (
              <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
                {pageConfig.description}
              </p>
            ) : null}
          </div>
          <div className="rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 text-sm text-slate-300">
            <span className="mr-2 inline-flex items-center gap-1 text-amber-300">
              <Sparkles className="h-4 w-4" />
              {pageConfig.badgeLabel}
            </span>
            {pageConfig.badgeText}
          </div>
        </div>

        {game.slug === 'hex-blitz' ? <HexBlitzPlayStage /> : <DouDizhuPlayStage />}
      </section>
    </main>
  );
}
