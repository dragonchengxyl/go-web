"use client";

import { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

const toneClassMap = {
  default: "bg-[#eaf4ee] text-[#476b61]",
  brand: "bg-[#daf4ee] text-[#1f7a73]",
  success: "bg-[#e5f6ea] text-[#2d8a62]",
  warning: "bg-[#fff2dc] text-[#a66a17]",
  danger: "bg-[#ffede6] text-[#bc5c3e]",
};

export function AdminMetricCard({
  label,
  value,
  hint,
  icon: Icon,
  tone = "default",
}: {
  label: string;
  value: React.ReactNode;
  hint?: string;
  icon: LucideIcon;
  tone?: keyof typeof toneClassMap;
}) {
  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-[0_16px_40px_rgba(15,23,42,0.06)]">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-sm font-medium text-slate-500">{label}</p>
          <p className="mt-3 text-3xl font-semibold tracking-tight text-slate-950">
            {value}
          </p>
          {hint ? (
            <p className="mt-2 text-xs leading-5 text-slate-400">{hint}</p>
          ) : null}
        </div>
        <span
          className={cn(
            "flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl",
            toneClassMap[tone],
          )}
        >
          <Icon className="h-5 w-5" />
        </span>
      </div>
    </div>
  );
}
