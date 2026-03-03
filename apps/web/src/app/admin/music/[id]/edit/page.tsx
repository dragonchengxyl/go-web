'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface AlbumForm {
  title: string
  slug: string
  artist: string
  description: string
  cover_image: string
  release_date: string
  price: number
  status: string
}

export default function AdminAlbumEditPage() {
  const params = useParams()
  const router = useRouter()
  const queryClient = useQueryClient()
  const albumId = params.id as string
  const isNew = albumId === 'new'

  const [formData, setFormData] = useState<AlbumForm>({
    title: '',
    slug: '',
    artist: '',
    description: '',
    cover_image: '',
    release_date: new Date().toISOString().split('T')[0],
    price: 0,
    status: 'draft',
  })

  const { data: album, isLoading } = useQuery({
    queryKey: ['admin-album', albumId],
    queryFn: async () => {
      const response = await apiClient.get(`/admin/albums/${albumId}`)
      return response.data.data
    },
    enabled: !isNew,
  })

  useEffect(() => {
    if (album) {
      setFormData({
        title: album.title || '',
        slug: album.slug || '',
        artist: album.artist || '',
        description: album.description || '',
        cover_image: album.cover_image || '',
        release_date: album.release_date?.split('T')[0] || '',
        price: album.price || 0,
        status: album.status || 'draft',
      })
    }
  }, [album])

  const saveMutation = useMutation({
    mutationFn: async (data: AlbumForm) => {
      if (isNew) {
        return await apiClient.post('/admin/albums', data)
      } else {
        return await apiClient.put(`/admin/albums/${albumId}`, data)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-albums'] })
      alert(isNew ? '创建成功！' : '更新成功！')
      router.push('/admin/music')
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
        {isNew ? '新增专辑' : '编辑专辑'}
      </h1>

      <form onSubmit={handleSubmit}>
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">专辑标题 *</label>
              <Input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="输入专辑标题"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Slug (URL标识) *</label>
              <Input
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                placeholder="album-slug"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">艺术家 *</label>
              <Input
                value={formData.artist}
                onChange={(e) => setFormData({ ...formData, artist: e.target.value })}
                placeholder="艺术家名称"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">专辑描述</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="介绍专辑内容、风格等"
                className="w-full px-3 py-2 border rounded-md min-h-[120px]"
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
              <label className="block text-sm font-medium mb-2">发行日期</label>
              <Input
                type="date"
                value={formData.release_date}
                onChange={(e) => setFormData({ ...formData, release_date: e.target.value })}
              />
            </div>
          </CardContent>
        </Card>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle>定价信息</CardTitle>
          </CardHeader>
          <CardContent>
            <div>
              <label className="block text-sm font-medium mb-2">价格 (¥)</label>
              <Input
                type="number"
                step="0.01"
                value={formData.price}
                onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) })}
                placeholder="0.00"
              />
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
