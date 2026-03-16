'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CalendarRange, CheckCircle2, Clock3, MapPin, RotateCcw } from 'lucide-react'
import { AdminEvent, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

const tabs = [
  { value: '', label: '全部活动' },
  { value: 'published', label: '已发布' },
  { value: 'cancelled', label: '已取消' },
  { value: 'completed', label: '已结束' },
]

const statusLabelMap: Record<string, string> = {
  draft: '草稿',
  published: '已发布',
  cancelled: '已取消',
  completed: '已结束',
}

export default function AdminEventsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-events', page, status],
    queryFn: () =>
      apiClient.getAdminEvents({
        page,
        page_size: 20,
        status: status || undefined,
      }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, nextStatus }: { id: string; nextStatus: AdminEvent['status'] }) =>
      apiClient.updateAdminEventStatus(id, nextStatus),
    onSuccess: (_, { nextStatus }) => {
      queryClient.invalidateQueries({ queryKey: ['admin-events'] })
      showAdminToast(`活动已更新为${statusLabelMap[nextStatus] ?? nextStatus}`, 'success')
    },
    onError: () => {
      showAdminToast('更新活动状态失败', 'error')
    },
  })

  const events = data?.events ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1
  const stats = {
    published: events.filter((item) => item.status === 'published').length,
    cancelled: events.filter((item) => item.status === 'cancelled').length,
    completed: events.filter((item) => item.status === 'completed').length,
  }

  const columns: AdminColumn<AdminEvent>[] = [
    {
      key: 'event',
      header: '活动',
      className: 'min-w-[320px]',
      render: (event) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={event.status} label={statusLabelMap[event.status] ?? event.status} />
            {event.is_online ? (
              <AdminStatusBadge value="reviewed" label="线上活动" />
            ) : (
              <AdminStatusBadge value="pending" label="线下活动" />
            )}
          </div>
          <div>
            <p className="font-medium text-slate-900">{event.title}</p>
            <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">
              {event.description || '暂无活动说明'}
            </p>
            <p className="mt-2 text-sm text-slate-400">
              发起人：{event.organizer_username ? `@${event.organizer_username}` : event.organizer_id}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'schedule',
      header: '时间 / 地点',
      className: 'min-w-[220px] text-slate-500',
      render: (event) => (
        <div className="space-y-1">
          <p>{new Date(event.start_time).toLocaleString('zh-CN')}</p>
          <p>{new Date(event.end_time).toLocaleString('zh-CN')}</p>
          <p className="inline-flex items-center gap-1">
            <MapPin className="h-3.5 w-3.5" />
            {event.location || (event.is_online ? '线上' : '待补充')}
          </p>
        </div>
      ),
    },
    {
      key: 'metrics',
      header: '报名',
      className: 'whitespace-nowrap text-slate-500',
      render: (event) => (
        <div className="space-y-1">
          <p>参与人数 {event.attendee_count}</p>
          <p>人数上限 {event.max_capacity || '不限'}</p>
          <p>{new Date(event.created_at).toLocaleDateString('zh-CN')}</p>
        </div>
      ),
    },
    {
      key: 'actions',
      header: '操作',
      className: 'w-[280px]',
      render: (event) => (
        <div className="flex flex-wrap gap-2">
          <Button asChild variant="outline" size="sm">
            <Link href={`/events/${event.id}`} target="_blank">
              查看活动
            </Link>
          </Button>
          {event.status !== 'published' ? (
            <Button
              size="sm"
              variant="outline"
              className="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
              disabled={updateMutation.isPending}
              onClick={() => updateMutation.mutate({ id: event.id, nextStatus: 'published' })}
            >
              <RotateCcw className="mr-1 h-4 w-4" />
              发布
            </Button>
          ) : null}
          {event.status !== 'completed' ? (
            <Button
              size="sm"
              variant="outline"
              className="border-slate-200 text-slate-700 hover:bg-slate-50"
              disabled={updateMutation.isPending}
              onClick={() => updateMutation.mutate({ id: event.id, nextStatus: 'completed' })}
            >
              <CheckCircle2 className="mr-1 h-4 w-4" />
              标记结束
            </Button>
          ) : null}
          {event.status !== 'cancelled' ? (
            <Button
              size="sm"
              variant="outline"
              className="border-rose-200 text-rose-700 hover:bg-rose-50"
              disabled={updateMutation.isPending}
              onClick={() => updateMutation.mutate({ id: event.id, nextStatus: 'cancelled' })}
            >
              取消活动
            </Button>
          ) : null}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Event Operations"
        title="活动运营"
        description="用于巡检活动状态、调整活动生命周期，以及查看报名和举办信息。"
      />

      <div className="grid gap-4 md:grid-cols-4">
        <AdminMetricCard
          label="活动结果数"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前筛选条件下的活动总数"
          icon={CalendarRange}
          tone="brand"
        />
        <AdminMetricCard
          label="本页已发布"
          value={stats.published.toLocaleString()}
          hint="对外可见的活动"
          icon={Clock3}
          tone="success"
        />
        <AdminMetricCard
          label="本页已取消"
          value={stats.cancelled.toLocaleString()}
          hint="需要额外关注说明"
          icon={CalendarRange}
          tone={stats.cancelled > 0 ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="本页已结束"
          value={stats.completed.toLocaleString()}
          hint="已闭环活动"
          icon={CheckCircle2}
          tone="default"
        />
      </div>

      <div className="flex flex-wrap gap-2">
        {tabs.map((tab) => (
          <button
            key={tab.value || 'all'}
            onClick={() => {
              setPage(1)
              setStatus(tab.value)
            }}
            className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
              status === tab.value
                ? 'border-slate-950 bg-slate-950 text-white'
                : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-900'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <AdminDataTable
        data={events}
        columns={columns}
        keyExtractor={(event) => event.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="当前筛选下没有活动"
            description="切换状态标签后再试，或者等待新的活动进入列表。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
