'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';

interface FollowUser {
  follower_id: string
  followee_id: string
  created_at: string
  username?: string
  furry_name?: string
  species?: string
}

export default function FollowersPage() {
  const params = useParams();
  const userId = params.id as string;

  const [followers, setFollowers] = useState<FollowUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [myId, setMyId] = useState<string | null>(null);
  const [followingSet, setFollowingSet] = useState<Set<string>>(new Set());

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      apiClient.setToken(token);
      apiClient.getMe().then(me => setMyId(me?.id ?? null)).catch(() => {});
    }
    loadFollowers(1);
  }, [userId]);

  async function loadFollowers(p: number) {
    setLoading(true);
    try {
      const data = await apiClient.getFollowers(userId, p, 20);
      if (p === 1) {
        setFollowers(data.followers ?? []);
      } else {
        setFollowers(prev => [...prev, ...(data.followers ?? [])]);
      }
      setTotal(data.total ?? 0);
      setPage(p);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  async function handleFollow(targetId: string) {
    if (!apiClient.getToken()) return;
    try {
      if (followingSet.has(targetId)) {
        await apiClient.unfollowUser(targetId);
        setFollowingSet(prev => { const s = new Set(prev); s.delete(targetId); return s; });
      } else {
        await apiClient.followUser(targetId);
        setFollowingSet(prev => new Set(prev).add(targetId));
      }
    } catch {
      // ignore
    }
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <div className="flex items-center gap-3 mb-6">
        <Link href={`/users/${userId}`}>
          <Button variant="ghost" size="sm">
            <ArrowLeft className="h-4 w-4 mr-1" />返回
          </Button>
        </Link>
        <h1 className="text-xl font-bold">粉丝 {total > 0 && <span className="text-muted-foreground text-base font-normal">({total})</span>}</h1>
      </div>

      {loading && followers.length === 0 ? (
        <div className="space-y-2">
          {[1, 2, 3, 4, 5].map(i => <div key={i} className="h-16 bg-muted animate-pulse rounded-xl" />)}
        </div>
      ) : followers.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">暂无粉丝</div>
      ) : (
        <>
          <div className="space-y-2">
            {followers.map(f => {
              const displayName = f.furry_name || f.username || f.follower_id;
              const isMe = myId === f.follower_id;
              const isFollowing = followingSet.has(f.follower_id);
              return (
                <div key={f.follower_id} className="flex items-center justify-between p-3 border rounded-xl hover:bg-accent/50 transition-colors">
                  <Link href={`/users/${f.follower_id}`} className="flex items-center gap-3 flex-1 min-w-0">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <span className="text-sm font-bold text-primary">{displayName[0]?.toUpperCase()}</span>
                    </div>
                    <div className="min-w-0">
                      <p className="text-sm font-medium truncate">{displayName}</p>
                      {f.furry_name && f.username && (
                        <p className="text-xs text-muted-foreground">@{f.username}</p>
                      )}
                      {f.species && <p className="text-xs text-primary">🐾 {f.species}</p>}
                    </div>
                  </Link>
                  {!isMe && apiClient.getToken() && (
                    <Button
                      variant={isFollowing ? 'outline' : 'default'}
                      size="sm"
                      className="ml-3 flex-shrink-0"
                      onClick={() => handleFollow(f.follower_id)}
                    >
                      {isFollowing ? '取消关注' : '关注'}
                    </Button>
                  )}
                </div>
              );
            })}
          </div>
          {followers.length < total && (
            <div className="text-center mt-6">
              <Button variant="outline" onClick={() => loadFollowers(page + 1)} disabled={loading}>
                {loading ? '加载中...' : '加载更多'}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
