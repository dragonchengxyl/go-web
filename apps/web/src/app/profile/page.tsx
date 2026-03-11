'use client'

import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Camera, Loader2 } from 'lucide-react'

interface UserProfile {
  id: string
  username: string
  email: string
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

export default function ProfilePage() {
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [profile, setProfile] = useState<UserProfile | null>(null)
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
    apiClient.getMe().then((u: UserProfile) => {
      setProfile(u)
      setFormData({
        bio: u.bio || '',
        website: u.website || '',
        location: u.location || '',
        furry_name: u.furry_name || '',
        species: u.species || '',
      })
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
    return <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">加载中...</div>
  }
  if (!profile) return null

  const av = avatarUrl(profile.avatar_key)
  const displayName = profile.furry_name || profile.username

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">个人中心</h1>
      </div>

      {/* Avatar + basic info */}
      <div className="flex items-start gap-5 mb-8">
        <div className="relative flex-shrink-0">
          <div className="w-20 h-20 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden">
            {av ? (
              <img src={av} alt="" className="w-full h-full object-cover" />
            ) : (
              <span className="text-2xl font-bold text-primary">{profile.username[0]?.toUpperCase()}</span>
            )}
          </div>
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={avatarUploading}
            className="absolute bottom-0 right-0 bg-primary rounded-full p-1.5 text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            {avatarUploading ? <Loader2 className="h-3 w-3 animate-spin" /> : <Camera className="h-3 w-3" />}
          </button>
          <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={handleAvatarChange} />
        </div>
        <div>
          <p className="text-xl font-bold">{displayName}</p>
          <p className="text-sm text-muted-foreground">@{profile.username}</p>
          {profile.species && <p className="text-sm text-muted-foreground mt-0.5">兽种: {profile.species}</p>}
          <p className="text-xs text-muted-foreground mt-1">{profile.email}</p>
        </div>
      </div>

      {/* Profile form */}
      {!isEditing ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            {profile.furry_name && (
              <div>
                <p className="text-xs text-muted-foreground">Furry 名</p>
                <p className="font-medium">{profile.furry_name}</p>
              </div>
            )}
            {profile.species && (
              <div>
                <p className="text-xs text-muted-foreground">兽种</p>
                <p className="font-medium">{profile.species}</p>
              </div>
            )}
          </div>
          {profile.bio && (
            <div>
              <p className="text-xs text-muted-foreground">简介</p>
              <p className="text-sm leading-relaxed whitespace-pre-wrap">{profile.bio}</p>
            </div>
          )}
          {profile.website && (
            <div>
              <p className="text-xs text-muted-foreground">网站</p>
              <a href={profile.website} target="_blank" rel="noopener noreferrer" className="text-sm text-primary hover:underline">{profile.website}</a>
            </div>
          )}
          {profile.location && (
            <div>
              <p className="text-xs text-muted-foreground">位置</p>
              <p className="text-sm">{profile.location}</p>
            </div>
          )}
          <Button onClick={() => setIsEditing(true)} className="mt-2">编辑资料</Button>
        </div>
      ) : (
        <form onSubmit={handleSave} className="space-y-4">
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
              rows={4}
              className="mt-1"
            />
          </div>
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
          {error && <p className="text-destructive text-sm">{error}</p>}
          <div className="flex gap-2">
            <Button type="submit" disabled={saving}>{saving ? '保存中...' : '保存'}</Button>
            <Button type="button" variant="outline" onClick={() => setIsEditing(false)}>取消</Button>
          </div>
        </form>
      )}
    </div>
  )
}
