"use client";

import { cn } from "@/lib/utils";

export function AdminFilterBar({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)]",
        className,
      )}
    >
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        {children}
      </div>
    </div>
  );
}
