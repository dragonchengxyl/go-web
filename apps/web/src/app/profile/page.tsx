'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api-client'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { LogOut } from 'lucide-react'

interface UserProfile {
  id: number
  username: string
  email: string
  nickname: string
  avatar: string
  bio: string
  website: string
  created_at: string
}

export default function ProfilePage() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState({
    nickname: '',
    bio: '',
    website: '',
    avatar: '',
  })

  const { data: profile, isLoading } = useQuery<UserProfile>({
    queryKey: ['profile'],
    queryFn: async () => {
      const response = await apiClient.get<UserProfile>('/users/me')
      return response
    },
  })

  useEffect(() => {
    if (profile) {
      setFormData({
        nickname: profile.nickname || '',
        bio: profile.bio || '',
        website: profile.website || '',
        avatar: profile.avatar || '',
      })
    }
  }, [profile])

  const updateMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      const response = await apiClient.put('/users/me', data)
      return response
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profile'] })
      setIsEditing(false)
      alert('更新成功！')
    },
    onError: () => {
      alert('更新失败，请重试')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate(formData)
  }

  const handleLogout = () => {
    apiClient.setToken(null)
    router.push('/login')
  }

  if (isLoading) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-8">
            <div className="text-center">加载中...</div>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  if (!profile) {
    router.push('/login')
    return null
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <div className="flex justify-between items-center mb-8">
            <h1 className="text-3xl font-bold">个人中心</h1>
            <Button variant="outline" onClick={handleLogout}>
              <LogOut className="h-4 w-4 mr-2" />
              退出登录
            </Button>
          </div>

          <Tabs defaultValue="profile" className="space-y-6">
            <TabsList>
              <TabsTrigger value="profile">个人资料</TabsTrigger>
              <TabsTrigger value="security">账号安全</TabsTrigger>
            </TabsList>

            <TabsContent value="profile">
              <Card>
                <CardHeader>
                  <div className="flex justify-between items-center">
                    <CardTitle>个人资料</CardTitle>
                    {!isEditing && (
                      <Button onClick={() => setIsEditing(true)}>编辑</Button>
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  {!isEditing ? (
                    <div className="space-y-4">
                      <div className="flex items-center gap-4">
                        <div className="w-20 h-20 rounded-full bg-primary flex items-center justify-center text-primary-foreground text-2xl font-bold">
                          {profile.username?.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <h2 className="text-xl font-bold">{profile.nickname || profile.username}</h2>
                          <p className="text-gray-500">@{profile.username}</p>
                        </div>
                      </div>
                      <div>
                        <p className="text-sm text-gray-500">邮箱</p>
                        <p className="font-medium">{profile.email}</p>
                      </div>
                      {profile.bio && (
                        <div>
                          <p className="text-sm text-gray-500">个人简介</p>
                          <p className="font-medium">{profile.bio}</p>
                        </div>
                      )}
                      {profile.website && (
                        <div>
                          <p className="text-sm text-gray-500">个人网站</p>
                          <a
                            href={profile.website}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:underline"
                          >
                            {profile.website}
                          </a>
                        </div>
                      )}
                      <div>
                        <p className="text-sm text-gray-500">注册时间</p>
                        <p className="font-medium">{new Date(profile.created_at).toLocaleDateString('zh-CN')}</p>
                      </div>
                    </div>
                  ) : (
                    <form onSubmit={handleSubmit} className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium mb-2">昵称</label>
                        <Input
                          value={formData.nickname}
                          onChange={(e) => setFormData({ ...formData, nickname: e.target.value })}
                          placeholder="输入昵称"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium mb-2">个人简介</label>
                        <textarea
                          value={formData.bio}
                          onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
                          placeholder="介绍一下自己..."
                          className="w-full px-3 py-2 border rounded-md min-h-[100px]"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium mb-2">个人网站</label>
                        <Input
                          value={formData.website}
                          onChange={(e) => setFormData({ ...formData, website: e.target.value })}
                          placeholder="https://example.com"
                        />
                      </div>
                      <div className="flex gap-2">
                        <Button type="submit" disabled={updateMutation.isPending}>
                          {updateMutation.isPending ? '保存中...' : '保存'}
                        </Button>
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => setIsEditing(false)}
                        >
                          取消
                        </Button>
                      </div>
                    </form>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="security">
              <Card>
                <CardHeader>
                  <CardTitle>账号安全</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex justify-between items-center py-3 border-b">
                    <div>
                      <p className="font-medium">登录密码</p>
                      <p className="text-sm text-gray-500">定期更换密码可以提高账号安全性</p>
                    </div>
                    <Button variant="outline" disabled>修改密码</Button>
                  </div>
                  <div className="flex justify-between items-center py-3 border-b">
                    <div>
                      <p className="font-medium">邮箱验证</p>
                      <p className="text-sm text-gray-500">{profile.email}</p>
                    </div>
                    <Button variant="outline" disabled>更换邮箱</Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </main>
      <Footer />
    </div>
  )
}
