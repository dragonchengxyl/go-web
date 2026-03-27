"use client";

import Link from "next/link";
import { ArrowRight, type LucideIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn, formatPrice } from "@/lib/utils";

type DashboardTone = "default" | "brand" | "success" | "warning" | "danger";

const toneClassMap: Record<DashboardTone, string> = {
  default: "border-[#d3e7dc] bg-[#eaf4ee] text-[#476b61]",
  brand: "border-[#bde7de] bg-[#daf4ee] text-[#1f7a73]",
  success: "border-[#ccead9] bg-[#e5f6ea] text-[#2d8a62]",
  warning: "border-[#f4dfb0] bg-[#fff2dc] text-[#a66a17]",
  danger: "border-[#f1cbbb] bg-[#ffede6] text-[#bc5c3e]",
};

export function AdminWorkspaceCard({
  href,
  title,
  description,
  icon: Icon,
  tone = "default",
  metrics = [],
  ctaLabel = "进入工作台",
}: {
  href: string;
  title: string;
  description: string;
  icon: LucideIcon;
  tone?: DashboardTone;
  metrics?: Array<{ label: string; value: React.ReactNode }>;
  ctaLabel?: string;
}) {
  return (
    <Link href={href} className="group block">
      <Card className="h-full rounded-[28px] border-slate-200 bg-white transition-all duration-200 hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-[0_20px_60px_rgba(15,23,42,0.08)]">
        <CardContent className="flex h-full flex-col gap-5 p-6">
          <div className="flex items-start justify-between gap-4">
            <span
              className={cn(
                "flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border",
                toneClassMap[tone],
              )}
            >
              <Icon className="h-5 w-5" />
            </span>
            <span className="inline-flex items-center gap-1 text-sm font-medium text-slate-400 transition-colors group-hover:text-slate-700">
              {ctaLabel}
              <ArrowRight className="h-4 w-4" />
            </span>
          </div>

          <div>
            <h3 className="text-lg font-semibold text-slate-950">{title}</h3>
            <p className="mt-2 text-sm leading-6 text-slate-500">{description}</p>
          </div>

          {metrics.length > 0 ? (
            <div className="mt-auto flex flex-wrap gap-2">
              {metrics.map((metric) => (
                <Badge
                  key={`${title}-${metric.label}`}
                  variant="outline"
                  className="rounded-full border-slate-200 bg-slate-50 px-3 py-1 text-[11px] font-medium text-slate-600"
                >
                  {metric.label} · {metric.value}
                </Badge>
              ))}
            </div>
          ) : null}
        </CardContent>
      </Card>
    </Link>
  );
}

export function AdminSectionCard({
  title,
  description,
  action,
  children,
  className,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <Card
      className={cn(
        "rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]",
        className,
      )}
    >
      <CardHeader className="flex flex-col gap-3 pb-4 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle className="text-lg text-slate-950">{title}</CardTitle>
          {description ? (
            <p className="mt-2 text-sm leading-6 text-slate-500">{description}</p>
          ) : null}
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}

export function formatDateTime(value?: string | null) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString("zh-CN");
}

export function formatCountRecord(
  value?: Record<string, number>,
  labels?: Record<string, string>,
) {
  if (!value) return "—";
  const items = Object.entries(value).filter(([, count]) => count > 0);
  if (items.length === 0) return "—";
  return items
    .sort(([, left], [, right]) => right - left)
    .map(([key, count]) => `${labels?.[key] ?? key}×${count}`)
    .join(" · ");
}

export function formatSponsorProgress(currentRaised = 0, monthlyGoal = 0) {
  if (monthlyGoal <= 0) return "未设置目标";
  const percent = Math.min(999, Math.round((currentRaised / monthlyGoal) * 100));
  return `${percent}% · ${formatPrice(currentRaised)} / ${formatPrice(monthlyGoal)}`;
}
