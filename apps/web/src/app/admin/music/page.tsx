'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { useState } from 'react'

interface Album {
  id: number
  slug: string
  title: string
  artist: string
  status: string
  price: number
  track_count: number
  created_at: string
}

export default function AdminMusicPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')

  const { data: albums, isLoading } = useQuery<Album[]>({
    queryKey: ['admin-albums', search],
    queryFn: async () => {
      const response = await apiClient.get('/admin/albums', {
        params: { search },
      })
      return response.data.data
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/admin/albums/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-albums'] })
      alert('删除成功！')
    },
    onError: () => {
      alert('删除失败，请重试')
    },
  })

  const handleDelete = (id: number, title: string) => {
    if (confirm(`确定要删除专辑"${title}"吗？`)) {
      deleteMutation.mutate(id)
    }
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <Link href="/admin">
          <Button variant="outline" className="mb-4">← 返回后台首页</Button>
        </Link>
        <div className="flex justify-between items-center mb-4">
          <h1 className="text-3xl font-bold">音乐管理</h1>
          <Link href="/admin/music/new">
            <Button>新增专辑</Button>
          </Link>
        </div>

        <div className="max-w-md">
          <Input
            placeholder="搜索专辑或艺术家..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>专辑列表</CardTitle>
        </CardHeader>
        <CardContent>
          {!albums || albums.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {search ? '未找到匹配的专辑' : '暂无专辑，点击右上角新增专辑'}
            </div>
          ) : (
            <div className="space-y-4">
              {albums.map((album) => (
                <div key={album.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="font-bold text-lg">{album.title}</h3>
                      <Badge className={
                        album.status === 'published' ? 'bg-green-500' :
                        album.status === 'draft' ? 'bg-gray-500' : 'bg-yellow-500'
                      }>
                        {album.status}
                      </Badge>
                    </div>
                    <div className="flex gap-6 text-sm text-gray-600">
                      <span>艺术家: {album.artist}</span>
                      <span>曲目: {album.track_count}</span>
                      <span>价格: ¥{album.price}</span>
                      <span>创建: {new Date(album.created_at).toLocaleDateString('zh-CN')}</span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link href={`/admin/music/${album.id}/edit`}>
                      <Button variant="outline" size="sm">编辑</Button>
                    </Link>
                    <Link href={`/admin/music/${album.id}/tracks`}>
                      <Button variant="outline" size="sm">曲目管理</Button>
                    </Link>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleDelete(album.id, album.title)}
                    >
                      删除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
