'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Loader2, ExternalLink } from 'lucide-react'

interface Comment {
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
  comments: Comment[]
  total: number
  page: number
  size: number
}

function toast(msg: string) {
  const el = document.createElement('div')
  el.className = 'fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm z-50'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

export default function AdminCommentsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

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
      queryClient.invalidateQueries({ queryKey: ['admin-comments'] })
      setDeleteConfirm(null)
      toast('评论已删除')
    },
    onError: () => toast('删除失败，请重试'),
  })

  const comments = data?.comments ?? []
  const totalPages = data ? Math.ceil(data.total / 20) : 1

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">评论管理</h1>
        {data && <span className="text-sm text-gray-400">共 {data.total} 条评论</span>}
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">评论列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="animate-spin text-gray-400" />
            </div>
          ) : comments.length === 0 ? (
            <div className="text-center py-10 text-gray-400">暂无评论</div>
          ) : (
            <div className="divide-y">
              {comments.map((comment) => (
                <div key={comment.id} className="py-4">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1.5 flex-wrap text-xs text-gray-400">
                        <span className="font-medium text-gray-700 dark:text-gray-300">
                          {comment.author_username
                            ? `@${comment.author_username}`
                            : `用户 ${comment.user_id.slice(0, 8)}…`}
                        </span>
                        <span>·</span>
                        <span>{comment.commentable_type}</span>
                        {comment.commentable_type === 'post' && (
                          <a
                            href={`/posts/${comment.commentable_id}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-500 hover:underline flex items-center gap-0.5"
                          >
                            查看帖子 <ExternalLink size={10} />
                          </a>
                        )}
                        <span>·</span>
                        <span>{new Date(comment.created_at).toLocaleString('zh-CN')}</span>
                        {comment.is_edited && <span className="text-yellow-500">(已编辑)</span>}
                        {comment.is_deleted && <span className="text-red-500">(已删除)</span>}
                      </div>
                      <p className="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap">
                        {comment.content}
                      </p>
                      <div className="flex gap-4 mt-1.5 text-xs text-gray-400">
                        <span>点赞 {comment.like_count}</span>
                        <span>回复 {comment.reply_count}</span>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      {deleteConfirm === comment.id ? (
                        <>
                          <Button
                            size="sm"
                            className="bg-red-600 hover:bg-red-700 text-white"
                            onClick={() => deleteMutation.mutate(comment.id)}
                            disabled={deleteMutation.isPending}
                          >
                            {deleteMutation.isPending
                              ? <Loader2 size={12} className="animate-spin" />
                              : '确认删除'}
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-gray-400"
                            onClick={() => setDeleteConfirm(null)}
                          >
                            取消
                          </Button>
                        </>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          onClick={() => setDeleteConfirm(comment.id)}
                        >
                          删除
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-6">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                上一页
              </Button>
              <span className="flex items-center px-3 text-sm text-gray-500">
                {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
              >
                下一页
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
