'use client';

import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams, usePathname, useRouter } from 'next/navigation';
import {
  ArrowLeft,
  AudioLines,
  Bookmark,
  Clock3,
  Flag,
  Heart,
  ListMusic,
  MessageCircle,
  Pause,
  Play,
  Repeat,
  Send,
  Shuffle,
  SkipBack,
  SkipForward,
  Sparkles,
  UserRound,
  Volume2,
  VolumeX,
  Waves,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import {
  audioWorkToPlayerTrack,
  audioWorksToPlayerTracks,
  type PlayerRepeatMode,
  usePlayerStore,
} from '@/components/music-player';
import { useAuth } from '@/contexts/auth-context';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';

function formatDateTime(value?: string) {
  if (!value) return '—';
  return new Date(value).toLocaleString('zh-CN');
}

function formatDuration(seconds: number) {
  if (!seconds || Number.isNaN(seconds)) return '0:00';
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function cycleRepeatMode(mode: PlayerRepeatMode): PlayerRepeatMode {
  switch (mode) {
    case 'off':
      return 'all';
    case 'all':
      return 'one';
    default:
      return 'off';
  }
}

function repeatLabel(mode: PlayerRepeatMode) {
  switch (mode) {
    case 'all':
      return '列表循环';
    case 'one':
      return '单曲循环';
    default:
      return '不循环';
  }
}

function chooseQueueIndex(
  queueLength: number,
  currentIndex: number,
  direction: 'next' | 'previous',
  repeatMode: PlayerRepeatMode,
  shuffle: boolean,
) {
  if (queueLength <= 0 || currentIndex < 0) {
    return -1;
  }

  if (direction === 'previous') {
    if (currentIndex > 0) return currentIndex - 1;
    return repeatMode === 'all' ? queueLength - 1 : currentIndex;
  }

  if (shuffle && queueLength > 1) {
    const choices = Array.from({ length: queueLength }, (_, index) => index).filter((index) => index !== currentIndex);
    return choices[Math.floor(Math.random() * choices.length)] ?? currentIndex;
  }

  if (currentIndex < queueLength - 1) return currentIndex + 1;
  return repeatMode === 'all' ? 0 : currentIndex;
}

function QueueStatusBadge({
  active,
  playing,
}: {
  active: boolean;
  playing: boolean;
}) {
  if (!active) return null;

  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/15 px-2 py-0.5 text-[11px] font-medium text-emerald-200">
      {playing ? <Waves className="h-3 w-3" /> : <Play className="h-3 w-3" />}
      {playing ? '播放中' : '当前项'}
    </span>
  );
}

export default function AudioWorkDetailPage() {
  const queryClient = useQueryClient();
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const pathname = usePathname();
  const workId = Array.isArray(params?.id) ? params.id[0] : params?.id;
  const { isLoggedIn, loading: authLoading } = useAuth();
  const openedEventRef = useRef<string | null>(null);
  const [commentDraft, setCommentDraft] = useState('');
  const [reportReason, setReportReason] = useState('违规内容');
  const [reportDescription, setReportDescription] = useState('');
  const [showReportForm, setShowReportForm] = useState(false);

  const {
    currentTrack,
    isPlaying,
    session,
    queue,
    history,
    currentIndex,
    repeatMode,
    shuffle,
    volume,
    currentTime,
    duration,
    togglePlay,
    setTrack,
    setVolume,
    seek,
    setRepeatMode,
    toggleShuffle,
  } = usePlayerStore();

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

  const creatorWorksQuery = useQuery({
    queryKey: ['audio-work-creator-works', workQuery.data?.author_id],
    queryFn: () => apiClient.listUserAudioWorks(workQuery.data!.author_id, { page: 1, page_size: 8 }),
    enabled: !!workQuery.data?.author_id,
  });

  const recommendationsQuery = useQuery({
    queryKey: ['audio-work-recommendations', workId],
    queryFn: () => apiClient.getRelatedAudioWorks(workId as string, 8),
    enabled: !!workId,
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

  useEffect(() => {
    const openedWorkId = workQuery.data?.id;
    if (!openedWorkId || openedEventRef.current === openedWorkId) return;
    openedEventRef.current = openedWorkId;
    void apiClient.recordAudioPlaybackEvent(openedWorkId, {
      event: 'open',
      position_sec: 0,
      source_kind: session?.sourceContext?.kind || 'audio_work_detail',
    }).catch(() => undefined);
  }, [session?.sourceContext?.kind, workQuery.data?.id]);

  if (workQuery.isLoading) {
    return (
      <div className="mx-auto max-w-7xl px-4 pb-20 pt-20">
        <div className="h-[640px] animate-pulse rounded-[32px] bg-muted/40" />
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
  const comments = commentsQuery.data?.comments ?? [];
  const meState = meStateQuery.data;
  const waveform = work.waveform_preview ?? [];
  const isCurrentTrack = currentTrack?.id === work.id;
  const sessionContainsWork = !!session?.queue.some((track) => track.id === work.id);
  const creatorWorks = creatorWorksQuery.data?.items ?? [];
  const detailQueueWorks = (() => {
    const source = creatorWorks.length > 0 ? creatorWorks : [work];
    const seen = new Set<string>();
    const deduped = source.filter((item) => {
      if (seen.has(item.id)) return false;
      seen.add(item.id);
      return true;
    });
    if (!deduped.some((item) => item.id === work.id)) {
      return [work, ...deduped];
    }
    return deduped;
  })();
  const detailQueueTracks = audioWorksToPlayerTracks(detailQueueWorks);
  const displayedQueue = sessionContainsWork && queue.length > 0 ? queue : detailQueueTracks;
  const activeTrackId = sessionContainsWork ? currentTrack?.id ?? work.id : work.id;
  const displayedCurrentIndex = Math.max(0, displayedQueue.findIndex((item) => item.id === activeTrackId));
  const pageTrackIndex = Math.max(0, displayedQueue.findIndex((item) => item.id === work.id));
  const playbackReferenceIndex = isCurrentTrack ? displayedCurrentIndex : pageTrackIndex;
  const activeDuration = isCurrentTrack ? duration || work.duration_sec : work.duration_sec;
  const activeCurrentTime = isCurrentTrack ? currentTime : 0;
  const progressMax = activeDuration > 0 ? activeDuration : 0;
  const playbackProgress = progressMax > 0 ? Math.min(1, activeCurrentTime / progressMax) : 0;
  const sourceContext = sessionContainsWork
    ? session?.sourceContext
    : {
        kind: 'audio_work_detail',
        label: '作品播放页',
        href: `/audio/works/${work.id}`,
        entityId: work.id,
      };

  const creatorOtherWorks = detailQueueWorks.filter((item) => item.id !== work.id).slice(0, 5);
  const recommendedWorks = (recommendationsQuery.data?.items ?? []).slice(0, 5);
  const recentHistory = history.filter((track) => track.id !== work.id).slice(0, 5);

  function startDetailQueueAt(index: number) {
    const track = displayedQueue[index];
    if (!track) return;

    setTrack(track, {
      queue: displayedQueue,
      currentIndex: index,
      repeatMode,
      shuffle,
      sourceContext,
      openedFrom: pathname,
      autoplay: true,
    });

    if (track.id !== work.id) {
      router.push(`/audio/works/${track.id}`);
    }
  }

  function handlePrimaryPlay() {
    if (isCurrentTrack) {
      togglePlay();
      return;
    }
    startDetailQueueAt(pageTrackIndex);
  }

  function handleQueueSelect(trackId: string) {
    const index = displayedQueue.findIndex((item) => item.id === trackId);
    if (index < 0) return;
    startDetailQueueAt(index);
  }

  function handlePrevious() {
    if (isCurrentTrack && activeCurrentTime > 3) {
      seek(0);
      return;
    }

    const targetIndex = chooseQueueIndex(displayedQueue.length, playbackReferenceIndex, 'previous', repeatMode, shuffle);
    if (targetIndex >= 0) {
      void apiClient.recordAudioPlaybackEvent(work.id, {
        event: 'skip_previous',
        position_sec: activeCurrentTime,
        source_kind: sourceContext?.kind,
      }).catch(() => undefined);
      startDetailQueueAt(targetIndex);
    }
  }

  function handleNext() {
    const targetIndex = chooseQueueIndex(displayedQueue.length, playbackReferenceIndex, 'next', repeatMode, shuffle);
    if (targetIndex >= 0) {
      void apiClient.recordAudioPlaybackEvent(work.id, {
        event: 'skip_next',
        position_sec: activeCurrentTime,
        source_kind: sourceContext?.kind,
      }).catch(() => undefined);
      startDetailQueueAt(targetIndex);
    }
  }

  return (
    <div className="mx-auto max-w-7xl px-4 pb-28 pt-20">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
          <Link href="/audio/works" className="inline-flex items-center gap-1.5 font-medium hover:text-foreground">
            <ArrowLeft className="h-4 w-4" />
            返回作品列表
          </Link>
          <span className="hidden text-border sm:inline">/</span>
          <span className="inline-flex items-center gap-2 rounded-full border border-border/70 bg-muted/20 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em]">
            <Sparkles className="h-3.5 w-3.5 text-emerald-500" />
            {sourceContext?.label || '专业播放页'}
          </span>
        </div>
        <div className="text-xs text-muted-foreground">当前队列 {displayedQueue.length} 首</div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.18fr)_360px]">
        <section className="overflow-hidden rounded-[34px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(16,185,129,0.17),transparent_24%),radial-gradient(circle_at_bottom_right,rgba(59,130,246,0.14),transparent_24%),linear-gradient(145deg,rgba(15,23,42,0.98),rgba(17,24,39,0.96))] p-6 text-white shadow-[0_40px_120px_-60px_rgba(15,23,42,0.9)]">
          <div className="grid gap-8 lg:grid-cols-[320px_minmax(0,1fr)]">
            <div className="space-y-4">
              <div className="relative aspect-square overflow-hidden rounded-[28px] border border-white/10 bg-gradient-to-br from-emerald-500/25 via-cyan-500/15 to-slate-900 shadow-[0_30px_70px_-40px_rgba(16,185,129,0.7)]">
                {work.cover_image_url ? (
                  <div
                    className="absolute inset-0 bg-cover bg-center"
                    style={{ backgroundImage: `url(${work.cover_image_url})` }}
                    aria-label={work.title}
                  />
                ) : null}
                <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_36%),linear-gradient(180deg,rgba(15,23,42,0.02),rgba(2,6,23,0.65))]" />
                <div className="absolute inset-x-5 bottom-5 rounded-[22px] border border-white/10 bg-black/25 px-4 py-4 backdrop-blur-md">
                  <div className="mb-3 flex items-center justify-between text-[11px] uppercase tracking-[0.24em] text-white/70">
                    <span className="inline-flex items-center gap-2">
                      <AudioLines className="h-3.5 w-3.5" />
                      Audio Work
                    </span>
                    <QueueStatusBadge active={isCurrentTrack} playing={isPlaying} />
                  </div>
                  <div className="flex h-20 items-end gap-1.5">
                    {(waveform.length > 0 ? waveform : [0.2, 0.32, 0.44, 0.6, 0.76, 0.58, 0.36, 0.24]).map((value, index, arr) => {
                      const ratio = arr.length > 1 ? index / (arr.length - 1) : 0;
                      const active = ratio <= playbackProgress;
                      return (
                        <div
                          key={`${work.id}-hero-wave-${index}`}
                          className={cn(
                            'min-w-0 flex-1 rounded-full transition-colors duration-300',
                            active ? 'bg-white shadow-[0_0_18px_-2px_rgba(255,255,255,0.7)]' : 'bg-white/30'
                          )}
                          style={{ height: `${Math.max(12, Math.min(100, value * 100))}%` }}
                        />
                      );
                    })}
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                  <p className="text-[11px] uppercase tracking-[0.2em] text-white/50">发布时间</p>
                  <p className="mt-2 text-sm">{formatDateTime(work.published_at)}</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                  <p className="text-[11px] uppercase tracking-[0.2em] text-white/50">作品时长</p>
                  <p className="mt-2 text-sm">{formatDuration(work.duration_sec)}</p>
                </div>
              </div>
            </div>

            <div className="flex flex-col justify-between gap-6">
              <div>
                <div className="mb-3 flex flex-wrap items-center gap-2">
                  <Badge className="border-white/10 bg-white/8 px-3 py-1 text-white hover:bg-white/8">
                    <Clock3 className="mr-1.5 h-3.5 w-3.5" />
                    {formatDuration(work.duration_sec)}
                  </Badge>
                  {work.author_username ? (
                    <Badge className="border-sky-300/20 bg-sky-300/10 px-3 py-1 text-sky-100 hover:bg-sky-300/10">
                      <UserRound className="mr-1.5 h-3.5 w-3.5" />
                      @{work.author_username}
                    </Badge>
                  ) : null}
                  <Badge className="border-rose-300/20 bg-rose-300/10 px-3 py-1 text-rose-100 hover:bg-rose-300/10">
                    <Heart className="mr-1.5 h-3.5 w-3.5" />
                    {work.like_count}
                  </Badge>
                  <Badge className="border-cyan-300/20 bg-cyan-300/10 px-3 py-1 text-cyan-100 hover:bg-cyan-300/10">
                    <MessageCircle className="mr-1.5 h-3.5 w-3.5" />
                    {work.comment_count}
                  </Badge>
                </div>

                <h1 className="max-w-3xl text-3xl font-semibold tracking-tight sm:text-4xl lg:text-5xl">{work.title}</h1>
                {work.description ? (
                  <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-200/85 sm:text-base">{work.description}</p>
                ) : (
                  <p className="mt-4 max-w-2xl text-sm leading-7 text-slate-300/70">这是一首来自社区的公开音频作品，进入播放页后可以连续收听、收藏并参与互动。</p>
                )}

                {work.tags && work.tags.length > 0 ? (
                  <div className="mt-5 flex flex-wrap gap-2">
                    {work.tags.map((tag) => (
                      <span key={tag} className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-slate-100">
                        #{tag}
                      </span>
                    ))}
                  </div>
                ) : null}
              </div>

              <div className="rounded-[28px] border border-white/10 bg-black/15 p-5 shadow-[0_20px_60px_-40px_rgba(15,23,42,0.85)]">
                <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-slate-400">播放控制</p>
                    <p className="mt-1 text-sm text-slate-300">{isCurrentTrack && isPlaying ? '正在播放当前作品' : '从这里开始收听整个队列'}</p>
                  </div>
                  <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-slate-300">
                    <ListMusic className="h-3.5 w-3.5" />
                    {displayedQueue.length} 首队列
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-3">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={handlePrevious}
                    className="border-white/10 bg-white/5 text-white hover:bg-white/12"
                    aria-label="上一首"
                  >
                    <SkipBack className="h-4 w-4" />
                  </Button>
                  <Button
                    onClick={handlePrimaryPlay}
                    className="h-12 rounded-full bg-white px-6 text-slate-950 hover:bg-white/90"
                  >
                    {isCurrentTrack && isPlaying ? <Pause className="mr-2 h-4 w-4" /> : <Play className="mr-2 h-4 w-4" />}
                    {isCurrentTrack && isPlaying ? '暂停播放' : '开始播放'}
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={handleNext}
                    className="border-white/10 bg-white/5 text-white hover:bg-white/12"
                    aria-label="下一首"
                  >
                    <SkipForward className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setRepeatMode(cycleRepeatMode(repeatMode))}
                    className={cn(
                      'border-white/10 bg-white/5 text-white hover:bg-white/12',
                      repeatMode !== 'off' && 'border-emerald-300/30 bg-emerald-300/10 text-emerald-100'
                    )}
                  >
                    <Repeat className="mr-2 h-4 w-4" />
                    {repeatLabel(repeatMode)}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={toggleShuffle}
                    className={cn(
                      'border-white/10 bg-white/5 text-white hover:bg-white/12',
                      shuffle && 'border-cyan-300/30 bg-cyan-300/10 text-cyan-100'
                    )}
                  >
                    <Shuffle className="mr-2 h-4 w-4" />
                    {shuffle ? '随机已开' : '随机播放'}
                  </Button>
                </div>

                <div className="mt-6 grid gap-4">
                  <div>
                    <div className="mb-2 flex items-center justify-between text-xs text-slate-400">
                      <span>{formatDuration(activeCurrentTime)}</span>
                      <span>{formatDuration(activeDuration)}</span>
                    </div>
                    <input
                      type="range"
                      min="0"
                      max={progressMax || 0}
                      value={Math.min(activeCurrentTime, progressMax)}
                      onChange={(event) => {
                        if (!isCurrentTrack) return;
                        seek(parseFloat(event.target.value));
                      }}
                      disabled={!isCurrentTrack || progressMax <= 0}
                      className="h-2 w-full cursor-pointer appearance-none rounded-full bg-white/10 accent-emerald-400"
                    />
                  </div>

                  <div className="flex items-center gap-3">
                    {volume <= 0.01 ? <VolumeX className="h-4 w-4 text-slate-300" /> : <Volume2 className="h-4 w-4 text-slate-300" />}
                    <input
                      type="range"
                      min="0"
                      max="1"
                      step="0.01"
                      value={volume}
                      onChange={(event) => setVolume(parseFloat(event.target.value))}
                      className="h-2 w-full max-w-[240px] cursor-pointer appearance-none rounded-full bg-white/10 accent-cyan-300"
                    />
                    <span className="w-10 text-right text-xs text-slate-400">{Math.round(volume * 100)}%</span>
                  </div>
                </div>
              </div>

              <div className="flex flex-wrap gap-3">
                <Button
                  variant={meState?.liked ? 'default' : 'outline'}
                  onClick={() => likeMutation.mutate()}
                  disabled={likeMutation.isPending || !isLoggedIn}
                  className={cn(
                    !meState?.liked && 'border-white/10 bg-white/5 text-white hover:bg-white/12',
                    meState?.liked && 'bg-rose-500 text-white hover:bg-rose-500/90'
                  )}
                >
                  <Heart className="mr-2 h-4 w-4" />
                  {meState?.liked ? '已点赞' : '点赞作品'}
                </Button>
                <Button
                  variant={meState?.bookmarked ? 'default' : 'outline'}
                  onClick={() => bookmarkMutation.mutate()}
                  disabled={bookmarkMutation.isPending || !isLoggedIn}
                  className={cn(
                    !meState?.bookmarked && 'border-white/10 bg-white/5 text-white hover:bg-white/12',
                    meState?.bookmarked && 'bg-cyan-500 text-white hover:bg-cyan-500/90'
                  )}
                >
                  <Bookmark className="mr-2 h-4 w-4" />
                  {meState?.bookmarked ? '已收藏' : '收藏作品'}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setShowReportForm((value) => !value)}
                  className="border-white/10 bg-white/5 text-white hover:bg-white/12"
                >
                  <Flag className="mr-2 h-4 w-4" />
                  举报作品
                </Button>
              </div>
            </div>
          </div>
        </section>

        <div className="space-y-6">
          <Card className="overflow-hidden rounded-[28px] border-border/70 bg-card/95 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <ListMusic className="h-5 w-5 text-emerald-500" />
                当前队列
              </CardTitle>
              <CardDescription>从列表进入时会保留上下文，支持连续收听。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 p-4">
              {displayedQueue.map((track, index) => {
                const active = track.id === activeTrackId;
                return (
                  <button
                    key={`${track.id}-${index}`}
                    type="button"
                    onClick={() => handleQueueSelect(track.id)}
                    className={cn(
                      'flex w-full items-center gap-3 rounded-2xl border px-3 py-3 text-left transition-colors',
                      active
                        ? 'border-emerald-300/40 bg-emerald-50/70 dark:bg-emerald-500/10'
                        : 'border-border/70 bg-background hover:bg-muted/40'
                    )}
                  >
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-muted">
                      {active && isCurrentTrack && isPlaying ? <Pause className="h-4 w-4 text-emerald-600" /> : <Play className="h-4 w-4 text-muted-foreground" />}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className={cn('truncate text-sm font-medium', active && 'text-emerald-700 dark:text-emerald-200')}>{track.title}</p>
                      <p className="mt-0.5 truncate text-xs text-muted-foreground">{track.artist}</p>
                    </div>
                    <QueueStatusBadge active={active} playing={active && isCurrentTrack && isPlaying} />
                  </button>
                );
              })}
            </CardContent>
          </Card>

          <Card className="overflow-hidden rounded-[28px] border-border/70 bg-card/95 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <UserRound className="h-5 w-5 text-sky-500" />
                创作者更多作品
              </CardTitle>
              <CardDescription>
                {work.author_username ? `来自 @${work.author_username} 的更多公开音频` : '同一创作者的更多公开音频'}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 p-4">
              {creatorWorksQuery.isLoading ? (
                <div className="space-y-3">
                  {[0, 1, 2].map((item) => (
                    <div key={item} className="h-16 animate-pulse rounded-2xl bg-muted/40" />
                  ))}
                </div>
              ) : creatorOtherWorks.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-6 text-sm text-muted-foreground">
                  这位创作者暂时还没有更多公开音频作品。
                </div>
              ) : (
                creatorOtherWorks.map((item) => {
                  const queueIndex = detailQueueWorks.findIndex((queueItem) => queueItem.id === item.id);
                  return (
                    <div key={item.id} className="rounded-2xl border border-border/70 bg-background px-4 py-3">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <Link href={`/audio/works/${item.id}`} className="line-clamp-1 text-sm font-medium hover:text-primary">
                            {item.title}
                          </Link>
                          <p className="mt-1 text-xs text-muted-foreground">{formatDuration(item.duration_sec)} · {formatDateTime(item.published_at)}</p>
                        </div>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            if (queueIndex >= 0) {
                              setTrack(audioWorkToPlayerTrack(item), {
                                queue: detailQueueTracks,
                                currentIndex: queueIndex,
                                repeatMode,
                                shuffle,
                                sourceContext,
                                openedFrom: pathname,
                                autoplay: true,
                              });
                            }
                            router.push(`/audio/works/${item.id}`);
                          }}
                        >
                          <Play className="mr-2 h-3.5 w-3.5" />
                          播放
                        </Button>
                      </div>
                    </div>
                  );
                })
              )}
            </CardContent>
          </Card>

          <Card className="overflow-hidden rounded-[28px] border-border/70 bg-card/95 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <Sparkles className="h-5 w-5 text-fuchsia-500" />
                推荐继续收听
              </CardTitle>
              <CardDescription>专门的相关推荐接口会优先返回同作者和相近标签的公开作品。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 p-4">
              {recommendationsQuery.isLoading ? (
                <div className="space-y-3">
                  {[0, 1, 2].map((item) => (
                    <div key={item} className="h-16 animate-pulse rounded-2xl bg-muted/40" />
                  ))}
                </div>
              ) : recommendedWorks.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-6 text-sm text-muted-foreground">
                  暂时没有更多推荐作品。
                </div>
              ) : (
                recommendedWorks.map((item) => (
                  <div key={item.id} className="rounded-2xl border border-border/70 bg-background px-4 py-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <Link href={`/audio/works/${item.id}`} className="line-clamp-1 text-sm font-medium hover:text-primary">
                          {item.title}
                        </Link>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {item.author_username ? `@${item.author_username}` : item.author_id} · {formatDuration(item.duration_sec)}
                        </p>
                      </div>
                      <Link href={`/audio/works/${item.id}`} className="text-xs text-primary hover:underline">
                        打开
                      </Link>
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>

          <Card className="overflow-hidden rounded-[28px] border-border/70 bg-card/95 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <AudioLines className="h-5 w-5 text-amber-500" />
                最近收听
              </CardTitle>
              <CardDescription>本地记录最近播放过的音频作品，为后续“继续收听”打基础。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 p-4">
              {recentHistory.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-6 text-sm text-muted-foreground">
                  暂时还没有最近收听记录。
                </div>
              ) : (
                recentHistory.map((track) => (
                  <button
                    key={track.id}
                    type="button"
                    onClick={() => {
                      setTrack(track, {
                        queue: [track],
                        currentIndex: 0,
                        sourceContext: { kind: 'recent_history', label: '最近收听', href: `/audio/works/${track.id}` },
                        openedFrom: pathname,
                        autoplay: true,
                      });
                      router.push(`/audio/works/${track.id}`);
                    }}
                    className="flex w-full items-center gap-3 rounded-2xl border border-border/70 bg-background px-4 py-3 text-left transition-colors hover:bg-muted/40"
                  >
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-muted">
                      <Play className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{track.title}</p>
                      <p className="mt-0.5 truncate text-xs text-muted-foreground">{track.artist}</p>
                    </div>
                  </button>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <Card className="overflow-hidden rounded-[28px] border-border/70 shadow-sm">
          <CardHeader className="border-b border-border/60 bg-muted/20">
            <CardTitle>作品信息</CardTitle>
            <CardDescription>播放页保留作品背景信息与社区互动语境。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5 p-6">
            <div className="rounded-2xl border border-border/70 bg-muted/10 px-4 py-4">
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">作品简介</p>
              <p className="mt-3 text-sm leading-7 text-foreground">
                {work.description || '创作者没有留下更多说明。你可以从评论区补充你的感受，帮助这首作品形成更多讨论。'}
              </p>
            </div>

            {waveform.length > 0 ? (
              <div className="rounded-2xl border border-border/70 bg-muted/10 px-4 py-4">
                <div className="mb-3 flex items-center gap-2 text-sm font-medium">
                  <Waves className="h-4 w-4 text-emerald-500" />
                  波形预览
                </div>
                <div className="flex h-20 items-end gap-1">
                  {waveform.map((value, index) => {
                    const ratio = waveform.length > 1 ? index / (waveform.length - 1) : 0;
                    const active = ratio <= playbackProgress;
                    return (
                      <div
                        key={`${index}-${value}`}
                        className={cn('min-w-0 flex-1 rounded-full transition-colors', active ? 'bg-emerald-500' : 'bg-muted-foreground/20')}
                        style={{ height: `${Math.max(8, Math.min(100, value * 100))}%` }}
                      />
                    );
                  })}
                </div>
              </div>
            ) : null}

            <div className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-2xl border border-border/70 bg-background px-4 py-4">
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">发布时间</p>
                <p className="mt-2 text-sm">{formatDateTime(work.published_at)}</p>
              </div>
              <div className="rounded-2xl border border-border/70 bg-background px-4 py-4">
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">源任务 ID</p>
                <p className="mt-2 break-all text-sm">{work.source_job_id}</p>
              </div>
            </div>

            {showReportForm ? (
              <div className="rounded-2xl border border-rose-200/70 bg-rose-50/70 p-4 dark:border-rose-400/20 dark:bg-rose-500/10">
                <div className="grid gap-3">
                  <select
                    value={reportReason}
                    onChange={(event) => setReportReason(event.target.value)}
                    className="h-10 rounded-md border border-input bg-background px-3 text-sm"
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
                    className="min-h-[100px]"
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
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card className="overflow-hidden rounded-[28px] border-border/70 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle>评论互动</CardTitle>
              <CardDescription>边听边留下你的感受，帮助作品获得更多反馈。</CardDescription>
            </CardHeader>
            <CardContent className="p-6">
              {isLoggedIn ? (
                <div className="space-y-3">
                  <Textarea
                    value={commentDraft}
                    onChange={(event) => setCommentDraft(event.target.value)}
                    placeholder="聊聊你的听感，比如情绪、编曲、人声、混音，或者你最喜欢的段落。"
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
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-6 text-sm text-muted-foreground">
                  登录后可以评论、点赞和收藏这首作品。
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="overflow-hidden rounded-[28px] border-border/70 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle>评论区</CardTitle>
              <CardDescription>{work.comment_count} 条评论</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 p-6">
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
                  <div key={comment.id} className="rounded-2xl border border-border/70 bg-background px-4 py-4">
                    <div className="flex items-center justify-between gap-3">
                      <p className="text-sm font-medium">@{comment.author_username || comment.user_id}</p>
                      <span className="text-xs text-muted-foreground">{formatDateTime(comment.created_at)}</span>
                    </div>
                    <p className="mt-3 text-sm leading-7 text-foreground">{comment.content}</p>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
