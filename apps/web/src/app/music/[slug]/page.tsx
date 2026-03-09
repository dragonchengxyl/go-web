'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Play,
  Pause,
  ShoppingCart,
  Music2,
  Clock,
  ChevronLeft,
} from 'lucide-react';
import Link from 'next/link';
import { formatPrice } from '@/lib/utils';
import { apiClient } from '@/lib/api-client';
import { useCartStore } from '@/lib/store/cart';

// ─── Types ─────────────────────────────────────────────────────────────────

interface Track {
  id: string;
  title: string;
  duration: number;         // seconds
  track_number: number;
  preview_url?: string;
}

interface Album {
  id: string;
  slug: string;
  title: string;
  artist?: string;
  cover_key?: string;
  description?: string;
  release_date?: string;
  genre?: string;
  total_tracks?: number;
}

// ─── Format helpers ────────────────────────────────────────────────────────

function formatDuration(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '0:00';
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, '0')}`;
}

function formatTotalDuration(tracks: Track[]): string {
  const total = tracks.reduce((sum, t) => sum + (t.duration || 0), 0);
  if (total < 60) return `${total} 秒`;
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  if (h > 0) return `${h} 小时 ${m} 分钟`;
  return `${m} 分钟`;
}

// ─── Audio Player Bar ──────────────────────────────────────────────────────

interface PlayerBarProps {
  streamUrl: string;
  trackTitle: string;
  albumTitle: string;
  artist?: string;
  onClose: () => void;
}

function PlayerBar({ streamUrl, trackTitle, albumTitle, artist, onClose }: PlayerBarProps) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);

  // Auto-play when URL changes
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;
    audio.src = streamUrl;
    audio.play().then(() => setIsPlaying(true)).catch(() => setIsPlaying(false));
  }, [streamUrl]);

  const togglePlay = () => {
    const audio = audioRef.current;
    if (!audio) return;
    if (isPlaying) {
      audio.pause();
      setIsPlaying(false);
    } else {
      audio.play().then(() => setIsPlaying(true)).catch(() => {});
    }
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const t = parseFloat(e.target.value);
    if (audioRef.current) {
      audioRef.current.currentTime = t;
      setCurrentTime(t);
    }
  };

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 bg-background/95 backdrop-blur border-t shadow-xl">
      <audio
        ref={audioRef}
        onTimeUpdate={() => setCurrentTime(audioRef.current?.currentTime || 0)}
        onLoadedMetadata={() => setDuration(audioRef.current?.duration || 0)}
        onEnded={() => setIsPlaying(false)}
      />
      <div className="container mx-auto px-4 py-3">
        <div className="flex items-center gap-4">
          {/* Track Info */}
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <div className="w-10 h-10 rounded bg-primary/10 flex items-center justify-center flex-shrink-0">
              <Music2 className="h-5 w-5 text-primary" />
            </div>
            <div className="min-w-0">
              <p className="font-medium text-sm truncate">{trackTitle}</p>
              <p className="text-xs text-muted-foreground truncate">
                {artist || albumTitle}
              </p>
            </div>
          </div>

          {/* Controls */}
          <div className="flex items-center gap-3 flex-[2] min-w-0">
            <Button variant="ghost" size="icon" onClick={togglePlay} className="flex-shrink-0">
              {isPlaying ? (
                <Pause className="h-5 w-5" />
              ) : (
                <Play className="h-5 w-5" />
              )}
            </Button>

            <div className="flex-1 flex items-center gap-2 min-w-0">
              <span className="text-xs text-muted-foreground w-10 text-right flex-shrink-0">
                {formatDuration(currentTime)}
              </span>
              <input
                type="range"
                min={0}
                max={duration || 0}
                value={currentTime}
                step={0.5}
                onChange={handleSeek}
                className="flex-1 accent-primary"
                aria-label="播放进度"
              />
              <span className="text-xs text-muted-foreground w-10 flex-shrink-0">
                {formatDuration(duration)}
              </span>
            </div>
          </div>

          {/* Close */}
          <Button variant="ghost" size="sm" onClick={onClose} className="flex-shrink-0 text-muted-foreground">
            关闭
          </Button>
        </div>
      </div>
    </div>
  );
}

// ─── Track Row ─────────────────────────────────────────────────────────────

interface TrackRowProps {
  track: Track;
  isActive: boolean;
  isLoading: boolean;
  onPlay: (trackId: string) => void;
}

function TrackRow({ track, isActive, isLoading, onPlay }: TrackRowProps) {
  return (
    <div
      className={`flex items-center gap-4 px-4 py-3 rounded-lg transition-colors group ${
        isActive
          ? 'bg-primary/10 text-primary'
          : 'hover:bg-muted/60 text-foreground'
      }`}
    >
      {/* Number / Play Button */}
      <div className="w-8 flex items-center justify-center flex-shrink-0">
        <span
          className={`text-sm tabular-nums group-hover:hidden ${
            isActive ? 'hidden' : 'block'
          }`}
        >
          {track.track_number}
        </span>
        <Button
          variant="ghost"
          size="icon"
          className={`h-7 w-7 ${isActive ? 'flex' : 'hidden group-hover:flex'}`}
          disabled={isLoading}
          onClick={() => onPlay(track.id)}
          aria-label={`播放 ${track.title}`}
        >
          {isLoading && isActive ? (
            <span className="block w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          ) : isActive ? (
            <Pause className="h-4 w-4" />
          ) : (
            <Play className="h-4 w-4" />
          )}
        </Button>
      </div>

      {/* Title */}
      <div className="flex-1 min-w-0">
        <p
          className={`font-medium text-sm truncate ${
            isActive ? 'text-primary' : ''
          }`}
        >
          {track.title}
        </p>
      </div>

      {/* Duration */}
      <span className="text-sm text-muted-foreground tabular-nums flex-shrink-0 flex items-center gap-1">
        <Clock className="h-3 w-3" />
        {formatDuration(track.duration)}
      </span>
    </div>
  );
}

// ─── Main Page ─────────────────────────────────────────────────────────────

export default function AlbumDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const addItem = useCartStore((state) => state.addItem);

  // Currently streaming track state
  const [activeTrackId, setActiveTrackId] = useState<string | null>(null);
  const [streamUrl, setStreamUrl] = useState<string | null>(null);
  const [loadingTrackId, setLoadingTrackId] = useState<string | null>(null);
  const [streamError, setStreamError] = useState<string | null>(null);

  // Fetch album by slug
  const { data: album, isLoading: albumLoading } = useQuery<Album>({
    queryKey: ['album', slug],
    queryFn: () => apiClient.getAlbumBySlug(slug),
    enabled: !!slug,
  });

  // Fetch tracks once album is loaded
  const { data: tracksData, isLoading: tracksLoading } = useQuery<{ tracks: Track[]; total: number }>({
    queryKey: ['album-tracks', album?.id],
    queryFn: () =>
      apiClient.get<{ tracks: Track[]; total: number }>(`/albums/${album!.id}/tracks`),
    enabled: !!album?.id,
  });

  // Fetch products for price
  const { data: productsData } = useQuery({
    queryKey: ['products', 'ost'],
    queryFn: () => apiClient.getProducts({ product_type: 'ost', is_active: true }),
  });

  // Fetch orders to check purchase status
  const { data: ordersData } = useQuery({
    queryKey: ['orders', 'mine'],
    queryFn: () => apiClient.getOrders(1, 100),
    retry: false,
  });

  const tracks = tracksData?.tracks || [];
  const product = productsData?.products?.find((p: any) => p.entity_id === album?.id);
  const price = product?.price_cents || 0;

  const hasPurchased = ordersData?.orders?.some((order: any) =>
    order.status === 'paid' &&
    order.items?.some((item: any) => item.product_id === product?.id)
  ) || false;

  // Play a track: fetch stream URL then play
  const handlePlayTrack = useCallback(
    async (trackId: string) => {
      // Toggle off if same track
      if (trackId === activeTrackId && streamUrl) {
        // handled by PlayerBar toggle — just signal stop
        setActiveTrackId(null);
        setStreamUrl(null);
        return;
      }

      setLoadingTrackId(trackId);
      setStreamError(null);

      try {
        const result = await apiClient.post<{ stream_url: string }>(
          `/tracks/${trackId}/stream`
        );
        if (result?.stream_url) {
          setStreamUrl(result.stream_url);
          setActiveTrackId(trackId);
        } else {
          setStreamError('未获取到播放链接');
        }
      } catch (err: any) {
        setStreamError(err?.message || '获取播放链接失败，请稍后重试');
      } finally {
        setLoadingTrackId(null);
      }
    },
    [activeTrackId, streamUrl]
  );

  const handleClosePlayer = () => {
    setActiveTrackId(null);
    setStreamUrl(null);
  };

  const handleAddToCart = () => {
    if (!product || !album) return;
    addItem({
      id: album.id,
      productId: product.id,
      name: album.title,
      price,
      coverImage: album.cover_key || '',
    });
  };

  const activeTrack = tracks.find((t) => t.id === activeTrackId);

  // ── Loading / Error States ──────────────────────────────────────────────

  if (albumLoading) {
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

  if (!album) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-24 text-center text-muted-foreground">
            专辑不存在或已下架
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  // ── Render ──────────────────────────────────────────────────────────────

  return (
    <div className={`min-h-screen ${streamUrl ? 'pb-24' : ''}`}>
      <Header />
      <main className="pt-16">
        {/* ── Hero / Album Header ── */}
        <section className="bg-gradient-to-br from-primary/10 to-secondary/10 py-16">
          <div className="container mx-auto px-4">
            {/* Back Link */}
            <Link
              href="/music"
              className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors mb-8"
            >
              <ChevronLeft className="h-4 w-4" />
              返回音乐列表
            </Link>

            <div className="flex flex-col md:flex-row gap-8 items-start">
              {/* Cover Art */}
              <div className="flex-shrink-0 w-56 h-56 md:w-64 md:h-64 rounded-xl overflow-hidden shadow-2xl bg-muted">
                {album.cover_key ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={album.cover_key}
                    alt={album.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-gradient-to-br from-primary/30 to-secondary/30 flex items-center justify-center">
                    <Music2 className="h-20 w-20 text-primary/50" />
                  </div>
                )}
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-2">
                  专辑
                </p>
                <h1 className="text-3xl md:text-4xl font-bold mb-3 break-words">
                  {album.title}
                </h1>
                <p className="text-lg text-muted-foreground mb-1">
                  {album.artist || '未知艺术家'}
                </p>

                {/* Meta */}
                <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground mb-4">
                  {album.release_date && (
                    <span>{new Date(album.release_date).getFullYear()}</span>
                  )}
                  {album.release_date && tracks.length > 0 && (
                    <span className="opacity-40">•</span>
                  )}
                  {tracks.length > 0 && (
                    <span>{tracks.length} 首曲目</span>
                  )}
                  {tracks.length > 0 && (
                    <>
                      <span className="opacity-40">•</span>
                      <span>{formatTotalDuration(tracks)}</span>
                    </>
                  )}
                  {album.genre && (
                    <>
                      <span className="opacity-40">•</span>
                      <Badge variant="secondary">{album.genre}</Badge>
                    </>
                  )}
                </div>

                {album.description && (
                  <p className="text-sm text-muted-foreground mb-6 leading-relaxed max-w-xl">
                    {album.description}
                  </p>
                )}

                {/* Actions */}
                <div className="flex flex-wrap gap-3 items-center">
                  {tracks.length > 0 && (
                    <Button
                      size="lg"
                      className="gap-2"
                      onClick={() => handlePlayTrack(tracks[0].id)}
                      disabled={loadingTrackId === tracks[0].id}
                    >
                      <Play className="h-5 w-5" />
                      {loadingTrackId === tracks[0].id ? '加载中...' : '立即播放'}
                    </Button>
                  )}

                  {hasPurchased ? (
                    <Button size="lg" variant="outline" disabled>
                      已购买
                    </Button>
                  ) : (
                    <Button
                      size="lg"
                      variant="outline"
                      className="gap-2"
                      onClick={handleAddToCart}
                      disabled={!product}
                    >
                      <ShoppingCart className="h-5 w-5" />
                      {product
                        ? `购买 ${price > 0 ? formatPrice(price) : ''}`
                        : '暂未上架'}
                    </Button>
                  )}
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* ── Track List ── */}
        <section className="container mx-auto px-4 py-10">
          <div className="max-w-3xl">
            <h2 className="text-xl font-bold mb-4">曲目列表</h2>

            {streamError && (
              <div className="mb-4 rounded-md bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
                {streamError}
              </div>
            )}

            {tracksLoading ? (
              <div className="text-center py-12 text-muted-foreground">加载曲目中...</div>
            ) : tracks.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                该专辑暂无曲目
              </div>
            ) : (
              <Card>
                <CardContent className="p-2">
                  {/* Header Row */}
                  <div className="flex items-center gap-4 px-4 py-2 text-xs font-medium text-muted-foreground uppercase tracking-wide border-b mb-1">
                    <div className="w-8 text-center">#</div>
                    <div className="flex-1">标题</div>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                    </div>
                  </div>

                  {/* Tracks */}
                  <div className="space-y-0.5">
                    {tracks.map((track) => (
                      <TrackRow
                        key={track.id}
                        track={track}
                        isActive={track.id === activeTrackId}
                        isLoading={loadingTrackId === track.id}
                        onPlay={handlePlayTrack}
                      />
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </section>
      </main>

      <Footer />

      {/* ── Floating Player Bar ── */}
      {streamUrl && activeTrack && (
        <PlayerBar
          streamUrl={streamUrl}
          trackTitle={activeTrack.title}
          albumTitle={album.title}
          artist={album.artist}
          onClose={handleClosePlayer}
        />
      )}
    </div>
  );
}
