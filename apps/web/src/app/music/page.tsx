'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useState } from 'react'
import { Search } from 'lucide-react'

export default function MusicPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const { data: albumsData, isLoading } = useQuery({
    queryKey: ['albums', search, page],
    queryFn: () => apiClient.getAlbums({
      search: search || undefined,
      page,
      page_size: 12,
    }),
  })

  // Fetch products to get prices
  const { data: productsData } = useQuery({
    queryKey: ['products', 'ost'],
    queryFn: () => apiClient.getProducts({ product_type: 'ost', is_active: true }),
  })

  const albums = albumsData?.albums?.map(album => {
    const product = productsData?.products?.find(p => p.entity_id === album.id);
    return {
      ...album,
      price: product?.price_cents || 0,
      productId: product?.id,
    };
  }) || [];

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* Hero Section */}
        <section className="bg-gradient-to-br from-primary/10 to-secondary/10 py-20">
          <div className="container mx-auto px-4">
            <div className="max-w-3xl mx-auto text-center">
              <h1 className="text-5xl font-bold mb-6">游戏原声音乐</h1>
              <p className="text-xl text-muted-foreground mb-8">
                探索我们工作室游戏的精彩原声音乐
              </p>

              {/* Search Bar */}
              <div className="flex gap-2 max-w-xl mx-auto">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    type="search"
                    placeholder="搜索专辑或艺术家..."
                    className="pl-10"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
                </div>
                <Button onClick={() => setPage(1)}>搜索</Button>
              </div>
            </div>
          </div>
        </section>

        {/* Albums Grid */}
        <section className="container mx-auto px-4 py-12">
          {isLoading ? (
            <div className="text-center py-12">加载中...</div>
          ) : albums.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground">暂无音乐专辑</p>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {albums.map((album) => (
                  <Link key={album.id} href={`/music/${album.slug}`}>
                    <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                      <div className="aspect-square relative overflow-hidden bg-muted">
                        <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
                      </div>
                      <CardContent className="p-4">
                        <h3 className="font-bold text-lg mb-1 truncate">{album.title}</h3>
                        <p className="text-muted-foreground text-sm mb-2">{album.artist || '未知艺术家'}</p>
                        <div className="flex justify-between items-center">
                          <span className="text-sm text-muted-foreground">专辑</span>
                          <span className="font-bold text-primary">¥{(album.price / 100).toFixed(2)}</span>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>

              {/* Pagination */}
              {albumsData && albumsData.total > 12 && (
                <div className="flex justify-center gap-2 mt-12">
                  <Button
                    variant="outline"
                    disabled={page === 1}
                    onClick={() => setPage(p => p - 1)}
                  >
                    上一页
                  </Button>
                  <Button variant="outline">{page}</Button>
                  <Button
                    variant="outline"
                    disabled={page * 12 >= albumsData.total}
                    onClick={() => setPage(p => p + 1)}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </>
          )}
        </section>
      </main>
      <Footer />
    </div>
  )
}
