'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ClipboardList, Download, Search } from 'lucide-react'
import { apiClient, AuditLog } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminFilterBar } from '@/components/admin/admin-filter-bar'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'

const actionOptions = [
  { value: '', label: '全部动作' },
  { value: 'create', label: '创建' },
  { value: 'update', label: '更新' },
  { value: 'delete', label: '删除' },
  { value: 'view', label: '查看' },
  { value: 'export', label: '导出' },
]

const resourceOptions = [
  { value: '', label: '全部资源' },
  { value: 'user', label: '用户' },
  { value: 'post', label: '帖子' },
  { value: 'comment', label: '评论' },
  { value: 'report', label: '举报' },
  { value: 'group', label: '圈子' },
  { value: 'event', label: '活动' },
  { value: 'order', label: '订单' },
  { value: 'assistant', label: 'AI 助手' },
]

function prettyJSON(value?: string | null) {
  if (!value) return '—'
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

export default function AdminAuditLogsPage() {
  const [page, setPage] = useState(1)
  const [userIdInput, setUserIdInput] = useState('')
  const [resourceIdInput, setResourceIdInput] = useState('')
  const [userId, setUserId] = useState('')
  const [resourceId, setResourceId] = useState('')
  const [action, setAction] = useState('')
  const [resource, setResource] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-audit-logs', page, userId, resourceId, action, resource],
    queryFn: () =>
      apiClient.getAdminAuditLogs({
        page,
        page_size: 20,
        user_id: userId || undefined,
        resource_id: resourceId || undefined,
        action: action || undefined,
        resource: resource || undefined,
      }),
  })

  const logs = data?.logs ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1

  const columns: AdminColumn<AuditLog>[] = [
    {
      key: 'action',
      header: '动作',
      className: 'min-w-[220px]',
      render: (log) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={log.action} label={log.action} />
            <AdminStatusBadge value={log.resource} label={log.resource} />
          </div>
          <div>
            <p className="font-medium text-slate-900">{log.username || 'anonymous'}</p>
            <p className="mt-1 text-sm text-slate-500">
              user_id: {log.user_id || '—'}
            </p>
            <p className="text-sm text-slate-500">
              resource_id: {log.resource_id || '—'}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'network',
      header: '来源',
      className: 'min-w-[220px] text-slate-500',
      render: (log) => (
        <div className="space-y-1">
          <p>IP: {log.ip_address}</p>
          <p className="line-clamp-2 break-all text-xs">{log.user_agent || '—'}</p>
        </div>
      ),
    },
    {
      key: 'payload',
      header: '变更数据',
      className: 'min-w-[320px]',
      render: (log) => (
        <div className="space-y-2">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-slate-400">After</p>
            <pre className="mt-1 max-h-28 overflow-auto rounded-2xl bg-slate-50 p-3 text-xs leading-5 text-slate-600">
              {prettyJSON(log.after_data)}
            </pre>
          </div>
          {log.before_data ? (
            <div>
              <p className="text-xs font-medium uppercase tracking-[0.18em] text-slate-400">Before</p>
              <pre className="mt-1 max-h-28 overflow-auto rounded-2xl bg-slate-50 p-3 text-xs leading-5 text-slate-600">
                {prettyJSON(log.before_data)}
              </pre>
            </div>
          ) : null}
        </div>
      ),
    },
    {
      key: 'created_at',
      header: '时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (log) => new Date(log.created_at).toLocaleString('zh-CN'),
    },
  ]

  async function handleExport() {
    const blob = await apiClient.exportAdminAuditLogsCsv({
      user_id: userId || undefined,
      resource_id: resourceId || undefined,
      action: action || undefined,
      resource: resource || undefined,
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `audit_logs_${new Date().toISOString().replace(/[:.]/g, '-')}.csv`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Audit Trail"
        title="审计日志"
        description="用于追踪后台关键动作的执行记录，适合排查误操作、回溯状态变化和做合规留痕。"
        actions={
          <Button variant="outline" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            导出 CSV
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="日志结果数"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前筛选条件下的审计记录"
          icon={ClipboardList}
          tone="brand"
        />
        <AdminMetricCard
          label="当前页记录"
          value={logs.length.toLocaleString()}
          hint="单页最多 20 条"
          icon={ClipboardList}
          tone="default"
        />
        <AdminMetricCard
          label="唯一动作类型"
          value={new Set(logs.map((item) => item.action)).size.toLocaleString()}
          hint="仅统计当前页"
          icon={ClipboardList}
          tone="default"
        />
      </div>

      <AdminFilterBar>
        <div className="flex flex-1 flex-col gap-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                value={userIdInput}
                onChange={(e) => setUserIdInput(e.target.value)}
                placeholder="按 user_id 过滤"
                className="pl-10"
              />
            </div>
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                value={resourceIdInput}
                onChange={(e) => setResourceIdInput(e.target.value)}
                placeholder="按 resource_id 过滤"
                className="pl-10"
              />
            </div>
            <select
              value={action}
              onChange={(e) => {
                setPage(1)
                setAction(e.target.value)
              }}
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
            >
              {actionOptions.map((option) => (
                <option key={option.value || 'all'} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <select
              value={resource}
              onChange={(e) => {
                setPage(1)
                setResource(e.target.value)
              }}
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
            >
              {resourceOptions.map((option) => (
                <option key={option.value || 'all'} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-800"
            onClick={() => {
              setPage(1)
              setUserId(userIdInput.trim())
              setResourceId(resourceIdInput.trim())
            }}
          >
            应用筛选
          </button>
          <button
            className="rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
            onClick={() => {
              setPage(1)
              setUserIdInput('')
              setResourceIdInput('')
              setUserId('')
              setResourceId('')
              setAction('')
              setResource('')
            }}
          >
            重置
          </button>
        </div>
      </AdminFilterBar>

      <AdminDataTable
        data={logs}
        columns={columns}
        keyExtractor={(log) => log.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有匹配的审计日志"
            description="当前筛选条件下没有记录，可以放宽过滤条件后再试。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
