'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Globe2, Lock, Search, Shapes, Users } from 'lucide-react'
import { AdminGroup, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminFilterBar } from '@/components/admin/admin-filter-bar'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminPagination } from '@/components/admin/admin-pagination'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

const privacyOptions = [
  { value: '', label: '全部可见性' },
  { value: 'public', label: '公开' },
  { value: 'private', label: '私密' },
]

export default function AdminGroupsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [privacy, setPrivacy] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-groups', page, search, privacy],
    queryFn: () =>
      apiClient.getAdminGroups({
        page,
        page_size: 20,
        search: search || undefined,
        privacy: privacy || undefined,
      }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, nextPrivacy }: { id: string; nextPrivacy: 'public' | 'private' }) =>
      apiClient.updateAdminGroup(id, { privacy: nextPrivacy }),
    onSuccess: (_, { nextPrivacy }) => {
      queryClient.invalidateQueries({ queryKey: ['admin-groups'] })
      showAdminToast(`圈子已调整为${nextPrivacy === 'public' ? '公开' : '私密'}`, 'success')
    },
    onError: () => {
      showAdminToast('更新圈子可见性失败', 'error')
    },
  })

  const groups = data?.groups ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1

  const stats = {
    publicCount: groups.filter((item) => item.privacy === 'public').length,
    privateCount: groups.filter((item) => item.privacy === 'private').length,
  }

  const columns: AdminColumn<AdminGroup>[] = [
    {
      key: 'group',
      header: '圈子',
      className: 'min-w-[300px]',
      render: (group) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={group.privacy} label={group.privacy === 'public' ? '公开' : '私密'} />
            <span className="text-xs text-slate-400">
              圈主：{group.owner_username ? `@${group.owner_username}` : group.owner_id}
            </span>
          </div>
          <div>
            <p className="font-medium text-slate-900">{group.name}</p>
            <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">
              {group.description || '暂无圈子简介'}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'stats',
      header: '规模',
      className: 'whitespace-nowrap text-slate-500',
      render: (group) => (
        <div className="space-y-1">
          <p>成员 {group.member_count}</p>
          <p>帖子 {group.post_count}</p>
          <p>{new Date(group.created_at).toLocaleDateString('zh-CN')}</p>
        </div>
      ),
    },
    {
      key: 'tags',
      header: '标签',
      className: 'min-w-[200px]',
      render: (group) =>
        group.tags?.length ? (
          <div className="flex flex-wrap gap-2">
            {group.tags.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600"
              >
                {tag}
              </span>
            ))}
          </div>
        ) : (
          <span className="text-sm text-slate-400">无标签</span>
        ),
    },
    {
      key: 'actions',
      header: '操作',
      className: 'w-[240px]',
      render: (group) => (
        <div className="flex flex-wrap gap-2">
          <Button asChild variant="outline" size="sm">
            <Link href={`/groups/${group.id}`} target="_blank">
              查看圈子
            </Link>
          </Button>
          {group.privacy === 'public' ? (
            <Button
              size="sm"
              variant="outline"
              className="border-slate-200 text-slate-700 hover:bg-slate-50"
              disabled={updateMutation.isPending}
              onClick={() => updateMutation.mutate({ id: group.id, nextPrivacy: 'private' })}
            >
              <Lock className="mr-1 h-4 w-4" />
              设为私密
            </Button>
          ) : (
            <Button
              size="sm"
              variant="outline"
              className="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
              disabled={updateMutation.isPending}
              onClick={() => updateMutation.mutate({ id: group.id, nextPrivacy: 'public' })}
            >
              <Globe2 className="mr-1 h-4 w-4" />
              设为公开
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Community Operations"
        title="圈子运营"
        description="用于巡检圈子规模、查看圈主信息，并在需要时调整圈子的公开或私密状态。"
      />

      <div className="grid gap-4 md:grid-cols-4">
        <AdminMetricCard
          label="圈子结果数"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前筛选条件下的圈子总数"
          icon={Shapes}
          tone="brand"
        />
        <AdminMetricCard
          label="本页公开圈子"
          value={stats.publicCount.toLocaleString()}
          hint="可被普通用户直接发现"
          icon={Globe2}
          tone="success"
        />
        <AdminMetricCard
          label="本页私密圈子"
          value={stats.privateCount.toLocaleString()}
          hint="仅成员可见"
          icon={Lock}
          tone={stats.privateCount > 0 ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="当前页记录"
          value={groups.length.toLocaleString()}
          hint="单页最多 20 条"
          icon={Users}
          tone="default"
        />
      </div>

      <AdminFilterBar>
        <div className="flex flex-1 flex-col gap-3 md:flex-row md:items-center">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <Input
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  setPage(1)
                  setSearch(searchInput.trim())
                }
              }}
              placeholder="搜索圈子名称"
              className="pl-10"
            />
          </div>
          <select
            value={privacy}
            onChange={(e) => {
              setPage(1)
              setPrivacy(e.target.value)
            }}
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          >
            {privacyOptions.map((option) => (
              <option key={option.value || 'all'} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center gap-2">
          <Button
            className="bg-slate-950 text-white hover:bg-slate-800"
            onClick={() => {
              setPage(1)
              setSearch(searchInput.trim())
            }}
          >
            搜索
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              setPage(1)
              setSearchInput('')
              setSearch('')
              setPrivacy('')
            }}
          >
            重置
          </Button>
        </div>
      </AdminFilterBar>

      <AdminDataTable
        data={groups}
        columns={columns}
        keyExtractor={(group) => group.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有匹配的圈子"
            description="调整搜索词或可见性筛选后再试一次。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
