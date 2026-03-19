'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { ArrowLeft, AudioLines, Clock3, UserRound, Waves } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

function formatDateTime(value?: string) {
  if (!value) return '—';
  return new Date(value).toLocaleString('zh-CN');
}

export default function AudioWorkDetailPage() {
  const params = useParams<{ id: string }>();
  const workId = Array.isArray(params?.id) ? params.id[0] : params?.id;

  const workQuery = useQuery({
    queryKey: ['audio-work', workId],
    queryFn: () => apiClient.getAudioWork(workId as string),
    enabled: !!workId,
  });

  if (workQuery.isLoading) {
    return (
      <div className="mx-auto max-w-5xl px-4 pb-12 pt-20">
        <div className="h-[520px] animate-pulse rounded-[28px] bg-muted/40" />
      </div>
    );
  }

  if (workQuery.isError || !workQuery.data) {
    return (
      <div className="mx-auto max-w-3xl px-4 pb-12 pt-20">
        <Card className="rounded-[28px] border-border/70">
          <CardHeader>
            <CardTitle>作品不存在</CardTitle>
            <CardDescription>该音频作品可能尚未公开，或者链接已经失效。</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/creator/audio" className="inline-flex items-center text-sm font-medium text-primary hover:underline">
              <ArrowLeft className="mr-1 h-4 w-4" />
              返回音频任务台
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  const work = workQuery.data;
  const waveform = work.waveform_preview ?? [];

  return (
    <div className="mx-auto max-w-5xl px-4 pb-12 pt-20">
      <div className="mb-6">
        <Link href="/creator/audio" className="inline-flex items-center text-sm font-medium text-muted-foreground hover:text-foreground">
          <ArrowLeft className="mr-1 h-4 w-4" />
          返回音频任务台
        </Link>
      </div>

      <div className="overflow-hidden rounded-[30px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(16,185,129,0.15),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.14),transparent_22%),linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,41,59,0.94))] p-6 text-white shadow-[0_30px_90px_-50px_rgba(15,23,42,0.85)]">
        <div className="grid gap-6 lg:grid-cols-[1.05fr_0.95fr]">
          <div>
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-emerald-200/80">Published Audio Work</p>
            <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">{work.title}</h1>
            {work.description ? <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-200/80 sm:text-base">{work.description}</p> : null}

            <div className="mt-5 flex flex-wrap gap-2">
              <Badge className="border-white/15 bg-white/10 px-3 py-1 text-white hover:bg-white/10">
                <AudioLines className="mr-1.5 h-3.5 w-3.5" />
                {work.visibility === 'public' ? '公开作品' : '私密作品'}
              </Badge>
              {work.duration_sec > 0 ? (
                <Badge className="border-emerald-300/30 bg-emerald-300/10 px-3 py-1 text-emerald-100 hover:bg-emerald-300/10">
                  <Clock3 className="mr-1.5 h-3.5 w-3.5" />
                  {work.duration_sec.toFixed(2)} 秒
                </Badge>
              ) : null}
              {work.author_username ? (
                <Badge className="border-sky-300/30 bg-sky-300/10 px-3 py-1 text-sky-100 hover:bg-sky-300/10">
                  <UserRound className="mr-1.5 h-3.5 w-3.5" />
                  @{work.author_username}
                </Badge>
              ) : null}
            </div>

            {work.tags && work.tags.length > 0 ? (
              <div className="mt-5 flex flex-wrap gap-2">
                {work.tags.map((tag) => (
                  <span key={tag} className="rounded-full border border-white/10 bg-white/8 px-3 py-1 text-xs text-slate-100">
                    {tag}
                  </span>
                ))}
              </div>
            ) : null}
          </div>

          <Card className="rounded-[24px] border-white/10 bg-white/5 text-white shadow-none backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="text-xl">播放与交付</CardTitle>
              <CardDescription className="text-slate-300">这是从任务结果发布出来的第一版作品详情。后续可以继续接评论、推荐和商业化。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <audio controls className="w-full">
                <source src={work.audio_url} />
              </audio>

              {waveform.length > 0 ? (
                <div>
                  <div className="mb-3 flex items-center gap-2 text-sm text-slate-300">
                    <Waves className="h-4 w-4" />
                    波形预览
                  </div>
                  <div className="flex h-24 items-end gap-1 rounded-2xl border border-white/10 bg-black/15 px-3 py-3">
                    {waveform.map((value, index) => (
                      <div
                        key={`${index}-${value}`}
                        className="min-w-0 flex-1 rounded-full bg-gradient-to-t from-emerald-300 via-cyan-300 to-sky-200"
                        style={{ height: `${Math.max(8, Math.min(100, value * 100))}%` }}
                      />
                    ))}
                  </div>
                </div>
              ) : null}

              <div className="grid gap-3 text-sm text-slate-200 sm:grid-cols-2">
                <div className="rounded-2xl border border-white/10 bg-black/10 px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.18em] text-slate-400">发布时间</p>
                  <p className="mt-2">{formatDateTime(work.published_at)}</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-black/10 px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.18em] text-slate-400">源任务</p>
                  <p className="mt-2 break-all">{work.source_job_id}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
