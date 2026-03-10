'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'

interface Post {
  id: string
  title: string
  content: string
  author_username: string
  tags: string[]
  moderation_status: string
  created_at: string
}

interface ListPostsOutput {
  posts: Post[]
  total: number
  page: number
}

const TABS = [
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已封禁', value: 'blocked' },
]

function toast(msg: string) {
  // Simple DOM toast
  const el = document.createElement('div')
  el.className = 'fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm z-50'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

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

  const [pendingAction, setPendingAction] = useState<{ id: string; action: string } | null>(null)

  const moderateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiClient.put(`/admin/posts/${id}/moderation`, { status }),
    onSuccess: (_, { status }) => {
      setPendingAction(null)
      toast(status === 'approved' ? '已通过审核' : '已封禁帖子')
      queryClient.invalidateQueries({ queryKey: ['admin-posts', tab, page] })
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] })
    },
    onError: () => {
      setPendingAction(null)
      toast('操作失败，请重试')
    },
  })

  const totalPages = data ? Math.ceil(data.total / pageSize) : 1
  const posts = data?.posts ?? []

  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white">内容审核</h1>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 dark:border-gray-700">
        {TABS.map(({ label, value }) => (
          <button
            key={value}
            onClick={() => { setTab(value); setPage(1) }}
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
      ) : posts.length === 0 ? (
        <div className="text-center py-12 text-gray-400">暂无帖子</div>
      ) : (
        <div className="space-y-3">
          {posts.map((p) => {
            const isActing = pendingAction?.id === p.id && moderateMutation.isPending
            return (
              <Card key={p.id}>
                <CardContent className="pt-4 pb-3">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      {p.title && (
                        <p className="font-semibold text-gray-900 dark:text-white mb-1">{p.title}</p>
                      )}
                      <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
                        {p.content.slice(0, 120)}{p.content.length > 120 ? '…' : ''}
                      </p>
                      <div className="flex items-center gap-3 mt-2 flex-wrap">
                        <span className="text-xs text-gray-400">@{p.author_username}</span>
                        <span className="text-xs text-gray-400">
                          {new Date(p.created_at).toLocaleString('zh-CN')}
                        </span>
                        {(p.tags ?? []).map((t) => (
                          <Badge key={t} variant="outline" className="text-xs h-5">{t}</Badge>
                        ))}
                      </div>
                    </div>
                    {tab === 'pending' && (
                      <div className="flex gap-2 shrink-0">
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-green-600 hover:text-green-700 hover:bg-green-50"
                          disabled={isActing}
                          onClick={() => {
                            setPendingAction({ id: p.id, action: 'approved' })
                            moderateMutation.mutate({ id: p.id, status: 'approved' })
                          }}
                        >
                          {isActing && pendingAction?.action === 'approved'
                            ? <Loader2 size={14} className="animate-spin" />
                            : <CheckCircle size={14} />}
                          <span className="ml-1">通过</span>
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          disabled={isActing}
                          onClick={() => {
                            setPendingAction({ id: p.id, action: 'blocked' })
                            moderateMutation.mutate({ id: p.id, status: 'blocked' })
                          }}
                        >
                          {isActing && pendingAction?.action === 'blocked'
                            ? <Loader2 size={14} className="animate-spin" />
                            : <XCircle size={14} />}
                          <span className="ml-1">封禁</span>
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
