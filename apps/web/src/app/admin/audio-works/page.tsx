'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircle2, Disc3, Loader2, XCircle } from 'lucide-react'
import { AdminAudioWork, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

const tabs = [
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已封禁', value: 'blocked' },
]

export default function AdminAudioWorksPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState('pending')
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading } = useQuery({
    queryKey: ['admin-audio-works', tab, page],
    queryFn: () => apiClient.getAdminAudioWorks({ status: tab, page, page_size: pageSize }),
  })

  const moderationMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'approved' | 'blocked' }) =>
      apiClient.updateAdminAudioWorkModeration(id, status),
    onSuccess: (_, { status }) => {
      showAdminToast(status === 'approved' ? '音频作品已通过审核' : '音频作品已封禁', 'success')
      queryClient.invalidateQueries({ queryKey: ['admin-audio-works'] })
      queryClient.invalidateQueries({ queryKey: ['audio-works-public'] })
    },
    onError: () => {
      showAdminToast('审核操作失败，请重试', 'error')
    },
  })

  const works = data?.works ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / pageSize)) : 1

  const columns: AdminColumn<AdminAudioWork>[] = [
    {
      key: 'work',
      header: '音频作品',
      className: 'min-w-[320px]',
      render: (work) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={work.moderation_status} />
            <span className="text-xs text-slate-400">@{work.author_username || work.author_id}</span>
          </div>
          <div>
            <p className="font-medium text-slate-900">{work.title}</p>
            {work.description ? (
              <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">{work.description}</p>
            ) : null}
            {work.moderation_note ? (
              <p className="mt-2 text-xs text-rose-500">备注：{work.moderation_note}</p>
            ) : null}
          </div>
        </div>
      ),
    },
    {
      key: 'meta',
      header: '元数据',
      className: 'min-w-[180px] text-slate-500',
      render: (work) => (
        <div className="space-y-1 text-sm">
          <p>{work.duration_sec.toFixed(2)} 秒</p>
          <p>{work.like_count} 赞 · {work.comment_count} 评</p>
          <p>{new Date(work.published_at).toLocaleString('zh-CN')}</p>
        </div>
      ),
    },
    {
      key: 'actions',
      header: '处理',
      className: 'w-[250px]',
      render: (work) =>
        tab === 'pending' ? (
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              className="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
              disabled={moderationMutation.isPending}
              onClick={() => moderationMutation.mutate({ id: work.id, status: 'approved' })}
            >
              <CheckCircle2 className="mr-1 h-4 w-4" />
              通过
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="border-rose-200 text-rose-700 hover:bg-rose-50"
              disabled={moderationMutation.isPending}
              onClick={() => moderationMutation.mutate({ id: work.id, status: 'blocked' })}
            >
              <XCircle className="mr-1 h-4 w-4" />
              封禁
            </Button>
          </div>
        ) : (
          <Button asChild size="sm" variant="outline">
            <Link href={`/audio/works/${work.id}`} target="_blank">
              查看详情
            </Link>
          </Button>
        ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Audio Governance"
        title="音频作品治理"
        description="集中处理音频作品的审核状态流转，补齐创作者发布到公开分发之间的治理环节。"
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="当前队列总数"
          value={(data?.total ?? 0).toLocaleString()}
          hint={`当前标签：${tabs.find((item) => item.value === tab)?.label ?? tab}`}
          icon={Disc3}
          tone={tab === 'pending' ? 'warning' : tab === 'blocked' ? 'danger' : 'success'}
        />
        <AdminMetricCard
          label="本页作品"
          value={works.length.toLocaleString()}
          hint="单页最多展示 20 条"
          icon={Disc3}
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
        data={works}
        columns={columns}
        keyExtractor={(work) => work.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="当前筛选下没有音频作品"
            description="可以切换审核状态，或者等待新的作品进入治理队列。"
          />
        }
      />

      <AdminPagination
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />

      {moderationMutation.isPending ? (
        <div className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <div className="flex items-center gap-3 text-sm text-slate-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            正在更新音频作品审核状态...
          </div>
        </div>
      ) : null}
    </div>
  )
}
