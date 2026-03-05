'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

interface Comment {
  id: string
  user_id: string
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

export default function AdminCommentsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery<ListCommentsOutput>({
    queryKey: ['admin-comments', page],
    queryFn: () => {
      const params = new URLSearchParams({ page: String(page), page_size: '20' })
      return apiClient.get(`/admin/comments?${params.toString()}`)
    },
  })

  const comments = data?.comments ?? []

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) =>
      apiClient.delete(`/admin/comments/${commentId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-comments'] })
      alert('评论已删除')
    },
    onError: () => alert('删除失败，请重试'),
  })

  const handleDelete = (commentId: string) => {
    if (confirm('确定要删除这条评论吗？')) {
      deleteMutation.mutate(commentId)
    }
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  const totalPages = data ? Math.ceil(data.total / 20) : 1

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <Link href="/admin">
          <Button variant="outline" className="mb-4">← 返回后台首页</Button>
        </Link>
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">评论审核</h1>
          {data && <span className="text-sm text-gray-500">共 {data.total} 条评论</span>}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>评论列表</CardTitle>
        </CardHeader>
        <CardContent>
          {comments.length === 0 ? (
            <div className="text-center py-12 text-gray-500">暂无评论</div>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <div key={comment.id} className="p-4 border rounded-lg">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2 text-sm text-gray-500">
                        <span className="font-medium text-gray-700">用户 {comment.user_id.slice(0, 8)}...</span>
                        <span>·</span>
                        <span>{comment.commentable_type}</span>
                        <span>·</span>
                        <span>{new Date(comment.created_at).toLocaleString('zh-CN')}</span>
                        {comment.is_edited && <span className="text-yellow-600">(已编辑)</span>}
                        {comment.is_deleted && <span className="text-red-600">(已删除)</span>}
                      </div>
                      <p className="text-gray-900 whitespace-pre-wrap">{comment.content}</p>
                      <div className="flex gap-4 mt-2 text-sm text-gray-500">
                        <span>👍 {comment.like_count}</span>
                        <span>💬 {comment.reply_count} 回复</span>
                      </div>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700 shrink-0"
                      onClick={() => handleDelete(comment.id)}
                      disabled={deleteMutation.isPending}
                    >
                      删除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
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
              <span className="flex items-center px-3 text-sm text-gray-600">
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
