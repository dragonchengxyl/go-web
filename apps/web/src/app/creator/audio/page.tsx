'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  ArrowRight,
  AudioLines,
  CheckCircle2,
  Clock3,
  Loader2,
  Mic2,
  PencilLine,
  PlayCircle,
  RefreshCcw,
  Sparkles,
  Trash2,
  UploadCloud,
  Wand2,
  Waves,
  AlertTriangle,
} from 'lucide-react';
import { useAuth } from '@/contexts/auth-context';
import { apiClient, type AudioJob, type AudioJobStatus, type AudioJobTaskType, type AudioWork } from '@/lib/api-client';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';

type UploadKind = 'source' | 'reference';

type AudioTaskOption = {
  value: AudioJobTaskType;
  label: string;
  description: string;
  icon: typeof Sparkles;
  tone: string;
};

const TASK_OPTIONS: AudioTaskOption[] = [
  {
    value: 'ai_music',
    label: 'AI 作曲',
    description: '根据提示词生成歌曲片段和创作灵感。',
    icon: Sparkles,
    tone: 'from-fuchsia-500/20 to-rose-500/10 border-fuchsia-500/30',
  },
  {
    value: 'voice_convert',
    label: '音色转换',
    description: '使用源音频和参考音频完成音色转换。',
    icon: Mic2,
    tone: 'from-sky-500/20 to-cyan-500/10 border-sky-500/30',
  },
  {
    value: 'voice_enhance',
    label: '人声增强',
    description: '对上传语音做降噪、清晰度修复和响度校准。',
    icon: Waves,
    tone: 'from-emerald-500/20 to-teal-500/10 border-emerald-500/30',
  },
  {
    value: 'audio_master',
    label: '母带处理',
    description: '优化响度、动态和整体听感，完成母带处理。',
    icon: Wand2,
    tone: 'from-amber-500/20 to-orange-500/10 border-amber-500/30',
  },
];

const STATUS_LABELS: Record<AudioJobStatus, string> = {
  queued: '排队中',
  running: '处理中',
  succeeded: '已完成',
  failed: '失败',
  dead_lettered: '死信',
};

function statusTone(status: AudioJobStatus) {
  switch (status) {
    case 'queued':
      return 'bg-slate-100 text-slate-700 border-slate-200';
    case 'running':
      return 'bg-sky-100 text-sky-700 border-sky-200';
    case 'succeeded':
      return 'bg-emerald-100 text-emerald-700 border-emerald-200';
    case 'failed':
      return 'bg-rose-100 text-rose-700 border-rose-200';
    case 'dead_lettered':
      return 'bg-amber-100 text-amber-800 border-amber-200';
    default:
      return 'bg-slate-100 text-slate-700 border-slate-200';
  }
}

function statusIcon(status: AudioJobStatus) {
  switch (status) {
    case 'queued':
      return Clock3;
    case 'running':
      return Loader2;
    case 'succeeded':
      return CheckCircle2;
    case 'failed':
      return AlertTriangle;
    case 'dead_lettered':
      return AlertTriangle;
    default:
      return Clock3;
  }
}

function taskMeta(taskType: AudioJobTaskType) {
  return TASK_OPTIONS.find((item) => item.value === taskType) ?? TASK_OPTIONS[0];
}

function formatDateTime(value?: string) {
  if (!value) return '—';
  return new Date(value).toLocaleString('zh-CN');
}

function readString(record: Record<string, unknown> | undefined, key: string) {
  const value = record?.[key];
  return typeof value === 'string' ? value : '';
}

function readNumber(record: Record<string, unknown> | undefined, key: string) {
  const value = record?.[key];
  return typeof value === 'number' ? value : null;
}

function readStringArray(record: Record<string, unknown> | undefined, key: string) {
  const value = record?.[key];
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];
}

function readRecord(record: Record<string, unknown> | undefined, key: string) {
  const value = record?.[key];
  if (!value || Array.isArray(value) || typeof value !== 'object') {
    return undefined;
  }
  return value as Record<string, unknown>;
}

