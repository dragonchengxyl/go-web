'use client'

import { useQuery } from '@tanstack/react-query'
import { Shield } from 'lucide-react'
import { apiClient, PermissionMatrixRole } from '@/lib/api-client'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'

export default function AdminPermissionsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin-permission-matrix'],
    queryFn: () => apiClient.getAdminPermissionMatrix(),
  })

  const rows = data?.roles ?? []
  const catalog = data?.catalog ?? {}
  const catalogEntries = Object.entries(catalog)

  const columns: AdminColumn<PermissionMatrixRole>[] = [
    {
      key: 'role',
      header: '角色',
      className: 'min-w-[180px]',
      render: (row) => (
        <div className="space-y-2">
          <AdminStatusBadge value={row.role} label={row.role} />
          <p className="text-sm text-slate-500">权限数 {row.count}</p>
        </div>
      ),
    },
    {
      key: 'permissions',
      header: '权限列表',
      className: 'min-w-[420px]',
      render: (row) => (
        <div className="flex flex-wrap gap-2">
          {row.permissions.map((permission) => (
            <span
              key={permission}
              className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600"
            >
              {permission}
            </span>
          ))}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Access Control"
        title="权限矩阵"
        description="把角色与权限的映射关系直接展示出来，方便研发、运营和面试时说明系统的访问控制设计。"
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="角色数量"
          value={rows.length.toLocaleString()}
          hint="当前参与权限映射的角色"
          icon={Shield}
          tone="brand"
        />
        <AdminMetricCard
          label="权限分类"
          value={catalogEntries.length.toLocaleString()}
          hint="按域拆分的权限组"
          icon={Shield}
          tone="default"
        />
        <AdminMetricCard
          label="总权限项"
          value={Object.values(catalog).flat().length.toLocaleString()}
          hint="权限目录中的原子权限数量"
          icon={Shield}
          tone="default"
        />
      </div>

      <AdminDataTable
        data={rows}
        columns={columns}
        keyExtractor={(row) => row.role}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有权限矩阵数据"
            description="请检查后台权限接口是否正常返回。"
          />
        }
      />

      <section className="space-y-3">
        <h2 className="text-lg font-semibold text-slate-950">权限目录</h2>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {catalogEntries.map(([key, values]) => (
            <div
              key={key}
              className="rounded-3xl border border-slate-200 bg-white p-5 shadow-[0_16px_40px_rgba(15,23,42,0.05)]"
            >
              <p className="text-sm font-semibold text-slate-900">{key}</p>
              <div className="mt-3 flex flex-wrap gap-2">
                {values.map((item) => (
                  <span
                    key={item}
                    className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600"
                  >
                    {item}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
