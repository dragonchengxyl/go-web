'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams, usePathname } from 'next/navigation';
import { useState } from 'react';
import { ArrowLeft, AudioLines, Bookmark, Clock3, Flag, Heart, MessageCircle, Play, Pause, Send, UserRound, Waves } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { audioWorkToPlayerTrack, usePlayerStore } from '@/components/music-player';
import { useAuth } from '@/contexts/auth-context';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';

function formatDateTime(value?: string) {
  if (!value) return '—';
  return new Date(value).toLocaleString('zh-CN');
}

export default function AudioWorkDetailPage() {
  const queryClient = useQueryClient();
  const params = useParams<{ id: string }>();
  const pathname = usePathname();
  const workId = Array.isArray(params?.id) ? params.id[0] : params?.id;
  const { isLoggedIn, loading: authLoading } = useAuth();
  const [commentDraft, setCommentDraft] = useState('');
  const [reportReason, setReportReason] = useState('违规内容');
  const [reportDescription, setReportDescription] = useState('');
  const [showReportForm, setShowReportForm] = useState(false);
  const { currentTrack, isPlaying, session, selectTrack, setTrack, togglePlay } = usePlayerStore();

  const workQuery = useQuery({
    queryKey: ['audio-work', workId],
    queryFn: () => apiClient.getAudioWork(workId as string),
    enabled: !!workId,
  });

  const commentsQuery = useQuery({
    queryKey: ['audio-work-comments', workId],
    queryFn: () => apiClient.getCommentsByTarget('audio_work', workId as string, 1, 50),
    enabled: !!workId,
  });

  const meStateQuery = useQuery({
    queryKey: ['audio-work-me-state', workId],
    queryFn: () => apiClient.getAudioWorkMeState(workId as string),
    enabled: !!workId && !authLoading && isLoggedIn,
  });

  const refreshDetail = () => {
    queryClient.invalidateQueries({ queryKey: ['audio-work', workId] });
    queryClient.invalidateQueries({ queryKey: ['audio-works-public'] });
    queryClient.invalidateQueries({ queryKey: ['audio-work-comments', workId] });
    queryClient.invalidateQueries({ queryKey: ['audio-work-me-state', workId] });
  };

  const likeMutation = useMutation({
    mutationFn: async () => {
      if (!workId) throw new Error('missing work id');
      if (meStateQuery.data?.liked) {
        return apiClient.unlikeAudioWork(workId);
      }
      return apiClient.likeAudioWork(workId);
    },
    onSuccess: refreshDetail,
  });

  const bookmarkMutation = useMutation({
    mutationFn: async () => {
      if (!workId) throw new Error('missing work id');
      if (meStateQuery.data?.bookmarked) {
        return apiClient.unbookmarkAudioWork(workId);
      }
      return apiClient.bookmarkAudioWork(workId);
    },
    onSuccess: refreshDetail,
  });

  const commentMutation = useMutation({
    mutationFn: async () => {
      if (!workId) throw new Error('missing work id');
      return apiClient.createCommentForTarget('audio_work', workId, commentDraft.trim());
    },
    onSuccess: () => {
      setCommentDraft('');
      refreshDetail();
    },
  });

  const reportMutation = useMutation({
    mutationFn: async () => {
      if (!workId) throw new Error('missing work id');
      return apiClient.createReport('audio_work', workId, reportReason, reportDescription.trim() || undefined);
    },
    onSuccess: () => {
      setReportDescription('');
      setShowReportForm(false);
    },
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
            <Link href="/audio/works" className="inline-flex items-center text-sm font-medium text-primary hover:underline">
              <ArrowLeft className="mr-1 h-4 w-4" />
              返回作品列表
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  const work = workQuery.data;
  const waveform = work.waveform_preview ?? [];
  const comments = commentsQuery.data?.comments ?? [];
  const meState = meStateQuery.data;
  const isCurrentTrack = currentTrack?.id === work.id;
  const sessionContainsWork = !!session?.queue.some((track) => track.id === work.id);
  const detailSourceContext = {
    kind: 'audio_work_detail',
    label: '作品播放页',
    href: `/audio/works/${work.id}`,
    entityId: work.id,
  } as const;

  return (
    <div className="mx-auto max-w-5xl px-4 pb-12 pt-20">
      <div className="mb-6">
        <Link href="/audio/works" className="inline-flex items-center text-sm font-medium text-muted-foreground hover:text-foreground">
          <ArrowLeft className="mr-1 h-4 w-4" />
          返回作品列表
        </Link>
      </div>

      <div className="overflow-hidden rounded-[30px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(16,185,129,0.15),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.14),transparent_22%),linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,41,59,0.94))] p-6 text-white shadow-[0_30px_90px_-50px_rgba(15,23,42,0.85)]">
        <div className="grid gap-6 lg:grid-cols-[1.05fr_0.95fr]">
          <div>
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-emerald-200/80">音频作品</p>
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
              <Badge className="border-rose-300/30 bg-rose-300/10 px-3 py-1 text-rose-100 hover:bg-rose-300/10">
                <Heart className="mr-1.5 h-3.5 w-3.5" />
                {work.like_count}
              </Badge>
              <Badge className="border-cyan-300/30 bg-cyan-300/10 px-3 py-1 text-cyan-100 hover:bg-cyan-300/10">
                <MessageCircle className="mr-1.5 h-3.5 w-3.5" />
                {work.comment_count}
              </Badge>
              {work.moderation_status !== 'approved' ? (
                <Badge className="border-amber-300/30 bg-amber-300/10 px-3 py-1 text-amber-100 hover:bg-amber-300/10">
                  {work.moderation_status === 'pending' ? '审核中' : '已封禁'}
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
              <CardTitle className="text-xl">播放与互动</CardTitle>
              <CardDescription className="text-slate-300">在线播放、收藏作品，或参与评论互动。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <audio controls className="w-full">
                <source src={work.audio_url} />
              </audio>

              {isLoggedIn ? (
                <div className="flex flex-wrap gap-3">
                  <Button
                    variant={isCurrentTrack && isPlaying ? 'default' : 'outline'}
                    onClick={() => {
                      if (isCurrentTrack) {
                        togglePlay();
                      } else if (sessionContainsWork) {
                        selectTrack(work.id);
                      } else {
                        setTrack(audioWorkToPlayerTrack(work), {
                          queue: [audioWorkToPlayerTrack(work)],
                          currentIndex: 0,
                          sourceContext: detailSourceContext,
                          openedFrom: pathname,
                        });
                      }
                    }}
                  >
                    {isCurrentTrack && isPlaying ? <Pause className="mr-2 h-4 w-4" /> : <Play className="mr-2 h-4 w-4" />}
                    {isCurrentTrack && isPlaying ? '暂停全局播放' : '全局播放'}
                  </Button>
                  <Button
                    variant={meState?.liked ? 'default' : 'outline'}
                    onClick={() => likeMutation.mutate()}
                    disabled={likeMutation.isPending}
                  >
                    <Heart className="mr-2 h-4 w-4" />
                    {meState?.liked ? '取消点赞' : '点赞作品'}
                  </Button>
                  <Button
                    variant={meState?.bookmarked ? 'default' : 'outline'}
                    onClick={() => bookmarkMutation.mutate()}
                    disabled={bookmarkMutation.isPending}
                  >
                    <Bookmark className="mr-2 h-4 w-4" />
                    {meState?.bookmarked ? '已收藏' : '收藏作品'}
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setShowReportForm((value) => !value)}
                  >
                    <Flag className="mr-2 h-4 w-4" />
                    举报作品
                  </Button>
                </div>
              ) : (
                <div className="flex flex-wrap gap-3">
                  <Button
                    variant={isCurrentTrack && isPlaying ? 'default' : 'outline'}
                    onClick={() => {
                      if (isCurrentTrack) {
                        togglePlay();
                      } else if (sessionContainsWork) {
                        selectTrack(work.id);
                      } else {
                        setTrack(audioWorkToPlayerTrack(work), {
                          queue: [audioWorkToPlayerTrack(work)],
                          currentIndex: 0,
                          sourceContext: detailSourceContext,
                          openedFrom: pathname,
                        });
                      }
                    }}
                  >
                    {isCurrentTrack && isPlaying ? <Pause className="mr-2 h-4 w-4" /> : <Play className="mr-2 h-4 w-4" />}
                    {isCurrentTrack && isPlaying ? '暂停全局播放' : '全局播放'}
                  </Button>
                  <p className="self-center text-sm text-slate-300">登录后可以点赞、评论和收藏这首作品。</p>
                </div>
              )}

              {showReportForm ? (
                <div className="rounded-2xl border border-rose-200/40 bg-rose-50/10 p-4">
                  <div className="grid gap-3">
                    <select
                      value={reportReason}
                      onChange={(event) => setReportReason(event.target.value)}
                      className="h-10 rounded-md border border-white/10 bg-black/10 px-3 text-sm text-white"
                    >
                      <option value="违规内容">违规内容</option>
                      <option value="侵权/搬运">侵权/搬运</option>
                      <option value="骚扰/攻击">骚扰/攻击</option>
                      <option value="其他">其他</option>
                    </select>
                    <Textarea
                      value={reportDescription}
                      onChange={(event) => setReportDescription(event.target.value)}
                      placeholder="补充说明举报原因，帮助版主更快处理。"
                      className="min-h-[100px] border-white/10 bg-black/10 text-white placeholder:text-slate-400"
                    />
                    <div className="flex gap-3">
                      <Button onClick={() => reportMutation.mutate()} disabled={reportMutation.isPending}>
                        <Flag className="mr-2 h-4 w-4" />
                        提交举报
                      </Button>
                      <Button variant="outline" onClick={() => setShowReportForm(false)}>
                        取消
                      </Button>
                    </div>
                  </div>
                </div>
              ) : null}

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

      <div className="mt-8 grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
        <Card className="rounded-[24px] border-border/70">
          <CardHeader>
            <CardTitle>作品互动</CardTitle>
            <CardDescription>评论体系复用了社区现有的多态评论模型，现在这首作品已经可以承接基础互动。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoggedIn ? (
              <div className="space-y-3">
                <Textarea
                  value={commentDraft}
                  onChange={(event) => setCommentDraft(event.target.value)}
                  placeholder="写下你对这首作品的想法，例如编曲、情绪、声线或混音建议。"
                  className="min-h-[120px]"
                />
                <Button
                  onClick={() => commentMutation.mutate()}
                  disabled={commentMutation.isPending || commentDraft.trim().length === 0}
                >
                  <Send className="mr-2 h-4 w-4" />
                  发表评论
                </Button>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">登录后可参与评论。</p>
            )}
          </CardContent>
        </Card>

        <Card className="rounded-[24px] border-border/70">
          <CardHeader>
            <CardTitle>评论区</CardTitle>
            <CardDescription>{work.comment_count} 条评论</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {commentsQuery.isLoading ? (
              <div className="space-y-3">
                {[0, 1, 2].map((item) => (
                  <div key={item} className="h-20 animate-pulse rounded-2xl bg-muted/30" />
                ))}
              </div>
            ) : comments.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-8 text-center text-sm text-muted-foreground">
                还没有评论，成为第一个留下反馈的人。
              </div>
            ) : (
              comments.map((comment) => (
                <div key={comment.id} className="rounded-2xl border border-border/70 bg-background px-4 py-3">
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-sm font-medium">@{comment.author_username || comment.user_id}</p>
                    <span className="text-xs text-muted-foreground">{formatDateTime(comment.created_at)}</span>
                  </div>
                  <p className="mt-2 text-sm leading-6 text-foreground">{comment.content}</p>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
