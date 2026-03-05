'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'

interface User {
  id: string
  username: string
  email: string
  nickname: string
  role: string
  status: string
  created_at: string
}

interface ListUsersOutput {
  users: User[]
  total: number
  page: number
  size: number
}

export default function AdminUsersPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery<ListUsersOutput>({
    queryKey: ['admin-users', search],
    queryFn: () => {
      const params = new URLSearchParams()
      if (search) params.append('search', search)
      return apiClient.get(`/admin/users?${params.toString()}`)
    },
  })

  const users = data?.users ?? []

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      apiClient.put(`/admin/users/${userId}/role`, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      alert('角色更新成功！')
    },
    onError: () => alert('更新失败，请重试'),
  })

  const banMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/ban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      alert('封禁成功！')
    },
    onError: () => alert('操作失败，请重试'),
  })

  const unbanMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/unban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      alert('解封成功！')
    },
    onError: () => alert('操作失败，请重试'),
  })

  const handleRoleChange = (userId: string, currentRole: string) => {
    const newRole = prompt('输入新角色 (user/admin/moderator):', currentRole)
    if (newRole && ['user', 'admin', 'moderator'].includes(newRole)) {
      updateRoleMutation.mutate({ userId, role: newRole })
    }
  }

  const handleBan = (userId: string, username: string) => {
    if (confirm(`确定要封禁用户 ${username} 吗？`)) {
      banMutation.mutate(userId)
    }
  }

  const handleUnban = (userId: string, username: string) => {
    if (confirm(`确定要解封用户 ${username} 吗？`)) {
      unbanMutation.mutate(userId)
    }
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <Link href="/admin">
          <Button variant="outline" className="mb-4">← 返回后台首页</Button>
        </Link>
        <h1 className="text-3xl font-bold mb-4">用户管理</h1>
        <div className="flex items-center gap-4">
          <div className="max-w-md flex-1">
            <Input
              placeholder="搜索用户名或邮箱..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          {data && (
            <span className="text-sm text-gray-500">共 {data.total} 个用户</span>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>用户列表</CardTitle>
        </CardHeader>
        <CardContent>
          {users.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {search ? '未找到匹配的用户' : '暂无用户'}
            </div>
          ) : (
            <div className="space-y-4">
              {users.map((user) => (
                <div key={user.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="font-bold">{user.nickname || user.username}</h3>
                      <Badge className={
                        user.role === 'admin' ? 'bg-red-500' :
                        user.role === 'moderator' ? 'bg-blue-500' : 'bg-gray-500'
                      }>
                        {user.role}
                      </Badge>
                      {user.status === 'banned' && (
                        <Badge className="bg-red-600">已封禁</Badge>
                      )}
                    </div>
                    <div className="flex gap-6 text-sm text-gray-600">
                      <span>@{user.username}</span>
                      <span>{user.email}</span>
                      <span>注册: {new Date(user.created_at).toLocaleDateString('zh-CN')}</span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRoleChange(user.id, user.role)}
                    >
                      修改角色
                    </Button>
                    {user.status === 'banned' ? (
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-green-600 hover:text-green-700"
                        onClick={() => handleUnban(user.id, user.username)}
                      >
                        解封
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-red-600 hover:text-red-700"
                        onClick={() => handleBan(user.id, user.username)}
                      >
                        封禁
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
