'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Loader2 } from 'lucide-react'

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

const ROLES = [
  { value: 'member', label: '普通用户', color: 'bg-gray-500' },
  { value: 'creator', label: '创作者', color: 'bg-purple-500' },
  { value: 'moderator', label: '版主', color: 'bg-blue-500' },
  { value: 'admin', label: '管理员', color: 'bg-red-500' },
]

function roleBadgeClass(role: string) {
  return ROLES.find(r => r.value === role)?.color ?? 'bg-gray-400'
}

function toast(msg: string) {
  const el = document.createElement('div')
  el.className = 'fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm z-50'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

export default function AdminUsersPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [roleDialogUser, setRoleDialogUser] = useState<User | null>(null)
  const [selectedRole, setSelectedRole] = useState('')
  const [banConfirm, setBanConfirm] = useState<string | null>(null)

  const { data, isLoading } = useQuery<ListUsersOutput>({
    queryKey: ['admin-users', search],
    queryFn: () => {
      const params = new URLSearchParams()
      if (search) params.append('search', search)
      return apiClient.get(`/admin/users?${params.toString()}`)
    },
  })

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      apiClient.put(`/admin/users/${userId}/role`, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setRoleDialogUser(null)
      toast('角色更新成功')
    },
    onError: () => toast('更新失败，请重试'),
  })

  const banMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/ban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setBanConfirm(null)
      toast('封禁成功')
    },
    onError: () => toast('操作失败，请重试'),
  })

  const unbanMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/unban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setBanConfirm(null)
      toast('解封成功')
    },
    onError: () => toast('操作失败，请重试'),
  })

  const users = data?.users ?? []

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h1>
        {data && <span className="text-sm text-gray-400">共 {data.total} 个用户</span>}
      </div>

      <div className="max-w-sm">
        <Input
          placeholder="搜索用户名或邮箱..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">用户列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="animate-spin text-gray-400" />
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-10 text-gray-400">
              {search ? '未找到匹配的用户' : '暂无用户'}
            </div>
          ) : (
            <div className="divide-y">
              {users.map((user) => (
                <div key={user.id} className="flex items-center justify-between py-3 gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <span className="font-medium text-sm">{user.nickname || user.username}</span>
                      <Badge className={`${roleBadgeClass(user.role)} text-white text-xs`}>
                        {ROLES.find(r => r.value === user.role)?.label ?? user.role}
                      </Badge>
                      {user.status === 'banned' && (
                        <Badge className="bg-red-600 text-white text-xs">已封禁</Badge>
                      )}
                    </div>
                    <div className="flex gap-4 text-xs text-gray-400 flex-wrap">
                      <span>@{user.username}</span>
                      <span>{user.email}</span>
                      <span>注册: {new Date(user.created_at).toLocaleDateString('zh-CN')}</span>
                    </div>
                  </div>
                  <div className="flex gap-2 shrink-0">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setSelectedRole(user.role)
                        setRoleDialogUser(user)
                      }}
                    >
                      修改角色
                    </Button>
                    {user.status === 'banned' ? (
                      banConfirm === user.id ? (
                        <Button
                          size="sm"
                          className="bg-green-600 hover:bg-green-700 text-white"
                          onClick={() => unbanMutation.mutate(user.id)}
                          disabled={unbanMutation.isPending}
                        >
                          {unbanMutation.isPending ? <Loader2 size={12} className="animate-spin" /> : '确认解封'}
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-green-600 hover:text-green-700"
                          onClick={() => setBanConfirm(user.id)}
                        >
                          解封
                        </Button>
                      )
                    ) : (
                      banConfirm === user.id ? (
                        <Button
                          size="sm"
                          className="bg-red-600 hover:bg-red-700 text-white"
                          onClick={() => banMutation.mutate(user.id)}
                          disabled={banMutation.isPending}
                        >
                          {banMutation.isPending ? <Loader2 size={12} className="animate-spin" /> : '确认封禁'}
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-red-600 hover:text-red-700"
                          onClick={() => setBanConfirm(user.id)}
                        >
                          封禁
                        </Button>
                      )
                    )}
                    {banConfirm === user.id && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-gray-400"
                        onClick={() => setBanConfirm(null)}
                      >
                        取消
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Role dialog */}
      <Dialog open={!!roleDialogUser} onOpenChange={(open) => !open && setRoleDialogUser(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>修改角色 — @{roleDialogUser?.username}</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 py-2">
            {ROLES.map(({ value, label, color }) => (
              <label
                key={value}
                className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                  selectedRole === value
                    ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20'
                    : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
              >
                <input
                  type="radio"
                  name="role"
                  value={value}
                  checked={selectedRole === value}
                  onChange={() => setSelectedRole(value)}
                  className="sr-only"
                />
                <Badge className={`${color} text-white text-xs`}>{label}</Badge>
                <span className="text-sm text-gray-600 dark:text-gray-300">{value}</span>
              </label>
            ))}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setRoleDialogUser(null)}>取消</Button>
            <Button
              onClick={() => {
                if (roleDialogUser && selectedRole) {
                  updateRoleMutation.mutate({ userId: roleDialogUser.id, role: selectedRole })
                }
              }}
              disabled={updateRoleMutation.isPending || !selectedRole}
            >
              {updateRoleMutation.isPending ? <Loader2 size={14} className="animate-spin mr-1" /> : null}
              确认修改
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
