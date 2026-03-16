'use client'

import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Receipt, Search, Wallet } from 'lucide-react'
import { apiClient, AdminOrder } from '@/lib/api-client'
import { formatPrice } from '@/lib/utils'
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

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'pending_payment', label: '待支付' },
  { value: 'paid', label: '已支付' },
  { value: 'fulfilled', label: '已完成' },
  { value: 'cancelled', label: '已取消' },
  { value: 'refunded', label: '已退款' },
  { value: 'failed', label: '支付失败' },
]

const methodOptions = [
  { value: '', label: '全部方式' },
  { value: 'alipay', label: '支付宝' },
  { value: 'wechat', label: '微信支付' },
]

const statusLabelMap: Record<string, string> = {
  pending_payment: '待支付',
  paid: '已支付',
  fulfilled: '已完成',
  cancelled: '已取消',
  refunded: '已退款',
  failed: '支付失败',
}

const methodLabelMap: Record<string, string> = {
  alipay: '支付宝',
  wechat: '微信支付',
}

export default function AdminOrdersPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [paymentMethod, setPaymentMethod] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-orders', page, search, status, paymentMethod],
    queryFn: () =>
      apiClient.getAdminOrders({
        page,
        page_size: 20,
        search: search || undefined,
        status: status || undefined,
        payment_method: paymentMethod || undefined,
        type: 'tip',
      }),
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, nextStatus }: { id: string; nextStatus: string }) =>
      apiClient.updateAdminOrderStatus(id, nextStatus),
    onSuccess: (_, { nextStatus }) => {
      queryClient.invalidateQueries({ queryKey: ['admin-orders'] })
      showAdminToast(`订单已更新为${statusLabelMap[nextStatus] ?? nextStatus}`, 'success')
    },
    onError: () => {
      showAdminToast('更新订单状态失败', 'error')
    },
  })

  const orders = data?.orders ?? []
  const totalPages = data ? Math.max(1, Math.ceil(data.total / 20)) : 1

  const pageStats = {
    pending: orders.filter((item) => item.status === 'pending_payment').length,
    paid: orders.filter((item) => item.status === 'paid').length,
    fulfilled: orders.filter((item) => item.status === 'fulfilled').length,
  }

  const columns: AdminColumn<AdminOrder>[] = [
    {
      key: 'order',
      header: '订单',
      className: 'min-w-[280px]',
      render: (order) => (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <AdminStatusBadge value={order.status} label={statusLabelMap[order.status] ?? order.status} />
            {order.payment_method ? (
              <AdminStatusBadge value={order.payment_method} label={methodLabelMap[order.payment_method] ?? order.payment_method} />
            ) : null}
            {order.order_type ? (
              <AdminStatusBadge value="reviewed" label={order.order_type} />
            ) : null}
          </div>
          <div>
            <p className="font-medium text-slate-900">{order.order_no}</p>
            <p className="mt-1 text-sm text-slate-500">
              付款人：{order.payer_username ? `@${order.payer_username}` : order.user_id}
            </p>
            <p className="text-sm text-slate-500">
              收款人：{order.recipient_username ? `@${order.recipient_username}` : order.recipient_user_id || '—'}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'amount',
      header: '金额',
      className: 'whitespace-nowrap',
      render: (order) => (
        <div>
          <p className="font-medium text-slate-900">
            {formatPrice(order.total_cents, order.currency || 'CNY')}
          </p>
          {order.discount_cents > 0 ? (
            <p className="text-xs text-slate-400">
              折扣 {formatPrice(order.discount_cents, order.currency || 'CNY')}
            </p>
          ) : null}
        </div>
      ),
    },
    {
      key: 'timeline',
      header: '时间',
      className: 'min-w-[180px] text-slate-500',
      render: (order) => (
        <div className="space-y-1">
          <p>创建：{new Date(order.created_at).toLocaleString('zh-CN')}</p>
          <p>支付：{order.paid_at ? new Date(order.paid_at).toLocaleString('zh-CN') : '—'}</p>
        </div>
      ),
    },
    {
      key: 'actions',
      header: '操作',
      className: 'w-[260px]',
      render: (order) => {
        const actions: Array<{ label: string; status: string; tone?: 'default' | 'danger' }> = []
        if (order.status === 'pending_payment') {
          actions.push({ label: '标记已支付', status: 'paid' })
          actions.push({ label: '取消订单', status: 'cancelled', tone: 'danger' })
        } else if (order.status === 'paid') {
          actions.push({ label: '标记已完成', status: 'fulfilled' })
          actions.push({ label: '标记已退款', status: 'refunded', tone: 'danger' })
        }

        if (actions.length === 0) {
          return <span className="text-sm text-slate-400">无需操作</span>
        }

        return (
          <div className="flex flex-wrap gap-2">
            {actions.map((action) => (
              <Button
                key={action.status}
                size="sm"
                variant="outline"
                className={
                  action.tone === 'danger'
                    ? 'border-rose-200 text-rose-700 hover:bg-rose-50'
                    : 'border-slate-200 text-slate-700 hover:bg-slate-50'
                }
                disabled={updateStatusMutation.isPending}
                onClick={() =>
                  updateStatusMutation.mutate({ id: order.id, nextStatus: action.status })
                }
              >
                {action.label}
              </Button>
            ))}
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Payment Operations"
        title="订单运营"
        description="面向打赏和支付链路的后台工作台，用于查看订单状态、筛查异常支付和完成状态流转。"
      />

      <div className="grid gap-4 md:grid-cols-4">
        <AdminMetricCard
          label="订单结果数"
          value={(data?.total ?? 0).toLocaleString()}
          hint="当前筛选条件下的订单总量"
          icon={Receipt}
          tone="brand"
        />
        <AdminMetricCard
          label="本页待支付"
          value={pageStats.pending.toLocaleString()}
          hint="需要关注未支付积压"
          icon={Wallet}
          tone={pageStats.pending > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="本页已支付"
          value={pageStats.paid.toLocaleString()}
          hint="待履约订单"
          icon={Wallet}
          tone="default"
        />
        <AdminMetricCard
          label="本页已完成"
          value={pageStats.fulfilled.toLocaleString()}
          hint="已闭环订单"
          icon={Receipt}
          tone="success"
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
              placeholder="搜索订单号、付款人或收款人 ID"
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
          <select
            value={paymentMethod}
            onChange={(e) => {
              setPage(1)
              setPaymentMethod(e.target.value)
            }}
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          >
            {methodOptions.map((option) => (
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
              setStatus('')
              setPaymentMethod('')
            }}
          >
            重置
          </Button>
        </div>
      </AdminFilterBar>

      <AdminDataTable
        data={orders}
        columns={columns}
        keyExtractor={(order) => order.id}
        loading={isLoading}
        empty={
          <AdminEmptyState
            title="没有匹配的订单"
            description="调整筛选条件后再试，或者等待新的支付记录进入系统。"
          />
        }
      />

      <AdminPagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
