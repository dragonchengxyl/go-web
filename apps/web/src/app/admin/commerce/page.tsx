'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { Disc3, Receipt, Sparkles, Wallet } from 'lucide-react'
import { AdminAudioWork, AdminOrder, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import {
  AdminSectionCard,
  AdminWorkspaceCard,
  formatDateTime,
  formatSponsorProgress,
} from '@/components/admin/admin-dashboard-kit'
import { formatPrice } from '@/lib/utils'

const statusLabelMap: Record<string, string> = {
  pending_payment: '待支付',
  paid: '已支付',
  fulfilled: '已完成',
  cancelled: '已取消',
  refunded: '已退款',
  failed: '支付失败',
}

const paymentMethodLabelMap: Record<string, string> = {
  alipay: '支付宝',
  wechat: '微信支付',
}

function CommerceItem({
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

export default function AdminCommercePage() {
  const { data: ordersData } = useQuery({
    queryKey: ['admin-dashboard-commerce-orders'],
    queryFn: () => apiClient.getAdminOrders({ page: 1, page_size: 5, type: 'tip' }),
  })

  const { data: pendingOrdersData } = useQuery({
    queryKey: ['admin-dashboard-commerce-orders-pending'],
    queryFn: () =>
      apiClient.getAdminOrders({
        page: 1,
        page_size: 1,
        type: 'tip',
        status: 'pending_payment',
      }),
  })

  const { data: failedOrdersData } = useQuery({
    queryKey: ['admin-dashboard-commerce-orders-failed'],
    queryFn: () =>
      apiClient.getAdminOrders({
        page: 1,
        page_size: 1,
        type: 'tip',
        status: 'failed',
      }),
  })

  const { data: paidOrdersData } = useQuery({
    queryKey: ['admin-dashboard-commerce-orders-paid'],
    queryFn: () =>
      apiClient.getAdminOrders({
        page: 1,
        page_size: 1,
        type: 'tip',
        status: 'paid',
      }),
  })

  const { data: audioData } = useQuery({
    queryKey: ['admin-dashboard-commerce-audio'],
    queryFn: () => apiClient.getAdminAudioWorks({ page: 1, page_size: 5, status: 'pending' }),
  })

  const { data: systemData } = useQuery({
    queryKey: ['admin-dashboard-commerce-system'],
    queryFn: () => apiClient.getAdminSystemConfig(),
  })

  const orders = (ordersData?.orders ?? []) as AdminOrder[]
  const works = (audioData?.works ?? []) as AdminAudioWork[]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Commerce Dashboard"
        title="商业与创作者盘"
        description="把打赏订单、音频分发和赞助进度收在同一个工作面里，优先发现支付异常、内容积压和创作者供给压力。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/system">查看赞助配置</Link>
            </Button>
            <Button asChild className="border border-[#c5dfd3] bg-[#edf8f2] text-[#21584e] hover:bg-[#e3f3eb]">
              <Link href="/admin/orders">进入订单工作台</Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <AdminMetricCard
          label="订单总量"
          value={(ordersData?.total ?? 0).toLocaleString()}
          hint="当前打赏订单总数"
          icon={Receipt}
          tone="brand"
        />
        <AdminMetricCard
          label="待支付订单"
          value={(pendingOrdersData?.total ?? 0).toLocaleString()}
          hint="需要留意支付链路是否积压"
          icon={Wallet}
          tone={(pendingOrdersData?.total ?? 0) > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="支付失败"
          value={(failedOrdersData?.total ?? 0).toLocaleString()}
          hint="可作为异常支付巡检入口"
          icon={Wallet}
          tone={(failedOrdersData?.total ?? 0) > 0 ? 'danger' : 'success'}
        />
        <AdminMetricCard
          label="待审音频"
          value={(audioData?.total ?? 0).toLocaleString()}
          hint="影响创作者公开分发的队列"
          icon={Disc3}
          tone={(audioData?.total ?? 0) > 0 ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="赞助进度"
          value={formatSponsorProgress(
            systemData?.sponsor.current_raised,
            systemData?.sponsor.monthly_goal,
          )}
          hint="当前对外赞助展示的完成度"
          icon={Sparkles}
          tone="success"
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-3">
        <AdminWorkspaceCard
          href="/admin/orders"
          title="订单运营"
          description="查看订单状态、处理支付异常并完成支付后的状态流转。"
          icon={Receipt}
          tone="brand"
          metrics={[
            { label: '总量', value: (ordersData?.total ?? 0).toLocaleString() },
            { label: '待支付', value: (pendingOrdersData?.total ?? 0).toLocaleString() },
            { label: '已支付', value: (paidOrdersData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/audio-works"
          title="音频分发"
          description="处理音频审核、补齐创作者从发布到公开分发的链路。"
          icon={Disc3}
          tone="warning"
          metrics={[
            { label: '待审', value: (audioData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/system"
          title="赞助与系统"
          description="维护赞助展示、支付配置和面向外部展示的商业信息。"
          icon={Sparkles}
          tone="success"
          metrics={[
            {
              label: '本月目标',
              value: formatPrice(systemData?.sponsor.monthly_goal ?? 0),
            },
          ]}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSectionCard
          title="近期订单"
          description="先看最近几笔订单的支付和履约状态，快速感知当前商业链路是否平稳。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/orders">查看全部订单</Link>
            </Button>
          }
        >
          <div className="space-y-3">
            {orders.slice(0, 5).map((order) => (
              <CommerceItem
                key={order.id}
                title={order.order_no}
                subtitle={`付款人：${order.payer_username ? `@${order.payer_username}` : order.user_id}`}
                badge={statusLabelMap[order.status] ?? order.status}
                meta={`${formatPrice(order.total_cents, order.currency || 'CNY')} · ${paymentMethodLabelMap[order.payment_method || ''] ?? '未指定方式'} · ${formatDateTime(order.created_at)}`}
              />
            ))}
            {orders.length === 0 ? (
              <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                当前没有订单数据。
              </p>
            ) : null}
          </div>
        </AdminSectionCard>

        <AdminSectionCard
          title="创作者发布压力"
          description="把待审音频和赞助目标放在一起看，判断创作者供给和商业转化是否同步。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/audio-works">查看音频治理</Link>
            </Button>
          }
        >
          <div className="space-y-4">
            <div className="rounded-2xl border border-[#c8e3d7] bg-[linear-gradient(135deg,#f7fffb_0%,#ebf8f2_56%,#def3ea_100%)] px-5 py-5 text-[#17342d]">
              <p className="text-[11px] uppercase tracking-[0.24em] text-[#64897e]">
                Sponsor Snapshot
              </p>
              <h3 className="mt-2 text-xl font-semibold">
                {formatSponsorProgress(
                  systemData?.sponsor.current_raised,
                  systemData?.sponsor.monthly_goal,
                )}
              </h3>
              <p className="mt-2 text-sm leading-6 text-[#5d8378]">
                {systemData?.sponsor.message || '当前尚未配置对外赞助文案。'}
              </p>
            </div>

            <div className="space-y-3">
              {works.slice(0, 4).map((work) => (
                <CommerceItem
                  key={work.id}
                  title={work.title}
                  subtitle={`@${work.author_username || work.author_id}`}
                  badge="待审"
                  meta={`${work.like_count} 赞 · ${work.comment_count} 评 · ${formatDateTime(work.published_at)}`}
                />
              ))}
              {works.length === 0 ? (
                <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-6 text-sm text-emerald-700">
                  当前没有待审音频作品，创作者分发链路相对顺畅。
                </div>
              ) : null}
            </div>
          </div>
        </AdminSectionCard>
      </div>

      <AdminSectionCard
        title="商业链路提醒"
        description="值班时只看这三条，就能快速判断是不是该先盯支付、分发，还是继续看社区供给。"
      >
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <p className="font-medium text-slate-950">支付积压</p>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              待支付订单 {(pendingOrdersData?.total ?? 0).toLocaleString()} 笔。
              如果持续抬升，优先检查支付回调和状态流转。
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <p className="font-medium text-slate-950">异常支付</p>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              支付失败 {(failedOrdersData?.total ?? 0).toLocaleString()} 笔。
              这块最适合拿来做每日巡检和人工补偿。
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
            <p className="font-medium text-slate-950">内容分发</p>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              待审音频 {(audioData?.total ?? 0).toLocaleString()} 个。
              如果堆积明显，创作者端会感知到发布延迟。
            </p>
          </div>
        </div>
      </AdminSectionCard>
    </div>
  )
}
