'use client';

import { useQuery } from '@tanstack/react-query';
import { AudioLines, Disc3 } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { AudioWorkCard } from '@/components/audio/audio-work-card';
import { Skeleton } from '@/components/ui/skeleton';

export default function AudioWorksPage() {
  const worksQuery = useQuery({
    queryKey: ['audio-works-public'],
    queryFn: () => apiClient.listAudioWorks({ page: 1, page_size: 24 }),
  });

  const works = worksQuery.data?.items ?? [];

  return (
    <div className="mx-auto max-w-6xl px-4 pb-12 pt-20">
      <div className="mb-8 overflow-hidden rounded-[28px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(16,185,129,0.16),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.16),transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.94))] p-6 text-white shadow-[0_24px_80px_-44px_rgba(15,23,42,0.8)]">
        <div className="max-w-3xl">
          <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-emerald-200/80">Public Audio Feed</p>
          <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">社区里的音频作品开始进入公开分发面</h1>
          <p className="mt-3 text-sm leading-6 text-slate-200/80 sm:text-base">
            这里收拢从音频任务台发布出来的作品。当前重点是建立公开消费入口，下一步会继续把它们接入首页、推荐流和互动系统。
          </p>
        </div>
      </div>

      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
            <Disc3 className="h-4 w-4" />
            Audio Works
          </div>
          <h2 className="text-2xl font-semibold">最新公开作品</h2>
        </div>
      </div>

      {worksQuery.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {[0, 1, 2, 3, 4, 5].map((item) => (
            <div key={item} className="space-y-3 rounded-2xl border border-border/60 bg-card p-3">
              <Skeleton className="aspect-[16/11] w-full rounded-xl" />
              <Skeleton className="h-5 w-2/3" />
              <Skeleton className="h-4 w-5/6" />
              <Skeleton className="h-4 w-1/2" />
            </div>
          ))}
        </div>
      ) : works.length === 0 ? (
        <div className="rounded-[24px] border border-dashed border-border bg-muted/10 px-6 py-14 text-center">
          <AudioLines className="mx-auto mb-4 h-10 w-10 text-muted-foreground/50" />
          <p className="text-base font-medium">还没有公开音频作品</p>
          <p className="mt-2 text-sm text-muted-foreground">先从创作者任务台发布第一批作品，这里会自动出现公开内容。</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {works.map((work) => (
            <AudioWorkCard key={work.id} work={work} />
          ))}
        </div>
      )}
    </div>
  );
}
