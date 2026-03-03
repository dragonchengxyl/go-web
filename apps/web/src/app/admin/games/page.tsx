'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface Game {
  id: number
  slug: string
  title: string
  status: string
  price: number
  discount_price: number
  downloads: number
  created_at: string
}

export default function AdminGamesPage() {
  const queryClient = useQueryClient()

  const { data: games, isLoading } = useQuery<Game[]>({
    queryKey: ['admin-games'],
    queryFn: async () => {
      const response = await apiClient.get('/admin/games')
      return response.data.data
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/admin/games/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-games'] })
      alert('删除成功！')
    },
    onError: () => {
      alert('删除失败，请重试')
    },
  })

  const handleDelete = (id: number, title: string) => {
    if (confirm(`确定要删除游戏"${title}"吗？`)) {
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
      <div className="flex justify-between items-center mb-8">
        <div>
          <Link href="/admin">
            <Button variant="outline" className="mb-4">← 返回后台首页</Button>
          </Link>
          <h1 className="text-3xl font-bold">游戏管理</h1>
        </div>
        <Link href="/admin/games/new">
          <Button>新增游戏</Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>游戏列表</CardTitle>
        </CardHeader>
        <CardContent>
          {!games || games.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              暂无游戏，点击右上角新增游戏
            </div>
          ) : (
            <div className="space-y-4">
              {games.map((game) => (
                <div key={game.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="font-bold text-lg">{game.title}</h3>
                      <Badge className={
                        game.status === 'published' ? 'bg-green-500' :
                        game.status === 'draft' ? 'bg-gray-500' : 'bg-yellow-500'
                      }>
                        {game.status}
                      </Badge>
                    </div>
                    <div className="flex gap-6 text-sm text-gray-600">
                      <span>Slug: {game.slug}</span>
                      <span>下载: {game.downloads}</span>
                      <span>价格: ¥{game.discount_price || game.price}</span>
                      <span>创建: {new Date(game.created_at).toLocaleDateString('zh-CN')}</span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link href={`/admin/games/${game.id}/edit`}>
                      <Button variant="outline" size="sm">编辑</Button>
                    </Link>
                    <Link href={`/admin/games/${game.id}/releases`}>
                      <Button variant="outline" size="sm">版本管理</Button>
                    </Link>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleDelete(game.id, game.title)}
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
