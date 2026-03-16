'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ExternalLink, Flag, ShieldAlert, ShieldCheck } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

interface ReportRow {
  id: string
  target_type: string
  target_id: string
  reason: string
  description: string
  reporter_username: string
  status: string
  action_taken?: string
  created_at: string
}

interface ListReportsOutput {
  reports: ReportRow[]
  total: number
  page: number
}

const tabs = [
  { label: '待处理', value: 'pending' },
  { label: '已处理', value: 'reviewed' },
  { label: '已忽略', value: 'dismissed' },
]

const targetLabels: Record<string, string> = {
  post: '帖子',
  comment: '评论',
  user: '用户',
}

const actionLabels: Record<string, string> = {
  block_post: '已封禁帖子',
  delete_comment: '已删除评论',
  ban_user: '已封禁用户',
}

export default function AdminReportsPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState('pending')
  const [page, setPage] = useState(1)
  const [confirming, setConfirming] = useState<{ id: string; status: 'reviewed' | 'dismissed'; action?: string } | null>(null)
  const pageSize = 20

  const { data, isLoading } = useQuery<ListReportsOutput>({
    queryKey: ['admin-reports', tab, page],
    queryFn: () =>
      apiClient.get(`/admin/reports?status=${tab}&page=${page}&page_size=${pageSize}`),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, status, action }: { id: string; status: string; action?: string }) =>
      apiClient.put(`/admin/reports/${id}`, { status, action }),
    onSuccess: (_, { status, action }) => {
      setConfirming(null)
      if (status === 'dismissed') {
        showAdminToast('举报已忽略', 'success')
      } else if (action && actionLabels[action]) {
        showAdminToast(actionLabels[action], 'success')
      } else {
        showAdminToast('举报已处理', 'success')
      }
      queryClient.invalidateQueries({ queryKey: ['admin-reports'] })
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] })
    },
    onError: () => {
      showAdminToast('处理失败，请重试', 'error')
    },
  })

  const reports = data?.reports ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / pageSize)) : 1

  const columns: AdminColumn<ReportRow>[] = [
    {
      key: 'report',
      header: '举报项',
      className: 'min-w-[320px]',
      render: (report) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={report.status} label={tabs.find((item) => item.value === report.status)?.label ?? report.status} />
            <AdminStatusBadge value={report.target_type} label={targetLabels[report.target_type] ?? report.target_type} />
            {report.action_taken ? (
              <AdminStatusBadge value="reviewed" label={actionLabels[report.action_taken] ?? report.action_taken} />
            ) : null}
          </div>
          <div>
            <p className="font-medium text-slate-900">{report.reason}</p>
            {report.description ? (
              <p className="mt-1 text-sm leading-6 text-slate-500">
                {report.description}
              </p>
            ) : null}
            <p className="mt-2 text-sm text-slate-400">
              举报人：{report.reporter_username || '未知用户'}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'created_at',
      header: '时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (report) => new Date(report.created_at).toLocaleString('zh-CN'),
    },
    {
      key: 'actions',
      header: '处理动作',
      className: 'w-[320px]',
      render: (report) => {
        if (tab !== 'pending') {
          return report.target_type === 'post' ? (
            <Button asChild variant="outline" size="sm">
              <Link href={`/posts/${report.target_id}`} target="_blank">
                查看原帖
              </Link>
            </Button>
          ) : (
            <span className="text-sm text-slate-400">已归档</span>
          )
        }

        const actionButton =
          report.target_type === 'post'
            ? { action: 'block_post', label: '封禁帖子' }
            : report.target_type === 'comment'
              ? { action: 'delete_comment', label: '删除评论' }
              : report.target_type === 'user'
                ? { action: 'ban_user', label: '封禁用户' }
                : null

        const isConfirming = confirming?.id === report.id

        return (
          <div className="flex flex-wrap items-center gap-2">
            {report.target_type === 'post' ? (
              <Button asChild variant="outline" size="sm">
                <Link href={`/posts/${report.target_id}`} target="_blank">
                  查看原帖
                  <ExternalLink className="ml-1 h-4 w-4" />
                </Link>
              </Button>
            ) : null}
            <Button
              size="sm"
              variant={isConfirming && confirming?.status === 'reviewed' && !confirming?.action ? 'default' : 'outline'}
              className={isConfirming && confirming?.status === 'reviewed' && !confirming?.action ? 'bg-slate-950 text-white hover:bg-slate-800' : ''}
              disabled={updateMutation.isPending}
              onClick={() => {
                if (isConfirming && confirming?.status === 'reviewed' && !confirming?.action) {
                  updateMutation.mutate({ id: report.id, status: 'reviewed' })
                } else {
                  setConfirming({ id: report.id, status: 'reviewed' })
                }
              }}
            >
              仅标记处理
            </Button>
            {actionButton ? (
              <Button
                size="sm"
                variant={isConfirming && confirming?.action === actionButton.action ? 'default' : 'outline'}
                className={isConfirming && confirming?.action === actionButton.action ? 'bg-rose-600 text-white hover:bg-rose-500' : 'border-rose-200 text-rose-700 hover:bg-rose-50'}
                disabled={updateMutation.isPending}
                onClick={() => {
                  if (isConfirming && confirming?.action === actionButton.action) {
                    updateMutation.mutate({
                      id: report.id,
                      status: 'reviewed',
                      action: actionButton.action,
                    })
                  } else {
                    setConfirming({
                      id: report.id,
                      status: 'reviewed',
                      action: actionButton.action,
                    })
                  }
                }}
              >
                {actionButton.label}
              </Button>
            ) : null}
            <Button
              size="sm"
              variant={isConfirming && confirming?.status === 'dismissed' ? 'default' : 'outline'}
              className={isConfirming && confirming?.status === 'dismissed' ? 'bg-slate-500 text-white hover:bg-slate-400' : ''}
              disabled={updateMutation.isPending}
              onClick={() => {
                if (isConfirming && confirming?.status === 'dismissed') {
                  updateMutation.mutate({ id: report.id, status: 'dismissed' })
                } else {
                  setConfirming({ id: report.id, status: 'dismissed' })
                }
              }}
            >
              忽略
            </Button>
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Report Workflow"
        title="举报处理"
        description="这里负责完成举报查看、动作执行和反馈闭环，适合高频治理场景。"
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="当前队列"
          value={(data?.total ?? 0).toLocaleString()}
          hint={`当前标签：${tabs.find((item) => item.value === tab)?.label ?? tab}`}
          icon={Flag}
          tone={tab === 'pending' ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="本页记录"
          value={reports.length.toLocaleString()}
          hint="单页最多展示 20 条"
          icon={ShieldAlert}
          tone="default"
        />
        <div className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <p className="text-sm font-medium text-slate-500">状态筛选</p>
          <div className="mt-4 flex flex-wrap gap-2">
            {tabs.map((item) => (
              <button
                key={item.value}
                onClick={() => {
                  setTab(item.value)
                  setPage(1)
                  setConfirming(null)
                }}
                className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
                  tab === item.value
                    ? 'border-slate-950 bg-slate-950 text-white'
                    : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-900'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <AdminDataTable
        data={reports}
        columns={columns}
        keyExtractor={(report) => report.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="当前筛选下没有举报"
            description="系统暂时没有新的举报要处理，可以继续巡检其他工作面。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
