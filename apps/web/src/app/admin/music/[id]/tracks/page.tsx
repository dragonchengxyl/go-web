'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Track {
  id: number
  title: string
  track_number: number
  duration: number
  audio_url: string
}

export default function AdminTracksPage() {
  const params = useParams()
  const router = useRouter()
  const queryClient = useQueryClient()
  const albumId = params.id as string

  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingTrack, setEditingTrack] = useState<Track | null>(null)
  const [formData, setFormData] = useState({
    title: '',
    track_number: 1,
    duration: 0,
    audio_url: '',
  })

  const { data: tracks, isLoading } = useQuery<Track[]>({
    queryKey: ['admin-tracks', albumId],
    queryFn: async () => {
      const response = await apiClient.get(`/admin/albums/${albumId}/tracks`)
      return response
    },
  })

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (editingTrack) {
        return await apiClient.put(`/admin/albums/${albumId}/tracks/${editingTrack.id}`, data)
      } else {
        return await apiClient.post(`/admin/albums/${albumId}/tracks`, data)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-tracks', albumId] })
      setIsDialogOpen(false)
      setEditingTrack(null)
      resetForm()
      alert(editingTrack ? '更新成功！' : '添加成功！')
    },
    onError: () => {
      alert('操作失败，请重试')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (trackId: number) => {
      await apiClient.delete(`/admin/albums/${albumId}/tracks/${trackId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-tracks', albumId] })
      alert('删除成功！')
    },
    onError: () => {
      alert('删除失败，请重试')
    },
  })

  const resetForm = () => {
    setFormData({
      title: '',
      track_number: (tracks?.length || 0) + 1,
      duration: 0,
      audio_url: '',
    })
  }

  const handleAdd = () => {
    setEditingTrack(null)
    resetForm()
    setIsDialogOpen(true)
  }

  const handleEdit = (track: Track) => {
    setEditingTrack(track)
    setFormData({
      title: track.title,
      track_number: track.track_number,
      duration: track.duration,
      audio_url: track.audio_url,
    })
    setIsDialogOpen(true)
  }

  const handleDelete = (track: Track) => {
    if (confirm(`确定要删除曲目"${track.title}"吗？`)) {
      deleteMutation.mutate(track.id)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    saveMutation.mutate(formData)
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
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
          <Button variant="outline" onClick={() => router.back()} className="mb-4">
            ← 返回专辑列表
          </Button>
          <h1 className="text-3xl font-bold">曲目管理</h1>
        </div>
        <Button onClick={handleAdd}>添加曲目</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>曲目列表</CardTitle>
        </CardHeader>
        <CardContent>
          {!tracks || tracks.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              暂无曲目，点击右上角添加曲目
            </div>
          ) : (
            <div className="space-y-2">
              {tracks.map((track) => (
                <div key={track.id} className="flex items-center justify-between p-3 border rounded-lg">
                  <div className="flex items-center gap-4">
                    <span className="text-gray-500 w-8 text-center font-medium">
                      {track.track_number}
                    </span>
                    <div>
                      <p className="font-medium">{track.title}</p>
                      <p className="text-sm text-gray-500">{formatDuration(track.duration)}</p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={() => handleEdit(track)}>
                      编辑
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleDelete(track)}
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

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingTrack ? '编辑曲目' : '添加曲目'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">曲目标题 *</label>
              <Input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="输入曲目标题"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">曲目序号 *</label>
              <Input
                type="number"
                value={formData.track_number}
                onChange={(e) => setFormData({ ...formData, track_number: parseInt(e.target.value) })}
                min="1"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">时长（秒）*</label>
              <Input
                type="number"
                value={formData.duration}
                onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) })}
                min="0"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">音频文件 URL *</label>
              <Input
                value={formData.audio_url}
                onChange={(e) => setFormData({ ...formData, audio_url: e.target.value })}
                placeholder="https://example.com/track.mp3"
                required
              />
            </div>

            <div className="flex gap-2 justify-end">
              <Button type="button" variant="outline" onClick={() => setIsDialogOpen(false)}>
                取消
              </Button>
              <Button type="submit" disabled={saveMutation.isPending}>
                {saveMutation.isPending ? '保存中...' : '保存'}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
