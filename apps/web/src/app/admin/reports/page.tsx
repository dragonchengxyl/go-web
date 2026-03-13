'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2, ExternalLink } from 'lucide-react'

interface Report {
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
  reports: Report[]
  total: number
  page: number
}

const TABS = [
  { label: '待处理', value: 'pending' },
  { label: '已处理', value: 'reviewed' },
  { label: '已忽略', value: 'dismissed' },
]

const TARGET_LABEL: Record<string, string> = {
  post: '帖子',
  comment: '评论',
  user: '用户',
}

const ACTION_LABEL: Record<string, string> = {
  block_post: '已封禁帖子',
  delete_comment: '已删除评论',
  ban_user: '已封禁用户',
}

function toast(msg: string) {
  const el = document.createElement('div')
  el.className = 'fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm z-50'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
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
        toast('举报已忽略')
      } else if (action && ACTION_LABEL[action]) {
        toast(`举报已处理，${ACTION_LABEL[action]}`)
      } else {
        toast('举报已处理')
      }
      queryClient.invalidateQueries({ queryKey: ['admin-reports'] })
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] })
    },
    onError: () => {
      setConfirming(null)
      toast('操作失败，请重试')
    },
  })

  const reports = data?.reports ?? []
  const totalPages = data ? Math.ceil(data.total / pageSize) : 1

  const handleAction = (id: string, status: 'reviewed' | 'dismissed', action?: string) => {
    if (confirming?.id === id && confirming.status === status && confirming.action === action) {
      updateMutation.mutate({ id, status, action })
    } else {
      setConfirming({ id, status, action })
    }
  }

  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white">举报处理</h1>

      <div className="flex gap-1 border-b border-gray-200 dark:border-gray-700">
        {TABS.map(({ label, value }) => (
          <button
            key={value}
            onClick={() => { setTab(value); setPage(1); setConfirming(null) }}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === value
                ? 'border-purple-600 text-purple-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="animate-spin text-gray-400" />
        </div>
      ) : reports.length === 0 ? (
        <div className="text-center py-12 text-gray-400">暂无举报</div>
      ) : (
        <div className="space-y-3">
          {reports.map((r) => {
            const isActing = updateMutation.isPending && confirming?.id === r.id
            const isConfirmingThis = confirming?.id === r.id
            const actionButton = r.target_type === 'post'
              ? { label: '处理并封禁帖子', action: 'block_post' }
              : r.target_type === 'comment'
                ? { label: '处理并删除评论', action: 'delete_comment' }
                : r.target_type === 'user'
                  ? { label: '处理并封禁用户', action: 'ban_user' }
                  : null
            return (
              <Card key={r.id}>
                <CardContent className="pt-4 pb-3">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1.5 flex-wrap">
                        <Badge variant="outline" className="text-xs">
                          {TARGET_LABEL[r.target_type] ?? r.target_type}
                        </Badge>
                        <span className="font-medium text-sm text-gray-900 dark:text-white">{r.reason}</span>
                      </div>
                      {r.description && (
                        <p className="text-sm text-gray-500 mb-1.5">{r.description}</p>
                      )}
                      <div className="flex items-center gap-3">
                        <span className="text-xs text-gray-400">举报人：{r.reporter_username || '未知'}</span>
                        <span className="text-xs text-gray-400">
                          {new Date(r.created_at).toLocaleString('zh-CN')}
                        </span>
                        {r.action_taken && (
                          <Badge variant="outline" className="text-xs">
                            {ACTION_LABEL[r.action_taken] ?? r.action_taken}
                          </Badge>
                        )}
                        {r.target_type === 'post' && (
                          <a
                            href={`/posts/${r.target_id}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-blue-500 hover:underline flex items-center gap-1"
                          >
                            查看原文 <ExternalLink size={10} />
                          </a>
                        )}
                      </div>
                    </div>
                    {tab === 'pending' && (
                      <div className="flex gap-2 shrink-0 items-center">
                        {isConfirmingThis && (
                          <span className="text-xs text-gray-400">再次点击确认</span>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          className={isConfirmingThis && confirming.status === 'reviewed' && !confirming.action
                            ? 'bg-blue-600 text-white border-blue-600 hover:bg-blue-700'
                            : 'text-blue-600 hover:text-blue-700'}
                          disabled={isActing}
                          onClick={() => handleAction(r.id, 'reviewed')}
                        >
                          {isActing && confirming?.status === 'reviewed' && !confirming?.action
                            ? <Loader2 size={12} className="animate-spin" />
                            : '处理'}
                        </Button>
                        {actionButton && (
                          <Button
                            size="sm"
                            variant="outline"
                            className={isConfirmingThis && confirming.status === 'reviewed' && confirming.action === actionButton.action
                              ? 'bg-red-600 text-white border-red-600 hover:bg-red-700'
                              : 'text-red-600 hover:text-red-700'}
                            disabled={isActing}
                            onClick={() => handleAction(r.id, 'reviewed', actionButton.action)}
                          >
                            {isActing && confirming?.status === 'reviewed' && confirming?.action === actionButton.action
                              ? <Loader2 size={12} className="animate-spin" />
                              : actionButton.label}
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          className={isConfirmingThis && confirming.status === 'dismissed'
                            ? 'bg-gray-600 text-white border-gray-600 hover:bg-gray-700'
                            : 'text-gray-500 hover:text-gray-700'}
                          disabled={isActing}
                          onClick={() => handleAction(r.id, 'dismissed')}
                        >
                          {isActing && confirming?.status === 'dismissed'
                            ? <Loader2 size={12} className="animate-spin" />
                            : '忽略'}
                        </Button>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex justify-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}>
            上一页
          </Button>
          <span className="flex items-center px-3 text-sm text-gray-500">{page} / {totalPages}</span>
          <Button variant="outline" size="sm" onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page === totalPages}>
            下一页
          </Button>
        </div>
      )}
    </div>
  )
}
