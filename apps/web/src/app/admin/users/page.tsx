'use client'

import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Loader2, RotateCcw, Search, Shield, UserCog, UserX } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminFilterBar } from '@/components/admin/admin-filter-bar'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

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

const roles = [
  { value: 'member', label: '普通用户' },
  { value: 'creator', label: '创作者' },
  { value: 'moderator', label: '版主' },
  { value: 'admin', label: '管理员' },
]

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'active', label: '正常' },
  { value: 'banned', label: '已封禁' },
]

function getRoleLabel(role: string) {
  return roles.find((item) => item.value === role)?.label ?? role
}

export default function AdminUsersPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [keywordInput, setKeywordInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [roleDialogUser, setRoleDialogUser] = useState<User | null>(null)
  const [selectedRole, setSelectedRole] = useState('')
  const [confirmAction, setConfirmAction] = useState<{ userId: string; type: 'ban' | 'unban' } | null>(null)

  const { data, isLoading } = useQuery<ListUsersOutput>({
    queryKey: ['admin-users', keyword, status, page],
    queryFn: () => {
      const params = new URLSearchParams({
        page: String(page),
        page_size: '20',
      })
      if (keyword) params.set('search', keyword)
      if (status) params.set('status', status)
      return apiClient.get(`/admin/users?${params.toString()}`)
    },
  })

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      apiClient.put(`/admin/users/${userId}/role`, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setRoleDialogUser(null)
      showAdminToast('角色更新成功', 'success')
    },
    onError: () => {
      showAdminToast('更新失败，请重试', 'error')
    },
  })

  const banMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/ban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setConfirmAction(null)
      showAdminToast('用户已封禁', 'success')
    },
    onError: () => {
      showAdminToast('操作失败，请重试', 'error')
    },
  })

  const unbanMutation = useMutation({
    mutationFn: (userId: string) => apiClient.post(`/admin/users/${userId}/unban`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setConfirmAction(null)
      showAdminToast('用户已解封', 'success')
    },
    onError: () => {
      showAdminToast('操作失败，请重试', 'error')
    },
  })

  const users = data?.users ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1
  const bannedOnPage = users.filter((user) => user.status === 'banned').length

  const columns: AdminColumn<User>[] = [
    {
      key: 'user',
      header: '用户',
      className: 'min-w-[240px]',
      render: (user) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-medium text-slate-900">
              {user.nickname || user.username}
            </p>
            <AdminStatusBadge value={user.role} label={getRoleLabel(user.role)} />
            <AdminStatusBadge value={user.status} label={user.status === 'banned' ? '已封禁' : '正常'} />
          </div>
          <div className="space-y-1 text-sm text-slate-500">
            <p>@{user.username}</p>
            <p>{user.email}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'created_at',
      header: '注册时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (user) => new Date(user.created_at).toLocaleString('zh-CN'),
    },
    {
      key: 'actions',
      header: '操作',
      className: 'w-[260px]',
      render: (user) => {
        const isConfirmingBan = confirmAction?.userId === user.id && confirmAction.type === 'ban'
        const isConfirmingUnban = confirmAction?.userId === user.id && confirmAction.type === 'unban'

        return (
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setSelectedRole(user.role)
                setRoleDialogUser(user)
              }}
            >
              <Shield className="mr-1 h-4 w-4" />
              修改角色
            </Button>

            {user.status === 'banned' ? (
              <Button
                size="sm"
                variant={isConfirmingUnban ? 'default' : 'outline'}
                className={isConfirmingUnban ? 'bg-emerald-600 text-white hover:bg-emerald-500' : 'border-emerald-200 text-emerald-700 hover:bg-emerald-50'}
                disabled={unbanMutation.isPending}
                onClick={() => {
                  if (isConfirmingUnban) {
                    unbanMutation.mutate(user.id)
                  } else {
                    setConfirmAction({ userId: user.id, type: 'unban' })
                  }
                }}
              >
                {unbanMutation.isPending && isConfirmingUnban ? (
                  <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                ) : (
                  <RotateCcw className="mr-1 h-4 w-4" />
                )}
                {isConfirmingUnban ? '再次确认解封' : '解封'}
              </Button>
            ) : (
              <Button
                size="sm"
                variant={isConfirmingBan ? 'default' : 'outline'}
                className={isConfirmingBan ? 'bg-rose-600 text-white hover:bg-rose-500' : 'border-rose-200 text-rose-700 hover:bg-rose-50'}
                disabled={banMutation.isPending}
                onClick={() => {
                  if (isConfirmingBan) {
                    banMutation.mutate(user.id)
                  } else {
                    setConfirmAction({ userId: user.id, type: 'ban' })
                  }
                }}
              >
                {banMutation.isPending && isConfirmingBan ? (
                  <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                ) : (
                  <UserX className="mr-1 h-4 w-4" />
                )}
                {isConfirmingBan ? '再次确认封禁' : '封禁'}
              </Button>
            )}
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="User Operations"
        title="用户管理"
        description="统一处理用户搜索、角色调整和封禁状态，是最常用的运营治理工作面。"
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="筛选结果"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前搜索条件下的用户总数"
          icon={UserCog}
          tone="brand"
        />
        <AdminMetricCard
          label="本页封禁用户"
          value={bannedOnPage.toLocaleString()}
          hint="用于快速发现异常账户"
          icon={UserX}
          tone={bannedOnPage > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="当前页容量"
          value={users.length.toLocaleString()}
          hint="单页最多展示 20 个用户"
          icon={Shield}
          tone="default"
        />
      </div>

      <AdminFilterBar>
        <div className="flex flex-1 flex-col gap-3 md:flex-row md:items-center">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <Input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  setPage(1)
                  setKeyword(keywordInput.trim())
                }
              }}
              placeholder="搜索用户名或邮箱"
              className="pl-10"
            />
          </div>
          <select
            value={status}
            onChange={(e) => {
              setPage(1)
              setStatus(e.target.value)
            }}
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          >
            {statusOptions.map((option) => (
              <option key={option.value || 'all'} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => {
              setPage(1)
              setKeyword(keywordInput.trim())
            }}
            className="bg-slate-950 text-white hover:bg-slate-800"
          >
            搜索
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              setPage(1)
              setKeywordInput('')
              setKeyword('')
              setStatus('')
            }}
          >
            重置
          </Button>
        </div>
      </AdminFilterBar>

      <AdminDataTable
        data={users}
        columns={columns}
        keyExtractor={(user) => user.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有匹配的用户"
            description="调整搜索关键词或状态筛选后再试一次。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />

      <Dialog open={!!roleDialogUser} onOpenChange={(open) => !open && setRoleDialogUser(null)}>
        <DialogContent className="max-w-md rounded-3xl">
          <DialogHeader>
            <DialogTitle>修改角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            {roles.map((role) => (
              <label
                key={role.value}
                className={`flex cursor-pointer items-center justify-between rounded-2xl border p-4 transition-colors ${
                  selectedRole === role.value
                    ? 'border-slate-950 bg-slate-950 text-white'
                    : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                }`}
              >
                <div>
                  <p className="font-medium">{role.label}</p>
                  <p className={`text-xs ${selectedRole === role.value ? 'text-slate-300' : 'text-slate-400'}`}>
                    {role.value}
                  </p>
                </div>
                <input
                  type="radio"
                  name="role"
                  checked={selectedRole === role.value}
                  onChange={() => setSelectedRole(role.value)}
                  className="h-4 w-4"
                />
              </label>
            ))}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setRoleDialogUser(null)}>
              取消
            </Button>
            <Button
              disabled={updateRoleMutation.isPending || !roleDialogUser || !selectedRole}
              onClick={() => {
                if (!roleDialogUser) return
                updateRoleMutation.mutate({ userId: roleDialogUser.id, role: selectedRole })
              }}
              className="bg-slate-950 text-white hover:bg-slate-800"
            >
              {updateRoleMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              保存角色
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
