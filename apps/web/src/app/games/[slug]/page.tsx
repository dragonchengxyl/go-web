'use client';

import { useState, useRef, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent } from '@/components/ui/card';
import {
  ShoppingCart,
  Download,
  Heart,
  Share2,
  ChevronLeft,
  ChevronRight,
  Send,
  User,
} from 'lucide-react';
import { formatPrice, formatDate } from '@/lib/utils';
import { apiClient } from '@/lib/api-client';
import { useCartStore } from '@/lib/store/cart';

// ─── Types ─────────────────────────────────────────────────────────────────

interface Release {
  id: string;
  version: string;
  platform: string;
  file_size: number;
}

interface SystemRequirements {
  minimum?: string;
  recommended?: string;
}

interface Game {
  id: string;
  title: string;
  slug: string;
  subtitle?: string;
  description: string;
  cover_key?: string;
  screenshots?: string[];
  tags?: string[];
  genre?: string[];
  engine?: string;
  release_date?: string;
  system_requirements?: SystemRequirements;
  releases?: Release[];
}

interface Comment {
  id: string;
  user_id: string;
  username?: string;
  content: string;
  created_at: string;
}

// ─── Screenshot Carousel ───────────────────────────────────────────────────

function ScreenshotCarousel({ screenshots }: { screenshots: string[] }) {
  const [current, setCurrent] = useState(0);

  if (!screenshots || screenshots.length === 0) return null;

  const prev = () => setCurrent((c) => (c - 1 + screenshots.length) % screenshots.length);
  const next = () => setCurrent((c) => (c + 1) % screenshots.length);

  return (
    <div className="relative aspect-video overflow-hidden rounded-lg bg-muted select-none">
      <div
        className="flex transition-transform duration-300"
        style={{ transform: `translateX(-${current * 100}%)` }}
      >
        {screenshots.map((src, idx) => (
          <div key={idx} className="min-w-full aspect-video bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={src}
              alt={`截图 ${idx + 1}`}
              className="w-full h-full object-cover"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
        ))}
      </div>

      {screenshots.length > 1 && (
        <>
          <button
            onClick={prev}
            className="absolute left-3 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white rounded-full p-2 transition-colors"
            aria-label="上一张"
          >
            <ChevronLeft className="h-5 w-5" />
          </button>
          <button
            onClick={next}
            className="absolute right-3 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white rounded-full p-2 transition-colors"
            aria-label="下一张"
          >
            <ChevronRight className="h-5 w-5" />
          </button>
          <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-1.5">
            {screenshots.map((_, idx) => (
              <button
                key={idx}
                onClick={() => setCurrent(idx)}
                className={`w-2 h-2 rounded-full transition-colors ${
                  idx === current ? 'bg-white' : 'bg-white/50'
                }`}
                aria-label={`跳转到第 ${idx + 1} 张`}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}

// ─── Comment Section ───────────────────────────────────────────────────────

function CommentSection({ gameId }: { gameId: string }) {
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState(false);

  const { data: commentsData, isLoading, refetch } = useQuery({
    queryKey: ['comments', 'game', gameId],
    queryFn: () =>
      apiClient.get<{ comments: Comment[]; total: number }>(
        `/comments?target_type=game&target_id=${gameId}`
      ),
    enabled: !!gameId,
  });

  const comments = commentsData?.comments || [];

  const token =
    typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;
  const isLoggedIn = !!token;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;

    setSubmitting(true);
    setSubmitError(null);
    setSubmitSuccess(false);

    try {
      await apiClient.post('/comments', {
        target_type: 'game',
        target_id: gameId,
        content: content.trim(),
      });
      setContent('');
      setSubmitSuccess(true);
      refetch();
    } catch (err: any) {
      setSubmitError(err?.message || '发表评论失败，请稍后重试');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">用户评论</h2>

      {/* Comment Form */}
      {isLoggedIn ? (
        <form onSubmit={handleSubmit} className="space-y-3">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="分享你对这款游戏的看法..."
            rows={4}
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            maxLength={1000}
          />
          {submitError && (
            <p className="text-sm text-destructive">{submitError}</p>
          )}
          {submitSuccess && (
            <p className="text-sm text-green-600">评论发表成功！</p>
          )}
          <div className="flex justify-between items-center">
            <span className="text-xs text-muted-foreground">{content.length}/1000</span>
            <Button type="submit" disabled={submitting || !content.trim()} size="sm">
              <Send className="h-4 w-4 mr-2" />
              {submitting ? '发表中...' : '发表评论'}
            </Button>
          </div>
        </form>
      ) : (
        <div className="rounded-md border border-dashed p-6 text-center text-muted-foreground">
          <User className="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p className="text-sm">
            请{' '}
            <a href="/login" className="text-primary hover:underline">
              登录
            </a>{' '}
            后发表评论
          </p>
        </div>
      )}

      {/* Comment List */}
      {isLoading ? (
        <div className="text-center py-8 text-muted-foreground">加载评论中...</div>
      ) : comments.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          暂无评论，成为第一个评论者吧
        </div>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <div key={comment.id} className="flex gap-3">
              <div className="flex-shrink-0 w-9 h-9 rounded-full bg-primary/10 flex items-center justify-center">
                <User className="h-4 w-4 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-baseline gap-2 mb-1">
                  <span className="font-medium text-sm">
                    {comment.username || `用户 ${comment.user_id.slice(0, 6)}`}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(comment.created_at)}
                  </span>
                </div>
                <p className="text-sm text-foreground/80 whitespace-pre-wrap break-words">
                  {comment.content}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Main Page ─────────────────────────────────────────────────────────────

export default function GameDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const addItem = useCartStore((state) => state.addItem);
  const [downloading, setDownloading] = useState<string | null>(null);

  // Fetch game
  const { data: game, isLoading } = useQuery<Game>({
    queryKey: ['game', slug],
    queryFn: () => apiClient.getGameBySlug(slug),
    enabled: !!slug,
  });

  // Fetch product for price
  const { data: productsData } = useQuery({
    queryKey: ['products', 'game'],
    queryFn: () => apiClient.getProducts({ product_type: 'game', is_active: true }),
  });

  // Fetch orders to check purchase status
  const { data: ordersData } = useQuery({
    queryKey: ['orders', 'mine'],
    queryFn: () => apiClient.getOrders(1, 100),
    retry: false,
  });

  const product = productsData?.products?.find((p: any) => p.entity_id === game?.id);
  const price = product?.price_cents || 0;

  // Check if user has purchased this game
  const hasPurchased = ordersData?.orders?.some((order: any) =>
    order.status === 'paid' &&
    order.items?.some((item: any) => item.product_id === product?.id)
  ) || false;

  const handleAddToCart = () => {
    if (!product || !game) return;
    addItem({
      id: game.id,
      productId: product.id,
      name: game.title,
      price,
      coverImage: game.cover_key || '',
    });
  };

  const handleDownload = async (releaseId: string) => {
    setDownloading(releaseId);
    try {
      const result = await apiClient.post<{ download_url: string }>(
        `/releases/${releaseId}/download`
      );
      if (result?.download_url) {
        window.open(result.download_url, '_blank');
      }
    } catch (err: any) {
      alert(err?.message || '获取下载链接失败');
    } finally {
      setDownloading(null);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-24 text-center text-muted-foreground">
            加载中...
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  if (!game) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-24 text-center text-muted-foreground">
            游戏不存在或已下架
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* ── Hero Section ── */}
        <section className="relative h-[60vh] overflow-hidden">
          {/* Background */}
          <div className="absolute inset-0 bg-gradient-to-br from-primary/20 to-secondary/20" />
          {game.cover_key && (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={game.cover_key}
              alt={game.title}
              className="absolute inset-0 w-full h-full object-cover opacity-30"
            />
          )}
          <div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />

          {/* Hero Content */}
          <div className="relative container mx-auto px-4 h-full flex items-end pb-12">
            <div className="max-w-3xl">
              {/* Tags */}
              {game.tags && game.tags.length > 0 && (
                <div className="flex flex-wrap gap-2 mb-4">
                  {game.tags.map((tag) => (
                    <Badge key={tag} variant="secondary">
                      {tag}
                    </Badge>
                  ))}
                </div>
              )}

              <h1 className="text-4xl md:text-5xl font-bold mb-3">{game.title}</h1>
              {game.subtitle && (
                <p className="text-xl text-muted-foreground mb-2">{game.subtitle}</p>
              )}

              {/* Price & Actions */}
              <div className="flex flex-wrap items-center gap-3 mt-6">
                {price > 0 && (
                  <span className="text-3xl font-bold">{formatPrice(price)}</span>
                )}

                {hasPurchased ? (
                  // Already purchased — show download buttons per release
                  game.releases && game.releases.length > 0 ? (
                    game.releases.map((release) => (
                      <Button
                        key={release.id}
                        size="lg"
                        className="gap-2"
                        disabled={downloading === release.id}
                        onClick={() => handleDownload(release.id)}
                      >
                        <Download className="h-5 w-5" />
                        {downloading === release.id
                          ? '获取中...'
                          : `下载 ${release.platform ? `(${release.platform})` : ''}`}
                      </Button>
                    ))
                  ) : (
                    <Button size="lg" variant="outline" disabled>
                      <Download className="h-5 w-5 mr-2" />
                      暂无可下载版本
                    </Button>
                  )
                ) : (
                  // Not purchased — add to cart
                  <Button
                    size="lg"
                    className="gap-2"
                    onClick={handleAddToCart}
                    disabled={!product}
                  >
                    <ShoppingCart className="h-5 w-5" />
                    {product ? '加入购物车' : '暂未上架'}
                  </Button>
                )}

                <Button size="lg" variant="outline" className="gap-2">
                  <Heart className="h-5 w-5" />
                  收藏
                </Button>
                <Button
                  size="lg"
                  variant="ghost"
                  className="gap-2"
                  onClick={() => {
                    if (typeof navigator !== 'undefined' && navigator.share) {
                      navigator.share({ title: game.title, url: window.location.href });
                    } else {
                      navigator.clipboard?.writeText(window.location.href);
                    }
                  }}
                >
                  <Share2 className="h-5 w-5" />
                  分享
                </Button>
              </div>
            </div>
          </div>
        </section>

        {/* ── Main Content ── */}
        <section className="container mx-auto px-4 py-12">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
            {/* Left / Main */}
            <div className="lg:col-span-2 space-y-8">
              {/* Screenshots */}
              {game.screenshots && game.screenshots.length > 0 && (
                <div>
                  <h2 className="text-xl font-bold mb-4">游戏截图</h2>
                  <ScreenshotCarousel screenshots={game.screenshots} />
                  {game.screenshots.length > 1 && (
                    <div className="flex gap-2 mt-3 overflow-x-auto pb-1">
                      {game.screenshots.map((src, idx) => (
                        <div
                          key={idx}
                          className="flex-shrink-0 w-20 h-12 rounded overflow-hidden bg-muted"
                        >
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img
                            src={src}
                            alt={`缩略图 ${idx + 1}`}
                            className="w-full h-full object-cover"
                          />
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Tabs: About / Requirements / Comments */}
              <Tabs defaultValue="about" className="w-full">
                <TabsList className="w-full justify-start border-b rounded-none bg-transparent px-0 h-auto gap-6">
                  <TabsTrigger
                    value="about"
                    className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent pb-2"
                  >
                    关于游戏
                  </TabsTrigger>
                  <TabsTrigger
                    value="requirements"
                    className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent pb-2"
                  >
                    系统要求
                  </TabsTrigger>
                  <TabsTrigger
                    value="comments"
                    className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent pb-2"
                  >
                    评论
                  </TabsTrigger>
                </TabsList>

                {/* About */}
                <TabsContent value="about" className="mt-6">
                  <div className="prose prose-lg dark:prose-invert max-w-none">
                    <p className="text-foreground/80 leading-relaxed whitespace-pre-wrap">
                      {game.description || '暂无介绍'}
                    </p>
                  </div>
                </TabsContent>

                {/* System Requirements */}
                <TabsContent value="requirements" className="mt-6">
                  {game.system_requirements ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      {game.system_requirements.minimum && (
                        <Card>
                          <CardContent className="p-6">
                            <h3 className="font-bold text-base mb-3 text-muted-foreground uppercase tracking-wide text-xs">
                              最低配置
                            </h3>
                            <p className="text-sm whitespace-pre-wrap leading-relaxed">
                              {game.system_requirements.minimum}
                            </p>
                          </CardContent>
                        </Card>
                      )}
                      {game.system_requirements.recommended && (
                        <Card>
                          <CardContent className="p-6">
                            <h3 className="font-bold text-base mb-3 text-muted-foreground uppercase tracking-wide text-xs">
                              推荐配置
                            </h3>
                            <p className="text-sm whitespace-pre-wrap leading-relaxed">
                              {game.system_requirements.recommended}
                            </p>
                          </CardContent>
                        </Card>
                      )}
                    </div>
                  ) : (
                    <p className="text-muted-foreground text-center py-10">
                      暂无系统要求信息
                    </p>
                  )}
                </TabsContent>

                {/* Comments */}
                <TabsContent value="comments" className="mt-6">
                  <CommentSection gameId={game.id} />
                </TabsContent>
              </Tabs>
            </div>

            {/* Right / Sidebar */}
            <div className="space-y-6">
              {/* Cover */}
              <div className="rounded-lg overflow-hidden border bg-muted aspect-video">
                {game.cover_key ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={game.cover_key}
                    alt={game.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
                )}
              </div>

              {/* Game Info */}
              <Card>
                <CardContent className="p-6">
                  <h3 className="font-bold text-base mb-4">游戏信息</h3>
                  <dl className="space-y-3 text-sm">
                    {game.genre && game.genre.length > 0 && (
                      <div>
                        <dt className="text-muted-foreground mb-1">类型</dt>
                        <dd className="font-medium">{game.genre.join('、')}</dd>
                      </div>
                    )}
                    {game.engine && (
                      <div>
                        <dt className="text-muted-foreground mb-1">引擎</dt>
                        <dd className="font-medium">{game.engine}</dd>
                      </div>
                    )}
                    {game.release_date && (
                      <div>
                        <dt className="text-muted-foreground mb-1">发行日期</dt>
                        <dd className="font-medium">
                          {new Date(game.release_date).toLocaleDateString('zh-CN')}
                        </dd>
                      </div>
                    )}
                    {game.tags && game.tags.length > 0 && (
                      <div>
                        <dt className="text-muted-foreground mb-1">标签</dt>
                        <dd className="flex flex-wrap gap-1">
                          {game.tags.map((tag) => (
                            <Badge key={tag} variant="outline" className="text-xs">
                              {tag}
                            </Badge>
                          ))}
                        </dd>
                      </div>
                    )}
                  </dl>
                </CardContent>
              </Card>

              {/* Purchase / Download Card */}
              <Card>
                <CardContent className="p-6">
                  <div className="text-center mb-4">
                    {price > 0 ? (
                      <p className="text-3xl font-bold">{formatPrice(price)}</p>
                    ) : (
                      <p className="text-lg font-medium text-muted-foreground">暂未定价</p>
                    )}
                  </div>

                  {hasPurchased ? (
                    <div className="space-y-2">
                      <p className="text-sm text-center text-green-600 font-medium mb-3">
                        已购买
                      </p>
                      {game.releases && game.releases.length > 0 ? (
                        game.releases.map((release) => (
                          <Button
                            key={release.id}
                            className="w-full gap-2"
                            disabled={downloading === release.id}
                            onClick={() => handleDownload(release.id)}
                          >
                            <Download className="h-4 w-4" />
                            {downloading === release.id
                              ? '获取中...'
                              : `下载 ${release.version || ''}${release.platform ? ` (${release.platform})` : ''}`}
                          </Button>
                        ))
                      ) : (
                        <Button className="w-full" variant="outline" disabled>
                          暂无可下载版本
                        </Button>
                      )}
                    </div>
                  ) : (
                    <Button
                      className="w-full gap-2"
                      onClick={handleAddToCart}
                      disabled={!product}
                    >
                      <ShoppingCart className="h-4 w-4" />
                      {product ? '加入购物车' : '暂未上架'}
                    </Button>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  );
}
