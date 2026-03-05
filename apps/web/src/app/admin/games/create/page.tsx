'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FileUpload } from '@/components/admin/file-upload'
import { apiClient } from '@/lib/api-client'

export default function CreateGamePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    title: '',
    slug: '',
    description: '',
    short_description: '',
    cover_image: '',
    price: 0,
    tags: [] as string[],
  })

  const handleCoverUpload = async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/upload/image`, {
        method: 'POST',
        body: formData,
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      })
      const data = await response.json()
      if (data.code === 0) {
        return data.data.url
      }
      throw new Error(data.message || '上传失败')
    } catch (error) {
      throw new Error('上传失败')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      await apiClient.post('/admin/games', formData)
      alert('游戏创建成功！')
      router.push('/admin/games')
    } catch (error) {
      alert('创建失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container mx-auto py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">创建新游戏</h1>
        <p className="text-gray-600 mt-2">填写游戏信息并上传封面图片</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
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
              <label className="block text-sm font-medium mb-2">URL Slug *</label>
              <Input
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                placeholder="game-slug"
                required
              />
              <p className="text-xs text-gray-500 mt-1">用于 URL，只能包含小写字母、数字和连字符</p>
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
                placeholder="详细介绍游戏内容、玩法等"
                className="w-full px-3 py-2 border rounded-md min-h-[150px]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">价格（分）</label>
              <Input
                type="number"
                value={formData.price}
                onChange={(e) => setFormData({ ...formData, price: parseInt(e.target.value) })}
                placeholder="0"
                min="0"
              />
              <p className="text-xs text-gray-500 mt-1">以分为单位，例如 9900 = 99.00 元</p>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">标签</label>
              <Input
                value={formData.tags.join(', ')}
                onChange={(e) => setFormData({ ...formData, tags: e.target.value.split(',').map(t => t.trim()) })}
                placeholder="动作, 冒险, 独立"
              />
              <p className="text-xs text-gray-500 mt-1">用逗号分隔多个标签</p>
            </div>
          </CardContent>
        </Card>

        <FileUpload
          label="游戏封面"
          accept="image/*"
          maxSize={5}
          onUpload={handleCoverUpload}
          currentUrl={formData.cover_image}
        />

        <div className="flex gap-4">
          <Button type="submit" disabled={loading}>
            {loading ? '创建中...' : '创建游戏'}
          </Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>
            取消
          </Button>
        </div>
      </form>
    </div>
  )
}
