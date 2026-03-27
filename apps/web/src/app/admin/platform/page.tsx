'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { Activity, Bot, ClipboardList, Shield, SlidersHorizontal } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import {
  AdminSectionCard,
  AdminWorkspaceCard,
  formatCountRecord,
  formatDateTime,
} from '@/components/admin/admin-dashboard-kit'

export default function AdminPlatformPage() {
  const { data: systemData } = useQuery({
    queryKey: ['admin-dashboard-platform-system'],
    queryFn: () => apiClient.getAdminSystemConfig(),
  })

  const { data: assistantOverview } = useQuery({
    queryKey: ['admin-dashboard-platform-assistant'],
    queryFn: () => apiClient.getAssistantOverview(),
  })

  const { data: permissionData } = useQuery({
    queryKey: ['admin-dashboard-platform-permissions'],
    queryFn: () => apiClient.getAdminPermissionMatrix(),
  })

  const { data: auditData } = useQuery({
    queryKey: ['admin-dashboard-platform-audit'],
    queryFn: () => apiClient.getAdminAuditLogs({ page: 1, page_size: 6 }),
  })

  const totalPermissionCount = (permissionData?.roles ?? []).reduce(
    (sum, item) => sum + item.count,
    0,
  )

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Platform Dashboard"
        title="平台配置盘"
        description="集中观察 AI 助手、系统配置、权限矩阵和最近后台动作，适合作为管理员和研发支持的统一平台视角。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/audit-logs">查看审计流</Link>
            </Button>
            <Button asChild className="border border-[#c5dfd3] bg-[#edf8f2] text-[#21584e] hover:bg-[#e3f3eb]">
              <Link href="/admin/system">进入系统配置</Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <AdminMetricCard
          label="助手状态"
          value={systemData?.assistant.configured ? '已配置' : '未配置'}
          hint="仅表示模型凭证和核心配置是否存在"
          icon={Bot}
          tone={systemData?.assistant.configured ? 'success' : 'warning'}
        />
        <AdminMetricCard
          label="权限角色"
          value={(permissionData?.roles.length ?? 0).toLocaleString()}
          hint="当前权限矩阵中的角色数量"
          icon={Shield}
          tone="brand"
        />
        <AdminMetricCard
          label="权限映射"
          value={totalPermissionCount.toLocaleString()}
          hint="所有角色权限项计数之和"
          icon={Shield}
          tone="default"
        />
        <AdminMetricCard
          label="审计记录"
          value={(auditData?.total ?? 0).toLocaleString()}
          hint="后台关键动作留痕总量"
          icon={ClipboardList}
          tone={(auditData?.total ?? 0) > 0 ? 'brand' : 'default'}
        />
        <AdminMetricCard
          label="播放事件"
          value={String(systemData?.audio_metrics.playback_events_total ?? 0)}
          hint="音频播放埋点累计事件数"
          icon={Activity}
          tone={(systemData?.audio_metrics.playback_events_total ?? 0) > 0 ? 'success' : 'default'}
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-4">
        <AdminWorkspaceCard
          href="/admin/assistant"
          title="AI 助手"
          description="查看检索索引、fallback、反馈和多模态运行态。"
          icon={Bot}
          tone="brand"
          metrics={[
            {
              label: '索引文档',
              value: (assistantOverview?.overview.indexed_documents ?? 0).toLocaleString(),
            },
            {
              label: 'Fallback',
              value: (assistantOverview?.metrics.fallback_total ?? 0).toLocaleString(),
            },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/system"
          title="系统配置"
          description="查看服务模式、限流配置、基础设施和赞助展示快照。"
          icon={SlidersHorizontal}
          tone="success"
          metrics={[
            { label: '模式', value: systemData?.server.mode || '—' },
            {
              label: '来源数',
              value: (systemData?.server.allow_origins.length ?? 0).toLocaleString(),
            },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/permissions"
          title="权限矩阵"
          description="查看角色与权限映射，快速确认后台访问边界。"
          icon={Shield}
          tone="default"
          metrics={[
            { label: '角色', value: (permissionData?.roles.length ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/audit-logs"
          title="审计留痕"
          description="回看最近后台动作，适合排查配置错误和误操作。"
          icon={ClipboardList}
          tone="warning"
          metrics={[
            { label: '总量', value: (auditData?.total ?? 0).toLocaleString() },
          ]}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSectionCard
          title="AI 助手运行态"
          description="把索引、检索、fallback 和多模态链路集中到一屏，方便判断助手是否健康。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/assistant">进入 AI 助手工作台</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
              <p className="text-sm font-medium text-slate-900">索引与检索</p>
              <ul className="mt-3 space-y-2 text-sm text-slate-600">
                <li>索引文档：{(assistantOverview?.overview.indexed_documents ?? 0).toLocaleString()}</li>
                <li>最近索引时间：{formatDateTime(assistantOverview?.overview.last_indexed_at)}</li>
                <li>检索总量：{(assistantOverview?.metrics.retrievals_total ?? 0).toLocaleString()}</li>
                <li>上次检索文档数：{(assistantOverview?.metrics.last_retrieved_documents ?? 0).toLocaleString()}</li>
              </ul>
            </div>
            <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
              <p className="text-sm font-medium text-slate-900">稳定性信号</p>
              <ul className="mt-3 space-y-2 text-sm text-slate-600">
                <li>Fallback：{(assistantOverview?.metrics.fallback_total ?? 0).toLocaleString()}</li>
                <li>反馈总量：{(assistantOverview?.metrics.feedback_total ?? 0).toLocaleString()}</li>
                <li>多模态请求：{(assistantOverview?.metrics.multimodal_requests_total ?? 0).toLocaleString()}</li>
                <li>最近索引错误：{assistantOverview?.metrics.last_index_error || '无'}</li>
              </ul>
            </div>
          </div>
        </AdminSectionCard>

        <AdminSectionCard
          title="运行配置快照"
          description="高频查看的配置项不需要翻完整页，在这里先看服务模式、支付、邮件和播放信号。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/system">查看系统配置明细</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
              <p className="text-sm font-medium text-slate-900">服务基础</p>
              <ul className="mt-3 space-y-2 text-sm text-slate-600">
                <li>运行模式：{systemData?.server.mode || '—'}</li>
                <li>前端地址：{systemData?.server.frontend_url || '—'}</li>
                <li>允许来源：{(systemData?.server.allow_origins.length ?? 0).toLocaleString()} 个</li>
                <li>邮件配置：{systemData?.email.configured ? '已配置' : '未配置'}</li>
              </ul>
            </div>
            <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
              <p className="text-sm font-medium text-slate-900">业务与基础设施</p>
              <ul className="mt-3 space-y-2 text-sm text-slate-600">
                <li>支付宝：{systemData?.payment.alipay_configured ? '已配置' : '未配置'}</li>
                <li>微信支付：{systemData?.payment.wechat_configured ? '已配置' : '未配置'}</li>
                <li>OSS Provider：{systemData?.oss.provider || '—'}</li>
                <li>播放事件分布：{formatCountRecord(systemData?.audio_metrics.events_by_type)}</li>
              </ul>
            </div>
          </div>
        </AdminSectionCard>
      </div>

      <AdminSectionCard
        title="权限与审计摘要"
        description="进入后台前先看一眼谁能访问、最近做了什么、最近是否有异常操作。"
      >
        <div className="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
          <div className="space-y-3">
            {(permissionData?.roles ?? []).map((role) => (
              <div
                key={role.role}
                className="flex items-center justify-between rounded-2xl border border-slate-200 bg-slate-50/70 px-4 py-4"
              >
                <div>
                  <p className="font-medium text-slate-950">{role.role}</p>
                  <p className="mt-1 text-sm text-slate-500">
                    {role.permissions.slice(0, 4).join('、') || '无权限'}
                  </p>
                </div>
                <Badge
                  variant="outline"
                  className="rounded-full border-slate-200 bg-white text-slate-600"
                >
                  {role.count} 项
                </Badge>
              </div>
            ))}
          </div>

          <div className="space-y-3">
            {(auditData?.logs ?? []).map((log) => (
              <div
                key={log.id}
                className="rounded-2xl border border-slate-200 bg-slate-50/70 px-4 py-4"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="font-medium text-slate-950">
                    {log.action} · {log.resource}
                  </p>
                  <span className="text-xs text-slate-400">
                    {formatDateTime(log.created_at)}
                  </span>
                </div>
                <p className="mt-2 text-sm text-slate-500">
                  {log.username || '未知操作者'} · {log.resource_id || '无资源 ID'}
                </p>
                {log.error_message ? (
                  <p className="mt-2 text-sm text-rose-600">
                    错误：{log.error_message}
                  </p>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      </AdminSectionCard>
    </div>
  )
}
