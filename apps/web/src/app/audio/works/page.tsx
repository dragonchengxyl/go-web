'use client';

import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AudioLines, Disc3 } from 'lucide-react';
import { apiClient, type AudioWork } from '@/lib/api-client';
import { AudioWorkCard } from '@/components/audio/audio-work-card';
import { Skeleton } from '@/components/ui/skeleton';

const EMPTY_WORKS: AudioWork[] = [];

export default function AudioWorksPage() {
  const [sort, setSort] = useState<'latest' | 'oldest' | 'popular'>('latest');
  const [tag, setTag] = useState('all');

  const worksQuery = useQuery({
    queryKey: ['audio-works-public', sort, tag],
    queryFn: () =>
      apiClient.listAudioWorks({
        page: 1,
        page_size: 60,
        sort,
        tag: tag === 'all' ? undefined : tag,
      }),
  });

  const works = worksQuery.data?.items ?? EMPTY_WORKS;
  const tags = useMemo(() => {
    const set = new Set<string>();
    works.forEach((work) => (work.tags ?? []).forEach((item) => set.add(item)));
    return ['all', ...Array.from(set).sort()];
  }, [works]);

  return (
    <div className="mx-auto max-w-6xl px-4 pb-12 pt-20">
      <div className="mb-8 overflow-hidden rounded-[28px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(16,185,129,0.16),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.16),transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.94))] p-6 text-white shadow-[0_24px_80px_-44px_rgba(15,23,42,0.8)]">
        <div className="max-w-3xl">
          <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-emerald-200/80">音频作品</p>
          <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">社区音频作品</h1>
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
        <div className="flex items-center gap-2">
          <select
            value={sort}
            onChange={(event) => setSort(event.target.value as 'latest' | 'oldest' | 'popular')}
            className="rounded-lg border bg-background px-3 py-2 text-sm"
          >
            <option value="latest">最新发布</option>
            <option value="oldest">最早发布</option>
            <option value="popular">最受欢迎</option>
          </select>
        </div>
      </div>

      {tags.length > 1 ? (
        <div className="mb-6 flex flex-wrap gap-2">
          {tags.map((item) => (
            <button
              key={item}
              onClick={() => setTag(item)}
              className={`rounded-full border px-3 py-1 text-sm transition-colors ${
                tag === item
                  ? 'border-transparent bg-emerald-500 text-white'
                  : 'border-border bg-background text-muted-foreground hover:text-foreground'
              }`}
            >
              {item === 'all' ? '全部标签' : item}
            </button>
          ))}
        </div>
      ) : null}

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
          <p className="text-base font-medium">当前筛选条件下没有作品</p>
          <p className="mt-2 text-sm text-muted-foreground">换个排序或标签试试看。</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {works.map((work) => (
            <AudioWorkCard
              key={work.id}
              work={work}
              queueWorks={works}
              sourceContext={{ kind: 'audio_works_public', label: '社区音频作品', href: '/audio/works' }}
            />
          ))}
        </div>
      )}
    </div>
  );
}
