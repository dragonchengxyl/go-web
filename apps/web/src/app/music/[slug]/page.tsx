'use client'

import { useQuery } from '@tanstack/react-query'
import { useParams } from 'next/navigation'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useCartStore } from '@/lib/store/cart'
import { usePlayerStore } from '@/components/music-player'

interface Track {
  id: number
  title: string
  duration: number
  track_number: number
  audio_url: string
}

interface Album {
  id: number
  slug: string
  title: string
  artist: string
  cover_image: string
  release_date: string
  price: number
  description: string
  tracks: Track[]
}

export default function AlbumDetailPage() {
  const params = useParams()
  const slug = params.slug as string
  const addItem = useCartStore((state) => state.addItem)
  const { setTrack, currentTrack } = usePlayerStore()

  const { data: album, isLoading } = useQuery<Album>({
    queryKey: ['album', slug],
    queryFn: async () => {
      const response = await apiClient.get(`/music/albums/${slug}`)
      return response.data.data
    },
  })

  const handleAddToCart = () => {
    if (!album) return
    addItem({
      id: album.id,
      name: album.title,
      price: album.price,
      quantity: 1,
      image: album.cover_image,
    })
    alert('已添加到购物车！')
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const handlePlayTrack = (track: Track) => {
    if (!album) return
    setTrack({
      id: track.id,
      title: track.title,
      artist: album.artist,
      cover: album.cover_image,
      audioUrl: track.audio_url,
    })
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  if (!album) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">专辑不存在</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        <div className="lg:col-span-1">
          <img
            src={album.cover_image}
            alt={album.title}
            className="w-full rounded-lg shadow-lg"
          />
        </div>

        <div className="lg:col-span-2">
          <h1 className="text-4xl font-bold mb-2">{album.title}</h1>
          <p className="text-xl text-gray-600 mb-4">{album.artist}</p>

          <div className="flex items-center gap-4 mb-6">
            <span className="text-gray-500">
              {new Date(album.release_date).getFullYear()}
            </span>
            <span className="text-gray-500">•</span>
            <span className="text-gray-500">{album.tracks.length} 首歌曲</span>
          </div>

          <p className="text-gray-700 mb-6 leading-relaxed">{album.description}</p>

          <div className="flex gap-4">
            <Button size="lg" onClick={handleAddToCart}>
              购买专辑 - ¥{album.price.toFixed(2)}
            </Button>
            <Button
              size="lg"
              variant="outline"
              onClick={() => album.tracks[0] && handlePlayTrack(album.tracks[0])}
            >
              试听
            </Button>
          </div>
        </div>
      </div>

      <Card>
        <CardContent className="p-6">
          <h2 className="text-2xl font-bold mb-4">曲目列表</h2>
          <div className="space-y-2">
            {album.tracks.map((track) => (
              <div
                key={track.id}
                className={`flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 cursor-pointer ${
                  currentTrack?.id === track.id ? 'bg-blue-50' : ''
                }`}
                onClick={() => handlePlayTrack(track)}
              >
                <div className="flex items-center gap-4">
                  <span className="text-gray-500 w-8 text-center">
                    {track.track_number}
                  </span>
                  <div>
                    <p className="font-medium">{track.title}</p>
                  </div>
                </div>
                <span className="text-gray-500">{formatDuration(track.duration)}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
