'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, Trash2 } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

interface CommentRow {
  id: string
  user_id: string
  author_username?: string
  commentable_type: string
  commentable_id: string
  content: string
  is_edited: boolean
  is_deleted: boolean
  like_count: number
  reply_count: number
  created_at: string
}

interface ListCommentsOutput {
  comments: CommentRow[]
  total: number
  page: number
  size: number
}

export default function AdminCommentsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null)

  const { data, isLoading } = useQuery<ListCommentsOutput>({
    queryKey: ['admin-comments', page],
    queryFn: () => {
      const params = new URLSearchParams({ page: String(page), page_size: '20' })
      return apiClient.get(`/admin/comments?${params.toString()}`)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) => apiClient.delete(`/admin/comments/${commentId}`),
    onSuccess: () => {
      setDeleteConfirmId(null)
      queryClient.invalidateQueries({ queryKey: ['admin-comments'] })
      showAdminToast('评论已删除', 'success')
    },
    onError: () => {
      showAdminToast('删除失败，请重试', 'error')
    },
  })

  const comments = data?.comments ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1

  const columns: AdminColumn<CommentRow>[] = [
    {
      key: 'content',
      header: '评论内容',
      className: 'min-w-[340px]',
      render: (comment) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm font-medium text-slate-900">
              {comment.author_username ? `@${comment.author_username}` : `用户 ${comment.user_id.slice(0, 8)}...`}
            </span>
            {comment.is_edited ? (
              <AdminStatusBadge value="reviewed" label="已编辑" />
            ) : null}
            {comment.is_deleted ? (
              <AdminStatusBadge value="blocked" label="已删除" />
            ) : null}
          </div>
          <p className="text-sm leading-6 text-slate-600">{comment.content}</p>
        </div>
      ),
    },
    {
      key: 'target',
      header: '关联对象',
      className: 'min-w-[220px]',
      render: (comment) => (
        <div className="space-y-2">
          <AdminStatusBadge value={comment.commentable_type} label={comment.commentable_type} />
          {comment.commentable_type === 'post' ? (
            <Button asChild variant="outline" size="sm">
              <Link href={`/posts/${comment.commentable_id}`} target="_blank">
                查看帖子
              </Link>
            </Button>
          ) : (
            <span className="text-sm text-slate-400">无直接跳转</span>
          )}
        </div>
      ),
    },
    {
      key: 'metrics',
      header: '互动',
      className: 'whitespace-nowrap text-slate-500',
      render: (comment) => (
        <div className="space-y-1">
          <p>点赞 {comment.like_count}</p>
          <p>回复 {comment.reply_count}</p>
          <p>{new Date(comment.created_at).toLocaleString('zh-CN')}</p>
        </div>
      ),
    },
    {
      key: 'actions',
      header: '操作',
      className: 'w-[180px]',
      render: (comment) => {
        const confirming = deleteConfirmId === comment.id
        return (
          <Button
            size="sm"
            variant={confirming ? 'default' : 'outline'}
            className={confirming ? 'bg-rose-600 text-white hover:bg-rose-500' : 'border-rose-200 text-rose-700 hover:bg-rose-50'}
            disabled={deleteMutation.isPending}
            onClick={() => {
              if (confirming) {
                deleteMutation.mutate(comment.id)
              } else {
                setDeleteConfirmId(comment.id)
              }
            }}
          >
            <Trash2 className="mr-1 h-4 w-4" />
            {confirming ? '再次确认删除' : '删除评论'}
          </Button>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Comment Patrol"
        title="评论管理"
        description="用于巡检评论区内容质量，快速下线违规评论，并跳回原帖查看上下文。"
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="评论总数"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前分页接口返回的总记录数"
          icon={MessageSquare}
          tone="brand"
        />
        <AdminMetricCard
          label="本页评论"
          value={comments.length.toLocaleString()}
          hint="单页最多展示 20 条"
          icon={MessageSquare}
          tone="default"
        />
        <AdminMetricCard
          label="已删除评论"
          value={comments.filter((item) => item.is_deleted).length.toLocaleString()}
          hint="仅统计当前页"
          icon={Trash2}
          tone="warning"
        />
      </div>

      <AdminDataTable
        data={comments}
        columns={columns}
        keyExtractor={(comment) => comment.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有评论记录"
            description="当前页面没有可管理的评论数据。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
