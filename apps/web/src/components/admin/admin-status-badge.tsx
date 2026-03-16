"use client";

import { cn } from "@/lib/utils";

const presetMap = {
  pending: "bg-amber-50 text-amber-700 ring-amber-200",
  approved: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  blocked: "bg-rose-50 text-rose-700 ring-rose-200",
  reviewed: "bg-sky-50 text-sky-700 ring-sky-200",
  dismissed: "bg-slate-100 text-slate-600 ring-slate-200",
  active: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  banned: "bg-rose-50 text-rose-700 ring-rose-200",
  admin: "bg-rose-50 text-rose-700 ring-rose-200",
  super_admin: "bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-200",
  moderator: "bg-sky-50 text-sky-700 ring-sky-200",
  creator: "bg-violet-50 text-violet-700 ring-violet-200",
  supporter: "bg-orange-50 text-orange-700 ring-orange-200",
  member: "bg-slate-100 text-slate-700 ring-slate-200",
  guest: "bg-slate-100 text-slate-600 ring-slate-200",
} as const;

export function AdminStatusBadge({
  value,
  label,
  className,
}: {
  value: string;
  label?: string;
  className?: string;
}) {
  const preset =
    presetMap[value as keyof typeof presetMap] ??
    "bg-slate-100 text-slate-700 ring-slate-200";

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset",
        preset,
        className,
      )}
    >
      {label ?? value}
    </span>
  );
}
