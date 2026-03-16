'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircle2, ShieldCheck, Tag, XCircle } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

interface PostRow {
  id: string
  title: string
  content: string
  author_username: string
  tags: string[]
  moderation_status: string
  created_at: string
}

interface ListPostsOutput {
  posts: PostRow[]
  total: number
  page: number
}

const tabs = [
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已封禁', value: 'blocked' },
]

export default function AdminModerationPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState('pending')
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading } = useQuery<ListPostsOutput>({
    queryKey: ['admin-posts', tab, page],
    queryFn: () =>
      apiClient.get(`/admin/posts?status=${tab}&page=${page}&page_size=${pageSize}`),
  })

  const moderateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiClient.put(`/admin/posts/${id}/moderation`, { status }),
    onSuccess: (_, { status }) => {
      showAdminToast(status === 'approved' ? '帖子已通过审核' : '帖子已封禁', 'success')
      queryClient.invalidateQueries({ queryKey: ['admin-posts'] })
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] })
    },
    onError: () => {
      showAdminToast('审核操作失败，请重试', 'error')
    },
  })

  const posts = data?.posts ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / pageSize)) : 1

  const columns: AdminColumn<PostRow>[] = [
    {
      key: 'content',
      header: '帖子内容',
      className: 'min-w-[320px]',
      render: (post) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={post.moderation_status} />
            <span className="text-xs text-slate-400">@{post.author_username || '未知作者'}</span>
          </div>
          <div>
            <p className="font-medium text-slate-900">
              {post.title || '未命名帖子'}
            </p>
            <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">
              {post.content}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'tags',
      header: '标签',
      className: 'min-w-[180px]',
      render: (post) =>
        post.tags?.length ? (
          <div className="flex flex-wrap gap-2">
            {post.tags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600"
              >
                <Tag className="mr-1 h-3 w-3" />
                {tag}
              </span>
            ))}
          </div>
        ) : (
          <span className="text-sm text-slate-400">无标签</span>
        ),
    },
    {
      key: 'created_at',
      header: '提交时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (post) => new Date(post.created_at).toLocaleString('zh-CN'),
    },
    {
      key: 'actions',
      header: '处理',
      className: 'w-[220px]',
      render: (post) =>
        tab === 'pending' ? (
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              className="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
              disabled={moderateMutation.isPending}
              onClick={() => moderateMutation.mutate({ id: post.id, status: 'approved' })}
            >
              <CheckCircle2 className="mr-1 h-4 w-4" />
              通过
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="border-rose-200 text-rose-700 hover:bg-rose-50"
              disabled={moderateMutation.isPending}
              onClick={() => moderateMutation.mutate({ id: post.id, status: 'blocked' })}
            >
              <XCircle className="mr-1 h-4 w-4" />
              封禁
            </Button>
          </div>
        ) : (
          <Button asChild variant="outline" size="sm">
            <Link href={`/posts/${post.id}`} target="_blank">
              查看原帖
            </Link>
          </Button>
        ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Content Governance"
        title="内容审核"
        description="围绕帖子内容做审核和状态流转，适合运营和版主值班使用。"
        actions={
          <Button asChild variant="outline">
            <Link href="/admin/reports">联动举报处理</Link>
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="当前队列总数"
          value={(data?.total ?? 0).toLocaleString()}
          hint={`当前标签：${tabs.find((item) => item.value === tab)?.label ?? tab}`}
          icon={ShieldCheck}
          tone={tab === 'pending' ? 'warning' : tab === 'blocked' ? 'danger' : 'success'}
        />
        <AdminMetricCard
          label="本页帖子"
          value={posts.length.toLocaleString()}
          hint="单页最多展示 20 条"
          icon={Tag}
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
        data={posts}
        columns={columns}
        keyExtractor={(post) => post.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="当前筛选下没有帖子"
            description="可以切换审核状态，或者等待新的帖子进入队列。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
