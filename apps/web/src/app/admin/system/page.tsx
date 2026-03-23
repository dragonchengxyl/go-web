'use client'

import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Activity, AudioLines, Radio, Settings2, ShieldCheck, SlidersHorizontal, Sparkles } from 'lucide-react'
import {
  AdminSponsorConfig,
  AdminSystemConfig,
  apiClient,
} from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { showAdminToast } from '@/components/admin/admin-toast'

function RuntimeCard({
  title,
  rows,
}: {
  title: string;
  rows: Array<{ label: string; value: React.ReactNode }>;
}) {
  return (
    <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg">{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {rows.map((row) => (
          <div key={row.label} className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3 last:border-0 last:pb-0">
            <span className="text-sm text-slate-500">{row.label}</span>
            <span className="max-w-[60%] break-all text-right text-sm font-medium text-slate-900">
              {row.value}
            </span>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

function formatEventDistribution(events?: Record<string, number>) {
  if (!events) return '—'
  const items = Object.entries(events).sort(([, left], [, right]) => right - left)
  if (items.length === 0) return '—'
  return items.map(([event, count]) => `${event}×${count}`).join(', ')
}

export default function AdminSystemPage() {
  const queryClient = useQueryClient()
  const [sponsorForm, setSponsorForm] = useState<AdminSponsorConfig>({
    monthly_goal: 0,
    current_raised: 0,
    alipay_qr_url: '',
    wechat_qr_url: '',
    message: '',
  })

  const { data, isLoading } = useQuery<AdminSystemConfig>({
    queryKey: ['admin-system-config'],
    queryFn: () => apiClient.getAdminSystemConfig(),
  })

  useEffect(() => {
    if (data?.sponsor) {
      setSponsorForm(data.sponsor)
    }
  }, [data])

  const updateSponsorMutation = useMutation({
    mutationFn: (payload: AdminSponsorConfig) => apiClient.updateAdminSponsorConfig(payload),
    onSuccess: (next) => {
      setSponsorForm(next)
      queryClient.invalidateQueries({ queryKey: ['admin-system-config'] })
      showAdminToast('赞助展示配置已更新', 'success')
    },
    onError: () => {
      showAdminToast('更新赞助配置失败', 'error')
    },
  })

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Runtime Configuration"
        title="系统配置"
        description="查看服务运行时配置快照，并维护对外展示的赞助配置。敏感凭证不会在这里回显。"
      />

      <div className="grid gap-4 md:grid-cols-4">
        <AdminMetricCard
          label="服务模式"
          value={data?.server.mode || '—'}
          hint="当前后端运行模式"
          icon={Settings2}
          tone="brand"
        />
        <AdminMetricCard
          label="助手状态"
          value={data?.assistant.configured ? '已配置' : '未配置'}
          hint="仅表示上游模型凭证是否存在"
          icon={Sparkles}
          tone={data?.assistant.configured ? 'success' : 'warning'}
        />
        <AdminMetricCard
          label="邮件能力"
          value={data?.email.configured ? '已配置' : '未配置'}
          hint="SMTP 当前是否可用"
          icon={ShieldCheck}
          tone={data?.email.configured ? 'success' : 'warning'}
        />
        <AdminMetricCard
          label="允许来源数"
          value={String(data?.server.allow_origins.length ?? 0)}
          hint="Server allow_origins 条目数"
          icon={SlidersHorizontal}
          tone="default"
        />
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="播放事件总量"
          value={String(data?.audio_metrics.playback_events_total ?? 0)}
          hint="播放器累计上报的播放事件数"
          icon={AudioLines}
          tone={(data?.audio_metrics.playback_events_total ?? 0) > 0 ? 'brand' : 'default'}
        />
        <AdminMetricCard
          label="最近事件"
          value={data?.audio_metrics.last_event || '—'}
          hint="最近一次捕获到的播放动作"
          icon={Activity}
          tone={data?.audio_metrics.last_event ? 'success' : 'default'}
        />
        <AdminMetricCard
          label="最近来源"
          value={data?.audio_metrics.last_source_kind || '—'}
          hint="最近一次事件的来源上下文"
          icon={Radio}
          tone={data?.audio_metrics.last_source_kind ? 'warning' : 'default'}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
        <div className="grid gap-6 md:grid-cols-2">
          <RuntimeCard
            title="服务基础"
            rows={[
              { label: '模式', value: data?.server.mode || '—' },
              { label: '端口', value: data?.server.port ?? '—' },
              { label: '前端地址', value: data?.server.frontend_url || '—' },
              {
                label: '允许来源',
                value: data?.server.allow_origins?.length
                  ? data.server.allow_origins.join(', ')
                  : '—',
              },
            ]}
          />
          <RuntimeCard
            title="限流设置"
            rows={[
              { label: '未登录', value: data?.ratelimit.unauthenticated ?? '—' },
              { label: '已登录', value: data?.ratelimit.authenticated ?? '—' },
              { label: '管理员', value: data?.ratelimit.admin ?? '—' },
            ]}
          />
          <RuntimeCard
            title="AI 助手"
            rows={[
              { label: 'Provider', value: data?.assistant.provider || '—' },
              { label: 'Model', value: data?.assistant.model || '—' },
              { label: 'Embedding', value: data?.assistant.embedding_model || '—' },
              { label: 'Vision', value: data?.assistant.vision_model || '—' },
              { label: 'Base URL', value: data?.assistant.base_url || '—' },
              { label: 'Embedding URL', value: data?.assistant.embedding_base_url || '—' },
              { label: 'Vision URL', value: data?.assistant.vision_base_url || '—' },
              { label: 'Timeout', value: data?.assistant.timeout_sec ? `${data.assistant.timeout_sec}s` : '—' },
              { label: 'Vision Timeout', value: data?.assistant.vision_timeout_sec ? `${data.assistant.vision_timeout_sec}s` : '—' },
              { label: 'Max Context', value: data?.assistant.max_context_items ?? '—' },
              { label: 'Retrieval Limit', value: data?.assistant.retrieval_limit ?? '—' },
              { label: 'Vector Scan', value: data?.assistant.vector_scan_limit ?? '—' },
              { label: 'Sync Interval', value: data?.assistant.sync_interval_sec ? `${data.assistant.sync_interval_sec}s` : '—' },
              { label: 'Persona', value: data?.assistant.persona_name || '—' },
            ]}
          />
          <RuntimeCard
            title="基础设施"
            rows={[
              { label: 'OSS Provider', value: data?.oss.provider || '—' },
              { label: 'Bucket', value: data?.oss.bucket || '—' },
              { label: 'Endpoint', value: data?.oss.endpoint || '—' },
              { label: 'Region', value: data?.oss.region || '—' },
              {
                label: 'Allowed Hosts',
                value: data?.oss.allowed_hosts?.length
                  ? data.oss.allowed_hosts.join(', ')
                  : '—',
              },
            ]}
          />
          <RuntimeCard
            title="邮件 / 支付"
            rows={[
              { label: 'SMTP Host', value: data?.email.host || '—' },
              { label: '发件人', value: data?.email.from || '—' },
              { label: '支付宝', value: data?.payment.alipay_configured ? '已配置' : '未配置' },
              { label: '微信支付', value: data?.payment.wechat_configured ? '已配置' : '未配置' },
            ]}
          />
          <RuntimeCard
            title="gRPC"
            rows={[
              { label: 'Stats Addr', value: data?.grpc.stats_addr || '—' },
              { label: 'Notification Addr', value: data?.grpc.notification_addr || '—' },
              { label: 'Moderation Addr', value: data?.grpc.moderation_addr || '—' },
              { label: 'Stats Port', value: data?.grpc.stats_port ?? '—' },
              { label: 'Notification Port', value: data?.grpc.notification_port ?? '—' },
              { label: 'Moderation Port', value: data?.grpc.moderation_port ?? '—' },
            ]}
          />
          <RuntimeCard
            title="音频播放"
            rows={[
              { label: '事件总量', value: data?.audio_metrics.playback_events_total ?? '—' },
              { label: '最近事件', value: data?.audio_metrics.last_event || '—' },
              { label: '最近来源', value: data?.audio_metrics.last_source_kind || '—' },
              {
                label: '最近位置',
                value:
                  typeof data?.audio_metrics.last_position_sec === 'number'
                    ? `${data.audio_metrics.last_position_sec.toFixed(1)}s`
                    : '—',
              },
              {
                label: '事件分布',
                value: formatEventDistribution(data?.audio_metrics.events_by_type),
              },
            ]}
          />
        </div>

        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader>
            <CardTitle className="text-lg">赞助展示配置</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoading ? (
              <div className="py-12 text-center text-sm text-slate-500">正在加载系统配置...</div>
            ) : (
              <>
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">月度目标</label>
                    <Input
                      type="number"
                      min="0"
                      value={sponsorForm.monthly_goal}
                      onChange={(e) =>
                        setSponsorForm((prev) => ({
                          ...prev,
                          monthly_goal: Number(e.target.value) || 0,
                        }))
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">当前已筹</label>
                    <Input
                      type="number"
                      min="0"
                      value={sponsorForm.current_raised}
                      onChange={(e) =>
                        setSponsorForm((prev) => ({
                          ...prev,
                          current_raised: Number(e.target.value) || 0,
                        }))
                      }
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">支付宝收款码</label>
                  <Input
                    value={sponsorForm.alipay_qr_url}
                    onChange={(e) =>
                      setSponsorForm((prev) => ({
                        ...prev,
                        alipay_qr_url: e.target.value,
                      }))
                    }
                    placeholder="https://..."
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">微信收款码</label>
                  <Input
                    value={sponsorForm.wechat_qr_url}
                    onChange={(e) =>
                      setSponsorForm((prev) => ({
                        ...prev,
                        wechat_qr_url: e.target.value,
                      }))
                    }
                    placeholder="https://..."
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">展示文案</label>
                  <Textarea
                    rows={6}
                    value={sponsorForm.message}
                    onChange={(e) =>
                      setSponsorForm((prev) => ({
                        ...prev,
                        message: e.target.value,
                      }))
                    }
                    placeholder="感谢支持社区运营..."
                  />
                </div>

                <div className="flex justify-end">
                  <Button
                    className="bg-slate-950 text-white hover:bg-slate-800"
                    disabled={updateSponsorMutation.isPending}
                    onClick={() => updateSponsorMutation.mutate(sponsorForm)}
                  >
                    保存赞助配置
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
