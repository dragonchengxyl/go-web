"use client";

import { Inbox } from "lucide-react";

export function AdminEmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center rounded-3xl border border-dashed border-slate-200 bg-slate-50 px-6 py-14 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm">
        <Inbox className="h-6 w-6" />
      </span>
      <h3 className="mt-4 text-lg font-semibold text-slate-900">{title}</h3>
      {description ? (
        <p className="mt-2 max-w-xl text-sm leading-6 text-slate-500">
          {description}
        </p>
      ) : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}
