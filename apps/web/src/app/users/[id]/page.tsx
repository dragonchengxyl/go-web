'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { UserPlus, UserMinus, MessageCircle, MapPin, Globe, ShieldX } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { PostCard } from '@/components/post/post-card';
import { apiClient, Post, FollowStats } from '@/lib/api-client';

interface UserProfile {
  id: string
  username: string
  email?: string
  bio?: string
  website?: string
  location?: string
  furry_name?: string
  species?: string
  avatar_key?: string
  role: string
  created_at: string
}

export default function UserProfilePage() {
  const params = useParams();
  const userId = params.id as string;

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<FollowStats | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isFollowing, setIsFollowing] = useState(false);
  const [isBlocked, setIsBlocked] = useState(false);
  const [myId, setMyId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [followLoading, setFollowLoading] = useState(false);
  const [blockLoading, setBlockLoading] = useState(false);

  useEffect(() => {
    if (!userId) return;
    const token = localStorage.getItem('access_token');
    if (token) {
      apiClient.setToken(token);
      apiClient.getMe().then(me => setMyId(me?.id ?? null)).catch(() => {});
    }
    loadProfile();
  }, [userId]);

  async function loadProfile() {
    try {
      const [profileData, statsData, postsData] = await Promise.all([
        apiClient.getUser(userId),
        apiClient.getFollowStats(userId),
        apiClient.getUserPosts(userId, 1, 20),
      ]);
      setProfile(profileData);
      setStats(statsData);
      setPosts(postsData.posts ?? []);

      if (apiClient.getToken()) {
        try {
          const me = await apiClient.getMe();
          if (me?.id !== userId) {
            const followersData = await apiClient.getFollowers(userId, 1, 1000);
            const isF = followersData.followers?.some(f => f.follower_id === me?.id) ?? false;
            setIsFollowing(isF);
          }
        } catch {
          // ignore
        }
      }
    } catch {
      // profile not found
    } finally {
      setLoading(false);
    }
  }

  async function handleFollow() {
    if (!apiClient.getToken()) return;
    setFollowLoading(true);
    try {
      if (isFollowing) {
        await apiClient.unfollowUser(userId);
        setIsFollowing(false);
        setStats(prev => prev ? { ...prev, follower_count: prev.follower_count - 1 } : prev);
      } else {
        await apiClient.followUser(userId);
        setIsFollowing(true);
        setStats(prev => prev ? { ...prev, follower_count: prev.follower_count + 1 } : prev);
      }
    } catch {
      // ignore
    } finally {
      setFollowLoading(false);
    }
  }

  async function handleBlock() {
    if (!apiClient.getToken()) return;
    setBlockLoading(true);
    try {
      if (isBlocked) {
        await apiClient.unblockUser(userId);
        setIsBlocked(false);
      } else {
        await apiClient.blockUser(userId);
        setIsBlocked(true);
        if (isFollowing) {
          await apiClient.unfollowUser(userId);
          setIsFollowing(false);
        }
      }
    } catch {
      // ignore
    } finally {
      setBlockLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-32 bg-muted animate-pulse rounded-xl mb-4" />
        <div className="space-y-3">
          {[1, 2, 3].map(i => <div key={i} className="h-24 bg-muted animate-pulse rounded-xl" />)}
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        用户不存在
      </div>
    );
  }

  const isSelf = myId === userId;
  const displayName = profile.furry_name || profile.username;

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      {/* Profile Header */}
      <div className="bg-card border rounded-xl p-6 mb-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
              <span className="text-2xl font-bold text-primary">
                {displayName[0]?.toUpperCase() || '?'}
              </span>
            </div>
            <div>
              <h1 className="text-xl font-bold">{displayName}</h1>
              {profile.furry_name && (
                <p className="text-sm text-muted-foreground">@{profile.username}</p>
              )}
              {profile.species && (
                <p className="text-sm text-primary">🐾 {profile.species}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {!isSelf && apiClient.getToken() && (
              <>
                <Button
                  variant={isFollowing ? 'outline' : 'default'}
                  size="sm"
                  onClick={handleFollow}
                  disabled={followLoading || isBlocked}
                >
                  {isFollowing ? (
                    <><UserMinus className="h-4 w-4 mr-1" />取消关注</>
                  ) : (
                    <><UserPlus className="h-4 w-4 mr-1" />关注</>
                  )}
                </Button>
                <Link href="/messages">
                  <Button variant="outline" size="sm">
                    <MessageCircle className="h-4 w-4" />
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleBlock}
                  disabled={blockLoading}
                  className={isBlocked ? 'text-muted-foreground' : 'text-red-500 hover:text-red-600'}
                  title={isBlocked ? '取消屏蔽' : '屏蔽用户'}
                >
                  <ShieldX className="h-4 w-4" />
                </Button>
              </>
            )}
            {isSelf && (
              <Link href="/profile">
                <Button variant="outline" size="sm">编辑资料</Button>
              </Link>
            )}
          </div>
        </div>

        {profile.bio && (
          <p className="mt-3 text-sm text-muted-foreground">{profile.bio}</p>
        )}

        <div className="flex flex-wrap gap-3 mt-3 text-xs text-muted-foreground">
          {profile.location && (
            <span className="flex items-center gap-1">
              <MapPin className="h-3 w-3" />{profile.location}
            </span>
          )}
          {profile.website && (
            <a href={profile.website} target="_blank" rel="noopener noreferrer"
              className="flex items-center gap-1 hover:text-primary">
              <Globe className="h-3 w-3" />{profile.website}
            </a>
          )}
        </div>

        {/* Follow stats — clickable links */}
        <div className="flex gap-6 mt-4 pt-4 border-t text-sm">
          <Link href={`/users/${userId}/followers`} className="hover:underline">
            <span className="font-bold">{stats?.follower_count ?? 0}</span>
            <span className="text-muted-foreground ml-1">粉丝</span>
          </Link>
          <Link href={`/users/${userId}/following`} className="hover:underline">
            <span className="font-bold">{stats?.following_count ?? 0}</span>
            <span className="text-muted-foreground ml-1">关注</span>
          </Link>
          <div>
            <span className="font-bold">{posts.length}</span>
            <span className="text-muted-foreground ml-1">帖子</span>
          </div>
        </div>
      </div>

      {/* Posts */}
      <div className="space-y-3">
        {posts.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">暂无帖子</div>
        ) : (
          posts.map(post => <PostCard key={post.id} post={post} />)
        )}
      </div>
    </div>
  );
}
