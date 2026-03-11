'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { UserPlus, UserMinus, MessageCircle, MapPin, Globe, ShieldX, Grid3X3 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
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

function avatarUrl(key?: string): string | null {
  if (!key) return null
  if (key.startsWith('http') || key.startsWith('/')) return key
  return `/uploads/images/${key}`
}

const ROLE_BADGE: Record<string, { label: string; color: string }> = {
  super_admin: { label: 'Super Admin', color: 'bg-red-500/10 text-red-500' },
  admin: { label: '管理员', color: 'bg-orange-500/10 text-orange-500' },
  moderator: { label: '审核员', color: 'bg-yellow-500/10 text-yellow-600' },
  creator: { label: '创作者', color: 'bg-brand-purple/10 text-brand-purple' },
  supporter: { label: '支持者', color: 'bg-brand-teal/10 text-brand-teal' },
  member: { label: '成员', color: 'bg-muted text-muted-foreground' },
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
        apiClient.getUserPosts(userId, 1, 30),
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
      <div className="max-w-4xl mx-auto pt-20 px-4">
        <div className="h-48 bg-muted animate-pulse rounded-2xl mb-4" />
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-8">
          {[1, 2, 3, 4, 5, 6].map(i => <div key={i} className="aspect-[4/3] bg-muted animate-pulse rounded-xl" />)}
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="max-w-4xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        <p className="text-xl font-medium mb-2">用户不存在</p>
        <p className="text-sm">该用户可能已注销或不存在</p>
      </div>
    );
  }

  const isSelf = myId === userId;
  const displayName = profile.furry_name || profile.username;
  const av = avatarUrl(profile.avatar_key);
  const roleBadge = ROLE_BADGE[profile.role];
  const totalLikes = posts.reduce((sum, p) => sum + p.like_count, 0);

  return (
    <div className="max-w-4xl mx-auto pt-20 px-4 pb-12">
      {/* Profile card */}
      <div className="bg-card border rounded-2xl overflow-hidden mb-8">
        {/* Cover banner */}
        <div className="h-32 bg-gradient-to-br from-brand-purple/40 via-brand-teal/30 to-brand-coral/20" />

        <div className="px-6 pb-6">
          {/* Avatar row */}
          <div className="flex items-end justify-between -mt-12 mb-4">
            <div className="w-24 h-24 rounded-full bg-background border-4 border-background overflow-hidden shadow-lg flex-shrink-0">
              {av ? (
                <img src={av} alt="" className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center">
                  <span className="text-3xl font-bold text-white">{displayName[0]?.toUpperCase() || '?'}</span>
                </div>
              )}
            </div>

            <div className="flex items-center gap-2">
              {!isSelf && apiClient.getToken() && (
                <>
                  <Button
                    variant={isFollowing ? 'outline' : 'default'}
                    size="sm"
                    onClick={handleFollow}
                    disabled={followLoading || isBlocked}
                    className={isFollowing ? '' : 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110'}
                  >
                    {isFollowing ? (
                      <><UserMinus className="h-4 w-4 mr-1" />取消关注</>
                    ) : (
                      <><UserPlus className="h-4 w-4 mr-1" />关注 TA</>
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

          {/* Name + badge */}
          <div className="mb-3">
            <div className="flex items-center gap-2 flex-wrap">
              <h1 className="text-2xl font-bold">{displayName}</h1>
              {roleBadge && (
                <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${roleBadge.color}`}>
                  {roleBadge.label}
                </span>
              )}
            </div>
            {profile.furry_name && (
              <p className="text-muted-foreground text-sm">@{profile.username}</p>
            )}
            {profile.species && (
              <p className="text-sm text-primary mt-0.5">🐾 {profile.species}</p>
            )}
          </div>

          {/* Bio */}
          {profile.bio && (
            <p className="text-sm text-muted-foreground leading-relaxed mb-3 whitespace-pre-wrap">{profile.bio}</p>
          )}

          {/* Links */}
          <div className="flex flex-wrap gap-4 text-xs text-muted-foreground mb-4">
            {profile.location && (
              <span className="flex items-center gap-1">
                <MapPin className="h-3.5 w-3.5" />{profile.location}
              </span>
            )}
            {profile.website && (
              <a href={profile.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-primary transition-colors">
                <Globe className="h-3.5 w-3.5" />{profile.website}
              </a>
            )}
          </div>

          {/* Stats row */}
          <div className="flex gap-6 text-sm border-t pt-4">
            <Link href={`/users/${userId}/followers`} className="hover:text-primary transition-colors">
              <span className="font-bold">{stats?.follower_count ?? 0}</span>
              <span className="text-muted-foreground ml-1">粉丝</span>
            </Link>
            <Link href={`/users/${userId}/following`} className="hover:text-primary transition-colors">
              <span className="font-bold">{stats?.following_count ?? 0}</span>
              <span className="text-muted-foreground ml-1">关注</span>
            </Link>
            <div>
              <span className="font-bold">{posts.length}</span>
              <span className="text-muted-foreground ml-1">帖子</span>
            </div>
            <div>
              <span className="font-bold">{totalLikes}</span>
              <span className="text-muted-foreground ml-1">获赞</span>
            </div>
          </div>
        </div>
      </div>

      {/* Posts gallery */}
      <div className="flex items-center gap-2 mb-5">
        <Grid3X3 className="h-4 w-4 text-muted-foreground" />
        <h2 className="font-semibold">TA 的帖子</h2>
      </div>

      {posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Grid3X3 className="h-8 w-8 opacity-40" />
          </div>
          <p>暂无帖子</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {posts.map(post => (
            <PostGalleryCard key={post.id} post={post} />
          ))}
        </div>
      )}
    </div>
  );
}
