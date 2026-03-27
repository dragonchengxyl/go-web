'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { CalendarRange, ShieldAlert, Shapes, UserX, Users } from 'lucide-react'
import { AdminEvent, AdminGroup, AdminUser, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import {
  AdminSectionCard,
  AdminWorkspaceCard,
  formatDateTime,
} from '@/components/admin/admin-dashboard-kit'

const roleLabelMap: Record<string, string> = {
  member: '普通用户',
  creator: '创作者',
  moderator: '版主',
  admin: '管理员',
  super_admin: '超级管理员',
}

const statusLabelMap: Record<string, string> = {
  active: '正常',
  banned: '已封禁',
  draft: '草稿',
  published: '已发布',
  cancelled: '已取消',
  completed: '已完成',
}

function SnapshotItem({
  title,
  subtitle,
  meta,
  badge,
}: {
  title: string
  subtitle?: string
  meta: string
  badge?: string
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="font-medium text-slate-950">{title}</p>
          {subtitle ? (
            <p className="mt-1 text-sm leading-6 text-slate-500">{subtitle}</p>
          ) : null}
        </div>
        {badge ? (
          <Badge
            variant="outline"
            className="shrink-0 rounded-full border-slate-200 bg-white text-slate-600"
          >
            {badge}
          </Badge>
        ) : null}
      </div>
      <p className="mt-3 text-xs text-slate-400">{meta}</p>
    </div>
  )
}

export default function AdminCommunityPage() {
  const { data: usersData } = useQuery({
    queryKey: ['admin-dashboard-community-users'],
    queryFn: () => apiClient.getAdminUsers({ page: 1, page_size: 6 }),
  })

  const { data: bannedUsersData } = useQuery({
    queryKey: ['admin-dashboard-community-banned-users'],
    queryFn: () =>
      apiClient.getAdminUsers({ page: 1, page_size: 3, status: 'banned' }),
  })

  const { data: privateGroupsData } = useQuery({
    queryKey: ['admin-dashboard-community-private-groups'],
    queryFn: () =>
      apiClient.getAdminGroups({ page: 1, page_size: 4, privacy: 'private' }),
  })

  const { data: groupsData } = useQuery({
    queryKey: ['admin-dashboard-community-groups'],
    queryFn: () => apiClient.getAdminGroups({ page: 1, page_size: 5 }),
  })

  const { data: publishedEventsData } = useQuery({
    queryKey: ['admin-dashboard-community-published-events'],
    queryFn: () =>
      apiClient.getAdminEvents({ page: 1, page_size: 4, status: 'published' }),
  })

  const { data: eventsData } = useQuery({
    queryKey: ['admin-dashboard-community-events'],
    queryFn: () => apiClient.getAdminEvents({ page: 1, page_size: 5 }),
  })

  const users = (usersData?.users ?? []) as AdminUser[]
  const bannedUsers = (bannedUsersData?.users ?? []) as AdminUser[]
  const groups = (groupsData?.groups ?? []) as AdminGroup[]
  const events = (eventsData?.events ?? []) as AdminEvent[]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Community Dashboard"
        title="社区运营盘"
        description="把用户、圈子和活动从单独工作台提升成一个运营视角，先看社区秩序和活跃面，再进入明细页操作。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/events">查看活动工作台</Link>
            </Button>
            <Button asChild className="border border-[#c5dfd3] bg-[#edf8f2] text-[#21584e] hover:bg-[#e3f3eb]">
              <Link href="/admin/users">进入用户管理</Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminMetricCard
          label="用户规模"
          value={(usersData?.total ?? 0).toLocaleString()}
          hint="当前筛选接口返回的累计用户数"
          icon={Users}
          tone="brand"
        />
        <AdminMetricCard
          label="封禁用户"
          value={(bannedUsersData?.total ?? 0).toLocaleString()}
          hint="需要持续巡检的高风险账号"
          icon={UserX}
          tone={(bannedUsersData?.total ?? 0) > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="私密圈子"
          value={(privateGroupsData?.total ?? 0).toLocaleString()}
          hint="适合重点巡检规则与成员状态"
          icon={Shapes}
          tone="default"
        />
        <AdminMetricCard
          label="已发布活动"
          value={(publishedEventsData?.total ?? 0).toLocaleString()}
          hint="当前在运营生命周期中的活动数"
          icon={CalendarRange}
          tone="success"
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-3">
        <AdminWorkspaceCard
          href="/admin/users"
          title="用户管理"
          description="处理搜索、角色调整和封禁状态，是社区秩序的第一工作面。"
          icon={Users}
          tone="brand"
          metrics={[
            { label: '总量', value: (usersData?.total ?? 0).toLocaleString() },
            { label: '封禁', value: (bannedUsersData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/groups"
          title="圈子运营"
          description="查看圈子规模、所有者和公开/私密状态，必要时做隐私调整。"
          icon={Shapes}
          tone="default"
          metrics={[
            { label: '总量', value: (groupsData?.total ?? 0).toLocaleString() },
            { label: '私密', value: (privateGroupsData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/events"
          title="活动运营"
          description="跟踪活动生命周期、报名压力和已发布活动的运营状态。"
          icon={CalendarRange}
          tone="success"
          metrics={[
            { label: '总量', value: (eventsData?.total ?? 0).toLocaleString() },
            { label: '已发布', value: (publishedEventsData?.total ?? 0).toLocaleString() },
          ]}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSectionCard
          title="近期注册与重点账号"
          description="值班时优先看新用户和已经被封禁的账号，确认是否存在异常集中的风险人群。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/users">查看全部用户</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">最新用户</p>
              {users.slice(0, 3).map((user) => (
                <SnapshotItem
                  key={user.id}
                  title={user.nickname || user.username}
                  subtitle={`${user.email} · @${user.username}`}
                  badge={roleLabelMap[user.role] ?? user.role}
                  meta={`${statusLabelMap[user.status] ?? user.status} · ${formatDateTime(user.created_at)}`}
                />
              ))}
              {users.length === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有用户数据。
                </p>
              ) : null}
            </div>

            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">重点账号</p>
              {bannedUsers.slice(0, 3).map((user) => (
                <SnapshotItem
                  key={user.id}
                  title={user.nickname || user.username}
                  subtitle={user.email}
                  badge="已封禁"
                  meta={`${roleLabelMap[user.role] ?? user.role} · ${formatDateTime(user.created_at)}`}
                />
              ))}
              {bannedUsers.length === 0 ? (
                <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-6 text-sm text-emerald-700">
                  当前没有封禁用户，社区秩序相对稳定。
                </div>
              ) : null}
            </div>
          </div>
        </AdminSectionCard>

        <AdminSectionCard
          title="圈子与活动快照"
          description="把圈子密度和活动生命周期放在一起看，快速判断社区供给是否健康。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/groups">进入圈子运营</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">圈子快照</p>
              {groups.slice(0, 3).map((group) => (
                <SnapshotItem
                  key={group.id}
                  title={group.name}
                  subtitle={`圈主：${group.owner_username || group.owner_id}`}
                  badge={group.privacy === 'private' ? '私密' : '公开'}
                  meta={`${group.member_count} 成员 · ${group.post_count} 帖子`}
                />
              ))}
              {groups.length === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有圈子数据。
                </p>
              ) : null}
            </div>

            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">活动快照</p>
              {events.slice(0, 3).map((event) => (
                <SnapshotItem
                  key={event.id}
                  title={event.title}
                  subtitle={`组织者：${event.organizer_username || event.organizer_id}`}
                  badge={statusLabelMap[event.status] ?? event.status}
                  meta={`${event.attendee_count} 报名 · ${formatDateTime(event.start_time)}`}
                />
              ))}
              {events.length === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有活动数据。
                </p>
              ) : null}
            </div>
          </div>
        </AdminSectionCard>
      </div>

      <AdminSectionCard
        title="社区秩序提醒"
        description="把需要值班注意的风险点浓缩成一屏摘要，避免只看总量而忽略结构变化。"
      >
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <div className="flex items-center gap-2 text-slate-900">
              <ShieldAlert className="h-4 w-4" />
              <p className="font-medium">账号风险</p>
            </div>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              当前封禁账号 {(bannedUsersData?.total ?? 0).toLocaleString()} 个。
              如果短时间内持续上涨，优先回到用户管理和举报工作台核查原因。
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <div className="flex items-center gap-2 text-slate-900">
              <Shapes className="h-4 w-4" />
              <p className="font-medium">圈子结构</p>
            </div>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              私密圈子 {(privateGroupsData?.total ?? 0).toLocaleString()} 个，
              建议定期巡检规则说明、圈主信息和成员活跃度。
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <div className="flex items-center gap-2 text-slate-900">
              <CalendarRange className="h-4 w-4" />
              <p className="font-medium">活动供给</p>
            </div>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              已发布活动 {(publishedEventsData?.total ?? 0).toLocaleString()} 个。
              如果数量持续偏低，建议结合增长盘补活动供给。
            </p>
          </div>
        </div>
      </AdminSectionCard>
    </div>
  )
}