export default function CreatorAudioPage() {
  const queryClient = useQueryClient();
  const { isLoggedIn, loading } = useAuth();

  const [title, setTitle] = useState('');
  const [taskType, setTaskType] = useState<AudioJobTaskType>('ai_music');
  const [prompt, setPrompt] = useState('');
  const [styleTags, setStyleTags] = useState('');
  const [notes, setNotes] = useState('');
  const [sourceAudioUrl, setSourceAudioUrl] = useState('');
  const [sourceAudioName, setSourceAudioName] = useState('');
  const [referenceAudioUrl, setReferenceAudioUrl] = useState('');
  const [referenceAudioName, setReferenceAudioName] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | AudioJobStatus>('all');
  const [taskFilter, setTaskFilter] = useState<'all' | AudioJobTaskType>('all');
  const [submitError, setSubmitError] = useState('');
  const [editingWorkId, setEditingWorkId] = useState<string | null>(null);
  const [workForm, setWorkForm] = useState({
    title: '',
    description: '',
    cover_image_url: '',
    visibility: 'public' as 'public' | 'private',
    tags: '',
  });

  const jobsQuery = useQuery({
    queryKey: ['audio-jobs', statusFilter, taskFilter],
    queryFn: () =>
      apiClient.listAudioJobs({
        page: 1,
        page_size: 20,
        status: statusFilter === 'all' ? undefined : statusFilter,
        task_type: taskFilter === 'all' ? undefined : taskFilter,
      }),
    enabled: !loading && isLoggedIn,
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data?.items) return false;
      return data.items.some((item) => item.status === 'queued' || item.status === 'running') ? 3000 : false;
    },
  });

  const worksQuery = useQuery({
    queryKey: ['my-audio-works'],
    queryFn: () => apiClient.listMyAudioWorks({ page: 1, page_size: 12 }),
    enabled: !loading && isLoggedIn,
  });

  const jobs = jobsQuery.data?.items ?? [];
  const works = worksQuery.data?.items ?? [];
  const publishedJobIDs = new Set(works.map((item) => item.source_job_id));
  const publishedWorkByJobId = new Map(works.map((item) => [item.source_job_id, item]));
  const hasActiveJobs = jobs.some((item) => item.status === 'queued' || item.status === 'running');
  const jobStats = {
    queued: jobs.filter((item) => item.status === 'queued').length,
    running: jobs.filter((item) => item.status === 'running').length,
    succeeded: jobs.filter((item) => item.status === 'succeeded').length,
    failed: jobs.filter((item) => item.status === 'failed' || item.status === 'dead_lettered').length,
  };

  const uploadMutation = useMutation({
    mutationFn: async ({ file, kind }: { file: File; kind: UploadKind }) => {
      const result = await apiClient.uploadAudio(file);
      return { kind, fileName: file.name, url: result.url };
    },
    onSuccess: ({ kind, fileName, url }) => {
      if (kind === 'source') {
        setSourceAudioUrl(url);
        setSourceAudioName(fileName);
      } else {
        setReferenceAudioUrl(url);
        setReferenceAudioName(fileName);
      }
    },
  });

  const createMutation = useMutation({
    mutationFn: () => {
      const params: Record<string, unknown> = {};
      const normalizedTags = styleTags
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean);
      if (normalizedTags.length > 0) {
        params.style_tags = normalizedTags;
      }
      if (notes.trim()) {
        params.notes = notes.trim();
      }

      return apiClient.createAudioJob({
        title,
        task_type: taskType,
        source_audio_url: sourceAudioUrl || undefined,
        reference_audio_url: referenceAudioUrl || undefined,
        prompt: prompt.trim() || undefined,
        params: Object.keys(params).length > 0 ? params : undefined,
      });
    },
    onSuccess: () => {
      setSubmitError('');
      setTitle('');
      setPrompt('');
      setStyleTags('');
      setNotes('');
      if (taskType !== 'voice_convert') {
        setReferenceAudioName('');
        setReferenceAudioUrl('');
      }
      if (taskType === 'ai_music') {
        setSourceAudioName('');
        setSourceAudioUrl('');
      }
      queryClient.invalidateQueries({ queryKey: ['audio-jobs'] });
    },
    onError: (error: Error) => {
      setSubmitError(error.message || '创建任务失败');
    },
  });

  const retryMutation = useMutation({
    mutationFn: (jobId: string) => apiClient.retryAudioJob(jobId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['audio-jobs'] });
    },
  });

  const publishMutation = useMutation({
    mutationFn: (job: AudioJob) =>
      apiClient.publishAudioJob(job.id, {
        title: job.title,
        description: readString(job.result, 'summary') || undefined,
        visibility: 'public',
        tags: readStringArray(job.result, 'style_tags'),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-audio-works'] });
    },
  });

  const updateWorkMutation = useMutation({
    mutationFn: () => {
      if (!editingWorkId) {
        throw new Error('missing work id');
      }
      return apiClient.updateAudioWork(editingWorkId, {
        title: workForm.title,
        description: workForm.description || undefined,
        cover_image_url: workForm.cover_image_url || undefined,
        visibility: workForm.visibility,
        tags: workForm.tags.split(',').map((item) => item.trim()).filter(Boolean),
      });
    },
    onSuccess: () => {
      setEditingWorkId(null);
      queryClient.invalidateQueries({ queryKey: ['my-audio-works'] });
      queryClient.invalidateQueries({ queryKey: ['audio-works-public'] });
      queryClient.invalidateQueries({ queryKey: ['audio-work'] });
    },
  });

  const deleteWorkMutation = useMutation({
    mutationFn: (workId: string) => apiClient.deleteAudioWork(workId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-audio-works'] });
      queryClient.invalidateQueries({ queryKey: ['audio-works-public'] });
    },
  });

  const currentTask = taskMeta(taskType);
  const CurrentTaskIcon = currentTask.icon;

  function requiresSourceAudio(type: AudioJobTaskType) {
    return type !== 'ai_music';
  }

  function requiresReferenceAudio(type: AudioJobTaskType) {
    return type === 'voice_convert';
  }

  function handleUpload(kind: UploadKind, file?: File | null) {
    if (!file) return;
    uploadMutation.mutate({ file, kind });
  }

  function handleCreateJob(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitError('');
    createMutation.mutate();
  }

  function beginEditWork(work: AudioWork) {
    setEditingWorkId(work.id);
    setWorkForm({
      title: work.title,
      description: work.description ?? '',
      cover_image_url: work.cover_image_url ?? '',
      visibility: work.visibility,
      tags: (work.tags ?? []).join(', '),
    });
  }

  function cancelEditWork() {
    setEditingWorkId(null);
    setWorkForm({
      title: '',
      description: '',
      cover_image_url: '',
      visibility: 'public',
      tags: '',
    });
  }

  if (loading) {
    return (
      <div className="mx-auto max-w-6xl px-4 pb-12 pt-20">
        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <div className="h-[520px] animate-pulse rounded-3xl bg-muted/50" />
          <div className="h-[520px] animate-pulse rounded-3xl bg-muted/50" />
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 pb-12 pt-20">
      <div className="mb-8 overflow-hidden rounded-[28px] border border-border/60 bg-[radial-gradient(circle_at_top_left,rgba(56,189,248,0.16),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(217,70,239,0.14),transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,41,59,0.92))] p-6 text-white shadow-[0_28px_80px_-44px_rgba(15,23,42,0.85)]">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-2xl">
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-cyan-200/80">音频创作</p>
            <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">音频任务台</h1>
            <p className="mt-3 text-sm leading-6 text-slate-200/80 sm:text-base">上传音频、选择任务，并随时查看处理进度与结果。</p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Badge className="border-white/15 bg-white/10 px-3 py-1 text-white hover:bg-white/10">
              <AudioLines className="mr-1.5 h-3.5 w-3.5" />
              四类任务
            </Badge>
            <Badge className="border-cyan-300/30 bg-cyan-300/10 px-3 py-1 text-cyan-100 hover:bg-cyan-300/10">
              {hasActiveJobs ? '任务更新中' : '等待新任务'}
            </Badge>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.05fr_0.95fr]">
        <div className="space-y-6">
          <Card className="overflow-hidden rounded-[24px] border-border/70 shadow-sm">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <Sparkles className="h-5 w-5 text-fuchsia-500" />
                新建音频任务
              </CardTitle>
              <CardDescription>先完成上传，再按任务类型提交。</CardDescription>
            </CardHeader>
            <CardContent className="p-6">
              <form className="space-y-6" onSubmit={handleCreateJob}>
                <div className="grid gap-3 sm:grid-cols-2">
                  {TASK_OPTIONS.map((option) => {
                    const Icon = option.icon;
                    const active = taskType === option.value;
                    return (
                      <button
                        key={option.value}
                        type="button"
                        onClick={() => setTaskType(option.value)}
                        className={cn(
                          'rounded-2xl border p-4 text-left transition-all',
                          active
                            ? `bg-gradient-to-br ${option.tone} shadow-[0_18px_40px_-28px_rgba(15,23,42,0.8)]`
                            : 'border-border/70 bg-background hover:border-primary/30 hover:bg-muted/20'
                        )}
                      >
                        <div className="mb-3 flex items-center justify-between">
                          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-background/80 shadow-sm">
                            <Icon className="h-5 w-5" />
                          </div>
                          {active ? <Badge className="border-primary/20 bg-primary/10 text-primary hover:bg-primary/10">当前</Badge> : null}
                        </div>
                        <p className="font-medium">{option.label}</p>
                        <p className="mt-1 text-sm text-muted-foreground">{option.description}</p>
                      </button>
                    );
                  })}
                </div>

                <div className="rounded-2xl border border-border/70 bg-muted/10 p-4">
                  <div className="mb-4 flex items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-background shadow-sm">
                      <CurrentTaskIcon className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">{currentTask.label}</p>
                      <p className="text-sm text-muted-foreground">{currentTask.description}</p>
                    </div>
                  </div>

                  <div className="grid gap-4">
                    <div>
                      <label className="mb-2 block text-sm font-medium text-foreground">任务标题</label>
                      <Input
                        value={title}
                        onChange={(event) => setTitle(event.target.value)}
                        placeholder="例如：清晨 synth-pop 风格单曲"
                        required
                      />
                    </div>

                    {taskType === 'ai_music' ? (
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">Prompt</label>
                        <Textarea
                          value={prompt}
                          onChange={(event) => setPrompt(event.target.value)}
                          placeholder="描述风格、节奏、情绪、配器和目标用户场景，例如：做一首偏动漫感的电子流行歌，女声，副歌要抓耳。"
                          className="min-h-[140px]"
                          required
                        />
                      </div>
                    ) : null}

                    <div className="grid gap-4 sm:grid-cols-2">
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">风格标签</label>
                        <Input
                          value={styleTags}
                          onChange={(event) => setStyleTags(event.target.value)}
                          placeholder="rock, anime, bright"
                        />
                      </div>
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">补充说明</label>
                        <Input
                          value={notes}
                          onChange={(event) => setNotes(event.target.value)}
                          placeholder="例如：目标发到活动页预热"
                        />
                      </div>
                    </div>
                  </div>
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <AudioUploadField
                    title="源音频"
                    description={requiresSourceAudio(taskType) ? '当前任务必需' : 'AI 作曲可不上传'}
                    fileName={sourceAudioName}
                    uploadedUrl={sourceAudioUrl}
                    required={requiresSourceAudio(taskType)}
                    loading={uploadMutation.isPending}
                    onChange={(file) => handleUpload('source', file)}
                  />
                  <AudioUploadField
                    title="参考音频"
                    description={requiresReferenceAudio(taskType) ? '音色转换必需' : '其余任务可选'}
                    fileName={referenceAudioName}
                    uploadedUrl={referenceAudioUrl}
                    required={requiresReferenceAudio(taskType)}
                    loading={uploadMutation.isPending}
                    onChange={(file) => handleUpload('reference', file)}
                  />
                </div>

                {submitError ? (
                  <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{submitError}</div>
                ) : null}

                <div className="flex flex-wrap items-center gap-3">
                  <Button type="submit" disabled={createMutation.isPending || uploadMutation.isPending} className="min-w-[156px]">
                    {createMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Sparkles className="mr-2 h-4 w-4" />}
                    提交任务
                  </Button>
                  <Link href="/creator" className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground">
                    返回创作者中心
                    <ArrowRight className="ml-1 h-4 w-4" />
                  </Link>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card className="rounded-[24px] border-border/70">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <AudioLines className="h-5 w-5 text-sky-500" />
                任务队列
              </CardTitle>
              <CardDescription>列表会在存在进行中任务时自动刷新。对本地上传文件，结果里会展示真实解析出的格式、时长、采样率等信息。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5 p-6">
              <div className="grid gap-3 sm:grid-cols-4">
                <MetricTile label="排队中" value={jobStats.queued} tone="slate" />
                <MetricTile label="处理中" value={jobStats.running} tone="sky" />
                <MetricTile label="已完成" value={jobStats.succeeded} tone="emerald" />
                <MetricTile label="失败" value={jobStats.failed} tone="rose" />
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <select
                  value={statusFilter}
                  onChange={(event) => setStatusFilter(event.target.value as 'all' | AudioJobStatus)}
                  className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="all">全部状态</option>
                  <option value="queued">排队中</option>
                  <option value="running">处理中</option>
                  <option value="succeeded">已完成</option>
                  <option value="failed">失败</option>
                  <option value="dead_lettered">死信</option>
                </select>
                <select
                  value={taskFilter}
                  onChange={(event) => setTaskFilter(event.target.value as 'all' | AudioJobTaskType)}
                  className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="all">全部类型</option>
                  {TASK_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>

              {jobsQuery.isLoading ? (
                <div className="space-y-3">
                  {[0, 1, 2].map((item) => (
                    <div key={item} className="h-32 animate-pulse rounded-2xl bg-muted/40" />
                  ))}
                </div>
              ) : jobs.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-6 py-10 text-center">
                  <p className="text-base font-medium">还没有音频任务</p>
                  <p className="mt-2 text-sm text-muted-foreground">先上传一个样本，提交第一条任务。当前最适合先测的是“人声增强”或“AI 作曲”。</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {jobs.map((job) => (
                    <JobCard
                      key={job.id}
                      job={job}
                      retrying={retryMutation.isPending}
                      onRetry={() => retryMutation.mutate(job.id)}
                      canPublish={
                        job.status === 'succeeded' &&
                        !!readString(job.result, 'output_audio_url') &&
                        !publishedJobIDs.has(job.id)
                      }
                      publishing={publishMutation.isPending}
                      onPublish={() => publishMutation.mutate(job)}
                      publishedWorkId={publishedWorkByJobId.get(job.id)?.id}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="rounded-[24px] border-border/70">
            <CardHeader className="border-b border-border/60 bg-muted/20">
              <CardTitle className="flex items-center gap-2 text-xl">
                <PlayCircle className="h-5 w-5 text-emerald-500" />
                已发布作品
              </CardTitle>
              <CardDescription>这里展示已经从任务发布出去的音频作品。现在可以继续编辑标题、描述、标签和可见性，也可以删除作品。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 p-6">
              {editingWorkId ? (
                <div className="rounded-2xl border border-border/70 bg-muted/10 p-4">
                  <div className="mb-4 flex items-center justify-between">
                    <div>
                      <p className="font-medium">编辑音频作品</p>
                      <p className="text-sm text-muted-foreground">修改展示信息，不会影响原始音频输出。</p>
                    </div>
                    <Button variant="ghost" size="sm" onClick={cancelEditWork}>
                      取消
                    </Button>
                  </div>

                  <form
                    className="space-y-4"
                    onSubmit={(event) => {
                      event.preventDefault();
                      updateWorkMutation.mutate();
                    }}
                  >
                    <div className="grid gap-4 sm:grid-cols-2">
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">标题</label>
                        <Input
                          value={workForm.title}
                          onChange={(event) => setWorkForm((prev) => ({ ...prev, title: event.target.value }))}
                          required
                        />
                      </div>
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">可见性</label>
                        <select
                          value={workForm.visibility}
                          onChange={(event) => setWorkForm((prev) => ({ ...prev, visibility: event.target.value as 'public' | 'private' }))}
                          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
                        >
                          <option value="public">公开</option>
                          <option value="private">私密</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <label className="mb-2 block text-sm font-medium text-foreground">简介</label>
                      <Textarea
                        value={workForm.description}
                        onChange={(event) => setWorkForm((prev) => ({ ...prev, description: event.target.value }))}
                        className="min-h-[100px]"
                      />
                    </div>

                    <div className="grid gap-4 sm:grid-cols-2">
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">封面地址</label>
                        <Input
                          value={workForm.cover_image_url}
                          onChange={(event) => setWorkForm((prev) => ({ ...prev, cover_image_url: event.target.value }))}
                          placeholder="/uploads/images/..."
                        />
                      </div>
                      <div>
                        <label className="mb-2 block text-sm font-medium text-foreground">标签</label>
                        <Input
                          value={workForm.tags}
                          onChange={(event) => setWorkForm((prev) => ({ ...prev, tags: event.target.value }))}
                          placeholder="ambient, night, vocal"
                        />
                      </div>
                    </div>

                    <div className="flex gap-3">
                      <Button type="submit" disabled={updateWorkMutation.isPending}>
                        {updateWorkMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PencilLine className="mr-2 h-4 w-4" />}
                        保存修改
                      </Button>
                      <Button type="button" variant="outline" onClick={cancelEditWork}>
                        取消
                      </Button>
                    </div>
                  </form>
                </div>
              ) : null}

              {worksQuery.isLoading ? (
                <div className="space-y-3">
                  {[0, 1].map((item) => (
                    <div key={item} className="h-28 animate-pulse rounded-2xl bg-muted/40" />
                  ))}
                </div>
              ) : works.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-6 py-8 text-center">
                  <p className="text-base font-medium">还没有发布作品</p>
                  <p className="mt-2 text-sm text-muted-foreground">先完成一个带输出音频的任务，然后点击“发布作品”。</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {works.map((work) => (
                    <PublishedWorkCard
                      key={work.id}
                      work={work}
                      onEdit={() => beginEditWork(work)}
                      onDelete={() => {
                        if (window.confirm(`确认删除作品《${work.title}》吗？`)) {
                          deleteWorkMutation.mutate(work.id);
                        }
                      }}
                      deleting={deleteWorkMutation.isPending}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function MetricTile({ label, value, tone }: { label: string; value: number; tone: 'slate' | 'sky' | 'emerald' | 'rose' }) {
  const toneClass = {
    slate: 'border-slate-200 bg-slate-50 text-slate-700',
    sky: 'border-sky-200 bg-sky-50 text-sky-700',
    emerald: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    rose: 'border-rose-200 bg-rose-50 text-rose-700',
  }[tone];

  return (
    <div className={cn('rounded-2xl border px-4 py-3', toneClass)}>
      <p className="text-xs font-medium uppercase tracking-[0.16em]">{label}</p>
      <p className="mt-2 text-2xl font-semibold">{value}</p>
    </div>
  );
}

function AudioUploadField({
  title,
  description,
  fileName,
  uploadedUrl,
  required,
  loading,
  onChange,
}: {
  title: string;
  description: string;
  fileName: string;
  uploadedUrl: string;
  required: boolean;
  loading: boolean;
  onChange: (file?: File | null) => void;
}) {
  return (
    <div className="rounded-2xl border border-border/70 bg-muted/10 p-4">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <p className="font-medium">{title}</p>
          <p className="mt-1 text-sm text-muted-foreground">{description}</p>
        </div>
        {required ? <Badge className="border-amber-200 bg-amber-50 text-amber-700 hover:bg-amber-50">必需</Badge> : null}
      </div>

      <label className="flex cursor-pointer flex-col items-center justify-center rounded-2xl border border-dashed border-border bg-background px-4 py-6 text-center transition-colors hover:border-primary/40 hover:bg-muted/20">
        <UploadCloud className="mb-3 h-5 w-5 text-muted-foreground" />
        <span className="text-sm font-medium">{loading ? '上传中...' : '点击选择音频文件'}</span>
        <span className="mt-1 text-xs text-muted-foreground">支持 mp3 / wav / flac / ogg / m4a</span>
        <input
          type="file"
          accept=".mp3,.wav,.flac,.ogg,.m4a,audio/*"
          className="hidden"
          onChange={(event) => onChange(event.target.files?.[0] ?? null)}
        />
      </label>

      {fileName ? <p className="mt-3 text-sm font-medium text-foreground">已上传：{fileName}</p> : null}
      {uploadedUrl ? <p className="mt-1 break-all text-xs text-muted-foreground">{uploadedUrl}</p> : null}
    </div>
  );
}

function JobCard({
  job,
  retrying,
  onRetry,
  canPublish,
  publishing,
  onPublish,
  publishedWorkId,
}: {
  job: AudioJob;
  retrying: boolean;
  onRetry: () => void;
  canPublish: boolean;
  publishing: boolean;
  onPublish: () => void;
  publishedWorkId?: string;
}) {
  const task = taskMeta(job.task_type);
  const StatusIcon = statusIcon(job.status);
  const summary = readString(job.result, 'summary');
  const outputAudioURL = readString(job.result, 'output_audio_url');
  const arrangement = readStringArray(job.result, 'arrangement');
  const sourceAnalysis = readRecord(job.result, 'source_analysis');
  const outputAnalysis = readRecord(job.result, 'output_analysis');

  return (
    <div className="rounded-2xl border border-border/70 bg-background p-4 shadow-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <p className="truncate text-base font-semibold">{job.title}</p>
            <Badge className={cn('border px-2.5 py-1 text-xs hover:bg-transparent', statusTone(job.status))}>
              <StatusIcon className={cn('mr-1.5 h-3.5 w-3.5', job.status === 'running' ? 'animate-spin' : '')} />
              {STATUS_LABELS[job.status]}
            </Badge>
            <Badge variant="outline">{task.label}</Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">{task.description}</p>
        </div>

        {job.status === 'failed' || job.status === 'dead_lettered' ? (
          <Button variant="outline" size="sm" disabled={retrying} onClick={onRetry}>
            <RefreshCcw className="mr-2 h-4 w-4" />
            重试
          </Button>
        ) : canPublish ? (
          <Button variant="outline" size="sm" disabled={publishing} onClick={onPublish}>
            {publishing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
            发布作品
          </Button>
        ) : publishedWorkId ? (
          <Link
            href={`/audio/works/${publishedWorkId}`}
            className="inline-flex items-center rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-700 transition-colors hover:bg-emerald-100"
          >
            查看作品
          </Link>
        ) : null}
      </div>

      <div className="mt-4 grid gap-3 text-sm text-muted-foreground sm:grid-cols-2">
        <div>
          <p>创建时间：{formatDateTime(job.created_at)}</p>
          <p>开始时间：{formatDateTime(job.started_at)}</p>
        </div>
        <div>
          <p>完成时间：{formatDateTime(job.finished_at)}</p>
          <p>尝试次数：{job.attempt_count} / {job.max_attempts}</p>
        </div>
      </div>

      <div className="mt-2 text-sm text-muted-foreground">
        {job.prompt ? <p className="truncate">Prompt：{job.prompt}</p> : null}
        {job.next_retry_at ? <p>下次重试：{formatDateTime(job.next_retry_at)}</p> : null}
      </div>

      {summary ? (
        <div className="mt-4 rounded-2xl border border-emerald-200/70 bg-emerald-50/60 px-4 py-3">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">处理摘要</p>
          <p className="mt-2 text-sm text-emerald-900">{summary}</p>
        </div>
      ) : null}

      {sourceAnalysis || outputAnalysis ? (
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {sourceAnalysis ? <AudioMetaPanel title="输入音频" meta={sourceAnalysis} /> : null}
          {outputAnalysis ? <AudioMetaPanel title="输出音频" meta={outputAnalysis} /> : null}
        </div>
      ) : null}

      {job.error_message ? (
        <div className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{job.error_message}</div>
      ) : null}

      {arrangement.length > 0 ? (
        <div className="mt-4 flex flex-wrap gap-2">
          {arrangement.map((item) => (
            <Badge key={item} variant="outline" className="bg-muted/40">
              {item}
            </Badge>
          ))}
        </div>
      ) : null}

      {outputAudioURL ? (
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <a
            href={outputAudioURL}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center text-sm font-medium text-primary hover:underline"
          >
            打开输出音频
            <ArrowRight className="ml-1 h-4 w-4" />
          </a>
        </div>
      ) : null}
    </div>
  );
}

function PublishedWorkCard({
  work,
  onEdit,
  onDelete,
  deleting,
}: {
  work: AudioWork;
  onEdit: () => void;
  onDelete: () => void;
  deleting: boolean;
}) {
  return (
    <div className="rounded-2xl border border-border/70 bg-background p-4 shadow-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <p className="truncate text-base font-semibold">{work.title}</p>
            <Badge className="border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-50">
              {work.visibility === 'public' ? '公开' : '私密'}
            </Badge>
          </div>
          {work.description ? <p className="mt-1 text-sm text-muted-foreground">{work.description}</p> : null}
          <p className="mt-2 text-xs text-muted-foreground">发布时间：{formatDateTime(work.published_at)}</p>
        </div>
        <Link
          href={`/audio/works/${work.id}`}
          className="inline-flex items-center rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted/30"
        >
          打开详情
          <ArrowRight className="ml-1 h-4 w-4" />
        </Link>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {work.duration_sec > 0 ? <Badge variant="outline">{work.duration_sec.toFixed(2)}s</Badge> : null}
        {(work.tags ?? []).map((tag) => (
          <Badge key={tag} variant="outline" className="bg-muted/30">
            {tag}
          </Badge>
        ))}
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button variant="outline" size="sm" onClick={onEdit}>
          <PencilLine className="mr-2 h-4 w-4" />
          编辑
        </Button>
        <Button variant="outline" size="sm" onClick={onDelete} disabled={deleting} className="border-rose-200 text-rose-700 hover:bg-rose-50">
          {deleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Trash2 className="mr-2 h-4 w-4" />}
          删除
        </Button>
      </div>
    </div>
  );
}

function AudioMetaPanel({ title, meta }: { title: string; meta: Record<string, unknown> }) {
  const format = readString(meta, 'format');
  const duration = readNumber(meta, 'duration_sec');
  const sampleRate = readNumber(meta, 'sample_rate');
  const bitDepth = readNumber(meta, 'bit_depth');
  const channels = readNumber(meta, 'channels');

  return (
    <div className="rounded-2xl border border-border/70 bg-muted/10 p-4">
      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">{title}</p>
      <div className="mt-3 flex flex-wrap gap-2">
        {format ? <Badge variant="outline">{format.toUpperCase()}</Badge> : null}
        {duration !== null ? <Badge variant="outline">{duration.toFixed(2)}s</Badge> : null}
        {sampleRate !== null ? <Badge variant="outline">{sampleRate} Hz</Badge> : null}
        {bitDepth !== null ? <Badge variant="outline">{bitDepth} bit</Badge> : null}
        {channels !== null ? <Badge variant="outline">{channels} ch</Badge> : null}
      </div>
    </div>
  );
}
