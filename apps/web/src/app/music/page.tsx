'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useState } from 'react'

interface Album {
  id: number
  slug: string
  title: string
  artist: string
  cover_image: string
  release_date: string
  price: number
  description: string
  track_count: number
}

export default function MusicPage() {
  const [search, setSearch] = useState('')

  const { data: albums, isLoading } = useQuery<Album[]>({
    queryKey: ['albums', search],
    queryFn: async () => {
      const response = await apiClient.get('/music/albums', {
        params: { search },
      })
      return response.data.data
    },
  })

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-4">游戏原声音乐</h1>
        <p className="text-gray-600 mb-6">探索我们工作室游戏的精彩原声音乐</p>

        <div className="max-w-md">
          <Input
            placeholder="搜索专辑或艺术家..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12">加载中...</div>
      ) : !albums || albums.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500">暂无音乐专辑</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {albums.map((album) => (
            <Link key={album.id} href={`/music/${album.slug}`}>
              <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                <div className="aspect-square relative overflow-hidden">
                  <img
                    src={album.cover_image}
                    alt={album.title}
                    className="w-full h-full object-cover"
                  />
                </div>
                <CardContent className="p-4">
                  <h3 className="font-bold text-lg mb-1 truncate">{album.title}</h3>
                  <p className="text-gray-600 text-sm mb-2">{album.artist}</p>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-500">{album.track_count} 首歌曲</span>
                    <span className="font-bold text-red-600">¥{album.price.toFixed(2)}</span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
