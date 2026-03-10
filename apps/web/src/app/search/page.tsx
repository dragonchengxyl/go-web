'use client';

import { Suspense, useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';

function Highlight({ text, query }: { text: string; query: string }) {
  if (!query.trim() || !text) return <>{text}</>;
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const parts = text.split(new RegExp(`(${escaped})`, 'gi'));
  return (
    <>
      {parts.map((part, i) =>
        part.toLowerCase() === query.toLowerCase()
          ? <mark key={i} className="bg-yellow-200 dark:bg-yellow-800 rounded px-0.5">{part}</mark>
          : part
      )}
    </>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<div className="max-w-2xl mx-auto pt-20 px-4 text-muted-foreground">搜索中...</div>}>
      <SearchContent />
    </Suspense>
  );
}

interface UserResult {
  id: string;
  username: string;
  furry_name?: string;
  species?: string;
  avatar_key?: string;
  bio?: string;
}

type TabType = 'posts' | 'users' | 'albums';

function SearchContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const query = searchParams.get('q') || '';
  const tabParam = (searchParams.get('tab') as TabType) || 'posts';

  const [tab, setTab] = useState<TabType>(tabParam);
  const [posts, setPosts] = useState<Post[]>([]);
  const [users, setUsers] = useState<UserResult[]>([]);
  const [albums, setAlbums] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); }
  }, []);

  useEffect(() => {
    if (!query) return;
    setLoading(true);
    apiClient.searchAll(query).then(res => {
      setPosts(res.posts || []);
      setUsers((res.users as UserResult[]) || []);
      setAlbums(res.albums || []);
    }).catch(() => {}).finally(() => setLoading(false));
  }, [query]);

  function handleTabChange(t: TabType) {
    setTab(t);
    const params = new URLSearchParams(searchParams.toString());
    params.set('tab', t);
    router.replace(`/search?${params}`);
  }

  if (!query) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 pb-8 text-center py-16 text-muted-foreground">
        输入关键词开始搜索
      </div>
    );
  }

  const tabs: { id: TabType; label: string; count: number }[] = [
    { id: 'posts', label: '动态', count: posts.length },
    { id: 'users', label: '用户', count: users.length },
    { id: 'albums', label: '音乐', count: albums.length },
  ];

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-xl font-bold mb-1">搜索结果</h1>
      <p className="text-sm text-muted-foreground mb-5">
        关于 <span className="font-medium text-foreground">"{query}"</span> 的结果
      </p>

      {/* Tabs */}
      <div className="flex gap-1 border-b mb-6">
        {tabs.map(t => (
          <button
            key={t.id}
            onClick={() => handleTabChange(t.id)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              tab === t.id
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            {t.label} {loading ? '' : `(${t.count})`}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => <div key={i} className="h-32 bg-muted animate-pulse rounded-lg" />)}
        </div>
      ) : (
        <>
          {tab === 'posts' && (
            <div className="space-y-4">
              {posts.length === 0 ? (
                <p className="text-center py-12 text-muted-foreground">未找到相关动态</p>
              ) : (
                posts.map(post => (
                  <PostCard key={post.id} post={post} />
                ))
              )}
            </div>
          )}

          {tab === 'users' && (
            <div className="space-y-3">
              {users.length === 0 ? (
                <p className="text-center py-12 text-muted-foreground">未找到相关用户</p>
              ) : (
                users.map(u => (
                  <Link key={u.id} href={`/users/${u.id}`} className="flex items-center gap-3 p-4 rounded-xl border hover:bg-muted/50 transition-colors">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <span className="font-bold text-primary">{(u.furry_name || u.username)[0]?.toUpperCase()}</span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-semibold truncate">
                        <Highlight text={u.furry_name || u.username} query={query} />
                      </p>
                      <p className="text-sm text-muted-foreground">@<Highlight text={u.username} query={query} /></p>
                      {u.species && <p className="text-xs text-muted-foreground">{u.species}</p>}
                      {u.bio && <p className="text-xs text-muted-foreground mt-0.5 line-clamp-1">{u.bio}</p>}
                    </div>
                  </Link>
                ))
              )}
            </div>
          )}

          {tab === 'albums' && (
            <div className="space-y-3">
              {albums.length === 0 ? (
                <p className="text-center py-12 text-muted-foreground">未找到相关音乐</p>
              ) : (
                albums.map((album: any) => (
                  <Link key={album.id} href={`/music/${album.slug}`} className="flex items-center gap-3 p-4 rounded-xl border hover:bg-muted/50 transition-colors">
                    <div className="w-14 h-14 rounded-lg bg-muted flex-shrink-0 overflow-hidden">
                      {album.cover_image_url && <img src={album.cover_image_url} alt="" className="w-full h-full object-cover" />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-semibold truncate">
                        <Highlight text={album.title} query={query} />
                      </p>
                      <p className="text-sm text-muted-foreground">{album.artist_name}</p>
                      {album.track_count > 0 && <p className="text-xs text-muted-foreground">{album.track_count} 首歌曲</p>}
                    </div>
                  </Link>
                ))
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}
