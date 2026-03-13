'use client'

import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { apiClient, Post, FollowStats } from '@/lib/api-client'
import { PostGalleryCard } from '@/components/post/post-gallery-card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Camera, Loader2, MapPin, Globe, Edit2, Grid3X3, Heart, MessageCircle } from 'lucide-react'

interface UserProfile {
  id: string
  username: string
  email: string
  status: string
  avatar_key?: string
  bio?: string
  website?: string
  location?: string
  furry_name?: string
  species?: string
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

const STATUS_NOTICE: Record<string, string> = {
  inactive: '你的账号尚未激活，部分功能可能不可用。',
  suspended: '你的账号当前处于暂停状态，请联系管理员了解详情。',
  banned: '你的账号已被封禁，如有疑问请联系管理员。',
};

export default function ProfilePage() {
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [stats, setStats] = useState<FollowStats | null>(null)
  const [posts, setPosts] = useState<Post[]>([])
  const [loading, setLoading] = useState(true)
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [avatarUploading, setAvatarUploading] = useState(false)
  const [error, setError] = useState('')
  const [formData, setFormData] = useState({
    bio: '',
    website: '',
    location: '',
    furry_name: '',
    species: '',
  })

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) { router.push('/login'); return }
    apiClient.setToken(token)

    Promise.all([
      apiClient.getMe(),
    ]).then(async ([u]: [UserProfile]) => {
      setProfile(u)
      setFormData({
        bio: u.bio || '',
        website: u.website || '',
        location: u.location || '',
        furry_name: u.furry_name || '',
        species: u.species || '',
      })
      const [statsData, postsData] = await Promise.all([
        apiClient.getFollowStats(u.id).catch(() => null),
        apiClient.getUserPosts(u.id, 1, 30).catch(() => ({ posts: [] })),
      ])
      setStats(statsData)
      setPosts(postsData.posts ?? [])
    }).catch(() => router.push('/login')).finally(() => setLoading(false))
  }, [router])

  async function handleAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setAvatarUploading(true)
    try {
      const { url } = await apiClient.uploadFile('/upload/image', file)
      const updated = await apiClient.updateProfile({ avatar_key: url })
      setProfile(updated)
    } catch (e: any) {
      setError(e.message || '头像上传失败')
    } finally {
      setAvatarUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      const updated = await apiClient.updateProfile({
        bio: formData.bio || undefined,
        website: formData.website || undefined,
        location: formData.location || undefined,
        furry_name: formData.furry_name || undefined,
        species: formData.species || undefined,
      })
      setProfile(updated)
      setIsEditing(false)
    } catch (e: any) {
      setError(e.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto pt-20 px-4">
        <div className="h-48 bg-muted animate-pulse rounded-2xl mb-4" />
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-8">
          {[1,2,3,4,5,6].map(i => <div key={i} className="aspect-[4/3] bg-muted animate-pulse rounded-xl" />)}
        </div>
      </div>
    )
  }
  if (!profile) return null

  const av = avatarUrl(profile.avatar_key)
  const displayName = profile.furry_name || profile.username
  const roleBadge = ROLE_BADGE[profile.role]
  const totalLikes = posts.reduce((sum, p) => sum + p.like_count, 0)
  const statusNotice = STATUS_NOTICE[profile.status]

  return (
    <div className="max-w-4xl mx-auto pt-20 px-4 pb-12">
      {/* Profile card */}
      <div className="bg-card border rounded-2xl overflow-hidden mb-8">
        {/* Cover banner */}
        <div className="h-32 bg-gradient-to-br from-brand-purple/40 via-brand-teal/30 to-brand-coral/20" />

        <div className="px-6 pb-6">
          {statusNotice && (
            <div className="mt-4 mb-2 rounded-xl border border-amber-500/20 bg-amber-500/10 px-4 py-3 text-sm text-amber-700 dark:text-amber-400">
              {statusNotice}
            </div>
          )}

          {/* Avatar row */}
          <div className="flex items-end justify-between -mt-12 mb-4">
            <div className="relative flex-shrink-0">
              <div className="w-24 h-24 rounded-full bg-background border-4 border-background overflow-hidden shadow-lg">
                {av ? (
                  <img src={av} alt="" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center">
                    <span className="text-3xl font-bold text-white">{profile.username[0]?.toUpperCase()}</span>
                  </div>
                )}
              </div>
              <button
                onClick={() => fileInputRef.current?.click()}
                disabled={avatarUploading}
                className="absolute bottom-1 right-1 bg-primary rounded-full p-1.5 text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm"
              >
                {avatarUploading ? <Loader2 className="h-3 w-3 animate-spin" /> : <Camera className="h-3 w-3" />}
              </button>
              <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={handleAvatarChange} />
            </div>

            {!isEditing && (
              <Button onClick={() => setIsEditing(true)} variant="outline" size="sm" className="flex items-center gap-1.5">
                <Edit2 className="h-3.5 w-3.5" />
                编辑资料
              </Button>
            )}
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
            <p className="text-muted-foreground text-sm">@{profile.username}</p>
            {profile.species && (
              <p className="text-sm text-primary mt-0.5">🐾 {profile.species}</p>
            )}
          </div>

          {/* Bio */}
          {profile.bio && !isEditing && (
            <p className="text-sm text-muted-foreground leading-relaxed mb-3 whitespace-pre-wrap">{profile.bio}</p>
          )}

          {/* Links */}
          {!isEditing && (
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
          )}

          {/* Stats row */}
          <div className="flex gap-6 text-sm border-t pt-4">
            <Link href={`/users/${profile.id}/followers`} className="hover:text-primary transition-colors">
              <span className="font-bold">{stats?.follower_count ?? 0}</span>
              <span className="text-muted-foreground ml-1">粉丝</span>
            </Link>
            <Link href={`/users/${profile.id}/following`} className="hover:text-primary transition-colors">
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

          {/* Edit form */}
          {isEditing && (
            <form onSubmit={handleSave} className="space-y-4 mt-4 border-t pt-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="furry_name">Furry 名</Label>
                  <Input
                    id="furry_name"
                    value={formData.furry_name}
                    onChange={e => setFormData({ ...formData, furry_name: e.target.value })}
                    placeholder="你的Furry角色名..."
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="species">兽种</Label>
                  <Input
                    id="species"
                    value={formData.species}
                    onChange={e => setFormData({ ...formData, species: e.target.value })}
                    placeholder="狼、狐、龙..."
                    className="mt-1"
                  />
                </div>
              </div>
              <div>
                <Label htmlFor="bio">个人简介</Label>
                <Textarea
                  id="bio"
                  value={formData.bio}
                  onChange={e => setFormData({ ...formData, bio: e.target.value })}
                  placeholder="介绍一下你自己和你的兽设..."
                  rows={3}
                  className="mt-1"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="website">个人网站</Label>
                  <Input
                    id="website"
                    value={formData.website}
                    onChange={e => setFormData({ ...formData, website: e.target.value })}
                    placeholder="https://..."
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="location">位置</Label>
                  <Input
                    id="location"
                    value={formData.location}
                    onChange={e => setFormData({ ...formData, location: e.target.value })}
                    placeholder="城市/地区"
                    className="mt-1"
                  />
                </div>
              </div>
              {error && <p className="text-destructive text-sm">{error}</p>}
              <div className="flex gap-2">
                <Button type="submit" disabled={saving}>{saving ? '保存中...' : '保存'}</Button>
                <Button type="button" variant="outline" onClick={() => setIsEditing(false)}>取消</Button>
              </div>
            </form>
          )}
        </div>
      </div>

      {/* Posts gallery */}
      <div className="flex items-center gap-2 mb-5">
        <Grid3X3 className="h-4 w-4 text-muted-foreground" />
        <h2 className="font-semibold">我的帖子</h2>
      </div>

      {posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Grid3X3 className="h-8 w-8 opacity-40" />
          </div>
          <p className="font-medium mb-2">还没有帖子</p>
          <p className="text-sm mb-6">分享你的兽设和创作吧</p>
          <Link href="/posts/create">
            <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
              发布第一条动态
            </Button>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {posts.map(post => (
            <PostGalleryCard key={post.id} post={post} />
          ))}
        </div>
      )}
    </div>
  )
}
