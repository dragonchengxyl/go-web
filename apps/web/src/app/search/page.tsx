'use client'

import { useSearchParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'

interface SearchResult {
  games: Array<{
    id: number
    slug: string
    title: string
    short_description: string
    cover_image: string
    price: number
    discount_price: number
    tags: string[]
  }>
  albums: Array<{
    id: number
    slug: string
    title: string
    artist: string
    cover_image: string
    price: number
    track_count: number
  }>
  query: string
}

export default function SearchPage() {
  const searchParams = useSearchParams()
  const query = searchParams.get('q') || ''

  const { data: results, isLoading } = useQuery({
    queryKey: ['search', query],
    queryFn: () => apiClient.searchAll(query),
    enabled: !!query,
  })

  if (!query) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-gray-500">请输入搜索关键词</p>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">加载中...</div>
      </div>
    )
  }

  const totalResults = (results?.games?.length || 0) + (results?.albums?.length || 0)

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">搜索结果</h1>
        <p className="text-gray-600">
          找到 <span className="font-medium">{totalResults}</span> 个关于 "
          <span className="font-medium">{query}</span>" 的结果
        </p>
      </div>

      {totalResults === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-gray-500 mb-4">未找到相关结果</p>
            <p className="text-sm text-gray-400">
              尝试使用不同的关键词或浏览我们的游戏和音乐库
            </p>
          </CardContent>
        </Card>
      ) : (
        <Tabs defaultValue="all" className="space-y-6">
          <TabsList>
            <TabsTrigger value="all">
              全部 ({totalResults})
            </TabsTrigger>
            <TabsTrigger value="games">
              游戏 ({results?.games?.length || 0})
            </TabsTrigger>
            <TabsTrigger value="music">
              音乐 ({results?.albums?.length || 0})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="all">
            {results?.games && results.games.length > 0 && (
              <div className="mb-8">
                <h2 className="text-xl font-bold mb-4">游戏</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {results.games.map((game) => (
                    <Link key={game.id} href={`/games/${game.slug}`}>
                      <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                        <div className="aspect-video relative overflow-hidden">
                          <img
                            src={game.cover_image}
                            alt={game.title}
                            className="w-full h-full object-cover"
                          />
                        </div>
                        <CardContent className="p-4">
                          <h3 className="font-bold text-lg mb-2">{game.title}</h3>
                          <p className="text-sm text-gray-600 mb-3 line-clamp-2">
                            {game.short_description}
                          </p>
                          <div className="flex flex-wrap gap-2 mb-3">
                            {game.tags?.slice(0, 3).map((tag, index) => (
                              <Badge key={index} variant="outline">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                          <div className="flex items-center gap-2">
                            {game.discount_price > 0 ? (
                              <>
                                <span className="text-lg font-bold text-red-600">
                                  ¥{game.discount_price.toFixed(2)}
                                </span>
                                <span className="text-sm text-gray-500 line-through">
                                  ¥{game.price.toFixed(2)}
                                </span>
                              </>
                            ) : (
                              <span className="text-lg font-bold">
                                ¥{game.price.toFixed(2)}
                              </span>
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {results?.albums && results.albums.length > 0 && (
              <div>
                <h2 className="text-xl font-bold mb-4">音乐</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  {results.albums.map((album) => (
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
                          <h3 className="font-bold mb-1 truncate">{album.title}</h3>
                          <p className="text-sm text-gray-600 mb-2">{album.artist}</p>
                          <div className="flex justify-between items-center">
                            <span className="text-sm text-gray-500">
                              {album.track_count} 首歌曲
                            </span>
                            <span className="font-bold text-red-600">
                              ¥{album.price.toFixed(2)}
                            </span>
                          </div>
                        </CardContent>
                      </Card>
                    </Link>
                  ))}
                </div>
              </div>
            )}
          </TabsContent>

          <TabsContent value="games">
            {results?.games && results.games.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {results.games.map((game) => (
                  <Link key={game.id} href={`/games/${game.slug}`}>
                    <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                      <div className="aspect-video relative overflow-hidden">
                        <img
                          src={game.cover_image}
                          alt={game.title}
                          className="w-full h-full object-cover"
                        />
                      </div>
                      <CardContent className="p-4">
                        <h3 className="font-bold text-lg mb-2">{game.title}</h3>
                        <p className="text-sm text-gray-600 mb-3 line-clamp-2">
                          {game.short_description}
                        </p>
                        <div className="flex flex-wrap gap-2 mb-3">
                          {game.tags?.slice(0, 3).map((tag, index) => (
                            <Badge key={index} variant="outline">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                        <div className="flex items-center gap-2">
                          {game.discount_price > 0 ? (
                            <>
                              <span className="text-lg font-bold text-red-600">
                                ¥{game.discount_price.toFixed(2)}
                              </span>
                              <span className="text-sm text-gray-500 line-through">
                                ¥{game.price.toFixed(2)}
                              </span>
                            </>
                          ) : (
                            <span className="text-lg font-bold">
                              ¥{game.price.toFixed(2)}
                            </span>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-12 text-center text-gray-500">
                  未找到相关游戏
                </CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="music">
            {results?.albums && results.albums.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {results.albums.map((album) => (
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
                        <h3 className="font-bold mb-1 truncate">{album.title}</h3>
                        <p className="text-sm text-gray-600 mb-2">{album.artist}</p>
                        <div className="flex justify-between items-center">
                          <span className="text-sm text-gray-500">
                            {album.track_count} 首歌曲
                          </span>
                          <span className="font-bold text-red-600">
                            ¥{album.price.toFixed(2)}
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-12 text-center text-gray-500">
                  未找到相关音乐
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      )}
    </div>
  )
}
