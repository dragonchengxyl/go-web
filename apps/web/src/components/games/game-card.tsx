import Link from 'next/link';
import { ArrowRight, Clock3, Gamepad2, Users2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { GameCatalogEntry, getStatusMeta } from '@/lib/games';

export function GameCard({
  game,
  featured = false,
}: {
  game: GameCatalogEntry;
  featured?: boolean;
}) {
  const status = getStatusMeta(game.status);

  return (
    <Card
      className={cn(
        'overflow-hidden border-white/10 bg-white/[0.04] text-white shadow-[0_30px_90px_-50px_rgba(0,0,0,0.85)] backdrop-blur-sm',
        featured ? 'lg:grid lg:grid-cols-[1.15fr_0.85fr]' : 'h-full',
      )}
    >
      <div
        className={cn(
          'relative overflow-hidden border-b border-white/10',
          featured ? 'min-h-[260px] lg:min-h-full lg:border-b-0 lg:border-r' : 'min-h-[210px]',
        )}
        style={{
          background: `radial-gradient(circle at top left, ${game.accentFrom}55, transparent 45%), linear-gradient(135deg, ${game.accentFrom} 0%, ${game.accentTo} 100%)`,
        }}
      >
        <div className="absolute inset-0 bg-[linear-gradient(120deg,rgba(255,255,255,0.22),transparent_30%,transparent_70%,rgba(0,0,0,0.18))]" />
        <div className="absolute inset-x-0 bottom-0 h-28 bg-gradient-to-t from-black/55 to-transparent" />
        <div className="relative flex h-full flex-col justify-between p-6">
          <div className="flex items-start justify-between gap-4">
            <Badge className={cn('border', status.className)}>{status.label}</Badge>
            <div className="rounded-full border border-white/20 bg-black/15 px-3 py-1 text-xs text-white/85">
              {game.genre}
            </div>
          </div>
          <div className="max-w-md">
            <p className="mb-2 text-xs uppercase tracking-[0.35em] text-white/70">
              {game.atmosphere}
            </p>
            <h3 className="text-3xl font-black tracking-tight text-white md:text-4xl">
              {game.title}
            </h3>
            <p className="mt-1 text-base text-white/82">{game.subtitle}</p>
            <p className="mt-4 max-w-lg text-sm leading-6 text-white/82">
              {game.shortDescription}
            </p>
          </div>
        </div>
      </div>

      <CardContent className="flex h-full flex-col p-6">
        <div className="grid gap-3 text-sm text-white/75 sm:grid-cols-3">
          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="mb-2 flex items-center gap-2 text-white/55">
              <Users2 className="h-4 w-4" />
              玩家规模
            </div>
            <div className="font-medium text-white">{game.playerMode}</div>
          </div>
          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="mb-2 flex items-center gap-2 text-white/55">
              <Clock3 className="h-4 w-4" />
              对局时长
            </div>
            <div className="font-medium text-white">{game.roundTime}</div>
          </div>
          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="mb-2 flex items-center gap-2 text-white/55">
              <Gamepad2 className="h-4 w-4" />
              当前热度
            </div>
            <div className="font-medium text-white">
              {game.onlineNow > 0 ? `${game.onlineNow} 人围观中` : '等待上线'}
            </div>
          </div>
        </div>

        <div className="mt-6 space-y-3">
          {game.highlights.slice(0, 3).map((highlight) => (
            <div
              key={highlight}
              className="rounded-2xl border border-white/8 bg-white/[0.03] px-4 py-3 text-sm leading-6 text-white/72"
            >
              {highlight}
            </div>
          ))}
        </div>

        <div className="mt-6 flex flex-wrap gap-3">
          <Button
            asChild
            className="border-0 text-slate-950 hover:brightness-110"
            style={{
              background: `linear-gradient(135deg, ${game.accentFrom}, ${game.accentTo})`,
            }}
          >
            <Link href={`/games/${game.slug}`}>
              {game.status === 'playable' ? '进入游戏页' : '查看企划'}
            </Link>
          </Button>
          {game.playPageEnabled && (
            <Button
              asChild
              variant="outline"
              className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
            >
              <Link href={`/games/${game.slug}/play`}>
                {game.slug === 'dou-dizhu'
                  ? '进入大厅'
                  : game.status === 'playable'
                    ? '直接开玩'
                    : '抢先看看'}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
