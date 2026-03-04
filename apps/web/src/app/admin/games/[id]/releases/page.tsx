'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface Release {
  id: number
  version: string
  branch: string
  status: string
  download_url: string
  file_size: number
  changelog: string
  created_at: string
}

export default function AdminGameReleasesPage() {
  const params = useParams()
  const router = useRouter()
  const queryClient = useQueryClient()
  const gameId = params.id as string

  const { data: releases, isLoading } = useQuery<Release[]>({
    queryKey: ['admin-game-releases', gameId],
    queryFn: async () => {
      const response = await apiClient.get(`/admin/games/${gameId}/releases`)
      return response
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (releaseId: number) => {
      await apiClient.delete(`/admin/games/${gameId}/releases/${releaseId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-game-releases', gameId] })
      alert('删除成功！')
    },
    onError: () => {
      alert('删除失败，请重试')
    },
  })

  const handleDelete = (releaseId: number, version: string) => {
    if (confirm(`确定要删除版本 ${version} 吗？`)) {
      deleteMutation.mutate(releaseId)
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
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
            ← 返回游戏列表
          </Button>
          <h1 className="text-3xl font-bold">版本管理</h1>
        </div>
        <Button onClick={() => router.push(`/admin/games/${gameId}/releases/new`)}>
          新增版本
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>版本列表</CardTitle>
        </CardHeader>
        <CardContent>
          {!releases || releases.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              暂无版本，点击右上角新增版本
            </div>
          ) : (
            <div className="space-y-4">
              {releases.map((release) => (
                <div key={release.id} className="p-4 border rounded-lg">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="font-bold text-lg">v{release.version}</h3>
                        <Badge className={
                          release.status === 'stable' ? 'bg-green-500' :
                          release.status === 'beta' ? 'bg-yellow-500' : 'bg-blue-500'
                        }>
                          {release.status}
                        </Badge>
                        <Badge variant="outline">{release.branch}</Badge>
                      </div>
                      <div className="flex gap-4 text-sm text-gray-600">
                        <span>大小: {formatFileSize(release.file_size)}</span>
                        <span>发布: {new Date(release.created_at).toLocaleString('zh-CN')}</span>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => window.open(release.download_url, '_blank')}
                      >
                        下载
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-red-600 hover:text-red-700"
                        onClick={() => handleDelete(release.id, release.version)}
                      >
                        删除
                      </Button>
                    </div>
                  </div>
                  {release.changelog && (
                    <div className="mt-3 p-3 bg-gray-50 rounded">
                      <p className="text-sm font-medium mb-1">更新日志：</p>
                      <p className="text-sm text-gray-700 whitespace-pre-wrap">
                        {release.changelog}
                      </p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
