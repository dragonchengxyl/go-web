'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface GameForm {
  title: string
  slug: string
  description: string
  short_description: string
  cover_image: string
  price: number
  discount_price: number
  status: string
  tags: string
}

export default function AdminGameEditPage() {
  const params = useParams()
  const router = useRouter()
  const queryClient = useQueryClient()
  const gameId = params.id as string
  const isNew = gameId === 'new'

  const [formData, setFormData] = useState<GameForm>({
    title: '',
    slug: '',
    description: '',
    short_description: '',
    cover_image: '',
    price: 0,
    discount_price: 0,
    status: 'draft',
    tags: '',
  })

  const { data: game, isLoading } = useQuery({
    queryKey: ['admin-game', gameId],
    queryFn: async () => {
      const response = await apiClient.get(`/admin/games/${gameId}`)
      return response.data.data
    },
    enabled: !isNew,
  })

  useEffect(() => {
    if (game) {
      setFormData({
        title: game.title || '',
        slug: game.slug || '',
        description: game.description || '',
        short_description: game.short_description || '',
        cover_image: game.cover_image || '',
        price: game.price || 0,
        discount_price: game.discount_price || 0,
        status: game.status || 'draft',
        tags: game.tags?.join(', ') || '',
      })
    }
  }, [game])

  const saveMutation = useMutation({
    mutationFn: async (data: GameForm) => {
      const payload = {
        ...data,
        tags: data.tags.split(',').map(t => t.trim()).filter(Boolean),
      }

      if (isNew) {
        return await apiClient.post('/admin/games', payload)
      } else {
        return await apiClient.put(`/admin/games/${gameId}`, payload)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-games'] })
      alert(isNew ? '创建成功！' : '更新成功！')
      router.push('/admin/games')
    },
    onError: () => {
      alert('保存失败，请重试')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    saveMutation.mutate(formData)
  }

  if (!isNew && isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      <Button variant="outline" onClick={() => router.back()} className="mb-4">
        ← 返回
      </Button>

      <h1 className="text-3xl font-bold mb-8">
        {isNew ? '新增游戏' : '编辑游戏'}
      </h1>

      <form onSubmit={handleSubmit}>
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">游戏标题 *</label>
              <Input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="输入游戏标题"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Slug (URL标识) *</label>
              <Input
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                placeholder="game-slug"
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                用于URL，只能包含小写字母、数字和连字符
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">简短描述</label>
              <Input
                value={formData.short_description}
                onChange={(e) => setFormData({ ...formData, short_description: e.target.value })}
                placeholder="一句话介绍游戏"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">详细描述</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="详细介绍游戏内容、玩法、特色等"
                className="w-full px-3 py-2 border rounded-md min-h-[150px]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">封面图片 URL</label>
              <Input
                value={formData.cover_image}
                onChange={(e) => setFormData({ ...formData, cover_image: e.target.value })}
                placeholder="https://example.com/cover.jpg"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">标签（逗号分隔）</label>
              <Input
                value={formData.tags}
                onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
                placeholder="动作, 冒险, 独立"
              />
            </div>
          </CardContent>
        </Card>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle>定价信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">原价 (¥)</label>
                <Input
                  type="number"
                  step="0.01"
                  value={formData.price}
                  onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) })}
                  placeholder="0.00"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">折扣价 (¥)</label>
                <Input
                  type="number"
                  step="0.01"
                  value={formData.discount_price}
                  onChange={(e) => setFormData({ ...formData, discount_price: parseFloat(e.target.value) })}
                  placeholder="0.00"
                />
                <p className="text-xs text-gray-500 mt-1">
                  留空表示无折扣
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle>发布状态</CardTitle>
          </CardHeader>
          <CardContent>
            <select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value })}
              className="w-full px-3 py-2 border rounded-md"
            >
              <option value="draft">草稿</option>
              <option value="published">已发布</option>
              <option value="archived">已归档</option>
            </select>
          </CardContent>
        </Card>

        <div className="flex gap-4">
          <Button type="submit" disabled={saveMutation.isPending}>
            {saveMutation.isPending ? '保存中...' : '保存'}
          </Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>
            取消
          </Button>
        </div>
      </form>
    </div>
  )
}
